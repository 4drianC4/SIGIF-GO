-- A category name is unique per level ignoring case: "Bebidas" and "BEBIDAS" are
-- the same name. The service already checks it; these indexes make it hold when
-- two requests arrive at the same time. The index names do not change.
DO $$
DECLARE
    duplicated text;
BEGIN
    SELECT string_agg(format('%s (company %s)', names, company_id), ', ' ORDER BY names COLLATE "C")
      INTO duplicated
      FROM (SELECT string_agg(name, ' / ' ORDER BY name COLLATE "C") AS names, company_id
              FROM categories
             WHERE deleted_at IS NULL
             GROUP BY company_id, parent_id, lower(name)
            HAVING count(*) > 1) AS d;
    IF duplicated IS NOT NULL THEN
        RAISE EXCEPTION 'category_names_ignore_case: category names repeated at the same level ignoring case: %. Rename all but one of them and run the migration again.', duplicated;
    END IF;
END $$;

DROP INDEX IF EXISTS idx_categories_company_root_name;
DROP INDEX IF EXISTS idx_categories_company_parent_name;
CREATE UNIQUE INDEX idx_categories_company_root_name ON categories (company_id, lower(name))
    WHERE deleted_at IS NULL AND parent_id IS NULL;
CREATE UNIQUE INDEX idx_categories_company_parent_name ON categories (company_id, parent_id, lower(name))
    WHERE deleted_at IS NULL AND parent_id IS NOT NULL;
