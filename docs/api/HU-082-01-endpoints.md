# Reporte de endpoints — [HU-082-01] Catálogo de permisos

## Información general

- HU: HU-082-01 — Catálogo de permisos (RF, P1).
- Rama: `feat/HU-082-01-permissions-catalog`.
- Prefijo base: `/api/v1`
- Requiere autenticación: token JWT válido en el esquema de autorización de portador (`Authorization: Bearer <access_token>`). El middleware global `AuthRequired` valida el token y que la sesión siga activa.
- Formato de respuesta exitosa: `{ "success": true, "data": ... }`. El listado incluye `meta`: `{ "page", "limit", "total", "total_pages" }`.
- Formato de respuesta de error: `{ "success": false, "error": { "code": "CODIGO", "message": "descripción", "details": { "campo": "motivo" } } }`.
- Colección Bruno: `bruno/Permission/` (001–019). El caso de creación acepta `201` o `409` para poder ejecutarse varias veces (no existe un par fijo previo). Los casos de edición y eliminación resuelven primero el `id` con `012-setup-resolve-permission-ids`.

---

## Resumen rápido

| Método | Ruta | Permiso | Descripción |
|---|---|---|---|
| `GET` | `/api/v1/permissions/modules` | `permissions.read` | Catálogo de módulos con sus operaciones soportadas. |
| `GET` | `/api/v1/permissions` | `permissions.read` | Lista paginada de permisos, filtrable por `module`. |
| `GET` | `/api/v1/permissions/export` | `permissions.export` | Exporta el catálogo a CSV, filtrable por `module`. |
| `POST` | `/api/v1/permissions` | `permissions.create` | Registra un permiso nuevo para asignarlo después a roles. |
| `PATCH` | `/api/v1/permissions/:id` | `permissions.update` | Edita la descripción de un permiso (solo ese campo). |
| `DELETE` | `/api/v1/permissions/:id` | `permissions.delete` | Elimina un permiso no perteneciente al sistema ni asignado a un rol. |

Los seis endpoints viven en el módulo `user` (el permiso es la unidad más pequeña de autorización: un par `module.operation`). No dependen de la empresa del token.

---

## Autenticación / permisos

| Header | Obligatorio | Descripción |
|---|---|---|
| `Authorization` | Sí | `Bearer <access_token>` de una sesión activa. |
| `Content-Type` | Sí (con body) | `application/json`. |

| Permiso | Endpoint | Roles sembrados que lo tienen |
|---|---|---|
| `permissions.read` | `GET /api/v1/permissions/modules`, `GET /api/v1/permissions` | `superadmin`, `business_admin` |
| `permissions.export` | `GET /api/v1/permissions/export` | `superadmin` |
| `permissions.create` | `POST /api/v1/permissions` | `superadmin` |
| `permissions.update` | `PATCH /api/v1/permissions/:id` | `superadmin` |
| `permissions.delete` | `DELETE /api/v1/permissions/:id` | `superadmin` |

`permissions.read`, `permissions.create`, `permissions.update`, `permissions.delete` y `permissions.export` se agregan al seed junto con los permisos existentes (incluido `roles.list` de HU-082-02). El rol `superadmin` recibe automáticamente todos los permisos del catálogo (ahora 25); el rol `business_admin` recibe `users.list`, `users.read` y `permissions.read` (solo lectura de usuarios y del catálogo), por lo que responde `403` en create/update/delete/export. El rol `employee` no recibe permisos.

---

## Catálogo de módulos y operaciones

Módulos soportados (fuente de verdad: `entity.AllPermissions()` y `entity.SupportedModuleCatalog()`):

| Módulo | Operaciones del seed |
|---|---|
| `users` | `create, list, read, update, delete, activate, deactivate, change_password` |
| `categories` | `create` |
| `products` | `create, read, update, delete` |
| `customers` | `create, list, read, update, status, delete` |
| `permissions` | `read, create, update, delete, export` |
| `roles` | `list` (módulo agregado por HU-082-02) |

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
    { "module": "permissions", "operations": ["read", "list", "create", "update", "delete", "export", "activate", "deactivate", "change_password", "status"] },
    { "module": "roles", "operations": ["read", "list", "create", "update", "delete", "export", "activate", "deactivate", "change_password", "status"] }
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

## 4) Editar permiso — `PATCH /api/v1/permissions/:id`

### Permiso
`permissions.update`

### Path param

| Parámetro | Tipo | Reglas |
|---|---|---|
| `id` | uuid | **Obligatorio**. Si no es un uuid válido responde `400`. |

### Body

| Campo | Tipo | Reglas |
|---|---|---|
| `description` | string | **Obligatorio**, máximo 200. Se recorta. |

Solo se puede editar la **descripción**. `module` y `operation` son inmutables: definen el código `module.operation`, y cambiarlos alteraría silenciosamente los permisos efectivos de todos los roles que ya lo tienen asignado. Enviar otros campos no los modifica.

### Ejemplo de request

```bash
curl -s -X PATCH "http://localhost:4600/api/v1/permissions/$ID" \
  -H "Authorization: Bearer $ACCESS_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"description":"Exportar productos actualizado"}'
```

