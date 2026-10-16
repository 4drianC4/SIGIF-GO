# Reporte de endpoints — [HU-082-04] Activar / desactivar roles

## Información general

- HU: HU-082-04 — Gestión del estado de roles (RF-82, Módulo Seguridad/Roles). Se apoya en el listado de HU-082-02 (`GET /api/v1/roles/`).
- Rama: `feat/HU-082-04-role-status-management`.
- Prefijo base: `/api/v1`
- Requiere autenticación: token JWT válido en el esquema de autorización de portador (`Authorization: Bearer <access_token>`). El middleware global `AuthRequired` valida el token y que la sesión siga activa.
- Empresa: el rol se busca por `id` **y** por la visibilidad de la empresa del token. Un rol global (sistema) es visible para todos; un rol propio de otra empresa responde `404`, igual que uno inexistente.
- Formato de respuesta exitosa: `{ "success": true, "data": { ... } }`.
- Formato de respuesta de error: `{ "success": false, "error": { "code": "CODIGO", "message": "descripción", "details": { "campo": "motivo" } } }`.
- Colección Bruno: `bruno/RoleStatus/` (001–016).

---

## Resumen rápido

| Método | Ruta | Permiso | Descripción |
|---|---|---|---|
| `PATCH` | `/api/v1/roles/:id/status` | `roles.status` | Activa o desactiva un rol sin borrarlo ni tocar sus permisos. **Nuevo en HU-082-04.** |

El endpoint vive en el módulo `user`. **No existe** `POST`/`DELETE` de roles: crear y eliminar roles son otras HU. El cambio de estado es la única escritura de roles de esta HU.

---

## Autenticación / permisos

| Header | Obligatorio | Descripción |
|---|---|---|
| `Authorization` | Sí | `Bearer <access_token>` de una sesión activa. |
| `Content-Type` | Sí | `application/json`. |

| Permiso | Endpoint | Roles sembrados que lo tienen |
|---|---|---|
| `roles.status` | `PATCH /api/v1/roles/:id/status` | `superadmin` |

`roles.status` (módulo `roles`, operación `status`) se siembra con esta HU y se suma al catálogo (ahora 26 permisos). El rol `superadmin` recibe todos los permisos del catálogo; el rol `business_admin` solo tiene `users.list`, `users.read` y `permissions.read`, por lo que responde `403`.

---

## 1) Cambiar el estado de un rol — `PATCH /api/v1/roles/:id/status`

Activa o desactiva un rol existente y devuelve el rol actualizado con el mismo contrato que el listado de HU-082-02. Nunca crea ni elimina un rol, ni modifica sus permisos.

### Permiso
`roles.status`

### Params de ruta
- `id`: UUID — identificador del rol (el que devuelve el listado). Un valor que no es UUID responde `400 invalid role id`.

### Body

```json
{
  "status": "inactive"
}
```

| Campo | Tipo | Reglas |
|---|---|---|
| `status` | enum | Obligatorio. Solo `active` o `inactive`. |

### Ejemplo de request

```bash
curl -s -X PATCH "http://localhost:4600/api/v1/roles/9a5c7d12-4b3f-4e8a-8c1d-2e3f4a5b6c7d/status" \
  -H "Authorization: Bearer $ACCESS_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"status": "inactive"}'
```

### Respuesta exitosa

**`200 OK`** — el rol con su nuevo estado, con la misma estructura del listado:

```json
{
  "success": true,
  "data": {
    "id": "9a5c7d12-4b3f-4e8a-8c1d-2e3f4a5b6c7d",
    "company_id": "2f4e6a8c-0b1d-4f3a-9c5e-7a9b1d3f5e70",
    "name": "Vendedor",
    "type": "custom",
    "status": "inactive",
    "permissions_count": 3,
    "description": "Vendedor de mostrador",
    "created_at": "2026-10-06T09:30:00Z"
  }
}
```

- `status` refleja el valor solicitado (`active` / `inactive`).
- `name`, `description`, `type`, `company_id` y `permissions_count` **no cambian**: el estado no es una edición de la configuración del rol.
- `description` se omite cuando está vacía.

### Códigos de error

