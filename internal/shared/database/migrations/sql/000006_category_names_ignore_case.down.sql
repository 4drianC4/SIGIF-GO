-- The names compare exactly again, as before. Nothing stored changes.
DROP INDEX IF EXISTS idx_categories_company_root_name;
DROP INDEX IF EXISTS idx_categories_company_parent_name;
CREATE UNIQUE INDEX idx_categories_company_root_name ON categories (company_id, name)
    WHERE deleted_at IS NULL AND parent_id IS NULL;
CREATE UNIQUE INDEX idx_categories_company_parent_name ON categories (company_id, parent_id, name)
    WHERE deleted_at IS NULL AND parent_id IS NOT NULL;
