//go:build integration

package migrations_test

import (
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/shopspring/decimal"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"github.com/sigif/sigif-go/internal/modules/product/infrastructure/persistence/model"
	"github.com/sigif/sigif-go/internal/shared/database/dbtest"
	"github.com/sigif/sigif-go/internal/shared/database/migrations"
)

const legacySchema = `
CREATE TABLE categories (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    company_id uuid NOT NULL,
    name varchar(100) NOT NULL,
    description varchar(500),
    status varchar(20) NOT NULL DEFAULT 'active',
    created_at timestamptz, updated_at timestamptz, deleted_at timestamptz
);
CREATE UNIQUE INDEX idx_categories_company_name ON categories (company_id, name) WHERE deleted_at IS NULL;
CREATE INDEX idx_categories_status ON categories (status);
CREATE INDEX idx_categories_deleted_at ON categories (deleted_at);
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
    created_at timestamptz, updated_at timestamptz, deleted_at timestamptz,
    CONSTRAINT fk_products_category FOREIGN KEY (category_id) REFERENCES categories (id) ON UPDATE CASCADE ON DELETE RESTRICT
);
CREATE UNIQUE INDEX idx_products_company_sku ON products (company_id, sku) WHERE deleted_at IS NULL;
CREATE UNIQUE INDEX idx_products_company_barcode ON products (company_id, barcode) WHERE deleted_at IS NULL AND barcode IS NOT NULL;
CREATE INDEX idx_products_category_id ON products (category_id);
CREATE INDEX idx_products_status ON products (status);
CREATE INDEX idx_products_deleted_at ON products (deleted_at);`

var catalogModels = []any{&model.UnitOfMeasureModel{}, &model.TaxModel{}, &model.CategoryModel{}, &model.ProductModel{}}

type env struct {
	t        *testing.T
	dsn      string
	db       *gorm.DB
	company  uuid.UUID
	category uuid.UUID
}

func newEnv(t *testing.T) *env {
	t.Helper()
	dsn := dbtest.NewSchemaDSN(t)
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{Logger: logger.Default.LogMode(logger.Silent)})
	if err != nil {
		t.Fatalf("open database: %v", err)
	}
	pool, err := db.DB()
	if err != nil {
		t.Fatalf("database pool: %v", err)
	}
	t.Cleanup(func() { _ = pool.Close() })
	return &env{t: t, dsn: dsn, db: db, company: uuid.New(), category: uuid.New()}
}

func (e *env) exec(query string, args ...any) {
	e.t.Helper()
	if err := e.db.Exec(query, args...).Error; err != nil {
		e.t.Fatalf("exec %q: %v", query, err)
	}
}

func (e *env) up() {
	e.t.Helper()
	if err := migrations.Up(e.dsn); err != nil {
		e.t.Fatalf("migrations.Up() error = %v", err)
	}
}

func (e *env) legacy() {
	e.t.Helper()
	e.exec(legacySchema)
	e.exec(`INSERT INTO categories (id, company_id, name, created_at) VALUES (?, ?, 'Mascotas', now())`, e.category, e.company)
}

func (e *env) legacyProduct(sku, name, unit, costPrice string, createdAt time.Time) uuid.UUID {
	e.t.Helper()
	id := uuid.New()
	e.exec(`INSERT INTO products (id, company_id, category_id, sku, name, unit_of_measure, cost_price, sale_price, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?::numeric, 300, ?, ?)`, id, e.company, e.category, sku, name, unit, costPrice, createdAt, createdAt.Add(time.Hour))
	return id
}

func (e *env) hasColumn(table, column string) bool {
	e.t.Helper()
	var n int64
	e.db.Raw(`SELECT count(*) FROM information_schema.columns
		WHERE table_schema = current_schema() AND table_name = ? AND column_name = ?`, table, column).Scan(&n)
	return n > 0
}

func (e *env) hasRelation(name string) bool {
	e.t.Helper()
	var exists bool
	e.db.Raw(`SELECT to_regclass(?) IS NOT NULL`, name).Scan(&exists)
	return exists
}

