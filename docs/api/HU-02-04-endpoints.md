# Reporte de endpoints — HU-02-04 Consultar y desactivar usuario

## Información general

- HU: HU-02-04 — Consultar y desactivar usuario
- Rama: `feature/HU-02-04-consult-and-deactivate-user` (parte de `feature/HU-02-03-view-and-preserve-user-history`)
- Prefijo base: `/api/v1`
- Requiere contexto autenticado del backend: `userId`, tomado del access token (`Authorization: Bearer <token>`). No se envía por header ni por body.
- Formato de respuesta exitosa: `{ "success": true, "data": ... }` — en listados paginados se agrega `meta` con `page`, `limit`, `total`, `total_pages`
- Formato de respuesta de error: `{ "success": false, "error": { "code": "CODIGO", "message": "descripción" } }`

---

## Resumen rápido

| Método | Ruta | Permiso | Estado | Descripción |
|---|---|---|---|---|
| `GET` | `/api/v1/users` | `users.list` | Ya existía | Lista usuarios con su estado actual, incluidos los desactivados. |
| `GET` | `/api/v1/users/:id` | `users.read` | Ya existía | Devuelve un usuario, también si está desactivado. |
| `GET` | `/api/v1/users/:id/history` | `users.read` | HU-02-03 | Historial del usuario; se conserva tras desactivarlo. |
| `POST` | `/api/v1/users/:id/deactivate` | `users.deactivate` | **Modificado** | Desactivación lógica. Ahora responde 200 con el usuario. |

No se creó ningún endpoint. Fuera del alcance de esta HU: `POST /users/:id/activate` y `DELETE /users/:id` siguen existiendo y no son parte del flujo.

---

## Autenticación / permisos

| Header | Obligatorio | Descripción |
|---|---|---|
| `Authorization` | Sí | `Bearer <access_token>` obtenido con `POST /api/v1/auth/login`. La sesión debe seguir activa. |

Permisos usados: `users.list`, `users.read`, `users.deactivate`. `superadmin` tiene los tres; `business_admin` solo los de lectura; `employee` ninguno (403).

---

## 1) Desactivar usuario — `POST /api/v1/users/:id/deactivate`

Desactivación lógica: el usuario pasa a `inactive` y se marca con `deleted_at`. La fila no se borra.

### Permiso
`users.deactivate`

### Path params
- `id` (obligatorio): uuid — usuario a desactivar.

### Body
No lleva body.

### Respuesta exitosa

**`200`** — antes respondía `204` sin cuerpo.
```json
{
  "success": true,
  "data": {
    "id": "de8cb7e2-692c-424a-880f-dc80eb48191b",
    "company_id": "11111111-1111-4111-8111-111111111111",
    "role_id": "3f0c1a52-7d0e-4f39-9a5e-2b8f0f6f1c11",
    "role": "employee",
    "first_name": "Juan",
    "last_name": "Pérez",
    "username": "juan@sigif.com",
    "email": "juan@sigif.com",
    "status": "inactive",
    "last_access": "2026-10-08T17:00:00Z",
    "created_at": "2026-10-01T12:00:00Z",
    "updated_at": "2026-10-08T18:30:00Z"
  }
}
```

Es el mismo objeto `User` del resto del módulo. `company_id`, `phone` y `last_access` se omiten si están vacíos. Nunca incluye `password_hash`, contraseñas ni `deleted_at`.

### Qué hace

1. Cambia `status` a `inactive` y guarda la fecha en `deleted_at`.
2. Conserva la fila y todos sus datos (nombre, email, rol, compañía, contraseña).
3. Registra el evento `deactivated` en el historial, con quién lo hizo, en la misma transacción.
4. No borra ni modifica el historial anterior del usuario.

### Códigos de error

| Code | HTTP | Causa |
|---|---|---|
| `BAD_REQUEST` | 400 | `id` no es un UUID válido. |
| `UNAUTHORIZED` | 401 | Falta el token, es inválido, expiró o su sesión se cerró. |
| `FORBIDDEN` | 403 | Usuario sin `users.deactivate`. |
| `NOT_FOUND` | 404 | El usuario no existe. |
| `CONFLICT` | 409 | El usuario ya está inactivo (`user is already inactive`). |
| `CONFLICT` | 409 | Se intenta desactivar la propia cuenta (`you cannot deactivate your own account`). |
| `INTERNAL_ERROR` | 500 | Error de persistencia. |

