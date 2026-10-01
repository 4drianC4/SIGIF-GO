# Reporte de endpoints — HU-API-001 Autenticación y gestión de usuarios

## Información general

- HU: HU-API-001 — Autenticación y gestión de usuarios
- Rama: `main`
- Prefijo base: `/api/v1` para usuarios; autenticación se registra actualmente bajo `/auth`
- Requiere contexto autenticado del backend: no existe middleware HTTP que valide el token Bearer, `userId` o permisos. Las rutas de usuarios solo leen `X-Tenant-ID`; `logout-all` espera un `user_id` en el contexto interno de Fiber.
- Formato de respuesta exitosa: `{ "success": true, "data": ... }`. En listados se agrega `meta` con `page`, `limit`, `total` y `total_pages`.
- Formato de respuesta de error: `{ "success": false, "error": { "code": "CODIGO", "message": "descripción" } }`.

---

## Resumen rápido

| Método | Ruta | Permiso | Descripción |
|---|---|---|---|
| `GET` | `/health` | — | Comprueba que el servidor está disponible. |
| `POST` | `/auth/login` | — | Autentica al usuario y genera access/refresh tokens. |
| `POST` | `/auth/refresh` | — | Rota un refresh token válido. |
| `POST` | `/auth/logout` | — | Revoca el refresh token enviado. |
| `POST` | `/auth/logout-all` | — | Revoca todos los refresh tokens del usuario del contexto. |
| `POST` | `/api/v1/users/` | — | Crea un usuario. |
| `GET` | `/api/v1/users/` | — | Lista usuarios del tenant con paginación. |
| `GET` | `/api/v1/users/by-email` | — | Busca un usuario por email dentro del tenant. |
| `GET` | `/api/v1/users/:id` | — | Obtiene un usuario por UUID. |
| `PUT` | `/api/v1/users/:id` | — | Actualiza datos, roles y configuración del usuario. |
| `PUT` | `/api/v1/users/:id/password` | — | Cambia la contraseña. |
| `DELETE` | `/api/v1/users/:id` | — | Elimina lógicamente un usuario. |
| `POST` | `/api/v1/users/:id/activate` | — | Cambia el estado a `active`. |
| `POST` | `/api/v1/users/:id/deactivate` | — | Cambia el estado a `inactive`. |
| `POST` | `/api/v1/users/:id/suspend` | — | Cambia el estado a `suspended`. |

---

## Autenticación / permisos (todas las rutas)

| Header | Obligatorio | Descripción |
|---|---|---|
| `X-Tenant-ID` | Requerido funcionalmente en login; opcional en el middleware | UUID del tenant. Si está presente y es válido, se guarda como `tenant_id`; si se omite, el contexto queda vacío/UUID nulo. No se usa `DEFAULT_COMPANY_ID` en el código HTTP revisado. |
| `Authorization` | No validado por el backend actual | Los ejemplos de Bruno envían `Bearer <access_token>`, pero no hay middleware que lo parsee o valide. |
| `X-User-ID` | No | No se lee en las rutas revisadas. `logout-all` espera `c.Locals("user_id")`, que no es poblado por el middleware actual. |

Permisos usados en esta HU: ninguno implementado en las rutas actuales.

---

## 1) Health check — `GET /health`

Comprueba la disponibilidad del servidor y devuelve su versión y hora UTC.

### Permiso
`—`

### Respuesta exitosa

**`200`**
```json
{
  "status": "ok",
  "version": "1.0.0",
  "time": "2026-10-01T12:00:00Z"
}
```

### Códigos de error

| Code | HTTP | Causa |
|---|---|---|
| `INTERNAL_ERROR` | 500 | Error no controlado al procesar la solicitud. |

---

## 2) Iniciar sesión — `POST /auth/login`

Autentica un usuario dentro del tenant y devuelve un par de tokens.

### Permiso
`—`

### Body

```json
{
  "email": "usuario@ejemplo.com",
  "password": "password123"
}
```

| Campo | Tipo | Obligatorio | Restricciones / notas |
|---|---|---|---|
| `email` | string | Sí | Debe ser un email válido según el DTO; la validación automática no está conectada al handler. |
| `password` | string | Sí | Se compara contra el hash almacenado. |