func (e *env) assertVersion(version int, dirty bool) {
	e.t.Helper()
	var row struct {
		Version int
		Dirty   bool
	}
	if err := e.db.Raw(`SELECT version, dirty FROM schema_migrations`).Scan(&row).Error; err != nil {
		e.t.Fatalf("read schema_migrations: %v", err)
	}
	if row.Version != version || row.Dirty != dirty {
		e.t.Fatalf("schema_migrations = %+v, want version %d dirty %v", row, version, dirty)
	}
}

func (e *env) unitID(abbreviation string) uuid.UUID {
	e.t.Helper()
	var unit model.UnitOfMeasureModel
	if err := e.db.First(&unit, "abbreviation = ?", abbreviation).Error; err != nil {
		e.t.Fatalf("unit %s: %v", abbreviation, err)
	}
	return unit.ID
}

func (e *env) newProduct(name string, sku *string) *model.ProductModel {
	return &model.ProductModel{
		ID: uuid.New(), CompanyID: e.company, CategoryID: e.category, UnitOfMeasureID: e.unitID("UNIT"),
		Name: name, SKU: sku, Cost: decimal.NewFromInt(230), SalePrice: decimal.NewFromInt(300),
		LastMovementAt: time.Now().UTC(),
	}
}

func isPgError(err error, code string) bool {
	var pgErr *pgconn.PgError
	return errors.As(err, &pgErr) && pgErr.Code == code
}

func TestFreshDatabase(t *testing.T) {
	e := newEnv(t)
	e.up()
	e.up()
	e.assertVersion(6, false)

	status, err := migrations.Status(e.dsn)
	if err != nil || len(status) != 6 || !status[0].Applied || !status[1].Applied || !status[2].Applied || !status[3].Applied || !status[4].Applied || !status[5].Applied ||
		status[0].Name != "catalog_baseline" || status[1].Name != "catalog_contract" || status[2].Name != "price_history_and_stock_balance" ||
		status[3].Name != "product_reserved_codes" || status[4].Name != "inventory_branches_and_movements" || status[5].Name != "category_names_ignore_case" {
		t.Fatalf("Status() = %+v, %v", status, err)
	}

	var units []string
	e.db.Model(&model.UnitOfMeasureModel{}).Order("sort_order").Pluck("abbreviation", &units)
	if strings.Join(units, ",") != "UNIT,BOX,PACK,KG,L,DOZ" {
		t.Errorf("units = %v", units)
	}
	var taxes []string
	e.db.Model(&model.TaxModel{}).Order("percentage DESC").Pluck("name", &taxes)
	if strings.Join(taxes, "|") != "IVA general 13%|IVA reducido 5%|Exempt" {
		t.Errorf("taxes = %v", taxes)
	}

	root := model.CategoryModel{ID: e.category, CompanyID: e.company, Name: "Mascotas"}
	if err := e.db.Create(&root).Error; err != nil {
		t.Fatalf("create category: %v", err)
	}
	if err := e.db.Create(&model.CategoryModel{CompanyID: e.company, ParentID: &root.ID, Name: "Mascotas"}).Error; err != nil {
		t.Errorf("same name under a parent must be allowed: %v", err)
	}
	if err := e.db.Create(&model.CategoryModel{CompanyID: e.company, Name: "Mascotas"}).Error; !isPgError(err, "23505") {
		t.Errorf("duplicate root category must be rejected, got %v", err)
	}
	if err := e.db.Create(&model.CategoryModel{CompanyID: e.company, Name: "MASCOTAS"}).Error; !isPgError(err, "23505") {
		t.Errorf("root category differing only in case must be rejected, got %v", err)
	}
	for _, name := range []string{"Sin SKU 1", "Sin SKU 2"} {
		if err := e.db.Create(e.newProduct(name, nil)).Error; err != nil {
			t.Errorf("product without sku %q: %v", name, err)
		}
	}
	if err := e.db.Create(e.newProduct("Sin SKU 1", nil)).Error; !isPgError(err, "23505") {
		t.Errorf("duplicate product name must be rejected, got %v", err)
	}
}

