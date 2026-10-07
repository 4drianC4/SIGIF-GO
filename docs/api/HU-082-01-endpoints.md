# Reporte de endpoints — [HU-082-01] Catálogo de permisos

## Información general

- HU: HU-082-01 — Catálogo de permisos (RF, P1).
- Rama: `feat/HU-082-01-permissions-catalog`.
- Prefijo base: `/api/v1`
- Requiere autenticación: token JWT válido en el esquema de autorización de portador (`Authorization: Bearer <access_token>`). El middleware global `AuthRequired` valida el token y que la sesión siga activa.
- Formato de respuesta exitosa: `{ "success": true, "data": ... }`. El listado incluye `meta`: `{ "page", "limit", "total", "total_pages" }`.
- Formato de respuesta de error: `{ "success": false, "error": { "code": "CODIGO", "message": "descripción", "details": { "campo": "motivo" } } }`.
- Colección Bruno: `bruno/Permission/` (001–011). El caso de creación acepta `201` o `409` para poder ejecutarse varias veces (no existe `DELETE` y el código es un par fijo).

---

## Resumen rápido

| Método | Ruta | Permiso | Descripción |
|---|---|---|---|
| `GET` | `/api/v1/permissions/modules` | `permissions.read` | Catálogo de módulos con sus operaciones soportadas. |
| `GET` | `/api/v1/permissions` | `permissions.read` | Lista paginada de permisos, filtrable por `module`. |
| `POST` | `/api/v1/permissions` | `permissions.create` | Registra un permiso nuevo para asignarlo después a roles. |

Los tres endpoints viven en el módulo `user` (el permiso es la unidad más pequeña de autorización: un par `module.operation`). No dependen de la empresa del token.

---

## Autenticación / permisos

| Header | Obligatorio | Descripción |
|---|---|---|
| `Authorization` | Sí | `Bearer <access_token>` de una sesión activa. |
| `Content-Type` | Sí (con body) | `application/json`. |

| Permiso | Endpoint | Roles sembrados que lo tienen |
|---|---|---|
| `permissions.read` | `GET /api/v1/permissions/modules`, `GET /api/v1/permissions` | `superadmin` |
| `permissions.create` | `POST /api/v1/permissions` | `superadmin` |

`permissions.read` y `permissions.create` se agregan al seed junto con los permisos existentes. El rol `superadmin` recibe automáticamente todos los permisos del catálogo (ahora 21); el rol `soporte` conserva solo `users.list` y `users.read` y por tanto responde `403`.

---

## Catálogo de módulos y operaciones

Módulos soportados (fuente de verdad: `entity.AllPermissions()` y `entity.SupportedModuleCatalog()`):

| Módulo | Operaciones del seed |
|---|---|
| `users` | `create, list, read, update, delete, activate, deactivate, change_password` |
| `categories` | `create` |
| `products` | `create, read, update, delete` |
| `customers` | `create, list, read, update, status, delete` |
| `permissions` | `read, create` |

Operaciones soportadas que puede usar **cualquier** permiso nuevo (catálogo extendido de HU-082-01 + las ya sembradas):

`read`, `list`, `create`, `update`, `delete`, `export`, `activate`, `deactivate`, `change_password`, `status`

**Convención:** el backend deriva el código a `module.operation`. Así, un permiso "Ver" se expresa como la operación `read`.

---

## 1) Módulos — `GET /api/v1/permissions/modules`

### Permiso
`permissions.read`

### Params
Sin parámetros.

### Respuesta exitosa — `200 OK`

```json
{
  "success": true,
  "data": [
    { "module": "users", "operations": ["read", "list", "create", "update", "delete", "export", "activate", "deactivate", "change_password", "status"] },
    { "module": "categories", "operations": ["read", "list", "create", "update", "delete", "export", "activate", "deactivate", "change_password", "status"] },
    { "module": "products", "operations": ["read", "list", "create", "update", "delete", "export", "activate", "deactivate", "change_password", "status"] },
    { "module": "customers", "operations": ["read", "list", "create", "update", "delete", "export", "activate", "deactivate", "change_password", "status"] },
    { "module": "permissions", "operations": ["read", "list", "create", "update", "delete", "export", "activate", "deactivate", "change_password", "status"] }
  ]
}
```

---

## 2) Listar permisos — `GET /api/v1/permissions`

### Permiso
`permissions.read`

### Params de query

