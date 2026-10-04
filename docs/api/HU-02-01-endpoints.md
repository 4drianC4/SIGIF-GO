# Reporte de endpoints — HU-02-01 Registrar usuario

## Información general

- HU: HU-02-01 — Registrar usuario (RF-02)
- Rama: `dev`
- Prefijo base: `/api/v1`
- Requiere contexto autenticado del backend: la ruta exige `Authorization: Bearer <access_token>` (validado por `AuthRequired`) y el permiso `users.create` (validado por `RequirePermission`). Solo un usuario con el rol adecuado (por defecto `superadmin`) puede registrar usuarios.
- Formato de respuesta exitosa: `{ "success": true, "data": ... }`.
- Formato de respuesta de error: `{ "success": false, "error": { "code": "CODIGO", "message": "descripción" } }`.

---

## Resumen rápido

| Método | Ruta | Permiso | Descripción |
|---|---|---|---|
| `POST` | `/api/v1/users` | `users.create` | Registra un nuevo usuario. |

---

## Autenticación / permisos (todas las rutas)

| Header | Obligatorio | Descripción |
|---|---|---|
| `Authorization` | Sí | `Bearer <access_token>` con una sesión activa. |
| `Content-Type` | Sí | `application/json`. |

Permisos usados en esta HU: `users.create`.

---

## 1) Registrar usuario — `POST /api/v1/users`

Crea un nuevo usuario validando los datos obligatorios, verificando que no exista un email duplicado y asignando el rol indicado.

### Permiso
`users.create`

### Body

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

| Campo | Tipo | Obligatorio | Restricciones / notas |
|---|---|---|---|
| `first_name` | string | Sí | Entre 1 y 80 caracteres. |
| `last_name` | string | Sí | Entre 1 y 80 caracteres. |
| `email` | string | Sí | Email válido y único en el sistema. |
| `password` | string | Sí | Mínimo 8 caracteres; se almacena hasheada con Argon2id. |
| `role` | string | Sí | `superadmin` o `soporte`. Si no existe el rol, se rechaza con `BAD_REQUEST`. |
| `area` | string | No | Máximo 120 caracteres. |

### Respuesta exitosa

**`201`**
```json
{
  "success": true,
  "data": {
    "id": "uuid",
    "role_id": "uuid",
    "role": "soporte",
    "first_name": "Juan",
    "last_name": "Pérez",
    "full_name": "Juan Pérez",
    "username": "soporte@sigif.com",
    "email": "soporte@sigif.com",
    "area": "Soporte técnico",
    "status": "active",
    "created_at": "2026-10-02T12:00:00Z",
    "updated_at": "2026-10-02T12:00:00Z"
  }
}
```

### Códigos de error

| Code | HTTP | Causa |
|---|---|---|
| `BAD_REQUEST` | 400 | Body ilegible, campos obligatorios faltantes o rol inexistente. |
| `CONFLICT` | 409 | Ya existe un usuario con ese email. |
| `FORBIDDEN` | 403 | El usuario autenticado no tiene el permiso `users.create`. |
| `UNAUTHORIZED` | 401 | Token ausente, inválido, expirado o sesión terminada. |
| `INTERNAL_ERROR` | 500 | Error de persistencia al guardar el usuario. |

---

## Estructura de respuesta de entidad `User`

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

Los campos `company_id`, `phone`, `area` y `last_access` se omiten cuando están vacíos o son nulos. El `password_hash` nunca se expone en la respuesta.

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
| `BAD_REQUEST` | 400 | Body ilegible, campos faltantes o rol inexistente. |
| `UNAUTHORIZED` | 401 | Token ausente, inválido, expirado o sesión terminada. |
| `FORBIDDEN` | 403 | Sin permiso `users.create`. |
| `CONFLICT` | 409 | Email duplicado. |
| `INTERNAL_ERROR` | 500 | Error de persistencia. |

---

## Casos borde / comportamiento no obvio

- El nuevo usuario se crea con estado `active` y `username` = `email` (no se solicita username por separado).
- `company_id` es opcional en esta HU: el registro no exige empresa, por lo que el usuario se crea sin `company_id` (se asignará cuando exista el flujo de empresa/suscripción).
- La validación de duplicados se basa en `email` (único global), no en `username`.
- Si se omite `role`, la validación lo rechaza por ser obligatorio; el rol debe existir en la tabla `role`.
- El registro no envía invitación ni contraseña temporal: el administrador define la contraseña inicial en el mismo request.

---

## Diferencias respecto a una versión anterior (si aplica)

- Antes el registro aceptaba `roles` (array) y `tenant_id`; ahora acepta un único `role` (`superadmin` o `soporte`) y no requiere tenant.
- Se eliminó el campo `settings` (idioma, tema, etc.); los datos personales se limitan a `first_name`, `last_name`, `phone` y `area`.