func TestLegacyCatalogKeepsItsData(t *testing.T) {
	e := newEnv(t)
	e.legacy()
	created := time.Date(2026, 1, 15, 10, 0, 0, 0, time.UTC)
	kg := e.legacyProduct("OLD-1", "Dog Chow 1kg", "kg", "20.50", created)
	grams := e.legacyProduct("OLD-2", "Arena gato", " G ", "10", created.Add(time.Minute))
	duplicate := e.legacyProduct("OLD-3", "Arena gato", "unit", "0", created.Add(2*time.Minute))
	deleted := e.legacyProduct("OLD-4", "Arena gato", "ml", "1", created.Add(3*time.Minute))
	e.exec(`UPDATE products SET deleted_at = now() WHERE id = ?`, deleted)

	e.up()
	e.assertVersion(6, false)

	type row struct {
		Name         string
		Abbreviation string
		Cost         string
		Stock        string
		LastMovement time.Time
		CreatedAt    time.Time
	}
	want := map[uuid.UUID]struct{ name, unit, cost string }{
		kg:        {"Dog Chow 1kg", "KG", "20.50"},
		grams:     {"Arena gato", "G", "10.00"},
		duplicate: {"Arena gato (OLD-3)", "UNIT", "0.00"},
		deleted:   {"Arena gato", "ML", "1.00"},
	}
	for id, w := range want {
		var r row
		if err := e.db.Raw(`SELECT p.name, u.abbreviation, p.cost::text AS cost, p.stock::text AS stock,
			p.last_movement_at AS last_movement, p.created_at
			FROM products p JOIN units_of_measure u ON u.id = p.unit_of_measure_id WHERE p.id = ?`, id).Scan(&r).Error; err != nil {
			t.Fatalf("read product: %v", err)
		}
		if r.Name != w.name || r.Abbreviation != w.unit || r.Cost != w.cost || r.Stock != "0.000" || !r.LastMovement.Equal(r.CreatedAt) {
			t.Errorf("product %s = %+v, want %+v", id, r, w)
		}
	}

	for _, column := range []string{"unit_of_measure", "cost_price"} {
		if e.hasColumn("products", column) {
			t.Errorf("legacy column %s still exists", column)
		}
	}
	if e.hasRelation("idx_categories_company_name") {
		t.Errorf("legacy category index still exists")
	}
	var units int64
	e.db.Model(&model.UnitOfMeasureModel{}).Count(&units)
	if units != 8 {
		t.Errorf("units = %d, want the 6 base units plus G and ML used by legacy products", units)
	}

	if err := e.db.Create(e.newProduct("Wiskas Gato adulto", nil)).Error; err != nil {
		t.Errorf("insert with the current model after migrating: %v", err)
	}
	if err := e.db.Create(&model.CategoryModel{CompanyID: e.company, ParentID: &e.category, Name: "Mascotas"}).Error; err != nil {
		t.Errorf("same category name under a parent must be allowed: %v", err)
	}
	e.up()
	e.assertVersion(6, false)
}

func TestHybridSchemaLeftByAutoMigrate(t *testing.T) {
	e := newEnv(t)
	e.legacy()
	if err := e.db.AutoMigrate(catalogModels...); err != nil {
		t.Fatalf("automigrate: %v", err)
	}
	e.exec(`INSERT INTO units_of_measure (id, name, abbreviation, sort_order) VALUES (gen_random_uuid(), 'Unit', 'UNIT', 1)`)

	sku := "WIA-879"
	var pgErr *pgconn.PgError
	err := e.db.Create(e.newProduct("Wiskas Gato adulto", &sku)).Error
	if !errors.As(err, &pgErr) || pgErr.Code != "23502" || pgErr.ColumnName != "unit_of_measure" {
		t.Fatalf("expected the reported NOT NULL violation on products.unit_of_measure, got %v", err)
	}

	e.up()
	if err := e.db.Create(e.newProduct("Wiskas Gato adulto", &sku)).Error; err != nil {
		t.Fatalf("insert after migrating: %v", err)
	}
	e.assertVersion(6, false)
}

