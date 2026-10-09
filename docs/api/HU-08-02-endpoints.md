# Reporte de endpoints — [HU-08-02] Búsqueda de clientes

## Información general

- HU: HU-08-02 — Búsqueda de clientes (RF-08, P0, Sprint 1). Depende de HU-08-01.
- Rama: `feat/HU-08-02-clients-search`
- Prefijo base: `/api/v1`
- Requiere autenticación: token JWT válido en el esquema de autorización de portador (`Authorization: Bearer <access_token>`). El middleware global `AuthRequired` valida el token y que la sesión siga activa.
- Empresa: la búsqueda se limita siempre a la empresa del claim `company_id` del token. La empresa **nunca** se recibe por query. Un usuario sin empresa (por ejemplo, el superadmin global `admin@sigif.com`) recibe `400 company is required`.
- Formato de respuesta exitosa: `{ "success": true, "data": [...], "meta": { ... } }`.
- Formato de respuesta de error: `{ "success": false, "error": { "code": "CODIGO", "message": "descripción", "details": { "campo": "motivo" } } }`. `details` solo aparece en errores de validación.
- Colección Bruno: `bruno/Customer/Search/` (001–021). Cada petición incluye sus `tests`.

---

## Resumen rápido

| Método | Ruta | Permiso | Descripción |
|---|---|---|---|
| `GET` | `/api/v1/customers/` | `customers.list` | Busca y lista clientes de la empresa con filtros, orden y paginación. |

---

## Autenticación / permisos

| Header | Obligatorio | Descripción |
|---|---|---|
| `Authorization` | Sí | `Bearer <access_token>` de una sesión activa. |

| Permiso | Endpoint | Roles sembrados que lo tienen |
|---|---|---|
| `customers.list` | `GET /api/v1/customers/` | `superadmin` |

El permiso ya existía en `AllPermissions()` (HU-08-01); esta HU no agrega permisos. `customers.read` **no** basta para listar (ese permiso es para `GET /api/v1/customers/:id`, HU-08-03). El rol `business_admin` no lo tiene y recibe **403**.

---

## 1) Buscar clientes — `GET /api/v1/customers/`

Devuelve los clientes no eliminados de la empresa autenticada que coinciden con el criterio. La operación es de **solo lectura**: no modifica ningún registro.

### Permiso
`customers.list`

### Params de ruta
Ninguno.

### Query params

| Parámetro | Obligatorio / por defecto | Tipo | Descripción |
|---|---|---|---|
| `q` | Opcional | string | Búsqueda general. Se recortan espacios; si queda vacía se ignora. Máximo 100 caracteres. Coincidencia **parcial** e **insensible a mayúsculas** sobre `legal_name`, `document_number`, `phone` y `email`. `%` y `_` se buscan literalmente (no son comodines). |
| `status` | Opcional, por defecto `active` | enum | `active` \| `inactive` \| `blocked` \| `all`. `all` devuelve todos los estados. |
| `page` | Opcional, por defecto `1` | integer ≥ 1 | Página solicitada. |
| `limit` | Opcional, por defecto `20` | integer 1–100 | Tamaño de página. El máximo sale de `pagination.max_limit` (100 por defecto). |
| `sort_by` | Opcional, por defecto `created_at` | enum | `legal_name` \| `created_at`. Solo esta lista blanca. |
| `sort_order` | Opcional, por defecto `desc` | enum | `asc` \| `desc`. |

Sin parámetros, el endpoint devuelve los clientes **activos**, del más reciente al más antiguo.

### Ejemplo de request

```bash
curl -s "http://localhost:4600/api/v1/customers/?q=ferreteria&status=active&page=1&limit=20&sort_by=legal_name&sort_order=asc" \
  -H "Authorization: Bearer $ACCESS_TOKEN"
```

### Respuesta exitosa

**`200 OK`** — con resultados

```json
{
  "success": true,
  "data": [
    {
      "id": "86d60c05-c7c6-4e4f-89b2-8a308817300e",
      "company_id": "7dcfcd06-724a-4dd7-a10e-6ba716edf3cc",
      "legal_name": "Ferreteria Central",
      "document_type": "tax_id",
      "document_number": "1020304050",
      "phone": "+59170000001",
      "email": "central@example.com",
      "status": "active"
    }
  ],
  "meta": {
    "page": 1,
    "limit": 20,
    "total": 1,
    "total_pages": 1
  }
}
```

El listado devuelve solo los campos de identificación (`CustomerSummary`). `id` es el valor que se usa en `GET /api/v1/customers/:id` para ver el detalle completo (HU-08-03). Los campos opcionales nulos (`document_number`, `phone`, `email`) se omiten.

**`200 OK`** — sin coincidencias (estado vacío, no es 404)

```json
{
  "success": true,
  "data": [],
  "meta": {
    "page": 1,
    "limit": 20,
    "total": 0,
    "total_pages": 0
  }
}
```

Una página posterior a la última también responde `200` con `data: []` y el `total` real.

### Códigos de error

| Code | HTTP | Causa |
|---|---|---|
| `VALIDATION_ERROR` | 400 | `q` > 100 caracteres, `page` < 1, `limit` < 1 o > máximo, `status`, `sort_by` o `sort_order` fuera de los valores permitidos. |
| `BAD_REQUEST` | 400 | `page` o `limit` no numéricos (`invalid query parameters`), o el token no tiene empresa (`company is required`). |
| `UNAUTHORIZED` | 401 | Falta el token o es inválido/expiró, o la sesión fue cerrada. |
| `FORBIDDEN` | 403 | El usuario no tiene el permiso `customers.list`. |
| `INTERNAL_ERROR` | 500 | Error inesperado de persistencia. |