### Respuesta exitosa

**`200`**
```json
{
  "success": true,
  "data": {
    "user": {
      "id": "uuid",
      "tenant_id": "uuid",
      "email": "usuario@ejemplo.com",
      "first_name": "Ana",
      "last_name": "Pérez",
      "full_name": "Ana Pérez",
      "roles": ["viewer"],
      "status": "active"
    },
    "tokens": {
      "access_token": "jwt",
      "refresh_token": "jwt",
      "expires_in": 3600,
      "token_type": "Bearer"
    }
  }
}
```

### Códigos de error

| Code | HTTP | Causa |
|---|---|---|
| `INVALID_TENANT` | 400 | Falta `X-Tenant-ID` válido en el contexto. |
| `INTERNAL_ERROR` | 400 | El parser no pudo leer el body. |
| `UNAUTHORIZED` | 401 | Credenciales inválidas. |
| `FORBIDDEN` | 401 | La cuenta no está activa; el handler conserva el status 401. |

---

## 3) Renovar tokens — `POST /auth/refresh`

Valida, revoca y reemplaza el refresh token enviado.

### Permiso
`—`

### Body

```json
{
  "refresh_token": "jwt-refresh-token"
}
```

| Campo | Tipo | Obligatorio | Restricciones / notas |
|---|---|---|---|
| `refresh_token` | string | Sí | Debe ser válido, no estar revocado ni expirado. |

### Respuesta exitosa

**`200`**
```json
{
  "success": true,
  "data": {
    "access_token": "jwt",
    "refresh_token": "jwt-nuevo",
    "expires_in": 3600,
    "token_type": "Bearer"
  }
}
```

### Códigos de error

| Code | HTTP | Causa |
|---|---|---|
| `INTERNAL_ERROR` | 400 | El parser no pudo leer el body. |
| `UNAUTHORIZED` | 401 | Refresh token inválido, expirado, revocado o usuario inactivo. |

---

## 4) Cerrar sesión — `POST /auth/logout`

Revoca el refresh token enviado; no devuelve un cuerpo de respuesta.

### Permiso
`—`

### Body

```json
{
  "refresh_token": "jwt-refresh-token"
}
```

| Campo | Tipo | Obligatorio | Restricciones / notas |
|---|---|---|---|
| `refresh_token` | string | Sí | Token cuyo hash se buscará para revocarlo. |

### Respuesta exitosa

**`204`** Sin contenido.

### Códigos de error

| Code | HTTP | Causa |
|---|---|---|
| `INTERNAL_ERROR` | 400 | El parser no pudo leer el body. |
| `INTERNAL_ERROR` | 500 | Fallo al revocar el token. |

---

## 5) Cerrar sesión en todos los dispositivos — `POST /auth/logout-all`

Revoca todos los refresh tokens asociados al usuario presente en `c.Locals("user_id")`.

### Permiso
`—`

### Body

No requiere body. El `user_id` no se obtiene de un header ni del body.

### Respuesta exitosa

**`204`** Sin contenido.

### Códigos de error

| Code | HTTP | Causa |
|---|---|---|
| `UNAUTHORIZED` | 401 | No existe un UUID de usuario en `c.Locals("user_id")`. |
| `INTERNAL_ERROR` | 500 | Fallo al revocar los tokens. |

---

## 6) Crear usuario — `POST /api/v1/users/`

Crea un usuario en el tenant guardado por el middleware a partir de `X-Tenant-ID`.

### Permiso
`—` (no se verifica en el backend actual).

### Body

```json
{
  "email": "cajero@ejemplo.com",
  "password": "password123",
  "first_name": "Juan",
  "last_name": "Pérez",
  "phone": "+57 300 123 4567",
  "roles": ["cashier"]
}
```