func TestSchemaAlreadyCreatedByAutoMigrate(t *testing.T) {
	e := newEnv(t)
	if err := e.db.AutoMigrate(catalogModels...); err != nil {
		t.Fatalf("automigrate: %v", err)
	}
	e.exec(`INSERT INTO units_of_measure (id, name, abbreviation, sort_order) VALUES (gen_random_uuid(), 'Unit', 'UNIT', 1)`)
	unit := e.unitID("UNIT")
	root := model.CategoryModel{ID: e.category, CompanyID: e.company, Name: "Aceites"}
	child := model.CategoryModel{CompanyID: e.company, ParentID: &root.ID, Name: "Aceites"}
	product := e.newProduct("Aceite Fino 1L", nil)
	product.Stock, product.MinStock = decimal.RequireFromString("12.345"), decimal.RequireFromString("1.5")
	for _, record := range []any{&root, &child, product} {
		if err := e.db.Create(record).Error; err != nil {
			t.Fatalf("seed modern data: %v", err)
		}
	}

	e.up()
	e.assertVersion(6, false)

	var got model.ProductModel
	if err := e.db.First(&got, "id = ?", product.ID).Error; err != nil {
		t.Fatalf("read product: %v", err)
	}
	if got.Name != "Aceite Fino 1L" || got.Stock.String() != "12.345" || got.MinStock.String() != "1.5" || got.UnitOfMeasureID != unit || got.SKU != nil {
		t.Errorf("modern data changed: %+v", got)
	}
	var categories, units int64
	e.db.Model(&model.CategoryModel{}).Count(&categories)
	e.db.Model(&model.UnitOfMeasureModel{}).Count(&units)
	if categories != 2 || units != 6 || e.unitID("UNIT") != unit {
		t.Errorf("categories = %d, units = %d", categories, units)
	}
}

func TestFailedMigrationIsRolledBackAndCanBeRetried(t *testing.T) {
	e := newEnv(t)
	e.legacy()
	id := e.legacyProduct("OLD-1", "Dog Chow 1kg", "barril", "20.50", time.Now().UTC())

	err := migrations.Up(e.dsn)
	if err == nil || !strings.Contains(err.Error(), "without an equivalent unit: 'barril'") {
		t.Fatalf("expected unknown unit error, got %v", err)
	}
	if !e.hasColumn("products", "unit_of_measure") || !e.hasColumn("products", "cost_price") {
		t.Errorf("failed migration removed legacy columns")
	}
	if e.hasColumn("products", "cost") || e.hasColumn("categories", "parent_id") || e.hasRelation("units_of_measure") || !e.hasRelation("idx_categories_company_name") {
		t.Errorf("failed migration left partial changes")
	}
	e.assertVersion(1, false)

	e.exec(`UPDATE products SET unit_of_measure = 'box' WHERE id = ?`, id)
	e.up()
	e.assertVersion(6, false)
	var unit string
	e.db.Raw(`SELECT u.abbreviation FROM products p JOIN units_of_measure u ON u.id = p.unit_of_measure_id WHERE p.id = ?`, id).Scan(&unit)
	if unit != "BOX" {
		t.Errorf("unit after retry = %q, want BOX", unit)
	}
}

func TestDirtyDatabaseIsReportedAndCanBeForced(t *testing.T) {
	e := newEnv(t)
	e.up()
	e.exec(`UPDATE schema_migrations SET dirty = true`)

	err := migrations.Up(e.dsn)
	if err == nil || !strings.Contains(err.Error(), "dirty at version 6") {
		t.Fatalf("expected dirty database error, got %v", err)
	}
	if err := migrations.Force(e.dsn, 6); err != nil {
		t.Fatalf("Force() error = %v", err)
	}
	e.up()
	e.assertVersion(6, false)
}

func TestCatalogContractCannotBeRolledBack(t *testing.T) {
	e := newEnv(t)
	e.up()

	// 000006 and 000004 only swap indexes, 000005 and 000003 only add tables, so
	// all four can be rolled back; 000002 cannot.
	for _, version := range []int{5, 4, 3, 2} {
		if err := migrations.Down(e.dsn); err != nil {
			t.Fatalf("rolling back to version %d: %v", version, err)
		}
		e.assertVersion(version, false)
	}

	err := migrations.Down(e.dsn)
	if err == nil || !strings.Contains(err.Error(), "cannot be rolled back automatically") {
		t.Fatalf("expected forward-only error, got %v", err)
	}
	e.assertVersion(2, false)
	if !e.hasColumn("products", "unit_of_measure_id") || !e.hasRelation("units_of_measure") {
		t.Errorf("rejected rollback changed the schema")
	}
	e.up()
	e.assertVersion(6, false)
}

