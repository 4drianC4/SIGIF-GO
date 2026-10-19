# Reporte de endpoints — HU-01 Autenticación de usuario y control de acceso por rol

## Información general

- HU: HU-01 — Autenticación de usuario y control de acceso por rol (RF-01)
- Rama: `dev`
- Autenticación del navegador: cookie HttpOnly `sigif_access_token` (ver sección [Autenticación / permisos](#autenticación--permisos-todas-las-rutas)).
- Prefijo base: `/api/v1`
- Requiere contexto autenticado del backend: el middleware `AuthRequired` obtiene el access token de la cookie HttpOnly `sigif_access_token` (navegador) o, si el cliente lo envía explícitamente, del header `Authorization: Bearer` (Bruno, pruebas, clientes no navegador); valida firma JWT + sesión activa en `user_session` y deja en el contexto `user_id`, `session_id`, `email`, `role` y `company_id` (si existe). El control de acceso por rol se aplica con el middleware `RequirePermission(module, operation)` en cada ruta protegida.
- Formato de respuesta exitosa: `{ "success": true, "data": ... }`.
- Formato de respuesta de error: `{ "success": false, "error": { "code": "CODIGO", "message": "descripción" } }`.

---

## Resumen rápido

| Método | Ruta | Permiso | Descripción |
|---|---|---|---|
| `POST` | `/api/v1/auth/login` | — | Inicia sesión con email y contraseña; crea una sesión y emite la cookie HttpOnly con el access token. |
| `POST` | `/api/v1/auth/logout` | — (token válido) | Cierra la sesión actual (revoca el token) y expira la cookie. |
| `GET` | `/api/v1/auth/me` | — (token válido) | Devuelve el usuario autenticado. |

---

## Autenticación / permisos (todas las rutas)

El access token es un JWT que se valida contra su firma **y** contra la sesión persistida en `user_session` (hash del token). Puede llegar por dos vías:

| Credencial | Cuándo | Descripción |
|---|---|---|
| Cookie `sigif_access_token` | Flujo navegador (Angular) | La emite el login como `HttpOnly`; el navegador la adjunta sola si la petición se hace con credenciales (`withCredentials: true`). JavaScript no puede leerla. |
| `Authorization: Bearer <access_token>` | Respaldo para Bruno, pruebas automatizadas y clientes no navegador | Se mantiene por compatibilidad. |

**Prioridad:** si la petición trae `Authorization`, se usa ese header y la cookie se ignora; si no lo trae, se usa la cookie. Un `Authorization` mal formado responde `401` sin recurrir a la cookie. Así un cliente que envía su credencial de forma explícita (por ejemplo Bruno con varios usuarios) nunca queda autenticado por una cookie que tenga guardada. El navegador nunca añade `Authorization` por su cuenta, por lo que el flujo cookie no se ve afectado.

Si la credencial que falla viene de la cookie (token inválido, expirado o sesión revocada), la respuesta `401` además expira la cookie para que el navegador la descarte.

| Header | Obligatorio | Descripción |
|---|---|---|
| `Cookie` | Sí en `logout` y `me` (flujo navegador) | `sigif_access_token=<jwt>`, lo gestiona el navegador. |
| `Authorization` | Alternativa a la cookie en `logout` y `me` | `Bearer <access_token>`. |
| `Origin` | Lo envía el navegador | Debe pertenecer a `cors.allow_origins` en métodos que cambian estado (ver [Protección CSRF](#cors-y-protección-csrf)). |
| `Content-Type` | Sí en `login` | `application/json`. |

### Cookie de sesión

| Atributo | Valor | Configuración |
|---|---|---|
| Nombre | `sigif_access_token` | `auth_cookie.name` / `SIGIF_AUTH_COOKIE_NAME` |
| `HttpOnly` | Siempre | — |
| `Path` | `/api/v1` | `auth_cookie.path` / `SIGIF_AUTH_COOKIE_PATH` |
| `Domain` | Vacío (host del backend) | `auth_cookie.domain` / `SIGIF_AUTH_COOKIE_DOMAIN` |
| `Max-Age` / `Expires` | Igual a `expires_in` del JWT | Deriva de `SIGIF_JWT_ACCESS_TOKEN_EXPIRY` |
| `Secure` | `false` en local; `true` en producción (HTTPS) | `auth_cookie.secure` / `SIGIF_AUTH_COOKIE_SECURE` |
| `SameSite` | `Lax` | `auth_cookie.same_site` / `SIGIF_AUTH_COOKIE_SAME_SITE` (`Lax`, `Strict` o `None`; `None` exige `Secure=true` o el servidor no arranca) |

`localhost:4200` y `localhost:4600` son el mismo *site* (el puerto no cuenta), por lo que `SameSite=Lax` funciona en local. Si en producción frontend y backend quedan en dominios registrables distintos, se necesita `SameSite=None` + `Secure=true`.

Permisos usados en esta HU: `login`, `logout` y `me` no requieren un permiso RBAC específico, solo un token válido y una sesión activa. El control de acceso por rol (`RequirePermission`) se describe en la sección [Control de acceso por rol (RBAC)](#control-de-acceso-por-rol-rbac).

---

## 1) Iniciar sesión — `POST /api/v1/auth/login`

Valida las credenciales, verifica el estado de la cuenta, crea una sesión en `user_session`, registra el intento en `login_attempt` y emite la cookie `sigif_access_token` con el access token. El token **no** se devuelve en el JSON para que no sea accesible desde JavaScript.

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
    "expires_in": 900
  }
}
```

Headers de respuesta:

```
Set-Cookie: sigif_access_token=<jwt>; max-age=900; expires=<fecha>; path=/api/v1; HttpOnly; SameSite=Lax
```

`expires_in` (segundos) indica cuándo caduca la sesión; al expirar, el usuario debe iniciar sesión de nuevo (no hay refresh token).

### Códigos de error

| Code | HTTP | Causa |
|---|---|---|
| `BAD_REQUEST` | 400 | Body ilegible o campos faltantes. |
| `UNAUTHORIZED` | 401 | Credenciales inválidas, cuenta inactiva/bloqueada o error interno de sesión. El mensaje es siempre genérico (`invalid credentials`) para evitar enumeración de cuentas. No se emite cookie. |
| `FORBIDDEN` | 403 | `Origin` no permitido (protección CSRF). No se emite cookie. |

> La razón real del fallo se guarda en `login_attempt.failure_reason` (`user_not_found`, `invalid_password`, `account_not_active`), pero no se expone al cliente.

---

## 2) Cerrar sesión — `POST /api/v1/auth/logout`

Marca la sesión actual como terminada (`ended_at` + `close_reason = 'manual'`) y expira la cookie `sigif_access_token` con el mismo nombre, `Path` y atributos con los que se creó. El token deja de ser válido en las siguientes peticiones porque el middleware verifica la sesión en base de datos.

### Permiso
`—` (token válido)

### Body

No requiere body.

### Respuesta exitosa

**`204`** Sin contenido.

```
Set-Cookie: sigif_access_token=; expires=Thu, 01 Jan 1970 00:00:00 GMT; path=/api/v1; HttpOnly; SameSite=Lax
```

### Códigos de error

| Code | HTTP | Causa |
|---|---|---|
| `UNAUTHORIZED` | 401 | Token ausente, inválido, expirado o sesión ya terminada. Si venía en la cookie, también se expira. |
| `FORBIDDEN` | 403 | `Origin` no permitido (protección CSRF). |
| `INTERNAL_ERROR` | 500 | Fallo al persistir el cierre de la sesión. |

---

## 3) Usuario autenticado — `GET /api/v1/auth/me`

Devuelve el usuario correspondiente al token actual. Autentica con la cookie igual que el resto de rutas protegidas, por lo que el frontend puede recuperar el usuario sin guardar el token.

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

## CORS y protección CSRF

- **CORS:** orígenes explícitos en `cors.allow_origins` y `cors.allow_credentials: true`. El servidor no arranca si se configura `"*"` junto con credenciales. El frontend debe hacer las peticiones con `withCredentials: true`.
- **CSRF:** como el navegador adjunta la cookie automáticamente, el middleware `OriginGuard` revisa toda petición `POST`, `PUT`, `PATCH` y `DELETE` (login incluido):
  - Si trae `Origin`, debe estar en `cors.allow_origins`; si no, `403 FORBIDDEN` (`origin not allowed`). `Origin: null` también se rechaza.
  - Si no trae `Origin` pero sí `Referer`, se valida el origen del `Referer` igual.
  - Si no trae ninguno de los dos, se acepta: los navegadores modernos siempre envían `Origin` en estos métodos, así que la petición viene de un cliente no navegador (Bruno, curl), que no puede ser usado por una página maliciosa.
  - `GET`, `HEAD` y `OPTIONS` no se bloquean (no deben cambiar estado).
- `SameSite=Lax` añade una segunda barrera: la cookie no viaja en peticiones `POST` iniciadas desde otros sites.

---

## Control de acceso por rol (RBAC)

El token transporta `user_id` y `session_id`; los permisos se resuelven por petición consultando `role_permission` (no se cachean en el token). El middleware `RequirePermission(module, operation)` se antepone a cada ruta protegida.

Roles semilla:

| Rol | Permisos |
|---|---|
| `superadmin` | Todos los permisos implementados. |
| `business_admin` | Solo lectura de usuarios (`users.list`, `users.read`). |

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
  "role": "superadmin | business_admin | employee",
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
| `FORBIDDEN` | 403 | El rol del usuario no tiene el permiso requerido, o `Origin` no permitido en un método que cambia estado. |
| `NOT_FOUND` | 404 | Usuario no encontrado en `me`. |
| `INTERNAL_ERROR` | 500 | Error de persistencia o caso de uso. |

---

## Casos borde / comportamiento no obvio

- El login responde el mismo mensaje para `user_not_found`, `invalid_password` y `account_not_active`; la diferencia solo queda en `login_attempt.failure_reason`.
- Un token firmado pero con sesión revocada/expirada es rechazado por el middleware (la sesión se valida contra `user_session` en cada petición).
- `company_id` es opcional: si el usuario no tiene empresa asignada, el campo se omite en la respuesta y el token no lo transporta.
- El access token expira según `SIGIF_JWT_ACCESS_TOKEN_EXPIRY` (minutos) y la cookie caduca a la vez; no existe refresh token en esta HU, así que al expirar el usuario vuelve a iniciar sesión.
- Si llegan cookie y `Authorization` a la vez, gana `Authorization`.

---

## Diferencias respecto a una versión anterior (si aplica)

- **Autenticación por cookie:** el login ya no devuelve `data.token` (`access_token`, `expires_in`, `token_type`); ahora responde `data.user` y `data.expires_in` y entrega el JWT en la cookie HttpOnly `sigif_access_token`. Los consumidores que leían `data.token.access_token` deben leer la cookie (Bruno lo hace desde `Set-Cookie`) o, en el navegador, simplemente enviar las peticiones con credenciales.
- `Authorization: Bearer` se mantiene como respaldo, con prioridad sobre la cookie cuando se envía.
- Se añadió la validación de `Origin` (CSRF) en métodos que cambian estado.

- Se eliminó el flujo de `refresh token` y `logout-all`; ahora hay un único access token por sesión persistida en `user_session`.
- El login ya no requiere `X-Tenant-ID`; la identidad es global (email + contraseña) y `company_id` se resuelve desde el usuario autenticado.