**`400`** — validación

```json
{
  "success": false,
  "error": {
    "code": "VALIDATION_ERROR",
    "message": "validation failed",
    "details": {
      "sort_by": "sort_by must be one of [legal_name created_at]",
      "status": "status must be one of [active inactive blocked all]"
    }
  }
}
```

**`401`**

```json
{
  "success": false,
  "error": {
    "code": "UNAUTHORIZED",
    "message": "unauthorized"
  }
}
```

**`403`**

```json
{
  "success": false,
  "error": {
    "code": "FORBIDDEN",
    "message": "forbidden"
  }
}
```

---

## Reglas de negocio

- **Empresa:** el filtro `company_id` se toma del token y se aplica siempre; los clientes de otra empresa nunca aparecen, aunque coincidan el nombre o el documento.
- **Estado (criterio 8):** por defecto solo se listan clientes `active`. Los `inactive` y `blocked` se consultan explícitamente con `status`, o todos con `status=all`. Los clientes eliminados (soft delete, `deleted_at` informado) nunca se listan.
- **Solo lectura:** conteo y página se ejecutan dentro de una única transacción `READ ONLY` con aislamiento `REPEATABLE READ`, así `total` y `data` corresponden a la misma foto de los datos y la base rechaza cualquier escritura.
- **Orden estable:** al ordenar se añade `customer_id` como desempate, para que la paginación no repita ni salte registros.

---

## Índices

El proyecto crea el esquema con `AutoMigrate` (`make migrate-up`), por lo que los índices se declaran en `CustomerModel`:

| Índice | Columnas | Uso |
|---|---|---|
| `idx_customer_company_status` | `(company_id, status)` | Filtro obligatorio por empresa + estado. **Nuevo.** |
| `idx_customer_company_document` | `(company_id, document_number)` | Búsqueda por documento dentro de la empresa. **Nuevo.** |
| `idx_customer_company_name` | `(company_id, legal_name)` | Ya existía (HU-08-01); sirve al orden por nombre. |

Nota: la coincidencia parcial `LOWER(col) LIKE '%texto%'` no puede usar índices B-tree; estos índices acotan las filas por empresa y estado. Si el volumen de clientes por empresa crece, conviene evaluar la extensión `pg_trgm` con índices GIN sobre `LOWER(legal_name)` y `document_number` (no activada en esta HU), y `unaccent` si se quiere ignorar tildes.

---

## Cobertura de criterios de aceptación

| Criterio HU-08-02 | Cómo se cubre |
|---|---|
| 1. Ver clientes registrados | `GET /api/v1/customers/` sin filtros lista los activos, paginados. |
| 2. Introducir un criterio | Parámetro `q` (más `status`, `sort_by`, `sort_order`). |
| 3. Solo los que coinciden | Filtro parcial sobre nombre, documento, teléfono y email, dentro de la empresa. |
| 4. Información para identificar | `CustomerSummary`: id, nombre, tipo y número de documento, teléfono, email, estado. |
| 5. Estado vacío | `200` con `data: []` y `meta.total: 0`. |
| 6. No modifica datos | Transacción de solo lectura; probado en tests (`updated_at` sin cambios). |
| 7. Seleccionar un resultado | Cada ítem trae `id` para `GET /api/v1/customers/:id`. |
| 8. Reglas de estado | Por defecto `active`; `inactive`, `blocked` o `all` bajo demanda; eliminados excluidos. |

---

## Pruebas

- Unitarias HTTP: `internal/modules/customer/interfaces/http/handler/list_handler_test.go` (`go test ./...`).
- Integración con PostgreSQL: `internal/modules/customer/infrastructure/persistence/gorm/repository_integration_test.go`:

```bash
SIGIF_TEST_DATABASE_DSN="host=localhost port=5432 user=sigif password=sigif dbname=sigif_test sslmode=disable" \
  go test -tags integration ./internal/modules/customer/...
```

- Bruno (requiere la API levantada y la base sembrada; usa `admin.empresa@sigif.com`, que sí tiene empresa):

```bash
cd bruno
bru run Customer/Search --env development \
  --env-var baseUrl=http://localhost:4600 \
  --env-var companyAdminPassword=<contraseña del seed>
```

La carpeta crea sus propios datos (dos clientes con un prefijo único `HU0802-<timestamp>` y un usuario `business_admin` sin permisos de clientes), por lo que puede ejecutarse varias veces.

---

## Diferencias respecto a la versión de HU-08-01

El listado ya existía en HU-08-01; esta HU define su contrato de búsqueda:

1. `status` por defecto pasa a ser `active` (antes devolvía todos los estados) y se agrega `status=all`.
2. Parámetros inválidos responden `400` (antes `status` inválido se ignoraba y `page`/`limit` fuera de rango se ajustaban en silencio).
3. Se agregan `sort_by` y `sort_order`, y el límite de 100 caracteres en `q`.
4. Los ítems del listado son `CustomerSummary` (sin `address`, `credit_limit`, `credit_balance`, `points_accrued`, `created_at`, `updated_at`); el detalle completo queda en `GET /api/v1/customers/:id`.
5. `%` y `_` en `q` se buscan literalmente.
6. Corrección: un token sin empresa ahora corta la petición con `400 company is required` en todas las rutas de clientes; antes se respondía 400 pero la operación seguía ejecutándose con `company_id` vacío.