func TestPriceHistoryAndStockBalanceMigration(t *testing.T) {
	e := newEnv(t)
	e.up()
	e.assertVersion(6, false)
	for _, table := range []string{"product_price_history", "stock_balance"} {
		if !e.hasRelation(table) {
			t.Fatalf("table %s was not created", table)
		}
	}

	e.exec(`INSERT INTO categories (id, company_id, name, created_at) VALUES (?, ?, 'Bebidas', now())`, e.category, e.company)
	product := e.newProduct("Producto con saldo", nil)
	if err := e.db.Create(product).Error; err != nil {
		t.Fatalf("seed product: %v", err)
	}
	insert := `INSERT INTO stock_balance (company_id, product_id, branch_id, location_id, quantity, reserved_quantity)
		VALUES (?, ?, ?, ?, ?::numeric, ?::numeric)`
	e.exec(insert, e.company, product.ID, 1, 1, "10", "4")
	rejected := map[string][]any{
		"reserved above quantity":  {e.company, product.ID, 1, 2, "5", "6"},
		"negative quantity":        {e.company, product.ID, 1, 3, "-1", "0"},
		"same branch and location": {e.company, product.ID, 1, 1, "1", "0"},
	}
	for name, args := range rejected {
		if err := e.db.Exec(insert, args...).Error; err == nil {
			t.Errorf("stock_balance accepted %s", name)
		}
	}

	for _, version := range []int{5, 4, 3, 2} {
		if err := migrations.Down(e.dsn); err != nil {
			t.Fatalf("Down() to version %d error = %v", version, err)
		}
	}
	e.assertVersion(2, false)
	if e.hasRelation("product_price_history") || e.hasRelation("stock_balance") {
		t.Errorf("rollback left the tables behind")
	}
	e.up()
	e.assertVersion(6, false)
}

func TestPriceHistoryAndStockBalanceAlreadyCreatedByAutoMigrate(t *testing.T) {
	e := newEnv(t)
	if err := e.db.AutoMigrate(catalogModels...); err != nil {
		t.Fatalf("automigrate catalog: %v", err)
	}
	if err := e.db.AutoMigrate(&model.ProductPriceHistoryModel{}, &model.StockBalanceModel{}); err != nil {
		t.Fatalf("automigrate history and stock: %v", err)
	}
	e.exec(`INSERT INTO units_of_measure (id, name, abbreviation, sort_order) VALUES (gen_random_uuid(), 'Unit', 'UNIT', 1)`)

	e.up()
	e.assertVersion(6, false)
	if !e.hasRelation("product_price_history") || !e.hasRelation("stock_balance") {
		t.Errorf("the tables created by AutoMigrate disappeared")
	}
}

// reservedCodesEnv leaves a database on version 3: deleted products still
// free their SKU and barcode there.
func reservedCodesEnv(t *testing.T) *env {
	t.Helper()
	e := newEnv(t)
	e.up()
	if err := migrations.Down(e.dsn); err != nil {
		t.Fatalf("Down() to version 5: %v", err)
	}
	e.assertVersion(5, false)
	if err := migrations.Down(e.dsn); err != nil {
		t.Fatalf("Down() to version 4: %v", err)
	}
	e.assertVersion(4, false)
	if err := migrations.Down(e.dsn); err != nil {
		t.Fatalf("Down() to version 3: %v", err)
	}
	e.assertVersion(3, false)
	e.exec(`INSERT INTO categories (id, company_id, name, created_at) VALUES (?, ?, 'Bebidas', now())`, e.category, e.company)
	return e
}

func (e *env) product(name, sku, barcode string) *model.ProductModel {
	e.t.Helper()
	p := e.newProduct(name, &sku)
	if barcode != "" {
		p.Barcode = &barcode
	}
	if err := e.db.Create(p).Error; err != nil {
		e.t.Fatalf("create %s: %v", name, err)
	}
	return p
}

func (e *env) softDelete(id uuid.UUID) {
	e.exec(`UPDATE products SET deleted_at = now() WHERE id = ?`, id)
}