| Code | HTTP | Causa |
|---|---|---|
| `VALIDATION_ERROR` | 400 | Falta `status`, es `null` o vacío, o no es `active`/`inactive` (`details.status`). |
| `BAD_REQUEST` | 400 | `id` no es UUID (`invalid role id`) o el body no es un objeto JSON legible. |
| `UNAUTHORIZED` | 401 | Falta el token, es inválido/expiró, o la sesión fue cerrada. |
| `FORBIDDEN` | 403 | El usuario no tiene el permiso `roles.status`. |
| `NOT_FOUND` | 404 | El rol no existe o pertenece a otra empresa (`role not found`). |
| `CONFLICT` | 409 | Se intenta desactivar un rol protegido (superadmin): `details.system`. |
| `INTERNAL_ERROR` | 500 | Error inesperado de persistencia. |

**`400`** — estado inválido o ausente

```json
{
  "success": false,
  "error": {
    "code": "VALIDATION_ERROR",
    "message": "validation failed",
    "details": { "status": "status must be one of [active inactive]" }
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

**`404`**

```json
{ "success": false, "error": { "code": "NOT_FOUND", "message": "role not found" } }
```

**`409`** — rol protegido

```json
{
  "success": false,
  "error": {
    "code": "CONFLICT",
    "message": "system role cannot be deactivated",
    "details": { "system": "system role cannot be deactivated" }
  }
}
```

---

## Reglas de negocio

- **No es un borrado.** Desactivar un rol no lo elimina ni lo oculta: el listado (`GET /api/v1/roles/?status=inactive`) lo sigue mostrando con `status: inactive` para poder reactivarlo. Se conservan `name`, `description`, `is_system`/`type`, `company_id` y `created_at`.
- **Los permisos se conservan.** La tabla `role_permission` no se toca: `permissions_count` es el mismo antes y después, y al reactivar el rol recupera exactamente los mismos permisos. El endpoint solo actualiza la columna `status`.
- **Protegido: superadmin.** El rol `superadmin` (`is_system = true` y `name = superadmin`) no se puede desactivar: responde `409 CONFLICT`. Es el rol reservado para el arranque del sistema (el mismo criterio de `IsUserAssignable`, que ya lo excluye de la asignación de usuarios). El resto de roles del sistema (`business_admin`, `employee`) sí se pueden desactivar.
- **Idempotente.** Cambiar un rol al estado que ya tiene responde `200` con el rol y **no** escribe en la base (no hay `UPDATE`). Repetir la operación es seguro.
- **El rol inactivo deja de autorizar de inmediato.** El estado del rol se lee en cada control de acceso:
  - `HasPermission` (middleware RBAC) devuelve `false` para un rol inactivo, **sin cerrar la sesión**: un usuario con sesión abierta deja de poder operar en cuanto se desactiva su rol (y vuelve a poder cuando se reactiva).
  - `PermissionCodesByRole` devuelve `[]` para un rol inactivo, por lo que `login` y `/auth/me` (`user.permissions`) no listan permisos efectivos de un rol desactivado. Los permisos siguen almacenados y reaparecen al reactivar.
- **No se asignan roles inactivos.** `POST /api/v1/users` (registro) y `PATCH /api/v1/users/:id` (edición) rechazan asignar un rol inactivo con `400 VALIDATION_ERROR` (`details.role`). Crear o editar un usuario con un rol inactivo no cambia su rol.
- **Aislamiento por empresa.** La empresa sale del token. Un rol global (`company_id IS NULL`) es visible para cualquier llamador autorizado; un rol propio de otra empresa responde `404`. Un token sin empresa solo puede ver y cambiar roles globales.
- **Sin permiso, sin acceso.** El permiso se valida por endpoint con `RequirePermission(checker, "roles", "status")`, antes de leer o escribir la base.
- **Auditoría implícita.** El `UPDATE` roza `updated_at` del rol; no se añaden columnas de auditoría en esta HU.

---

## Esquema relacionado

```sql
-- HU-082-02
ALTER TABLE role ADD COLUMN status varchar(20) NOT NULL DEFAULT 'active';
-- HU-082-04: no agrega columnas ni tablas; solo escribe role.status
UPDATE role SET status = 'inactive' WHERE id = :id;
```

- Valores válidos de `role.status`: `active`, `inactive`.
- `role_permission` (id, role_id, permission_id) **no** se modifica.
- El permiso se siembra en `permission` (`module = 'roles'`, `operation = 'status'`) y se asigna a `superadmin` con el seed idempotente (`make migrate-up`).

---

## Cobertura de criterios de aceptación

| Criterio | Cómo se cubre |
|---|---|
| Activar un rol | `PATCH /api/v1/roles/:id/status` con `{"status":"active"}` → `200` y `status: active`. |
| Desactivar un rol | Mismo endpoint con `{"status":"inactive"}` → `200` y `status: inactive`. |
| Conservar configuración y permisos | El servicio solo cambia `status`; `Update` escribe la columna `status`; `permissions_count` y los campos del rol no varían. |
| El rol inactivo no puede usarse | `HasPermission` y `PermissionCodesByRole` devuelven vacío/false para roles inactivos; registrar/editar usuario con rol inactivo → `400`. |
| Impacto inmediato en sesiones abiertas | El estado se lee en cada chequeo RBAC: la misma sesión pasa de `200` a `403` al desactivar el rol, sin invalidarla. |
| Reactivar restaura los permisos | Al volver a `active`, el rol autoriza de nuevo con sus mismos permisos. |
| Idempotencia | Repetir el mismo estado responde `200` sin escribir (verificado con el contador de escrituras). |
| Rol protegido | Desactivar `superadmin` → `409 CONFLICT`. |
| Aislamiento por empresa | Rol de otra empresa → `404`; rol global visible para cualquier empresa; token sin empresa solo ve globales. |
| Validaciones | `id` no UUID → `400 BAD_REQUEST`; `status` inválido/ausente → `400 VALIDATION_ERROR`; inexistente → `404`. |
| RBAC correcto | `401` sin token, `403` sin `roles.status`. |
| El listado sigue siendo consistente | Los roles inactivos aparecen en `GET /api/v1/roles/?status=inactive` (HU-082-02). |

---

## Pruebas

- Unitarias de servicio: `internal/modules/user/domain/service/change_role_status_test.go` (activar/desactivar, idempotencia sin escritura, superadmin protegido, visibilidad por empresa, `HasPermission` y `PermissionCodesByRole` con rol inactivo).
- Unitarias HTTP: `internal/modules/user/interfaces/http/handler/role_status_handler_test.go` (contrato de respuesta y `permissions_count`, idempotencia, `409` de superadmin, `404`, `400`, `401`/`403`, aislamiento entre empresas y revocación inmediata de permisos con el servicio real como `PermissionChecker`).
- Ajuste de pruebas existentes: `authorization`/`permission_codes`/`edit` (`internal/modules/user/domain/service/*_test.go`) y los stubs HTTP (`user_form_test.go`, `permission_handler_test.go`).
- Bruno:

```bash
cd bruno
bru run RoleStatus --env development \
  --env-var baseUrl=http://localhost:4600 \
  --env-var companyAdminPassword=<contraseña del seed> \
  --env-var businessAdminPassword=<contraseña del seed>
```

La carpeta reactiva `employee` al final (016), por lo que puede ejecutarse varias veces. Los casos "rol de otra empresa" y "sesión abierta que pierde permisos" no están en Bruno (el seed tiene una sola empresa): están cubiertos en las pruebas de Go.

---

## Notas / decisiones de ambigüedad

- **Rol protegido = superadmin.** El requerimiento no enumeraba qué roles no pueden desactivarse. Se optó por proteger únicamente `superadmin`, alineado con `IsUserAssignable` (que ya lo reserva para el arranque). Si se quiere proteger también `business_admin`/`employee`, basta ampliar `Role.IsProtected`.
- **`409` vs `400` para rol protegido.** Se usa `409 CONFLICT` (el recurso existe pero la operación entra en conflicto con su naturaleza), consistente con el `409` de permisos del sistema/asignados de HU-082-01.
- **Sin borrado lógico.** `status = inactive` no es un `deleted_at`: el rol sigue existiendo y visible.