Un 409 no cambia nada ni registra evento.

---

## 2) Listar usuarios — `GET /api/v1/users?page=1&limit=100`

Sin cambios de contrato. Permiso `users.list`.

Los usuarios desactivados **siguen apareciendo** en el listado, con `"status": "inactive"`. El frontend distingue activos de inactivos por ese campo.

---

## 3) Consultar usuario — `GET /api/v1/users/:id`

Sin cambios de contrato. Permiso `users.read`.

Un usuario desactivado se sigue devolviendo con `200` y `"status": "inactive"`. Solo responde `404` si el `id` no existe.

---

## 4) Consultar historial — `GET /api/v1/users/:id/history?page=1&limit=20`

Definido en [HU-02-03](HU-02-03-endpoints.md). Después de desactivar, el historial completo sigue disponible y su primer elemento es el evento `deactivated`:

```json
{
  "action": "deactivated",
  "description": "User deactivated",
  "changes": {},
  "performed_by": { "id": "...", "full_name": "Admin Empresa", "email": "admin.empresa@sigif.com" },
  "created_at": "2026-10-08T18:30:00Z"
}
```

---

## Acceso después de desactivar

| Situación | Respuesta |
|---|---|
| `POST /api/v1/auth/login` de un usuario desactivado | `401 UNAUTHORIZED`, `invalid credentials` |
| Usuario desactivado con un token que ya tenía | `403 FORBIDDEN` en toda ruta que pida permiso |

- El login responde igual que con una contraseña incorrecta, a propósito: no revela que la cuenta existe ni su estado. Es el contrato de autenticación que ya tenía el backend. El intento queda en `login_attempt` con el motivo `account_not_active`.
- La desactivación **no cierra las sesiones abiertas**. Un token emitido antes sigue siendo válido hasta que expira (15 minutos por defecto), pero el usuario ya no tiene ningún permiso. Las rutas que solo piden sesión, como `GET /api/v1/auth/me`, sí le siguen respondiendo durante ese tiempo.

---

## Cambios de comportamiento a tener en cuenta

1. **`POST /users/:id/deactivate` responde `200` con el usuario**, no `204`. Es el único cambio de contrato.
2. **Desactivar dos veces responde `409`.** Antes la segunda llamada respondía `204`.
3. **No se puede desactivar la propia cuenta** (`409`). No estaba en el contrato; se agregó para que un administrador no se quede sin acceso por error.
4. **El email de un usuario desactivado sigue reservado.** Registrar otro usuario con ese email responde `409 CONFLICT`.
5. **`POST /users/:id/activate` revierte la desactivación**: vuelve a `active` y limpia `deleted_at`. Sigue respondiendo `204`.
6. **`DELETE /users/:id` ahora guarda `deleted_at`.** Antes lo calculaba pero no lo persistía. Su efecto es el mismo que desactivar; sigue fuera del flujo de esta HU.
7. **`make migrate-up` no duplica administradores desactivados.** El admin inicial (`admin@sigif.com`) se reactiva al ejecutar las migraciones, como ya estaba previsto en el seed.

## Base de datos

No hay tablas ni columnas nuevas. Se usa la columna `deleted_at` que `app_user` ya tenía. Las consultas del módulo de usuarios incluyen las filas con `deleted_at`, para que un usuario desactivado se pueda consultar y listar.

## Pruebas

- Tests de Go: `go test ./internal/modules/user/... ./internal/modules/auth/...`
- Tests de integración con Postgres: `SIGIF_TEST_DATABASE_DSN='host=localhost port=5432 user=sigif password=sigif dbname=sigif sslmode=disable' go test -tags=integration ./internal/modules/user/infrastructure/...` (usan un esquema temporal; no tocan las tablas de `public`).
- Bruno: carpeta `bruno/Users/Deactivation` (16 peticiones). Es autónoma: inicia sesión con `companyAdminEmail` / `companyAdminPassword` y crea su propio usuario.
