# Reporte de endpoints — HU-03-01 Registro de producto y categoría

## Información general

- HU: HU-03-01 — Registro de producto y categoría
- Rama: `feature/HU-03-01-product-and-category-registration`
- Prefijo base: `/api/v1`
- Autenticación: todas las rutas requieren `Authorization: Bearer <access_token>` obtenido con `POST /api/v1/auth/login`. El middleware global `AuthRequired` valida el token y que la sesión siga activa (tras `logout` el token deja de servir).
- Empresa: los productos y categorías pertenecen a la empresa del usuario, que se toma del claim `company_id` del token. Un usuario sin empresa (por ejemplo, un superadmin global) recibe `400 company is required`.
- Formato de respuesta exitosa: `{ "success": true, "data": ... }`.
- Formato de respuesta de error: `{ "success": false, "error": { "code": "CODIGO", "message": "descripción", "details": { "campo": "motivo" } } }`. `details` solo aparece en errores de validación.
- Colección Bruno: `bruno/Products/` (001–009). Cada petición incluye sus `tests`. Usa las variables de entorno `baseUrl` (`http://localhost:8080`) y `accessToken` (pegar el `access_token` del login).

---

## Resumen rápido

| Método | Ruta | Permiso | Descripción |
|---|---|---|---|
| `POST` | `/api/v1/categories` | `categories.create` | Registra una categoría activa. |
| `POST` | `/api/v1/products` | `products.create` | Registra un producto activo dentro de una categoría. |

---

## Autenticación / permisos (todas las rutas)

| Header | Obligatorio | Descripción |
|---|---|---|
| `Authorization` | Sí | `Bearer <access_token>` de una sesión activa. |
| `Content-Type` | Sí | `application/json` |

Los permisos se comprueban con el RBAC de la base de datos (`RequirePermission`):

| Permiso | Endpoint | Roles sembrados que lo tienen |
|---|---|---|
| `categories.create` | `POST /api/v1/categories` | `superadmin` |
| `products.create` | `POST /api/v1/products` | `superadmin` |

Ambos permisos se añadieron al catálogo de `AllPermissions()`; `make migrate-up` los crea y se los asigna al superadmin también en bases ya existentes. El rol `soporte` no los tiene y recibe **403**. Para darlos a otro rol, asignarlos en `role_permission`.

---

## 1) Registrar categoría — `POST /api/v1/categories`

### Body

```json
{
  "name": "Bebidas",
  "description": "Gaseosas, jugos y aguas"
}
```

| Campo | Tipo | Obligatorio | Restricciones / notas |
|---|---|---|---|
| `name` | string | Sí | 2–100 caracteres tras quitar espacios. Los espacios repetidos se colapsan (`"  Bebidas   frías "` → `"Bebidas frías"`). **Único por empresa sin distinguir mayúsculas.** |
| `description` | string | No | Máximo 500 caracteres. |

### Respuesta exitosa

**`201`**
```json
{
  "success": true,
  "data": {
    "id": "b45978d0-37f8-4051-92a8-218441dac948",
    "company_id": "aaaaaaaa-aaaa-4aaa-8aaa-aaaaaaaaaaaa",
    "name": "Bebidas",
    "description": "Gaseosas, jugos y aguas",
    "status": "active",
    "created_at": "2026-10-02T23:40:27Z",
    "updated_at": "2026-10-02T23:40:27Z"
  }
}
```

### Códigos de error

| Code | HTTP | Causa |
|---|---|---|
| `BAD_REQUEST` | 400 | El body no es JSON válido. |
| `VALIDATION_ERROR` | 400 | `name` vacío, demasiado corto o largo; `description` demasiado larga. |
| `BAD_REQUEST` | 400 | El token no tiene empresa (`company is required`). |
| `UNAUTHORIZED` | 401 | Falta el token, es inválido, expiró o su sesión se cerró. |
| `FORBIDDEN` | 403 | El usuario no tiene el permiso `categories.create`. |
| `CONFLICT` | 409 | Ya existe una categoría con ese nombre en la empresa (`"BEBIDAS"` = `"bebidas"`). |
| `INTERNAL_ERROR` | 500 | Error no controlado. |

---

## 2) Registrar producto — `POST /api/v1/products`

### Body

```json
{
  "category_id": "b45978d0-37f8-4051-92a8-218441dac948",
  "sku": "coca-600",
  "barcode": "7750182000123",
  "name": "Coca Cola 600ml",
  "description": "Botella PET 600ml",
  "unit_of_measure": "unit",
  "cost_price": 3.20,
  "sale_price": "5.50"
}
```

| Campo | Tipo | Obligatorio | Restricciones / notas |
|---|---|---|---|
| `category_id` | uuid | Sí | Debe existir, pertenecer a la empresa y estar activa. |
| `sku` | string | Sí | Máximo 50 caracteres. Solo letras, números, `.`, `-` y `_`; debe empezar por letra o número. **Se guarda en mayúsculas** (`coca-600` → `COCA-600`). **Único por empresa.** |
| `barcode` | string | No | Solo dígitos, 8–14 caracteres (EAN-8, UPC-A, EAN-13, GTIN-14). **Único por empresa** si se envía. |
| `name` | string | Sí | 2–150 caracteres tras quitar espacios. Los espacios repetidos se colapsan. |
| `description` | string | No | Máximo 1000 caracteres. |
| `unit_of_measure` | enum | Sí | `unit` \| `kg` \| `g` \| `l` \| `ml` \| `m` \| `box` \| `pack` |
| `cost_price` | number \| string | No | ≥ 0, como máximo 2 decimales. Por defecto `0`. |
| `sale_price` | number \| string | Sí | > 0, como máximo 2 decimales. |

