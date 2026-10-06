# Reporte de endpoints — HU-02-02 Editar usuario

## Información general

- HU: HU-02-02 — Editar usuario (RF-02). Depende de HU-02-01 (Registrar usuario).
- Rama: `dev`
- Prefijo base: `/api/v1`
- Requiere contexto autenticado del backend: la ruta exige `Authorization: Bearer <access_token>` (validado por `AuthRequired`, sesión activa en `user_session`) y el permiso `users.update` (validado por `RequirePermission`).
- Formato de respuesta exitosa: `{ "success": true, "data": ... }`.
- Formato de respuesta de error: `{ "success": false, "error": { "code": "CODIGO", "message": "descripción", "details": { "campo": "motivo" } } }`. `details` solo aparece en errores de validación.

---

## Resumen rápido

| Método | Ruta | Permiso | Descripción |
|---|---|---|---|
| `PATCH` | `/api/v1/users/:id` | `users.update` | Edita los datos permitidos de un usuario existente. |

---

## Autenticación / permisos (todas las rutas)

| Header | Obligatorio | Descripción |
|---|---|---|
| `Authorization` | Sí | `Bearer <access_token>` con una sesión activa. |
| `Content-Type` | Sí | `application/json`. |

Permisos usados en esta HU: `users.update` (lo tiene el rol `superadmin`; `soporte` recibe `403`).

---

## 1) Editar usuario — `PATCH /api/v1/users/:id`

Edita un usuario existente con semántica de parche: **solo se aplican los campos enviados**; los omitidos se conservan. El cuerpo acepta exactamente los **mismos campos que el registro** (HU-02-01), todos opcionales.

### Permiso
`users.update`

### Params de ruta
- `id`: UUID — identificador del usuario.

### Body

```json
{
  "first_name": "Juan Carlos",
  "last_name": "Pérez Gómez",
  "email": "nuevo@sigif.com",
  "password": "nuevaPassword123",
  "role": "superadmin",
  "area": "Sistemas"
}
```

| Campo | Tipo | Obligatorio | Restricciones / notas |
|---|---|---|---|
| `first_name` | string | No | Entre 1 y 80 caracteres si se envía. |
| `last_name` | string | No | Entre 1 y 80 caracteres si se envía. |
| `email` | string | No | Formato email. Si cambia, debe ser único en el sistema; al cambiarlo también se actualiza `username`. |
| `password` | string | No | Mínimo 8 caracteres si se envía; se almacena hasheada con Argon2id. Es un restablecimiento por parte del admin: no exige la contraseña actual (para eso existe `PUT /users/:id/password` del propio usuario). |
| `role` | string | No | `superadmin` o `soporte`. Si el rol no existe, se rechaza con `VALIDATION_ERROR` (400). |
| `area` | string | No | Máximo 120 caracteres; permite vaciarlo (`""`). |

### Respuesta exitosa

**`200`**

```json
{
  "success": true,
  "data": {
    "id": "uuid",
    "role_id": "uuid",
    "role": "superadmin",
    "first_name": "Juan Carlos",
    "last_name": "Pérez Gómez",
    "full_name": "Juan Carlos Pérez Gómez",
    "username": "nuevo@sigif.com",
    "email": "nuevo@sigif.com",
    "area": "Sistemas",
    "status": "active",
    "created_at": "2026-10-02T12:00:00Z",
    "updated_at": "2026-10-03T12:30:00Z"
  }
}
```

### Códigos de error

| Code | HTTP | Causa |
|---|---|---|
| `BAD_REQUEST` | 400 | JSON ilegible; validación de formato fallida (campos en `details`). |
| `NOT_FOUND` | 404 | El usuario no existe. |
| `CONFLICT` | 409 | El nuevo email ya pertenece a otro usuario. |
| `FORBIDDEN` | 403 | El usuario autenticado no tiene el permiso `users.update`. |
| `UNAUTHORIZED` | 401 | Token ausente, inválido, expirado o sesión terminada. |
| `INTERNAL_ERROR` | 500 | Error de persistencia al guardar el usuario. |

---

## Estructura de respuesta de entidad `User`

Misma estructura que en HU-02-01:

```json
{
  "id": "uuid",
  "company_id": "uuid | null",
  "role_id": "uuid",
  "role": "superadmin | soporte",
  "first_name": "string",
  "last_name": "string",
  "full_name": "string",
  "username": "string",
  "email": "string",
  "phone": "string",
  "area": "string",
  "status": "active | inactive | invited | locked",
  "last_access": "date | null",
  "created_at": "date",
  "updated_at": "date"
}
```

Los campos `company_id`, `phone`, `area` y `last_access` se omiten cuando están vacíos o son nulos. El `password_hash` nunca se expone.

---

## Tabla de errores consolidada

```json
{
  "success": false,
  "error": {
    "code": "CODIGO",
    "message": "descripción"
  }
}
```

| Code | HTTP | Origen |
|---|---|---|
| `BAD_REQUEST` | 400 | Body ilegible, validación o rol inexistente. |
| `UNAUTHORIZED` | 401 | Token ausente/inválido/expirado o sesión terminada. |
| `FORBIDDEN` | 403 | Sin permiso `users.update`. |
| `NOT_FOUND` | 404 | Usuario inexistente. |
| `CONFLICT` | 409 | Email duplicado. |
| `INTERNAL_ERROR` | 500 | Error de persistencia. |

---

## Casos borde / comportamiento no obvio

- **Cambios parciales**: un body `{}` o con solo algunos campos aplica únicamente esos campos; el resto conserva su valor (el servicio hace read-modify-write y persiste todo el agregado).
- **Email**: si se envía uno distinto al actual, se verifica unicidad (`CONFLICT` 409). El `username` se mantiene igual al email (no existe ingreso de username separado).
- **Contraseña**: enviar `password` sustituye la contraseña (re-hash Argon2id). No requiere la contraseña anterior porque es un cambio administrativo; la sesión actual del usuario editado no se invalida automáticamente.
- **Role**: el nombre del rol se resuelve a su `role_id`; un nombre inexistente responde `400`.
- **Estado**: esta HU no cubre activación/desactivación; eso se hace con `POST /users/:id/activate|deactivate` (dueño del estado separado del PATCH).
- **`phone`** no forma parte de los campos del panel de registro/edición de esta HU.

---

## Diferencias respecto a una versión anterior (si aplica)

- Se reemplazó el `PUT /api/v1/users/:id` (solo aceptaba `first_name`, `last_name`, `phone`, `area`) por este `PATCH`, que acepta **todos los campos del registro** (`email`, `role`, `password`, `area`, `first_name`, `last_name`) como opcionales.