| Parámetro | Tipo | Reglas |
|---|---|---|
| `module` | string | Opcional. Máximo 60. Si se envía debe ser un módulo soportado (`400` si no lo es) y la lista trae solo ese módulo. |
| `page` | int | Opcional, mínimo 1. Por defecto `1`. |
| `limit` | int | Opcional, mínimo 1, máximo `max_limit` (100). Por defecto `default_limit` (20). |

### Respuesta exitosa — `200 OK`

```json
{
  "success": true,
  "data": [
    {
      "id": "164e83e5-e30a-488d-9204-469fd54ce525",
      "module": "users",
      "operation": "create",
      "code": "users.create",
      "description": "Registrar usuarios"
    }
  ],
  "meta": { "page": 1, "limit": 20, "total": 22, "total_pages": 2 }
}
```

- El listado se ordena por `module ASC, operation ASC`.
- `code` es `module.operation`, derivado por el backend.
- `total` incluye todos los permisos que cumplen el filtro (la paginación no lo limita).

---

## 3) Registrar permiso — `POST /api/v1/permissions`

### Permiso
`permissions.create`

### Body

| Campo | Tipo | Reglas |
|---|---|---|
| `module` | string | **Obligatorio**, 1–60. Se recorta y pasa a minúsculas. Debe ser un módulo soportado. |
| `operation` | string | **Obligatorio**, 1–60. Se recorta y pasa a minúsculas. Debe ser una operación soportada. |
| `code` | string | Opcional, máximo 80. Se normaliza (recorte + minúsculas). **Si se envía, debe ser `module.operation`**; de lo contrario `400`. |
| `description` | string | Opcional, máximo 200. Se recorta. |

El `id` lo genera el backend (`uuid`). No hay campos de auditoría (`created_at`/`updated_at`).

### Ejemplo de request

```bash
curl -s -X POST "http://localhost:4600/api/v1/permissions/" \
  -H "Authorization: Bearer $ACCESS_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"module":"products","operation":"export","description":"Exportar productos"}'
```

### Respuesta exitosa — `201 Created`

```json
{
  "success": true,
  "data": {
    "id": "e3f876fc-8444-4150-9ea4-ba5c797f2cdd",
    "module": "products",
    "operation": "export",
    "code": "products.export",
    "description": "Exportar productos"
  }
}
```

---

## Códigos de error

| Code | HTTP | Causa |
|---|---|---|
| `VALIDATION_ERROR` | 400 | Módulo u operación fuera del catálogo (`details.module` o `details.operation`), `code` distinto de `module.operation` (`details.code`), campos obligatorios faltantes, formato/longitud inválidos. |
| `BAD_REQUEST` | 400 | El body no es un objeto JSON (malformado, array, `null`). |
| `UNAUTHORIZED` | 401 | Falta el token o es inválido/expiró, o la sesión fue cerrada. |
| `FORBIDDEN` | 403 | El usuario no tiene el permiso requerido (`permissions.read` o `permissions.create`). |
| `CONFLICT` | 409 | El par `(module, operation)` ya está registrado (`details.code = "already registered"`). |
| `INTERNAL_ERROR` | 500 | Error inesperado de persistencia. |

**`400`** — módulo u operación fuera del catálogo

```json
{
  "success": false,
  "error": {
    "code": "VALIDATION_ERROR",
    "message": "unknown module",
    "details": { "module": "not a supported module" }
  }
}
```

**`400`** — operación fuera del catálogo

```json
{
  "success": false,
  "error": {
    "code": "VALIDATION_ERROR",
    "message": "unknown operation",
    "details": { "operation": "not a supported operation" }
  }
}
```

**`400`** — `code` no coincide

```json
{
  "success": false,
  "error": {
    "code": "VALIDATION_ERROR",
    "message": "code does not match module.operation",
    "details": { "code": "must match module.operation" }
  }
}
```

**`400`** — campos obligatorios faltantes

```json
{
  "success": false,
  "error": {
    "code": "VALIDATION_ERROR",
    "message": "validation failed",
    "details": {
      "module": "module is a required field",
      "operation": "operation is a required field"
    }
  }
}
```

**`409`** — par repetido

```json
{
  "success": false,
  "error": {
    "code": "CONFLICT",
    "message": "permission code is already registered",
    "details": { "code": "already registered" }
  }
}
```

**`401`**

