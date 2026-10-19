# Reporte de endpoints — [HU-082-02] Listado de roles

## Información general

- HU: HU-082-02 — Listado de roles (RF, P1).
- Rama: `feat/HU-082-02-roles-list`.
- Prefijo base: `/api/v1`
- Requiere autenticación: token JWT válido en el esquema de autorización de portador (`Authorization: Bearer <access_token>`). El middleware global `AuthRequired` valida el token y que la sesión siga activa.
- Formato de respuesta exitosa: `{ "success": true, "data": ... }`. El listado incluye `meta`: `{ "page", "limit", "total", "total_pages" }`.
- Formato de respuesta de error: `{ "success": false, "error": { "code": "CODIGO", "message": "descripción", "details": { "campo": "motivo" } } }`.
- Colección Bruno: `bruno/Role/` (001–020).

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
| `roles.list` | `GET /api/v1/roles/` | `superadmin`, `business_admin` |

`roles.list` (módulo `roles`, operación `list`) se siembra con esta HU. El rol `superadmin` recibe todos los permisos del catálogo; el rol `business_admin` recibe `roles.list` junto con `users.list`, `users.read` y `permissions.read` (solo lectura). El rol `employee` no recibe este permiso, por lo que un usuario con rol `employee` responde `403`.

---

## Correcciones de la revisión del frontend

1. **`business_admin` recibía `403`: resuelto.** El seed ahora asigna `roles.list` al rol `business_admin` (además de `superadmin`). Se conservan las restricciones para los demás roles y no se otorgan permisos de escritura de roles.
2. **Acciones disponibles según autorización:** HU-082-02 sigue siendo solo de lectura. El backend ya expone los permisos efectivos en `GET /api/v1/auth/me` y `POST /api/v1/auth/login` (`data.permissions`), que el frontend puede usar para la visibilidad de acciones; no se agregó contrato nuevo. Ver «Acciones del frontend y autorización».
3. **Nombres visibles en español:** el listado agrega el campo aditivo `display_name`; el buscador también admite las etiquetas visibles. El identificador técnico `name` no cambia.
- El endpoint y su contrato no cambian salvo el campo aditivo `display_name`.

---

## Params de query

Todos son opcionales y combinables.

