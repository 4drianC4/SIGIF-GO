DO $$
BEGIN
    IF to_regclass('categories') IS NULL THEN
        CREATE TABLE categories (
            id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
            company_id uuid NOT NULL,
            name varchar(100) NOT NULL,
            description varchar(500),
            status varchar(20) NOT NULL DEFAULT 'active',
            created_at timestamptz,
            updated_at timestamptz,
            deleted_at timestamptz
        );
        CREATE UNIQUE INDEX idx_categories_company_name ON categories (company_id, name) WHERE deleted_at IS NULL;
        CREATE INDEX idx_categories_status ON categories (status);
        CREATE INDEX idx_categories_deleted_at ON categories (deleted_at);
    END IF;

    IF to_regclass('products') IS NULL THEN
        CREATE TABLE products (
            id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
            company_id uuid NOT NULL,
            category_id uuid NOT NULL,
            sku varchar(50) NOT NULL,
            barcode varchar(14),
            name varchar(150) NOT NULL,
            description varchar(1000),
            unit_of_measure varchar(10) NOT NULL,
            cost_price numeric(12,2) NOT NULL DEFAULT 0,
            sale_price numeric(12,2) NOT NULL,
            status varchar(20) NOT NULL DEFAULT 'active',
            created_at timestamptz,
            updated_at timestamptz,
            deleted_at timestamptz,
            CONSTRAINT fk_products_category FOREIGN KEY (category_id) REFERENCES categories (id) ON UPDATE CASCADE ON DELETE RESTRICT
        );
        CREATE UNIQUE INDEX idx_products_company_sku ON products (company_id, sku) WHERE deleted_at IS NULL;
        CREATE UNIQUE INDEX idx_products_company_barcode ON products (company_id, barcode) WHERE deleted_at IS NULL AND barcode IS NOT NULL;
        CREATE INDEX idx_products_category_id ON products (category_id);
        CREATE INDEX idx_products_status ON products (status);
        CREATE INDEX idx_products_deleted_at ON products (deleted_at);
    END IF;
END $$;
