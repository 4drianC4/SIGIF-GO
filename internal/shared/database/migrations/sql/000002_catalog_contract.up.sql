CREATE TABLE IF NOT EXISTS units_of_measure (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    name varchar(50) NOT NULL,
    abbreviation varchar(10) NOT NULL,
    sort_order bigint NOT NULL DEFAULT 0,
    created_at timestamptz
);
CREATE UNIQUE INDEX IF NOT EXISTS idx_units_of_measure_name ON units_of_measure (name);
CREATE UNIQUE INDEX IF NOT EXISTS idx_units_of_measure_abbreviation ON units_of_measure (abbreviation);

CREATE TABLE IF NOT EXISTS taxes (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    name varchar(60) NOT NULL,
    percentage numeric(5,2) NOT NULL,
    created_at timestamptz
);
CREATE UNIQUE INDEX IF NOT EXISTS idx_taxes_name ON taxes (name);

INSERT INTO units_of_measure (name, abbreviation, sort_order, created_at) VALUES
    ('Unit', 'UNIT', 1, now()),
    ('Box', 'BOX', 2, now()),
    ('Pack', 'PACK', 3, now()),
    ('Kilogram', 'KG', 4, now()),
    ('Liter', 'L', 5, now()),
    ('Dozen', 'DOZ', 6, now())
ON CONFLICT DO NOTHING;

INSERT INTO taxes (name, percentage, created_at) VALUES
    ('IVA general 13%', 13, now()),
    ('IVA reducido 5%', 5, now()),
    ('Exempt', 0, now())
ON CONFLICT DO NOTHING;

ALTER TABLE categories
    ADD COLUMN IF NOT EXISTS parent_id uuid,
    ADD COLUMN IF NOT EXISTS default_tax_id uuid,
    ADD COLUMN IF NOT EXISTS target_margin numeric(5,2);

ALTER TABLE products
    ADD COLUMN IF NOT EXISTS unit_of_measure_id uuid,
    ADD COLUMN IF NOT EXISTS cost numeric(12,2),
    ADD COLUMN IF NOT EXISTS stock numeric(14,3) NOT NULL DEFAULT 0,
    ADD COLUMN IF NOT EXISTS min_stock numeric(14,3) NOT NULL DEFAULT 0,
    ADD COLUMN IF NOT EXISTS last_movement_at timestamptz;

DO $$
DECLARE
    unknown_units text;
BEGIN
    IF EXISTS (SELECT 1 FROM information_schema.columns
               WHERE table_schema = current_schema() AND table_name = 'products' AND column_name = 'unit_of_measure') THEN
        INSERT INTO units_of_measure (name, abbreviation, sort_order, created_at)
        SELECT extra.name, extra.abbreviation, extra.sort_order, now()
        FROM (VALUES ('Gram', 'G', 7), ('Milliliter', 'ML', 8), ('Meter', 'M', 9)) AS extra (name, abbreviation, sort_order)
        WHERE EXISTS (SELECT 1 FROM products p
                      WHERE p.unit_of_measure_id IS NULL AND upper(btrim(p.unit_of_measure)) = extra.abbreviation)
        ON CONFLICT DO NOTHING;

        UPDATE products p
        SET unit_of_measure_id = u.id
        FROM units_of_measure u
        WHERE p.unit_of_measure_id IS NULL AND u.abbreviation = upper(btrim(p.unit_of_measure));

        SELECT string_agg(DISTINCT quote_literal(unit_of_measure), ', ') INTO unknown_units
        FROM products
        WHERE unit_of_measure_id IS NULL;
        IF unknown_units IS NOT NULL THEN
            RAISE EXCEPTION 'products.unit_of_measure has values without an equivalent unit: %', unknown_units;
        END IF;

        ALTER TABLE products DROP COLUMN unit_of_measure;
    END IF;

    IF EXISTS (SELECT 1 FROM information_schema.columns
               WHERE table_schema = current_schema() AND table_name = 'products' AND column_name = 'cost_price') THEN
        UPDATE products SET cost = cost_price WHERE cost IS NULL;
        ALTER TABLE products DROP COLUMN cost_price;
    END IF;
