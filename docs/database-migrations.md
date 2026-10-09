# Migraciones versionadas de base de datos

El catálogo (productos, categorías, unidades de medida e impuestos) ya no cambia su esquema con `AutoMigrate`. Cada cambio es un archivo SQL numerado que se aplica una sola vez, en orden, en todas las bases. Usamos [golang-migrate](https://github.com/golang-migrate/migrate).

## Por qué

`AutoMigrate` añade columnas nuevas pero nunca renombra ni borra las viejas. Al cambiar `cost_price` por `cost` y `unit_of_measure` por `unit_of_measure_id`, las bases que ya existían quedaron con las dos versiones mezcladas y `POST /api/v1/products` respondía 500. Con migraciones versionadas, cada base se pone al día sola y conserva sus datos.

## Alcance

| Tablas | Quién cambia el esquema |
|---|---|
| `products`, `categories`, `units_of_measure`, `taxes` | SQL versionado (`internal/shared/database/migrations/sql/`) |
| `app_user`, `role`, `permission`, `role_permission`, `company`, `customer`, `user_session`, `login_attempt` | `AutoMigrate`, como hasta ahora |

Los demás módulos siguen con `AutoMigrate` para no romper las ramas que están en curso. Para pasar un módulo a SQL versionado: crear una migración que deje sus tablas como están hoy (con `IF NOT EXISTS`), quitar sus modelos de `autoMigrate()` en `cmd/migrate/main.go` y, desde ahí, hacer cada cambio con un archivo nuevo.

## Comandos

| Comando | Qué hace |
|---|---|
| `make migrate-up` | Aplica `AutoMigrate` de los otros módulos, las migraciones pendientes y los seeds. Es el de siempre. |
| `make migrate-status` | Lista las migraciones y si están aplicadas o pendientes. |
| `make migrate-create NAME=add_brand_to_products` | Crea el siguiente par de archivos vacíos (`.up.sql` y `.down.sql`). |
| `make migrate-down` | Deshace la última migración aplicada. |
| `make migrate-force VERSION=n` | Fija la versión sin ejecutar SQL. Solo para el caso descrito en "Si queda marcada como dirty". |

Todos usan la base configurada en `configs/config.yaml` o en las variables `SIGIF_DATABASE_*`.

Ejemplo de `make migrate-status`:

```
000001  applied  catalog_baseline
000002  applied  catalog_contract
```

## Si ya tenías una base local

No hay que borrar nada. Con `git pull` y `make migrate-up` basta: la migración `000002` convierte el esquema anterior y conserva los productos y categorías.

## Cómo agregar un cambio de esquema

1. Crear los archivos: `make migrate-create NAME=descripcion_corta`.
2. Escribir el cambio en el `.up.sql` y cómo deshacerlo en el `.down.sql`.
3. Actualizar el modelo GORM y el mapper para que coincidan con el SQL.
4. Probar con `make migrate-up` y agregar un caso al test de integración si el cambio mueve datos.

Reglas:

- **Nunca editar una migración que ya está en `dev`.** Para corregirla, crear otra.
- **No usar `BEGIN`/`COMMIT` ni `CREATE INDEX CONCURRENTLY`.** Cada archivo se ejecuta como una sola transacción: si una sentencia falla, Postgres deshace el archivo entero. Esas sentencias rompen esa garantía.
- **Si un cambio no se puede deshacer sin perder datos**, el `.down.sql` debe fallar con un mensaje claro (ver `000002_catalog_contract.down.sql`) en lugar de aparentar que funcionó.
- Los datos de referencia que otras tablas necesitan (unidades, impuestos) van en la migración, con `ON CONFLICT DO NOTHING`.

## Si una migración falla

No queda nada a medias: el archivo se deshace completo y la versión vuelve a la última aplicada. Corregir el dato o el SQL que indica el error y ejecutar `make migrate-up` otra vez.

### Si queda marcada como dirty

Solo pasa si el proceso se corta mientras se ejecuta una migración (por ejemplo, con Ctrl+C). El siguiente `make migrate-up` responde `database is marked dirty at version N`. Postgres ya deshizo esa migración, así que:

```sh
make migrate-status              # ver cuál es la última realmente aplicada
make migrate-force VERSION=1     # el número de esa última (o -1 si no hay ninguna)
make migrate-up
```

## Qué hacen las migraciones actuales

### 000001_catalog_baseline

Crea `categories` y `products` como estaban en `dev` antes de la HU-03-01 (con `unit_of_measure` y `cost_price`). Si las tablas ya existen, no hace nada. Su `down` las borra.

### 000002_catalog_contract

Lleva el catálogo al modelo actual. Funciona sobre una base vacía, sobre el esquema anterior con o sin datos, y sobre una base que ya tenía el modelo nuevo creado por `AutoMigrate`.

- Crea `units_of_measure` y `taxes` e inserta las 6 unidades y los 3 impuestos del contrato.
- Convierte `products.unit_of_measure` (`kg`, `unit`, ...) en `unit_of_measure_id`. Si había productos en `g`, `ml` o `m`, crea esas unidades para no perder el dato.
- Copia `cost_price` a `cost` y borra las dos columnas viejas.
- Agrega `stock`, `min_stock` y `last_movement_at` (para los productos existentes, su fecha de creación).
- Deja `sku` como opcional.
- Agrega a `categories` `parent_id`, `default_tax_id` y `target_margin`.
- Reemplaza los índices únicos: nombre de producto por empresa, SKU y código de barras solo cuando existen, y nombre de categoría por nivel.

Dos decisiones a tener en cuenta:

- **Nombres de producto repetidos.** El modelo anterior los permitía y el actual no. El más antiguo conserva su nombre; a los demás se les agrega el SKU: `Arena gato` pasa a `Arena gato (OLD-3)`.
- **Unidad desconocida.** Si un producto tiene una unidad que no es ninguna de las anteriores, la migración falla e indica el valor, sin adivinar.

No tiene `down` automático: reconstruir las columnas viejas perdería información. Para volver atrás hay que restaurar un respaldo.

## Pruebas

```sh
go test ./...

SIGIF_TEST_DATABASE_DSN='host=localhost port=5432 user=sigif password=sigif dbname=sigif sslmode=disable' \
  go test -tags=integration ./internal/shared/database/... ./internal/modules/product/infrastructure/...
```

Los tests de integración crean un esquema de Postgres con nombre aleatorio por cada test y lo borran al terminar, así que no tocan las tablas de `public`. Sin `SIGIF_TEST_DATABASE_DSN` se omiten.

Cubren: base vacía, esquema anterior con datos, el esquema mezclado que causaba el 500, una base ya creada con el modelo nuevo, fallo con vuelta atrás completa y reintento, base marcada como dirty, y `down`.
