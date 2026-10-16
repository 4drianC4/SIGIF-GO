-- Deleted products free their SKU and barcode again. The normalized values
-- are kept: they are valid under both rules.
DROP INDEX IF EXISTS idx_products_company_sku;
DROP INDEX IF EXISTS idx_products_company_barcode;
CREATE UNIQUE INDEX idx_products_company_sku ON products (company_id, sku)
    WHERE deleted_at IS NULL AND sku IS NOT NULL;
CREATE UNIQUE INDEX idx_products_company_barcode ON products (company_id, barcode)
    WHERE deleted_at IS NULL AND barcode IS NOT NULL;