END $$;

UPDATE products SET last_movement_at = COALESCE(created_at, now()) WHERE last_movement_at IS NULL;

ALTER TABLE products
    ALTER COLUMN unit_of_measure_id SET NOT NULL,
    ALTER COLUMN cost SET NOT NULL,
    ALTER COLUMN last_movement_at SET NOT NULL,
    ALTER COLUMN sku DROP NOT NULL;

UPDATE products p
SET name = left(p.name, 150 - char_length(d.suffix)) || d.suffix
FROM (
    SELECT id,
           ' (' || COALESCE(sku, left(id::text, 8)) || ')' AS suffix,
           row_number() OVER (PARTITION BY company_id, name ORDER BY created_at NULLS LAST, id) AS position
    FROM products
    WHERE deleted_at IS NULL
) d
WHERE p.id = d.id AND d.position > 1;

DROP INDEX IF EXISTS idx_categories_company_name;
DROP INDEX IF EXISTS idx_categories_company_root_name;
DROP INDEX IF EXISTS idx_categories_company_parent_name;
CREATE UNIQUE INDEX idx_categories_company_root_name ON categories (company_id, name)
    WHERE deleted_at IS NULL AND parent_id IS NULL;
CREATE UNIQUE INDEX idx_categories_company_parent_name ON categories (company_id, parent_id, name)
    WHERE deleted_at IS NULL AND parent_id IS NOT NULL;

DROP INDEX IF EXISTS idx_products_company_name;
DROP INDEX IF EXISTS idx_products_company_sku;
DROP INDEX IF EXISTS idx_products_company_barcode;
CREATE UNIQUE INDEX idx_products_company_name ON products (company_id, name)
    WHERE deleted_at IS NULL;
CREATE UNIQUE INDEX idx_products_company_sku ON products (company_id, sku)
    WHERE deleted_at IS NULL AND sku IS NOT NULL;
CREATE UNIQUE INDEX idx_products_company_barcode ON products (company_id, barcode)
    WHERE deleted_at IS NULL AND barcode IS NOT NULL;

CREATE INDEX IF NOT EXISTS idx_categories_parent_id ON categories (parent_id);
CREATE INDEX IF NOT EXISTS idx_categories_default_tax_id ON categories (default_tax_id);
CREATE INDEX IF NOT EXISTS idx_products_unit_of_measure_id ON products (unit_of_measure_id);
CREATE INDEX IF NOT EXISTS idx_products_last_movement_at ON products (last_movement_at);

DO $$
BEGIN
    IF NOT EXISTS (SELECT 1 FROM pg_constraint WHERE conrelid = 'categories'::regclass AND conname = 'fk_categories_parent') THEN
        ALTER TABLE categories ADD CONSTRAINT fk_categories_parent
            FOREIGN KEY (parent_id) REFERENCES categories (id) ON UPDATE CASCADE ON DELETE RESTRICT;
    END IF;
    IF NOT EXISTS (SELECT 1 FROM pg_constraint WHERE conrelid = 'categories'::regclass AND conname = 'fk_categories_default_tax') THEN
        ALTER TABLE categories ADD CONSTRAINT fk_categories_default_tax
            FOREIGN KEY (default_tax_id) REFERENCES taxes (id) ON UPDATE CASCADE ON DELETE RESTRICT;
    END IF;
    IF NOT EXISTS (SELECT 1 FROM pg_constraint WHERE conrelid = 'products'::regclass AND conname = 'fk_products_unit_of_measure') THEN
        ALTER TABLE products ADD CONSTRAINT fk_products_unit_of_measure
            FOREIGN KEY (unit_of_measure_id) REFERENCES units_of_measure (id) ON UPDATE CASCADE ON DELETE RESTRICT;
    END IF;
END $$;