func TestReservedCodesKeepDeletedValuesTaken(t *testing.T) {
	e := reservedCodesEnv(t)
	deleted := e.product("Descontinuado", "RES-1", "7750000000001")
	e.softDelete(deleted.ID)

	e.up()
	e.assertVersion(6, false)
	for name, p := range map[string]*model.ProductModel{
		"same SKU": e.newProduct("Nuevo 1", ptr("RES-1")),
		"same barcode": func() *model.ProductModel {
			p := e.newProduct("Nuevo 2", ptr("RES-2"))
			p.Barcode = ptr("7750000000001")
			return p
		}(),
	} {
		if err := e.db.Create(p).Error; err == nil {
			t.Errorf("%s of a deleted product was accepted", name)
		}
	}
	for i, name := range []string{"Sin SKU 1", "Sin SKU 2"} {
		p := e.newProduct(name, nil)
		if err := e.db.Create(p).Error; err != nil {
			t.Errorf("product %d without SKU and barcode: %v", i, err)
		}
	}

	// Rolling back frees the values of deleted products again.
	for _, version := range []int{5, 4, 3} {
		if err := migrations.Down(e.dsn); err != nil {
			t.Fatalf("Down() to version %d: %v", version, err)
		}
		e.assertVersion(version, false)
	}
	e.product("Reutiliza", "RES-1", "7750000000001")
}

func TestReservedCodesRefuseValuesAlreadyShared(t *testing.T) {
	e := reservedCodesEnv(t)
	deleted := e.product("Descontinuado", "DUP-1", "7750000000001")
	e.softDelete(deleted.ID)
	live := e.product("Vigente", "dup-1 ", "7750000000001")

	err := migrations.Up(e.dsn)
	if err == nil || !strings.Contains(err.Error(), "SKUs shared by several products") || !strings.Contains(err.Error(), "DUP-1 (2 products)") {
		t.Fatalf("expected the list of shared SKUs, got %v", err)
	}
	e.assertVersion(3, false)
	var sku string
	e.db.Raw(`SELECT sku FROM products WHERE id = ?`, live.ID).Scan(&sku)
	if sku != "dup-1 " {
		t.Errorf("the failed migration changed data: sku = %q", sku)
	}

	e.exec(`UPDATE products SET sku = 'DUP-2' WHERE id = ?`, live.ID)
	err = migrations.Up(e.dsn)
	if err == nil || !strings.Contains(err.Error(), "barcodes shared by several products") {
		t.Fatalf("expected the list of shared barcodes, got %v", err)
	}
	e.assertVersion(3, false)

	e.exec(`UPDATE products SET barcode = '7750000000002' WHERE id = ?`, live.ID)
	e.up()
	e.assertVersion(6, false)
}

func TestReservedCodesNormalizeStoredValues(t *testing.T) {
	e := reservedCodesEnv(t)
	legacy := e.product("Antiguo", " abc-1 ", " 7750000000009 ")
	blank := e.product("En blanco", "  ", "   ")

	e.up()
	var rows []struct {
		ID      uuid.UUID
		SKU     *string
		Barcode *string
	}
	e.db.Raw(`SELECT id, sku, barcode FROM products WHERE id IN (?, ?)`, legacy.ID, blank.ID).Scan(&rows)
	for _, r := range rows {
		switch r.ID {
		case legacy.ID:
			if r.SKU == nil || *r.SKU != "ABC-1" || r.Barcode == nil || *r.Barcode != "7750000000009" {
				t.Errorf("legacy row = %v / %v, want ABC-1 / 7750000000009", deref(r.SKU), deref(r.Barcode))
			}
		case blank.ID:
			if r.SKU != nil || r.Barcode != nil {
				t.Errorf("blank values must become NULL, got %q / %q", deref(r.SKU), deref(r.Barcode))
			}
		}
	}
}

func ptr(s string) *string { return &s }

func deref(s *string) string {
	if s == nil {
		return "<nil>"
	}
	return *s
}

func TestInventoryBranchesAndMovementsMigration(t *testing.T) {
	e := newEnv(t)
	e.up()
	e.assertVersion(6, false)
	for _, table := range []string{"branches", "locations", "stock_movements"} {
		if !e.hasRelation(table) {
			t.Fatalf("table %s was not created", table)
		}
	}
}