### Respuesta exitosa — `200 OK`

```json
{
  "success": true,
  "data": {
    "id": "e3f876fc-8444-4150-9ea4-ba5c797f2cdd",
    "module": "products",
    "operation": "export",
    "code": "products.export",
    "description": "Exportar productos actualizado",
    "is_system": false
  }
}
```

---

## 5) Eliminar permiso — `DELETE /api/v1/permissions/:id`

### Permiso
`permissions.delete`

### Path param

| Parámetro | Tipo | Reglas |
|---|---|---|
| `id` | uuid | **Obligatorio**. Si no es un uuid válido responde `400`. |

### Reglas

1. **Los permisos del sistema no se borran.** Un permiso es "del sistema" cuando pertenece al seed (`entity.AllPermissions()`), por ejemplo `permissions.read`. Responde `409 CONFLICT` con `details.system`. Para que el front oculte el botón, la lista (`GET /permissions`) y el export marcan estos permisos con `is_system: true`.
2. **Un permiso asignado a un rol no se borra.** Primero hay que quitarle la asignación (HU de roles). Responde `409 CONFLICT` con `details.roles`.
3. Si no es del sistema y no está asignado, se elimina y responde `204 No Content`.

### Respuesta exitosa — `204 No Content`

Sin cuerpo.

---

## 6) Exportar permisos — `GET /api/v1/permissions/export`

### Permiso
`permissions.export`

### Params de query

| Parámetro | Tipo | Reglas |
|---|---|---|
| `module` | string | Opcional. Máximo 60. Si se envía debe ser un módulo soportado (`400` si no lo es). |

### Respuesta exitosa — `200 OK`

Devuelve un archivo CSV (`Content-Type: text/csv; charset=utf-8`), con BOM UTF-8 para que las hojas de cálculo lean bien los acentos. `Content-Disposition: attachment; filename="permissions.csv"` (o `permissions_<module>.csv` al filtrar).

Columnas: `id`, `module`, `operation`, `code`, `description`, `is_system`.

```csv
﻿id,module,operation,code,description,is_system
164e83e5-e30a-488d-9204-469fd54ce525,users,create,users.create,Registrar usuarios,true
e3f876fc-8444-4150-9ea4-ba5c797f2cdd,products,export,products.export,Exportar productos,false
```

### Ejemplo de request

```bash
curl -s "http://localhost:4600/api/v1/permissions/export?module=permissions" \
  -H "Authorization: Bearer $ACCESS_TOKEN" -o permissions_permissions.csv
```

---

## Permisos del usuario (login y `/auth/me`)

`POST /api/v1/auth/login` y `GET /api/v1/auth/me` devuelven, además del nombre del rol, la lista de códigos de permiso del usuario (`user.permissions`), para que el front pueda habilitar/ocultar acciones sin adivinar por rol:

```json
{
  "success": true,
  "data": {
    "user": {
      "id": "9c9f...",
      "role": "business_admin",
      "permissions": ["users.list", "users.read", "permissions.read"],
      "first_name": "Usuario",
      "last_name": "Negocio"
    },
    "token": { "access_token": "...", "expires_in": 3600, "token_type": "Bearer" }
  }
}
```

`superadmin` recibe los 24 códigos del catálogo.

---

## Códigos de error

| Code | HTTP | Causa |
|---|---|---|
| `VALIDATION_ERROR` | 400 | Módulo u operación fuera del catálogo (`details.module` o `details.operation`), `code` distinto de `module.operation` (`details.code`), campos obligatorios faltantes, formato/longitud inválidos, `id` no uuid en PATCH/DELETE. |
| `BAD_REQUEST` | 400 | El body no es un objeto JSON (malformado, array, `null`). |
| `UNAUTHORIZED` | 401 | Falta el token o es inválido/expiró, o la sesión fue cerrada. |
| `FORBIDDEN` | 403 | El usuario no tiene el permiso requerido (`permissions.read`, `permissions.create`, `permissions.update`, `permissions.delete` o `permissions.export`). |
| `NOT_FOUND` | 404 | El `:id` de PATCH/DELETE no corresponde a un permiso registrado. |
| `CONFLICT` | 409 | El par `(module, operation)` ya está registrado (`details.code = "already registered"`); se intenta borrar un permiso del sistema (`details.system`) o uno asignado a un rol (`details.roles`). |
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

**`404`** — permiso inexistente (PATCH/DELETE)

```json
{
  "success": false,
  "error": { "code": "NOT_FOUND", "message": "permission not found" }
}
```

**`409`** — permiso del sistema protegido

```json
{
  "success": false,
  "error": {
    "code": "CONFLICT",
    "message": "system permission cannot be deleted",
    "details": { "system": "system permission cannot be deleted" }
  }
}
```

**`409`** — permiso asignado a un rol

```json
{
  "success": false,
  "error": {
    "code": "CONFLICT",
    "message": "permission is assigned to a role",
    "details": { "roles": "permission is assigned to one or more roles" }
  }
}
```

**`400`** — falta `description` en PATCH

