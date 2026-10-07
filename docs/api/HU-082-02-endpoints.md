# Reporte de endpoints — [HU-082-02] Listado de roles

## Información general

- HU: HU-082-02 — Listado de roles (RF, P1).
- Rama: `feat/HU-082-02-roles-list`.
- Prefijo base: `/api/v1`
- Requiere autenticación: token JWT válido en el esquema de autorización de portador (`Authorization: Bearer <access_token>`). El middleware global `AuthRequired` valida el token y que la sesión siga activa.
- Formato de respuesta exitosa: `{ "success": true, "data": ... }`. El listado incluye `meta`: `{ "page", "limit", "total", "total_pages" }`.
- Formato de respuesta de error: `{ "success": false, "error": { "code": "CODIGO", "message": "descripción", "details": { "campo": "motivo" } } }`.
- Colección Bruno: `bruno/Role/` (001–016).

---

## Resumen rápido

| Método | Ruta | Permiso | Descripción |
|---|---|---|---|
| `GET` | `/api/v1/roles/` | `roles.list` | Lista paginada de los roles visibles, con búsqueda, filtros y conteo de permisos. |

La HU es **solo de lectura**: no existe `POST`/`PATCH`/`DELETE` de roles; crear roles y asignarles permisos son HU posteriores. El endpoint vive en el módulo `user`.

---

## Autenticación / permisos

| Header | Obligatorio | Descripción |
|---|---|---|
| `Authorization` | Sí | `Bearer <access_token>` de una sesión activa. |
| `Content-Type` | No | No se envía body (GET). |

| Permiso | Endpoint | Roles sembrados que lo tienen |
|---|---|---|
| `roles.list` | `GET /api/v1/roles/` | `superadmin` |

`roles.list` (módulo `roles`, operación `list`) se siembra con esta HU. El rol `superadmin` recibe todos los permisos del catálogo (ahora 22); el rol `soporte` conserva solo `users.list` y `users.read`, por lo que un usuario con `soporte` responde `403`.

---

## Params de query

Todos son opcionales y combinables.

