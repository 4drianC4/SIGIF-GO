package dbtest

import (
	"database/sql"
	"net/url"
	"os"
	"strings"
	"testing"

	"github.com/google/uuid"
	_ "github.com/jackc/pgx/v5/stdlib"
)

const EnvDSN = "SIGIF_TEST_DATABASE_DSN"

func NewSchemaDSN(t *testing.T) string {
	t.Helper()
	dsn := os.Getenv(EnvDSN)
	if dsn == "" {
		t.Skip(EnvDSN + " no está definido")
	}

	admin, err := sql.Open("pgx", dsn)
	if err != nil {
		t.Fatalf("open test database: %v", err)
	}
	schema := "test_" + strings.ReplaceAll(uuid.NewString(), "-", "")
	if _, err := admin.Exec(`CREATE SCHEMA "` + schema + `"`); err != nil {
		_ = admin.Close()
		t.Fatalf("create test schema: %v", err)
	}
	t.Cleanup(func() {
		_, _ = admin.Exec(`DROP SCHEMA "` + schema + `" CASCADE`)
		_ = admin.Close()
	})
	return withSearchPath(dsn, schema)
}

func withSearchPath(dsn, schema string) string {
	if strings.Contains(dsn, "://") {
		if u, err := url.Parse(dsn); err == nil {
			q := u.Query()
			q.Set("search_path", schema)
			u.RawQuery = q.Encode()
			return u.String()
		}
	}
	return dsn + " search_path=" + schema
}
