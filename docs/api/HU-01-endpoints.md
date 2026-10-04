# Reporte de endpoints — HU-01 Autenticación de usuario y control de acceso por rol

## Información general

- HU: HU-01 — Autenticación de usuario y control de acceso por rol (RF-01)
- Rama: `dev`
- Prefijo base: `/api/v1`
- Requiere contexto autenticado del backend: el middleware `AuthRequired` valida el token Bearer (firma JWT + sesión activa en `user_session`) y deja en el contexto `user_id`, `session_id`, `email`, `role` y `company_id` (si existe). El control de acceso por rol se aplica con el middleware `RequirePermission(module, operation)` en cada ruta protegida.
- Formato de respuesta exitosa: `{ "success": true, "data": ... }`.
- Formato de respuesta de error: `{ "success": false, "error": { "code": "CODIGO", "message": "descripción" } }`.

---

## Resumen rápido

| Método | Ruta | Permiso | Descripción |
|---|---|---|---|
| `POST` | `/api/v1/auth/login` | — | Inicia sesión con email y contraseña; crea una sesión y devuelve un access token. |
| `POST` | `/api/v1/auth/logout` | — (token válido) | Cierra la sesión actual (revoca el token). |
| `GET` | `/api/v1/auth/me` | — (token válido) | Devuelve el usuario autenticado. |

---

## Autenticación / permisos (todas las rutas)

| Header | Obligatorio | Descripción |
|---|---|---|
| `Authorization` | Sí en `logout` y `me`; no en `login` | `Bearer <access_token>`. El token se valida contra su firma JWT **y** contra la sesión persistida en `user_session` (hash del token). |
| `Content-Type` | Sí en `login` | `application/json`. |

Permisos usados en esta HU: `login`, `logout` y `me` no requieren un permiso RBAC específico, solo un token válido y una sesión activa. El control de acceso por rol (`RequirePermission`) se describe en la sección [Control de acceso por rol (RBAC)](#control-de-acceso-por-rol-rbac).

---

## 1) Iniciar sesión — `POST /api/v1/auth/login`

Valida las credenciales, verifica el estado de la cuenta, crea una sesión en `user_session`, registra el intento en `login_attempt` y devuelve un access token.

### Permiso
`—`

### Body

```json
{
  "email": "admin@sigif.com",
  "password": "admin123"
}
```

| Campo | Tipo | Obligatorio | Restricciones / notas |
|---|---|---|---|
| `email` | string | Sí | Email válido. |
| `password` | string | Sí | Se compara contra el hash almacenado (Argon2id). |

### Respuesta exitosa

**`200`**
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

### Códigos de error

| Code | HTTP | Causa |
|---|---|---|
| `BAD_REQUEST` | 400 | Body ilegible o campos faltantes. |
| `UNAUTHORIZED` | 401 | Credenciales inválidas, cuenta inactiva/bloqueada o error interno de sesión. El mensaje es siempre genérico (`invalid credentials`) para evitar enumeración de cuentas. |

> La razón real del fallo se guarda en `login_attempt.failure_reason` (`user_not_found`, `invalid_password`, `account_not_active`), pero no se expone al cliente.

---

## 2) Cerrar sesión — `POST /api/v1/auth/logout`

Marca la sesión actual como terminada (`ended_at` + `close_reason = 'manual'`). El token deja de ser válido en las siguientes peticiones porque el middleware verifica la sesión en base de datos.

### Permiso
`—` (token válido)

### Body

No requiere body.

### Respuesta exitosa

**`204`** Sin contenido.

### Códigos de error

| Code | HTTP | Causa |
|---|---|---|
| `UNAUTHORIZED` | 401 | Token ausente, inválido, expirado o sesión ya terminada. |
| `INTERNAL_ERROR` | 500 | Fallo al persistir el cierre de la sesión. |

---

## 3) Usuario autenticado — `GET /api/v1/auth/me`

Devuelve el usuario correspondiente al token actual.

### Permiso
`—` (token válido)

### Respuesta exitosa

**`200`** Mismo shape de `user` mostrado en el login.

### Códigos de error

| Code | HTTP | Causa |
|---|---|---|
| `UNAUTHORIZED` | 401 | Token ausente, inválido, expirado o sesión terminada. |
| `NOT_FOUND` | 404 | El usuario ya no existe. |

---

## Control de acceso por rol (RBAC)

El token transporta `user_id` y `session_id`; los permisos se resuelven por petición consultando `role_permission` (no se cachean en el token). El middleware `RequirePermission(module, operation)` se antepone a cada ruta protegida.

Roles semilla:

| Rol | Permisos |
|---|---|
| `superadmin` | Todos los permisos implementados. |
| `soporte` | Solo lectura de usuarios (`users.list`, `users.read`). |

Cuando el usuario autenticado no tiene el permiso requerido, se responde `403 FORBIDDEN`.

### Registro de sesiones

- Cada login exitoso crea una fila en `user_session` con el **hash** del token (`token_hash`), `ip_address`, `device`, `started_at` y `expires_at`. Nunca se almacena el token en claro.
- Cada intento de login (exitoso o no) crea una fila en `login_attempt` (`email`, `ip_address`, `successful`, `failure_reason`).

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
  "status": "active | inactive | invited | locked"
}
```

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
| `BAD_REQUEST` | 400 | Body ilegible o campos faltantes. |
| `UNAUTHORIZED` | 401 | Credenciales inválidas, token inválido/expirado o sesión terminada. |
| `FORBIDDEN` | 403 | El rol del usuario no tiene el permiso requerido. |
| `NOT_FOUND` | 404 | Usuario no encontrado en `me`. |
| `INTERNAL_ERROR` | 500 | Error de persistencia o caso de uso. |

---

## Casos borde / comportamiento no obvio

- El login responde el mismo mensaje para `user_not_found`, `invalid_password` y `account_not_active`; la diferencia solo queda en `login_attempt.failure_reason`.
- Un token firmado pero con sesión revocada/expirada es rechazado por el middleware (la sesión se valida contra `user_session` en cada petición).
- `company_id` es opcional: si el usuario no tiene empresa asignada, el campo se omite en la respuesta y el token no lo transporta.
- El access token expira según `SIGIF_JWT_ACCESS_TOKEN_EXPIRY` (minutos); no existe refresh token en esta HU.

---

## Diferencias respecto a una versión anterior (si aplica)

- Se eliminó el flujo de `refresh token` y `logout-all`; ahora hay un único access token por sesión persistida en `user_session`.
- El login ya no requiere `X-Tenant-ID`; la identidad es global (email + contraseña) y `company_id` se resuelve desde el usuario autenticado.