> Los precios aceptan número (`5.5`) o string (`"5.50"`). En la respuesta siempre se devuelven como **string con 2 decimales** para no perder precisión.

El producto se crea siempre con `status: "active"`.

### Respuesta exitosa

**`201`**
```json
{
  "success": true,
  "data": {
    "id": "e1844a92-0665-45bc-b7fe-a2516fbc458d",
    "company_id": "aaaaaaaa-aaaa-4aaa-8aaa-aaaaaaaaaaaa",
    "category_id": "b45978d0-37f8-4051-92a8-218441dac948",
    "sku": "COCA-600",
    "barcode": "7750182000123",
    "name": "Coca Cola 600ml",
    "description": "Botella PET 600ml",
    "unit_of_measure": "unit",
    "cost_price": "3.20",
    "sale_price": "5.50",
    "status": "active",
    "created_at": "2026-10-02T23:40:27Z",
    "updated_at": "2026-10-02T23:40:27Z"
  }
}
```

### Ejemplo de error de validación

**`400`**
```json
{
  "success": false,
  "error": {
    "code": "VALIDATION_ERROR",
    "message": "validation failed",
    "details": {
      "barcode": "barcode must be a valid numeric value",
      "category_id": "category_id is a required field",
      "name": "name is a required field",
      "sale_price": "sale_price is a required field",
      "sku": "sku is a required field",
      "unit_of_measure": "unit_of_measure must be one of [unit kg g l ml m box pack]"
    }
  }
}
```

### Códigos de error

| Code | HTTP | Causa |
|---|---|---|
| `BAD_REQUEST` | 400 | El body no es JSON válido o un precio no es numérico (`"abc"`). |
| `VALIDATION_ERROR` | 400 | Campos obligatorios vacíos o con formato inválido; `sale_price` ≤ 0; `cost_price` negativo; precio con más de 2 decimales; SKU con caracteres no permitidos. Ver `details`. |
| `BAD_REQUEST` | 400 | La categoría existe pero está inactiva (`category is inactive`). |
| `BAD_REQUEST` | 400 | El token no tiene empresa (`company is required`). |
| `UNAUTHORIZED` | 401 | Falta el token, es inválido, expiró o su sesión se cerró. |
| `FORBIDDEN` | 403 | El usuario no tiene el permiso `products.create`. |
| `NOT_FOUND` | 404 | La categoría no existe en la empresa del usuario. |
| `CONFLICT` | 409 | Ya existe un producto con ese SKU, o con ese código de barras, en la empresa. |
| `INTERNAL_ERROR` | 500 | Error no controlado. |

---

## Reglas de unicidad (BE-04)

| Regla | Comprobación previa (servicio) | Respaldo en base de datos |
|---|---|---|
| Nombre de categoría por empresa | `LOWER(name)`, sin distinguir mayúsculas | Índice único parcial `idx_categories_company_name (company_id, name) WHERE deleted_at IS NULL` |
| SKU por empresa | SKU normalizado a mayúsculas | Índice único parcial `idx_products_company_sku (company_id, sku) WHERE deleted_at IS NULL` |
| Código de barras por empresa | Solo si se envía | Índice único parcial `idx_products_company_barcode (company_id, barcode) WHERE deleted_at IS NULL AND barcode IS NOT NULL` |

- El servicio comprueba **antes de persistir**. Si dos peticiones idénticas llegan a la vez y ambas pasan la comprobación, el índice único rechaza la segunda, y el repositorio la traduce a **409**, no a 500. Probado con 10 peticiones simultáneas: 1 × 201 y 9 × 409.
- Los índices son **parciales**: un producto o categoría borrado lógicamente libera su SKU, código de barras o nombre.
- El mismo SKU, código de barras o nombre **sí** se puede usar en otra empresa.

---

## Pruebas (BE-06)

| Nivel | Archivo | Qué cubre |
|---|---|---|
| Servicio | `internal/modules/product/domain/service/catalog_service_test.go` | Registro válido, duplicados (SKU sin distinguir mayúsculas, código de barras, nombre de categoría), categoría inexistente, de otra empresa o inactiva, reglas de precio y SKU, aislamiento por empresa. |
| HTTP | `internal/modules/product/interfaces/http/handler/catalog_handler_test.go` | 201 / 400 / 401 / 403 / 404 / 409 con `AuthRequired` real; cada permiso habilita solo su endpoint; empresa tomada del token. |
| Integración (Postgres) | `internal/modules/product/infrastructure/persistence/gorm/repository_integration_test.go` | Índices únicos parciales → 409, varios productos sin código de barras, SKU reutilizable tras borrado. |
| API manual | `bruno/Products/001–009` | Flujo completo con `tests` en cada petición. |

```bash
# Unitarias y HTTP (sin base de datos)
go test ./internal/modules/product/... ./internal/shared/middleware/

# Integración (requiere Postgres)
SIGIF_TEST_DATABASE_DSN="host=localhost port=5432 user=sigif password=sigif dbname=sigif sslmode=disable" \
  go test -tags=integration ./internal/modules/product/infrastructure/persistence/gorm/
```