| Parámetro | Tipo | Reglas |
|---|---|---|
| `q` | string | Máximo 100 caracteres (se recorta antes de validar). Búsqueda parcial e insensible a mayúsculas sobre `name` y, para los roles del sistema, también sobre su `display_name` en español (p. ej. `empleado`, `administrador de empresa`). Los comodines `%`, `_` y `\` del usuario se tratan como texto literal (no como comodines). |
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
      "name": "superadmin",
      "display_name": "Superadministrador",
      "type": "system",
      "status": "active",
      "permissions_count": 3,
      "description": "Acceso total al sistema",
      "created_at": "2026-10-05T12:00:00Z"
    },
    {
      "id": "9a5c7d12-4b3f-4e8a-8c1d-2e3f4a5b6c7d",
      "company_id": "2f4e6a8c-0b1d-4f3a-9c5e-7a9b1d3f5e70",
      "name": "Vendedor",
      "display_name": "Vendedor",
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
- `display_name` es el nombre amigable en español para mostrar en el frontend. Para los roles del sistema (`is_system = true`) se deriva del catálogo oficial; para los roles personalizados conserva el nombre ingresado por la empresa. El campo `name` siempre contiene el identificador técnico.
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
| `FORBIDDEN` | 403 | El usuario no tiene el permiso `roles.list` (p. ej. un usuario `employee`). |
| `INTERNAL_ERROR` | 500 | Error inesperado de persistencia. |

No hay `404`: una lista vacía es un `200` legítimo.

---

## Reglas de negocio

- **Solo lectura.** `RoleHTTPHandler` únicamente expone `List`; no toca la base salvo con `SELECT`. El seed se mantiene idempotente (el `status` de los roles existentes se fija a `active` la próxima vez que corra `migrate-up`).
- **Sin acciones por fila.** No existen endpoints de creación, edición, eliminación, activación o desactivación de roles: son alcance de otras HU. Esta HU no incorpora ninguna de esas operaciones.
- **Filtros combinables y opcionales.** `q`, `type` y `status` se aplican juntos (p. ej. `?q=vendedor&type=custom&status=active`). `status=all` y `status` vacío significan lo mismo: sin filtro.
- **`type` mapea a `is_system`.** No hay columna nueva para el tipo: `system` ⇔ `is_system = true`.
- **Nombres visibles.** `display_name` se resuelve en la capa de presentación desde el catálogo `systemRoleDisplayLabels` (dominio `user`); **no se almacena en la base de datos**. Los roles personalizados devuelven su propio `name` como `display_name`. Renombrar un rol del sistema o traducirlo no cambia el RBAC, que depende de `name`.
- **Búsqueda segura.** El patrón se arma como `%<texto>%` sobre `lower(name)`, escapando los comodines de `LIKE`; una búsqueda por `100%` busca literalmente `100%`. Para los roles del sistema se amplía a los identificadores técnicos cuyo `display_name` coincide, usando una lista cerrada del dominio (nunca texto del usuario concatenado en SQL).
- **Conteo exacto.** `permissions_count` cuenta las filas de `role_permission` de ese rol, incluidas las que sembró `migrate-up`.
- **Sin permiso, sin acceso.** El permiso se valida por endpoint con `RequirePermission(checker, "roles", "list")`, antes de leer la base. `business_admin` recibe `roles.list` solo por el seed; no se relaja el middleware.

---

## Acciones del frontend y autorización (Corrección 2)

HU-082-02 es de solo lectura, por lo que no hay acciones por fila que definir. El backend ya expone los permisos efectivos del usuario autenticado, que el frontend puede usar como única fuente para decidir qué acciones mostrar cuando existan endpoints de escritura:

| Método | Ruta | Descripción |
|---|---|---|
| `GET` | `/api/v1/auth/me` | Devuelve el usuario autenticado y `data.permissions`: arreglo de códigos `module.operation` de su rol. |
| `POST` | `/api/v1/auth/login` | En `data.permissions` devuelve el mismo arreglo al iniciar sesión. |

- Para HU-082-02, un cliente puede comprobar `roles.list` en `data.permissions` antes de mostrar la pantalla de listado, pero **la visibilidad de botones nunca sustituye la autorización**: el backend sigue validando `RequirePermission` en el endpoint.
- Cuando otras HU agreguen creación/edición/eliminación de roles, deberán exponer sus permisos (`roles.create`, `roles.update`, `roles.delete`, etc.) en el catálogo y validarlos en el backend. Mientras no existan esos endpoints y permisos, la colección de acciones disponibles es una **dependencia de otras HU**.
- No se agregan campos nuevos al listado para señalizar acciones: el contrato de `roles.list` se mantiene y el frontend deduce las acciones de `data.permissions`.

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
- `display_name` **no agrega columnas ni migraciones**: se resuelve en tiempo de presentación desde el catálogo de dominio. No se modificó `role.name` ni ninguna tabla.
- El seed (`internal/modules/user/infrastructure/seed/seed.go`) asigna `roles.list` a `business_admin` mediante `assignPermissions`, que verifica la existencia de cada fila de `role_permission` antes de insertarla. En bases existentes, volver a ejecutar `make migrate-up` agrega únicamente la asignación faltante; no duplica filas ni elimina asignaciones previas.

---

## Compatibilidad del contrato

- Se conservan `GET /api/v1/roles/` y sus parámetros `q`, `type`, `status`, `page`, `limit`.
- Se conservan campos, ordenamiento determinista (`lower(name) ASC, id ASC`), paginación, `meta`, `permissions_count`, visibilidad por empresa y respuestas de error.
- `display_name` es un **campo nuevo y aditivo**: no elimina ni renombra campos existentes, por lo que los clientes que consumían el contrato anterior siguen funcionando.
- El nombre técnico `role.name` **no se renombra en la base de datos** y sigue siendo el valor usado por RBAC, seeds y búsqueda.

---

## Cobertura de criterios de aceptación

| Criterio | Cómo se cubre |
|---|---|
| Listar los roles con paginación y conteo | `GET /api/v1/roles/` devuelve `data` + `meta` con `permissions_count` por rol. |
| Búsqueda por nombre | `?q=` parcial e insensible a mayúsculas, con comodines escapados; `q` recortado. |
| Búsqueda por nombre visible | Los roles del sistema también coinciden por su `display_name` en español (`empleado`, `administrador de empresa`), sin alterar la búsqueda por nombre técnico. |
| Nombres amigables sin perder identificadores | `display_name` en español para roles del sistema; `name` conserva el identificador técnico y sigue siendo la clave de RBAC y de búsqueda. |
| Filtros `type` y `status` | `?type=system/custom`, `?status=active/inactive/all`; combinables entre sí y con `q`. |
| Validaciones de paginación | `page >= 1`, `1 <= limit <= 100` → `400 VALIDATION_ERROR` con `details.<campo>`; valores no numéricos → `400 BAD_REQUEST`. |
| Aislamiento por empresa | Filtro de visibilidad basado en el token; roles de otra empresa no aparecen ni se encuentran por `q`. |
| Rol global / sin empresa | Un token sin empresa solo recibe `company_id IS NULL`. |
| Lista vacía | `200` con `data: []` y `total: 0` (nunca `404`). |
| Solo lectura | No hay endpoints de escritura de roles y el listado no modifica filas (verificado en las pruebas de Go). |
| RBAC correcto | `401` sin token, `200` para `superadmin` y `business_admin`, `403` sin `roles.list` (`employee`). |
| Seed idempotente | Re-ejecutar `migrate-up` no duplica permisos, roles ni asignaciones, y fija `status = active`. |

---

## Pruebas

- Unitarias HTTP: `internal/modules/user/interfaces/http/handler/role_handler_test.go` (`go test ./...`). Cubren campos y `permissions_count`, búsqueda (mayúsculas, espacios, combinada con `status`), `display_name` de roles del sistema y personalizados, búsqueda por nombre técnico y por etiqueta visible en español, lista vacía, filtros `type`/`status`, paginación (páginas 1–4 sobre 25 roles), los `400` de validación, aislamiento entre empresas, token sin empresa, `401`/`403` y que el listado no escribe.
- Unitarias de dominio: `internal/modules/user/domain/entity/role_test.go` cubre el catálogo de etiquetas (`DisplayName`), la expansión de búsqueda por etiqueta (`SystemRoleNamesByDisplayPrefix`) y la conservación de nombres técnicos.
- Complemento: `internal/modules/user/domain/service/list_roles.go` se ejercita desde esas mismas pruebas (normalización de filtros inválidos a sin filtro).
- Integración del seed: `internal/modules/user/infrastructure/seed/seed_test.go` (build tag `integration`) verifica que `superadmin` y `business_admin` tienen `roles.list`, que `business_admin` no tiene `roles.create/update/delete`, que `employee` no lo tiene, que `business_admin` conserva `users.list`/`users.read`/`permissions.read` y que ejecutar el seed dos veces no duplica `role_permission` ni roles/permisos.

```bash
SIGIF_TEST_DATABASE_DSN="postgres://<user>:<pass>@<host>:<port>/<db>?sslmode=disable" \
  go test -tags integration -run TestSeedBusinessAdminCanListRoles \
  ./internal/modules/user/infrastructure/seed/
```

- Bruno (`bruno/Role/`, 001–020):

```bash
cd bruno
bru run Role --env development \
  --env-var baseUrl=http://localhost:4600 \
  --env-var companyAdminPassword=<contraseña del seed> \
  --env-var businessAdminPassword=<contraseña del business_admin del seed>
```

El caso `403` crea primero (002) un usuario con rol `employee` (que no tiene `roles.list`) y luego ingresa (003) con la contraseña generada; el rol `employee` responde `403`. El caso `017`/`018` inicia sesión como `business_admin` y comprueba que el listado responde `200`. El aislamiento por empresa se verifica estructuralmente: todo elemento devuelto tiene `company_id` nulo o igual a la empresa del token. Los casos `019`/`020` buscan por etiqueta visible en español.