| Parámetro | Tipo | Reglas |
|---|---|---|
| `q` | string | Máximo 100 caracteres (se recorta antes de validar). Búsqueda parcial e insensible a mayúsculas sobre `name`; `%`, `_` y `\` del usuario se tratan como texto literal (no como comodines). |
| `type` | string | `system` (roles sembrados, disponibles para toda empresa) o `custom` (propios de una empresa). Sin valor o vacío: sin filtro. |
| `status` | string | `active`, `inactive` o `all`. Por defecto `all`: sin filtro. |
| `page` | int | Mínimo 1. Por defecto `1`. |
| `limit` | int | Mínimo 1, máximo `max_limit` (100). Por defecto `default_limit` (20). Un `limit` mayor a 100 responde `400`. |

---

## Respuesta exitosa — `200 OK`

```json
{
  "success": true,
  "data": [
    {
      "id": "3f1d3b8e-6d2e-4a7e-9d0a-1b2c3d4e5f60",
      "company_id": null,
      "name": "soporte",
      "type": "system",
      "status": "active",
      "permissions_count": 2,
      "description": "Soporte con acceso de solo lectura a usuarios",
      "created_at": "2026-10-05T12:00:00Z"
    },
    {
      "id": "9a5c7d12-4b3f-4e8a-8c1d-2e3f4a5b6c7d",
      "company_id": "2f4e6a8c-0b1d-4f3a-9c5e-7a9b1d3f5e70",
      "name": "Vendedor",
      "type": "custom",
      "status": "active",
      "permissions_count": 3,
      "created_at": "2026-10-06T09:30:00Z"
    }
  ],
  "meta": { "page": 1, "limit": 20, "total": 4, "total_pages": 1 }
}
```

- El listado se ordena por `lower(name) ASC, id ASC` (insensible a mayúsculas y determinista en empates).
- `permissions_count` es el número de permisos asignados al rol. Se calcula con una subconsulta sobre `role_permission` en la misma consulta: sin N+1 y sin coste extra en el `COUNT` de `meta.total`.
- `company_id` es `null` en los roles del sistema (y en cualquier otro rol global); en los roles de una empresa lleva su `id`.
- `type` se deriva de `is_system` (`system` / `custom`).
- `description` se omite cuando está vacía.
- `total` cuenta todos los roles que cumplen el filtro (la paginación no lo limita) y `total_pages` se deriva de `total` y `limit`.
- **Sin resultados responde `200` con `data: []`, `total: 0` y `total_pages: 0`**; nunca `404`.

---

## Visibilidad (aislamiento por empresa)

La empresa **siempre** sale del token (`company_id` del JWT), nunca del query string.

| Token | Roles que ve |
|---|---|
| Con `company_id` | `company_id IS NULL` (sistemas/globales) **o** `company_id = <empresa del token>`. |
| Sin `company_id` (superadmin global) | Solo `company_id IS NULL`: los roles del sistema. |

Un rol con el mismo nombre en otra empresa no aparece ni se cruza en la búsqueda: el filtro de empresa se aplica antes de `q`.

---

## Códigos de error

| Code | HTTP | Causa |
|---|---|---|
| `VALIDATION_ERROR` | 400 | `q` con más de 100 caracteres (`details.q`), `type` fuera de `system/custom` (`details.type`), `status` fuera de `active/inactive/all` (`details.status`), `page` menor que 1 (`details.page`), `limit` menor que 1 (`details.limit`) o `limit` mayor que 100 (`details.limit`, mensaje `limit must be 100 or less`). |
| `BAD_REQUEST` | 400 | Query string ilegible: `page`/`limit` no numéricos (`page=abc`, `limit=ten`). |
| `UNAUTHORIZED` | 401 | Falta el token, es inválido/expiró o la sesión fue cerrada. |
| `FORBIDDEN` | 403 | El usuario no tiene el permiso `roles.list` (p. ej. un usuario `soporte`). |
| `INTERNAL_ERROR` | 500 | Error inesperado de persistencia. |

No hay `404`: una lista vacía es un `200` legítimo.

---

## Reglas de negocio

- **Solo lectura.** `RoleHTTPHandler` únicamente expone `List`; no toca la base salvo con `SELECT`. El seed se mantiene idempotente (el `status` de los roles existentes se fija a `active` la próxima vez que corra `migrate-up`).
- **Filtros combinables y opcionales.** `q`, `type` y `status` se aplican juntos (p. ej. `?q=vendedor&type=custom&status=active`). `status=all` y `status` vacío significan lo mismo: sin filtro.
- **`type` mapea a `is_system`.** No hay columna nueva para el tipo: `system` ⇔ `is_system = true`.
- **Búsqueda segura.** El patrón se arma como `%<texto>%` sobre `lower(name)`, escapando los comodines de `LIKE`; una búsqueda por `100%` busca literalmente `100%`.
- **Conteo exacto.** `permissions_count` cuenta las filas de `role_permission` de ese rol, incluidas las que sembró `migrate-up`.
- **Sin permiso, sin acceso.** El permiso se valida por endpoint con `RequirePermission(checker, "roles", "list")`, antes de leer la base.

---

## Esquema relacionado

```sql
ALTER TABLE role ADD COLUMN status varchar(20) NOT NULL DEFAULT 'active';
CREATE INDEX idx_roles_company_id_name ON role (company_id, name);
```

- La columna `status` se agrega con `AutoMigrate` (`make migrate-up`); las filas existentes reciben `active` por el `DEFAULT`. Valores válidos: `active`, `inactive`.
- El índice compuesto `(company_id, name)` es **no único** y acelera el filtro de empresa (los `NULL` de `company_id` son distintos entre sí, por eso no reemplaza la unicidad).
- `role.name` **mantiene su índice único global** (`idx_roles_company_name`, equivalente a `unique(name)`): no se tocó, es decir, dos empresas no pueden usar el mismo nombre de rol. Tampoco se modificaron `updated_at` ni `deleted_at`.
- `RoleListModel` (la fila proyectada con `permissions_count`) es solo de lectura y no está en el `AutoMigrate`.

---

## Cobertura de criterios de aceptación

| Criterio | Cómo se cubre |
|---|---|
| Listar los roles con paginación y conteo | `GET /api/v1/roles/` devuelve `data` + `meta` con `permissions_count` por rol. |
| Búsqueda por nombre | `?q=` parcial e insensible a mayúsculas, con comodines escapados; `q` recortado. |
| Filtros `type` y `status` | `?type=system/custom`, `?status=active/inactive/all`; combinables entre sí y con `q`. |
| Validaciones de paginación | `page >= 1`, `1 <= limit <= 100` → `400 VALIDATION_ERROR` con `details.<campo>`; valores no numéricos → `400 BAD_REQUEST`. |
| Aislamiento por empresa | Filtro de visibilidad basado en el token; roles de otra empresa no aparecen ni se encuentran por `q`. |
| Rol global / sin empresa | Un token sin empresa solo recibe `company_id IS NULL`. |
| Lista vacía | `200` con `data: []` y `total: 0` (nunca `404`). |
| Solo lectura | No hay endpoints de escritura de roles y el listado no modifica filas (verificado en las pruebas de Go). |
| RBAC correcto | `401` sin token, `403` sin `roles.list`. |
| Seed idempotente | Re-ejecutar `migrate-up` no duplica permisos ni roles y fija `status = active`. |

---

## Pruebas

- Unitarias HTTP: `internal/modules/user/interfaces/http/handler/role_handler_test.go` (`go test ./...`). Cubren campos y `permissions_count`, búsqueda (mayúsculas, espacios, combinada con `status`), lista vacía, filtros `type`/`status`, paginación (páginas 1–4 sobre 25 roles), los `400` de validación, aislamiento entre empresas, token sin empresa, `401`/`403` y que el listado no escribe.
- Complemento: `internal/modules/user/domain/service/list_roles.go` se ejercita desde esas mismas pruebas (normalización de filtros inválidos a sin filtro).
- Bruno:

```bash
cd bruno
bru run Role --env development \
  --env-var baseUrl=http://localhost:4600 \
  --env-var companyAdminPassword=<contraseña del seed>
```

El caso `403` crea primero (002) y luego ingresa (003) un usuario con rol `soporte`, que no tiene `roles.list`. El aislamiento por empresa se verifica estructuralmente: todo elemento devuelto tiene `company_id` nulo o igual a la empresa del token.
