CREATE TABLE IF NOT EXISTS product_price_history (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    product_id uuid NOT NULL,
    user_id uuid NOT NULL,
    previous_cost numeric(12,2) NOT NULL,
    new_cost numeric(12,2) NOT NULL,
    previous_price numeric(12,2) NOT NULL,
    new_price numeric(12,2) NOT NULL,
    reason varchar(160),
    created_at timestamptz
);
CREATE INDEX IF NOT EXISTS idx_product_price_history_product_id ON product_price_history (product_id);
CREATE INDEX IF NOT EXISTS idx_product_price_history_created_at ON product_price_history (created_at);

CREATE TABLE IF NOT EXISTS stock_balance (
    stock_balance_id bigserial PRIMARY KEY,
    company_id uuid NOT NULL,
    product_id uuid NOT NULL,
    branch_id bigint NOT NULL,
    location_id bigint NOT NULL,
    quantity numeric(14,3) NOT NULL DEFAULT 0,
    reserved_quantity numeric(14,3) NOT NULL DEFAULT 0,
    updated_at timestamptz NOT NULL DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT chk_stock_balance_quantity CHECK (quantity >= 0),
    CONSTRAINT chk_stock_balance_reserved CHECK (reserved_quantity >= 0 AND reserved_quantity <= quantity)
);
CREATE INDEX IF NOT EXISTS idx_stock_balance_company_product ON stock_balance (company_id, product_id);
CREATE UNIQUE INDEX IF NOT EXISTS idx_stock_balance_product_branch_location ON stock_balance (product_id, branch_id, location_id);
CREATE INDEX IF NOT EXISTS idx_stock_balance_branch_product ON stock_balance (branch_id, product_id);