| Campo | Tipo | Obligatorio | Restricciones / notas |
|---|---|---|---|
| `email` | string | Sí | Email válido según el DTO. |
| `password` | string | Sí | Mínimo 8 caracteres según el DTO; se almacena hasheada. |
| `first_name` | string | Sí | Entre 1 y 100 caracteres. |
| `last_name` | string | Sí | Entre 1 y 100 caracteres. |
| `phone` | string | No | Máximo 50 caracteres. |
| `roles` | enum[] | No | `super_admin` \| `tenant_admin` \| `company_admin` \| `manager` \| `cashier` \| `inventory` \| `sales` \| `viewer`. Si se omite, el dominio asigna `viewer`. |

### Respuesta exitosa

**`201`**
```json
{
  "success": true,
  "data": {
    "id": "uuid",
    "tenant_id": "uuid",
    "email": "cajero@ejemplo.com",
    "first_name": "Juan",
    "last_name": "Pérez",
    "full_name": "Juan Pérez",
    "roles": ["cashier"],
    "status": "pending",
    "settings": {
      "language": "es",
      "timezone": "UTC",
      "theme": "light",
      "notifications": true
    },
    "created_at": "2026-10-01T12:00:00Z",
    "updated_at": "2026-10-01T12:00:00Z"
  }
}
```

### Códigos de error

| Code | HTTP | Causa |
|---|---|---|
| `INTERNAL_ERROR` | 400 | El parser no pudo leer el body. |
| `INTERNAL_ERROR` | 500 | Fallo al crear el usuario. |

---

## 7) Listar usuarios — `GET /api/v1/users/`

Lista usuarios filtrados por el tenant del contexto.

### Permiso
`—`

### Query params

| Nombre | Obligatorio | Tipo | Descripción |
|---|---|---|---|
| `page` | No, por defecto `1` | integer | Si es menor que 1, se ajusta a `1`. |
| `limit` | No, por defecto `20` | integer | Valores menores que 1 o mayores que `100` se ajustan silenciosamente a `20`. |

### Respuesta exitosa

**`200`**
```json
{
  "success": true,
  "data": [
    {
      "id": "uuid",
      "tenant_id": "uuid",
      "email": "usuario@ejemplo.com",
      "first_name": "Ana",
      "last_name": "Pérez",
      "full_name": "Ana Pérez",
      "roles": ["viewer"],
      "status": "active",
      "settings": {
        "language": "es",
        "timezone": "UTC",
        "theme": "light",
        "notifications": true
      },
      "created_at": "2026-10-01T12:00:00Z",
      "updated_at": "2026-10-01T12:00:00Z"
    }
  ],
  "meta": {
    "page": 1,
    "limit": 20,
    "total": 1,
    "total_pages": 1
  }
}
```

### Códigos de error

| Code | HTTP | Causa |
|---|---|---|
| `INTERNAL_ERROR` | 500 | Fallo al consultar usuarios. |

---

## 8) Obtener usuario por email — `GET /api/v1/users/by-email`

Busca un usuario por email dentro del tenant del contexto.

### Permiso
`—`

### Query params

| Nombre | Obligatorio | Tipo | Descripción |
|---|---|---|---|
| `email` | Sí | string | Email exacto que se buscará. |

### Respuesta exitosa

**`200`**
```json
{
  "success": true,
  "data": {
    "id": "uuid",
    "tenant_id": "uuid",
    "email": "usuario@ejemplo.com",
    "first_name": "Ana",
    "last_name": "Pérez",
    "full_name": "Ana Pérez",
    "roles": ["viewer"],
    "status": "active",
    "settings": {
      "language": "es",
      "timezone": "UTC",
      "theme": "light",
      "notifications": true
    },
    "created_at": "2026-10-01T12:00:00Z",
    "updated_at": "2026-10-01T12:00:00Z"
  }
}
```

### Códigos de error

| Code | HTTP | Causa |
|---|---|---|
| `INTERNAL_ERROR` | 400 | Falta `email` en la query. |
| `NOT_FOUND` | 404 | No existe un usuario con ese email en el tenant. |
| `INTERNAL_ERROR` | 500 | Fallo al consultar el usuario. |

---

## 9) Obtener usuario por ID — `GET /api/v1/users/:id`

Obtiene un usuario por UUID.

### Permiso
`—`

### Params de ruta
- `id`: UUID — identificador del usuario.

### Respuesta exitosa

**`200`** Mismo shape de usuario mostrado en `GET /api/v1/users/by-email`.

