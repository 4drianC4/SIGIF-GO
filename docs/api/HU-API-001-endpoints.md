# Endpoints — Autenticación (HU-01) y Registro de usuarios (HU-02-01)

- Prefijo base: `/api/v1`
- Formato exitoso: `{ "success": true, "data": ... }` (listados agregan `meta`).
- Formato de error: `{ "success": false, "error": { "code": "...", "message": "..." } }`.
- Autenticación: `Authorization: Bearer <access_token>`. El token se valida contra la sesión persistida en `user_session` (hash), no solo contra su firma.

## Resumen

| Método | Ruta | Permiso | Descripción |
|---|---|---|---|
| `POST` | `/api/v1/auth/login` | — | Inicia sesión (email + contraseña). |
| `POST` | `/api/v1/auth/logout` | — (token válido) | Revoca la sesión actual. |
| `GET`  | `/api/v1/auth/me` | — (token válido) | Devuelve el usuario autenticado. |
| `POST` | `/api/v1/users` | `users.create` | Registra un usuario. |
| `GET`  | `/api/v1/users` | `users.list` | Lista usuarios (paginado). |
| `GET`  | `/api/v1/users/by-email` | `users.read` | Busca usuario por email. |
| `GET`  | `/api/v1/users/:id` | `users.read` | Obtiene usuario por ID. |
| `PUT`  | `/api/v1/users/:id` | `users.update` | Actualiza datos del usuario. |
| `PUT`  | `/api/v1/users/:id/password` | `users.change_password` | Cambia la contraseña. |
| `DELETE` | `/api/v1/users/:id` | `users.delete` | Elimina (soft delete) al usuario. |
| `POST` | `/api/v1/users/:id/activate` | `users.activate` | Activa al usuario. |
| `POST` | `/api/v1/users/:id/deactivate` | `users.deactivate` | Desactiva al usuario. |

## 1) Login — `POST /api/v1/auth/login`

Body:

```json
{ "email": "admin@sigif.com", "password": "admin123" }
```

| Campo | Tipo | Obligatorio | Notas |
|---|---|---|---|
| `email` | string | Sí | Email válido. |
| `password` | string | Sí | Se compara contra el hash almacenado (Argon2id). |

Respuesta `200`:

```json
{
  "success": true,
  "data": {
    "user": {
      "id": "uuid",
      "role": "superadmin",
      "role_id": "uuid",
      "first_name": "Admin",
      "last_name": "SIGIF",
      "full_name": "Admin SIGIF",
      "username": "admin@sigif.com",
      "email": "admin@sigif.com",
      "status": "active"
    },
    "token": {
      "access_token": "jwt",
      "expires_in": 900,
      "token_type": "Bearer"
    }
  }
}
```

Errores:

| Code | HTTP | Causa |
|---|---|---|
| `BAD_REQUEST` | 400 | Body inválido o campos faltantes. |
| `UNAUTHORIZED` | 401 | Credenciales inválidas, cuenta inactiva o bloqueada (mensaje genérico). |

> El login devuelve siempre el mismo mensaje genérico (`invalid credentials`) para evitar enumeración de cuentas. La razón real se guarda en `login_attempt.failure_reason`.

Cada login exitoso crea una fila en `user_session` (token hasheado, IP, dispositivo, expiración) y un registro en `login_attempt`.

## 2) Logout — `POST /api/v1/auth/logout`

Requiere token válido. Marca la sesión actual como terminada (`ended_at` + `close_reason = 'manual'`).

Respuesta `204` sin contenido.

## 3) Me — `GET /api/v1/auth/me`

Devuelve el usuario autenticado (mismo shape de `user` del login).

## 4) Registrar usuario — `POST /api/v1/users`

Body:

```json
{
  "first_name": "Juan",
  "last_name": "Pérez",
  "email": "soporte@sigif.com",
  "password": "password123",
  "role": "soporte",
  "area": "Soporte técnico"
}
```

| Campo | Tipo | Obligatorio | Notas |
|---|---|---|---|
| `first_name` | string | Sí | 1-80 caracteres. |
| `last_name` | string | Sí | 1-80 caracteres. |
| `email` | string | Sí | Email válido y único. |
| `password` | string | Sí | Mínimo 8 caracteres. |
| `role` | string | Sí | `superadmin` o `soporte`. |
| `area` | string | No | Máximo 120 caracteres. |

Respuesta `201` con el usuario creado (sin el hash de contraseña).

Errores:

| Code | HTTP | Causa |
|---|---|---|
| `BAD_REQUEST` | 400 | Body inválido, campos faltantes o rol inexistente. |
| `CONFLICT` | 409 | Ya existe un usuario con ese email. |
| `FORBIDDEN` | 403 | El usuario autenticado no tiene el permiso `users.create`. |

## 5) Listar usuarios — `GET /api/v1/users`

Query: `page` (default 1), `limit` (default 20, máx 100).

Respuesta `200` con `data` (array) y `meta` (`page`, `limit`, `total`, `total_pages`).

## 6) Otros endpoints de usuario

- `GET /api/v1/users/:id` / `by-email` → `200` con el usuario, o `404`.
- `PUT /api/v1/users/:id` → actualiza `first_name`, `last_name`, `phone`, `area`.
- `PUT /api/v1/users/:id/password` → verifica `current_password` y guarda `new_password`.
- `DELETE /api/v1/users/:id` → soft delete (marca `deleted_at` y estado `inactive`).
- `POST /api/v1/users/:id/activate|deactivate` → cambia el estado.

## Permisos (RBAC)

| Rol | Permisos |
|---|---|
| `superadmin` | Todos (`users.create`, `users.list`, `users.read`, `users.update`, `users.delete`, `users.activate`, `users.deactivate`, `users.change_password`). |
| `soporte` | `users.list`, `users.read` (solo lectura). |

Los permisos se resuelven por petición vía `role_permission`; el middleware `RequirePermission(module, operation)` los aplica por ruta.

## Modelo de datos (tablas nuevas)

- `app_user`: usuarios (UUID PK, `role_id`, `status`, `password_algorithm`, soft delete).
- `role`, `permission`, `role_permission`: RBAC.
- `user_session`: sesiones (token hasheado, IP, dispositivo, expiración).
- `login_attempt`: auditoría de intentos de login.