func TestDownWithoutMigrationsIsANoOp(t *testing.T) {
	e := newEnv(t)
	if err := migrations.Down(e.dsn); err != nil {
		t.Fatalf("Down() on an empty database error = %v", err)
	}
	status, err := migrations.Status(e.dsn)
	if err != nil || len(status) != 6 || status[0].Applied || status[1].Applied || status[2].Applied || status[3].Applied || status[4].Applied || status[5].Applied {
		t.Fatalf("Status() = %+v, %v", status, err)
	}
}

func TestCategoryNamesIgnoreCase(t *testing.T) {
	e := newEnv(t)
	e.up()
	other := uuid.New()
	root := model.CategoryModel{CompanyID: e.company, Name: "Mascotas"}
	if err := e.db.Create(&root).Error; err != nil {
		t.Fatal(err)
	}
	child := model.CategoryModel{CompanyID: e.company, ParentID: &root.ID, Name: "Perros"}
	if err := e.db.Create(&child).Error; err != nil {
		t.Fatal(err)
	}

	for name, category := range map[string]model.CategoryModel{
		"root":  {CompanyID: e.company, Name: "mascotas"},
		"child": {CompanyID: e.company, ParentID: &root.ID, Name: "PERROS"},
	} {
		if err := e.db.Create(&category).Error; !isPgError(err, "23505") {
			t.Errorf("%s repeated with another case must be rejected, got %v", name, err)
		}
	}
	// Other level, other company and deleted rows do not conflict.
	for name, category := range map[string]model.CategoryModel{
		"same name under another parent": {CompanyID: e.company, ParentID: &child.ID, Name: "MASCOTAS"},
		"same name in another company":   {CompanyID: other, Name: "MASCOTAS"},
	} {
		if err := e.db.Create(&category).Error; err != nil {
			t.Errorf("%s: %v", name, err)
		}
	}
	e.exec(`UPDATE categories SET deleted_at = now() WHERE id = ?`, child.ID)
	if err := e.db.Create(&model.CategoryModel{CompanyID: e.company, ParentID: &root.ID, Name: "PERROS"}).Error; err != nil {
		t.Errorf("name of a deleted category must be free: %v", err)
	}

	// Rolling back compares the exact name again.
	if err := migrations.Down(e.dsn); err != nil {
		t.Fatal(err)
	}
	e.assertVersion(5, false)
	if err := e.db.Create(&model.CategoryModel{CompanyID: e.company, Name: "MASCOTAS"}).Error; err != nil {
		t.Errorf("after rolling back, a different case is a different name: %v", err)
	}
}

func TestCategoryNamesRefuseDuplicatesAlreadyStored(t *testing.T) {
	e := newEnv(t)
	e.up()
	if err := migrations.Down(e.dsn); err != nil {
		t.Fatal(err)
	}
	e.assertVersion(5, false)

	parent := uuid.New()
	e.exec(`INSERT INTO categories (id, company_id, name, created_at) VALUES (?, ?, 'Padre', now())`, parent, e.company)
	insert := `INSERT INTO categories (id, company_id, parent_id, name, deleted_at, created_at) VALUES (gen_random_uuid(), ?, ?, ?, ?, now())`
	e.exec(insert, e.company, nil, "Bebidas", nil)
	e.exec(insert, e.company, nil, "BEBIDAS", nil)
	e.exec(insert, e.company, parent, "Aceites", nil)
	e.exec(insert, e.company, nil, "bebidas", time.Now()) // deleted: does not count

	err := migrations.Up(e.dsn)
	if err == nil || !strings.Contains(err.Error(), "repeated at the same level ignoring case") || !strings.Contains(err.Error(), "BEBIDAS / Bebidas (company") {
		t.Fatalf("expected the list of repeated names, got %v", err)
	}
	if strings.Contains(err.Error(), "Aceites") {
		t.Errorf("a name used once under a parent must not be reported: %v", err)
	}
	e.assertVersion(5, false)

	e.exec(`UPDATE categories SET name = 'Bebidas frías' WHERE name = 'BEBIDAS' AND deleted_at IS NULL`)
	e.up()
	e.assertVersion(6, false)
}