```json
{ "success": false, "error": { "code": "UNAUTHORIZED", "message": "unauthorized" } }
```

**`403`**

```json
{ "success": false, "error": { "code": "FORBIDDEN", "message": "forbidden" } }
```

---

## Reglas de negocio

- **El código lo deriva el backend.** `code == module.operation` siempre. El `POST` acepta `code` opcional solo para compatibilidad con clientes; si llega y no coincide con `module.operation`, responde `400`.
- **Normalización.** `module`, `operation` y `code` se recortan y pasan a minúsculas antes de validar y persistir. `" USERS.CREATE "` es duplicado de `users.create` (409).
- **Catálogo cerrado pero extensible.** Un permiso solo puede ser `(módulo soportado, operación del catálogo)`. La fuente de verdad son `entity.AllPermissions()` (módulos y seed) y `entity.SupportedOperations` (operaciones). Para habilitar un módulo nuevo basta agregarlo al catálogo y al seed; el índice y el seed son idempotentes.
- **Unicidad real en la base.** Índice único `idx_permissions_module_operation` sobre `(module, operation)`. El servicio además chequea antes con `GetByModuleOperation`; la violación concurrente (`SQLSTATE 23505`) también se traduce a `409`.
- **Afecta solo a permisos.** No hay roles involucrados: asignar permisos a roles es de una HU posterior. Los permisos nuevos quedan disponibles en `GET /api/v1/permissions` para futura asignación.
- **Sin permiso, sin acceso.** `GET /permissions/*` exige `permissions.read` y `POST /permissions` exige `permissions.create` por separado (RBAC del módulo `user`).

---

## Índice único y seed

```sql
CREATE UNIQUE INDEX idx_permissions_module_operation ON permission (module, operation);
```

Se declara en `PermissionModel` y lo crea `AutoMigrate` (`make migrate-up`), igual que el resto del esquema. El seed inserta `permissions.read` y `permissions.create` y le da a `superadmin` **todos** los permisos del catálogo dinámicamente (ya no se fijan 19: al sembrar un módulo nuevo, `superadmin` lo recibe). Al migrar una base existente limpia, `superadmin` pasa de 19 a 21 permisos; `soporte` se mantiene en 2.

---

## Cobertura de criterios de aceptación

| Criterio | Cómo se cubre |
|---|---|
| Consultar los módulos y permisos | `GET /api/v1/permissions/modules` (catálogo) y `GET /api/v1/permissions` (registrados). |
| Registro con validaciones | `POST` valida módulo y operación contra el catálogo, normaliza y aplica reglas de longitud. |
| Código derivado `module.operation` | `code` se expone derivado; si el cliente envía uno distinto responde `400`. |
| Rechazar duplicados | Chequeo del servicio + índice único → `409 CONFLICT` (también con espacios/mayúsculas). |
| El permiso queda disponible para roles | La respuesta `201` trae el `id`; el permiso aparece en `GET /api/v1/permissions?module=...`. |
| RBAC correcto | `401` sin token, `403` sin `permissions.read`/`permissions.create`, endpoints aislados por permiso. |
| Seed idempotente | Re-ejecutar `migrate-up` no duplica filas (verificado contra la base real de desarrollo). |

---

## Pruebas

- Unitarias HTTP: `internal/modules/user/interfaces/http/handler/permission_handler_test.go` (`go test ./...`). Cubren 201+listado, 409 (incluidas variantes mayúsculas/espacios), 400 (operación inválida, módulo inexistente, campos faltantes, `code` desalineado), catálogo de módulos, filtro, paginación, 401 y 403.
- E2E contra PostgreSQL (contenedor de desarrollo `sigif-go-dev-db`, puerto 5437): `SIGIF_DATABASE_PORT=5437 make migrate-up` + servidor con el mismo override, validando el índice creado, 21 permisos, `superadmin` = 21, y toda la matriz HTTP anterior vía `curl`.
- Bruno:

```bash
cd bruno
bru run Permission --env development \
  --env-var baseUrl=http://localhost:4600 \
  --env-var companyAdminPassword=<contraseña del seed>
```

El caso de creación usa el par `products.export` (no sembrado): la primera ejecución responde `201`; las siguientes `409`. Por eso el `tests` acepta ambos. El `409` determinista usa `users.create` (sembrado). El caso `403` se cubre en las pruebas de Go (el seed solo tiene usuarios superadmin, no un usuario `soporte`).