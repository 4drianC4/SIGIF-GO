-- HU-06-01: a SKU or barcode stays taken while any product of the company
-- holds it, also an inactive or deleted one. SKUs are stored trimmed and in
-- upper case so the comparison ignores case and outer spaces.
DO $$
DECLARE
    duplicated text;
BEGIN
    UPDATE products SET sku = NULL WHERE sku IS NOT NULL AND btrim(sku) = '';
    UPDATE products SET barcode = NULL WHERE barcode IS NOT NULL AND btrim(barcode) = '';

    SELECT string_agg(format('%s (%s products)', code, holders), ', ' ORDER BY code)
      INTO duplicated
      FROM (SELECT upper(btrim(sku)) AS code, count(*) AS holders
              FROM products
             WHERE sku IS NOT NULL
             GROUP BY company_id, upper(btrim(sku))
            HAVING count(*) > 1) AS d;
    IF duplicated IS NOT NULL THEN
        RAISE EXCEPTION 'product_reserved_codes: SKUs shared by several products, counting deleted ones: %. Change the SKU of all but one of them and run the migration again.', duplicated;
    END IF;

    SELECT string_agg(format('%s (%s products)', code, holders), ', ' ORDER BY code)
      INTO duplicated
      FROM (SELECT btrim(barcode) AS code, count(*) AS holders
              FROM products
             WHERE barcode IS NOT NULL
             GROUP BY company_id, btrim(barcode)
            HAVING count(*) > 1) AS d;
    IF duplicated IS NOT NULL THEN
        RAISE EXCEPTION 'product_reserved_codes: barcodes shared by several products, counting deleted ones: %. Change the barcode of all but one of them and run the migration again.', duplicated;
    END IF;

    UPDATE products SET sku = upper(btrim(sku)) WHERE sku IS NOT NULL AND sku <> upper(btrim(sku));
    UPDATE products SET barcode = btrim(barcode) WHERE barcode IS NOT NULL AND barcode <> btrim(barcode);
END $$;

DROP INDEX IF EXISTS idx_products_company_sku;
DROP INDEX IF EXISTS idx_products_company_barcode;
CREATE UNIQUE INDEX idx_products_company_sku ON products (company_id, sku)
    WHERE sku IS NOT NULL;
CREATE UNIQUE INDEX idx_products_company_barcode ON products (company_id, barcode)
    WHERE barcode IS NOT NULL;