```json
{
  "success": false,
  "error": {
    "code": "VALIDATION_ERROR",
    "message": "validation failed",
    "details": { "description": "description is a required field" }
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
- **Afecta solo a la administración del permiso.** No hay asignación de roles en estos endpoints (es de una HU posterior). Los permisos nuevos quedan disponibles en `GET /api/v1/permissions` para futura asignación.
- **Descripción editable, código inmutable.** `PATCH` solo cambia `description`; `module` y `operation` no se tocan porque definen el código y con él el permiso efectivo de los roles.
- **Permisos del sistema protegidos.** Los que provienen del seed (`entity.IsSystemPermission`) no se pueden borrar (`409`). La lista y el export los marcan con `is_system: true` para que el front oculte el botón.
- **Borrado seguro.** Un permiso asignado a un rol no se borra (`409`); primero se quita la asignación. Un permiso creado por API y no asignado sí se borra (`204`).
- **Sin permiso, sin acceso.** Cada endpoint exige su permiso propio: `permissions.read` (módulos/listado), `permissions.export` (export), `permissions.create` (POST), `permissions.update` (PATCH) y `permissions.delete` (DELETE) (RBAC del módulo `user`).

---

## Índice único y seed

```sql
CREATE UNIQUE INDEX idx_permissions_module_operation ON permission (module, operation);
```

Se declara en `PermissionModel` y lo crea `AutoMigrate` (`make migrate-up`), igual que el resto del esquema. El seed inserta los permisos del catálogo (incluidos `permissions.read/create/update/delete/export` de HU-082-01 y `roles.list` de HU-082-02) y le da a `superadmin` **todos** los permisos del catálogo dinámicamente (al sembrar un módulo nuevo, `superadmin` lo recibe). Adicionalmente:

- `business_admin` recibe `users.list`, `users.read` y `permissions.read` (solo lectura).
- El seed crea un usuario de prueba `business_admin` (`negocio@sigif.com` / `negocio123` por defecto, configurable con `SIGIF_SEED_BUSINESS_ADMIN_*`) ligado a la empresa demo, para poder probar el "solo ver".
- Al migrar una base existente limpia, `superadmin` queda con **25** permisos; `business_admin` con 3.

---

## Cobertura de criterios de aceptación

| Criterio | Cómo se cubre |
|---|---|
| Consultar los módulos y permisos | `GET /api/v1/permissions/modules` (catálogo) y `GET /api/v1/permissions` (registrados). |
| Registro con validaciones | `POST` valida módulo y operación contra el catálogo, normaliza y aplica reglas de longitud. |
| Código derivado `module.operation` | `code` se expone derivado; si el cliente envía uno distinto responde `400`. |
| Rechazar duplicados | Chequeo del servicio + índice único → `409 CONFLICT` (también con espacios/mayúsculas). |
| Editar un permiso | `PATCH` cambia la descripción (`200`) y responde `404`/`400` según el caso. |
| Eliminar un permiso | `DELETE` responde `204`; `409` si es del sistema o está asignado; `404` si no existe. |
| Exportar el catálogo | `GET /export` devuelve CSV con BOM, respetando el filtro `module`. |
| El permiso queda disponible para roles | La respuesta `201` trae el `id`; el permiso aparece en `GET /api/v1/permissions?module=...`. |
| Front conoce los permisos del usuario | `login` y `/auth/me` devuelven `user.permissions`. |
| RBAC correcto | `401` sin token, `403` sin el permiso del endpoint, endpoints aislados por permiso. |
| Seed idempotente | Re-ejecutar `migrate-up` no duplica filas (verificado contra la base real de desarrollo). |

---

## Pruebas

- Unitarias HTTP: `internal/modules/user/interfaces/http/handler/permission_handler_test.go` (`go test ./...`). Cubren 201+listado, 409 (incluidas variantes mayúsculas/espacios), 400 (operación inválida, módulo inexistente, campos faltantes, `code` desalineado), catálogo de módulos, filtro, paginación, **PATCH (200/404/400/403)**, **DELETE (204/404/409 sistema/409 asignado/403)**, **export CSV (200/filtro/400/403)**, 401 y 403.
- E2E contra PostgreSQL (contenedor de desarrollo `sigif-go-dev-db`, puerto 5437): `SIGIF_DATABASE_PORT=5437 make migrate-up` + servidor con el mismo override, validando el índice creado, 25 permisos, `superadmin` = 25, `business_admin` = 3 (y su usuario de prueba), y toda la matriz HTTP anterior vía `curl`.
- Bruno:

```bash
cd bruno
bru run Permission --env development \
  --env-var baseUrl=http://localhost:4600 \
  --env-var companyAdminPassword=<contraseña del seed>
```

El caso de creación usa el par `products.export` (no sembrado): la primera ejecución responde `201`; las siguientes `409`. Por eso el `tests` acepta ambos. El `409` determinista de duplicado usa `users.create` (sembrado). `012` resuelve los ids usados por `013`–`016`. El caso `403` se cubre en las pruebas de Go (Bruno autentica con superadmin); alternativamente puede ejecutarse con el usuario `business_admin` sembrado, que solo tiene permisos de lectura.