### Códigos de error

| Code | HTTP | Causa |
|---|---|---|
| `INTERNAL_ERROR` | 400 | `id` no es un UUID válido. |
| `NOT_FOUND` | 404 | El usuario no existe. |
| `INTERNAL_ERROR` | 500 | Fallo al consultar el usuario. |

---

## 10) Actualizar usuario — `PUT /api/v1/users/:id`

Actualiza datos personales, roles y configuración del usuario.

### Permiso
`—`

### Params de ruta
- `id`: UUID — identificador del usuario.

### Body

```json
{
  "first_name": "Juan Carlos",
  "last_name": "Pérez Gómez",
  "phone": "+57 300 123 4567",
  "avatar_url": "https://ejemplo.com/avatar.png",
  "roles": ["cashier", "inventory"],
  "settings": {
    "language": "es",
    "timezone": "America/Bogota",
    "theme": "dark",
    "notifications": true
  }
}
```

| Campo | Tipo | Obligatorio | Restricciones / notas |
|---|---|---|---|
| `first_name` | string | Sí | Entre 1 y 100 caracteres según el DTO. |
| `last_name` | string | Sí | Entre 1 y 100 caracteres según el DTO. |
| `phone` | string | No | Máximo 50 caracteres. |
| `avatar_url` | string | No | URL válida según el DTO. |
| `roles` | enum[] | No | Roles soportados: `super_admin`, `tenant_admin`, `company_admin`, `manager`, `cashier`, `inventory`, `sales`, `viewer`. |
| `settings` | object | Sí en el struct | `language`, `timezone`, `theme`, `notifications`. El handler no ejecuta validación automática. |

### Respuesta exitosa

**`200`** Devuelve `{ "success": true, "data": <usuario> }` con el shape de entidad documentado al final.

### Códigos de error

| Code | HTTP | Causa |
|---|---|---|
| `INTERNAL_ERROR` | 400 | UUID inválido o body ilegible. |
| `INTERNAL_ERROR` | 500 | Fallo al actualizar el usuario. |

---

## 11) Cambiar contraseña — `PUT /api/v1/users/:id/password`

Verifica la contraseña actual y almacena la nueva contraseña hasheada.

### Permiso
`—`

### Params de ruta
- `id`: UUID — identificador del usuario.

### Body

```json
{
  "current_password": "password123",
  "new_password": "newpassword456"
}
```

| Campo | Tipo | Obligatorio | Restricciones / notas |
|---|---|---|---|
| `current_password` | string | Sí | Debe coincidir con la contraseña actual. |
| `new_password` | string | Sí | Mínimo 8 caracteres según el DTO. |

### Respuesta exitosa

**`204`** Sin contenido.

### Códigos de error

| Code | HTTP | Causa |
|---|---|---|
| `INTERNAL_ERROR` | 400 | UUID inválido o body ilegible. |
| `INTERNAL_ERROR` | 500 | Fallo al verificar o guardar la contraseña. |

---

## 12) Eliminar usuario — `DELETE /api/v1/users/:id`

Realiza un borrado lógico: marca el usuario como eliminado e inactivo.

### Permiso
`—`

### Params de ruta
- `id`: UUID — identificador del usuario.

### Respuesta exitosa

**`204`** Sin contenido.

### Códigos de error

| Code | HTTP | Causa |
|---|---|---|
| `INTERNAL_ERROR` | 400 | `id` no es un UUID válido. |
| `INTERNAL_ERROR` | 500 | Fallo al eliminar el usuario. |

---

## 13) Activar usuario — `POST /api/v1/users/:id/activate`

Cambia el estado del usuario a `active`.

### Permiso
`—`

### Params de ruta
- `id`: UUID — identificador del usuario.

### Respuesta exitosa

**`204`** Sin contenido.

### Códigos de error

| Code | HTTP | Causa |
|---|---|---|
| `INTERNAL_ERROR` | 400 | `id` no es un UUID válido. |
| `INTERNAL_ERROR` | 500 | Fallo al actualizar el estado. |

---

## 14) Desactivar usuario — `POST /api/v1/users/:id/deactivate`

Cambia el estado del usuario a `inactive`.

