CREATE TABLE IF NOT EXISTS branches (
    id bigserial PRIMARY KEY,
    company_id uuid NOT NULL,
    name varchar(150) NOT NULL,
    created_at timestamptz NOT NULL DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT uq_branches_company_name UNIQUE (company_id, name)
);
CREATE INDEX IF NOT EXISTS idx_branches_company ON branches (company_id);

CREATE TABLE IF NOT EXISTS locations (
    id bigserial PRIMARY KEY,
    branch_id bigint NOT NULL REFERENCES branches (id) ON DELETE CASCADE,
    company_id uuid NOT NULL,
    name varchar(150) NOT NULL,
    created_at timestamptz NOT NULL DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT uq_locations_branch_name UNIQUE (branch_id, name)
);
CREATE INDEX IF NOT EXISTS idx_locations_company ON locations (company_id);

CREATE TABLE IF NOT EXISTS stock_movements (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    company_id uuid NOT NULL,
    product_id uuid NOT NULL,
    branch_id bigint NOT NULL,
    location_id bigint NOT NULL,
    type varchar(20) NOT NULL,
    quantity numeric(14,3) NOT NULL,
    reason varchar(200),
    user_id uuid,
    created_at timestamptz NOT NULL DEFAULT CURRENT_TIMESTAMP
);
CREATE INDEX IF NOT EXISTS idx_stock_movements_company_product ON stock_movements (company_id, product_id);
CREATE INDEX IF NOT EXISTS idx_stock_movements_created_at ON stock_movements (created_at);
