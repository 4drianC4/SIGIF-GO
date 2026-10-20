DO $$
BEGIN
    RAISE EXCEPTION 'catalog_contract cannot be rolled back automatically: it drops the legacy columns after converting their data. Restore a backup or write a new migration.';
END $$;