### Permiso
`—`

### Params de ruta
- `id`: UUID — identificador del usuario.

### Respuesta exitosa

**`204`** Sin contenido.

### Códigos de error

| Code | HTTP | Causa |
|---|---|---|
| `INTERNAL_ERROR` | 400 | `id` no es un UUID válido. |
| `INTERNAL_ERROR` | 500 | Fallo al actualizar el estado. |

---

## 15) Suspender usuario — `POST /api/v1/users/:id/suspend`

Cambia el estado del usuario a `suspended`.

### Permiso
`—`

### Params de ruta
- `id`: UUID — identificador del usuario.

### Respuesta exitosa

**`204`** Sin contenido.

### Códigos de error

| Code | HTTP | Causa |
|---|---|---|
| `INTERNAL_ERROR` | 400 | `id` no es un UUID válido. |
| `INTERNAL_ERROR` | 500 | Fallo al actualizar el estado. |

---

## Estructura de respuesta de entidad `Usuario`

```json
{
  "id": "uuid",
  "tenant_id": "uuid",
  "email": "usuario@ejemplo.com",
  "first_name": "Ana",
  "last_name": "Pérez",
  "full_name": "Ana Pérez",
  "phone": "string",
  "avatar_url": "https://ejemplo.com/avatar.png",
  "roles": ["viewer"],
  "status": "active | inactive | pending | suspended",
  "last_login_at": "date | null",
  "settings": {
    "language": "es",
    "timezone": "UTC",
    "theme": "light",
    "notifications": true
  },
  "created_at": "date",
  "updated_at": "date"
}
```

Los campos `phone`, `avatar_url` y `last_login_at` se omiten cuando están vacíos o son nulos.

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
| `INVALID_TENANT` | 400 | `X-Tenant-ID` ausente o inválido en login. |
| `NOT_FOUND` | 404 | Usuario no encontrado en consultas que sí lo detectan. |
| `UNAUTHORIZED` | 401 | Credenciales/token inválido o `logout-all` sin `user_id` en contexto. |
| `FORBIDDEN` | 401 | Cuenta inactiva durante login; el handler devuelve 401. |
| `INTERNAL_ERROR` | 400 | Error de parseo, UUID inválido o query obligatoria ausente; el handler no conserva un código específico. |
| `INTERNAL_ERROR` | 500 | Error de persistencia o de caso de uso en handlers de usuarios/logout. |

---

## Casos borde / comportamiento no obvio

- El middleware lee `X-Tenant-ID`, no `x-company-id`. Los nombres de headers HTTP no distinguen mayúsculas y minúsculas, pero el significado implementado es tenant.
- La autenticación está montada en `/auth/*`, no en `/api/v1/auth/*`, aunque los archivos de Bruno usan el segundo prefijo.
- No se valida `Authorization: Bearer ...` en el pipeline HTTP actual; enviarlo no convierte una ruta en autenticada.
- Las etiquetas `validate:"..."` de los DTO existen, pero los handlers revisados solo ejecutan `BodyParser`; no invocan el validador compartido.
- En paginación, `page < 1` se convierte en `1`; `limit < 1` o `limit > 100` se convierte silenciosamente en `20`.
- El listado devuelve el array directamente en `data` y la paginación en `meta`, no campos `total`, `page` y `limit` al mismo nivel de `data`.
- Los usuarios nuevos empiezan en estado `pending`; si no se envían roles, se asigna `viewer`.
- Las búsquedas por ID no filtran explícitamente por tenant en el handler; la búsqueda por email sí recibe el tenant del contexto.
- Los errores de dominio de varias rutas de usuarios se envuelven como `INTERNAL_ERROR` y HTTP 500 por el handler, aunque el dominio pueda representar otra causa.
- `logout-all` requiere que algún componente externo haya poblado `c.Locals("user_id")`; el middleware incluido en este servidor no lo hace.
- Las rutas de cambio de estado, eliminación y cambio de contraseña responden 204 sin cuerpo.

---

## Diferencias respecto a una versión anterior (si aplica)

Esta rama documenta el estado actual de `main`; no se proporcionó una rama o versión anterior para comparar.