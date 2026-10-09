# Reporte de endpoints — HU-082-06 Controlar operaciones autorizadas por rol

## Información general

- HU: HU-082-06 — Controlar operaciones autorizadas por rol (RF-82, Seguridad / Roles y Permisos)
- Rama: `feature/HU-082-06-control-operations-authorized-by-role` (parte de `feat/HU-082-04-role-status-management`, que incluye HU-082-01 y HU-082-02)
- Prefijo base: `/api/v1`
- Requiere autenticación: `Authorization: Bearer <access_token>`. Esta HU no cambia el login ni la validación del token; usa la identidad que ya entrega el middleware `AuthRequired`.
- Formato de respuesta exitosa: `{ "success": true, "data": { ... } }`
- Formato de respuesta de error: `{ "success": false, "error": { "code": "CODIGO", "message": "descripción" } }`
- Colección Bruno: `bruno/Authorization/` (001–016)

---

## Resumen rápido

| Método | Ruta | Permiso | Descripción |
|---|---|---|---|
| `GET` | `/api/v1/auth/me/permissions` | — (solo sesión) | Códigos de permiso del usuario actual. **Nuevo.** |

El resto de la HU no es un endpoint: es el guard de autorización que se ejecuta antes de cada operación protegida.

---

## 1) Permisos del usuario actual — `GET /api/v1/auth/me/permissions`

Devuelve los códigos de permiso efectivos del usuario autenticado. El frontend lo usa para ocultar o deshabilitar acciones. Es solo una ayuda visual: la seguridad real la aplica el guard en cada operación.

### Permiso
Ninguno. Solo requiere sesión activa.

### Respuesta exitosa

**`200`**
```json
{
  "success": true,
  "data": {
    "role_id": "dcd04646-ef9e-416d-89cd-6929e6ac0881",
    "role": "business_admin",
    "permissions": ["permissions.read", "users.list", "users.read"]
  }
}
```

| Campo | Tipo | Descripción |
|---|---|---|
| `role_id` | uuid | Rol del usuario. |
| `role` | string | Nombre del rol. |
| `permissions` | string[] | Códigos con formato `modulo.operacion`, ordenados por módulo y operación. Siempre es un arreglo; `[]` si no tiene ninguno. |

Devuelve `"permissions": []` cuando:
- el rol no tiene permisos asignados (por ejemplo `employee`);
- el rol está desactivado (HU-082-04);
- el usuario está inactivo.

Son los mismos casos en los que el guard responde 403, así que lo que el frontend muestra coincide con lo que el backend permite.

`GET /api/v1/auth/me` y `POST /api/v1/auth/login` siguen devolviendo `permissions` dentro del usuario. Este endpoint sirve para refrescarlos sin pedir el perfil completo, por ejemplo después de que un administrador cambie los permisos del rol.

### Códigos de error

| Code | HTTP | Causa |
|---|---|---|
| `UNAUTHORIZED` | 401 | Falta el token, es inválido, expiró, su sesión se cerró o el usuario ya no existe. |
| `INTERNAL_ERROR` | 500 | No se pudieron consultar los permisos. |

---

## 2) Guard de autorización

No es un endpoint. Cada ruta protegida declara el código de permiso que exige, y el guard decide antes de ejecutar la operación.

### Orden de validación

| Paso | Comprobación | Si falla |
|---|---|---|
| 1 | Autenticación: token válido y sesión activa (`AuthRequired`) | `401 UNAUTHORIZED` |
| 2 | El usuario existe y está activo | `403 FORBIDDEN` |
| 3 | Su rol existe y está activo | `403 FORBIDDEN` |
| 4 | El rol tiene asignado el código de permiso de la ruta | `403 FORBIDDEN` |
| 5 | — | Se ejecuta la operación |

- **El permiso no sustituye la autenticación.** Sin token o con token inválido la respuesta es siempre 401, nunca 403, y los permisos ni se consultan.
- **Una operación rechazada no se ejecuta.** El guard corta la petición antes del handler: no se valida el body, no se lee ni se escribe nada.
- **Los permisos se leen de la base en cada petición.** Quitar un permiso, desactivar un rol o desactivar un usuario tiene efecto inmediato, sin esperar a que expire el token.
- **Si los permisos no se pueden consultar, la operación no se ejecuta** y responde 500.
- **Cada rechazo queda en el log del servidor** como `permission denied`, con `user_id`, `company_id`, los permisos exigidos, método, ruta, IP y `request_id`. No se guarda en una tabla.

### Formato único del 403

Todas las rutas protegidas responden lo mismo:

**`403`**
```json
{
  "success": false,
  "error": {
    "code": "FORBIDDEN",
    "message": "you do not have permission to perform this action"
  }
}
```

El frontend debe decidir por el estado HTTP `403` y el código `FORBIDDEN`, y mostrar su propio texto ("No tienes permiso para esta acción").

### Diferencias con la propuesta del frontend

| Propuesta | Implementado | Motivo |
|---|---|---|
| `{ "status": 403, "code": "PERMISO_DENEGADO", "message": "..." }` | `{ "success": false, "error": { "code": "FORBIDDEN", "message": "..." } }` | Es el formato de error estándar del backend, igual que 400, 401, 404 y 409. Un formato distinto solo para el 403 obligaría al frontend a tratar dos estructuras. |
| `GET /api/auth/me/permisos` | `GET /api/v1/auth/me/permissions` | Prefijo `/api/v1` y rutas en inglés, como el resto de la API. |
| `GET /api/permisos` (catálogo, opcional) | Ya existe: `GET /api/v1/permissions` | Lo entrega HU-082-01. |

El mensaje del 403 cambió de `forbidden` a `you do not have permission to perform this action`. Los documentos de otras HU todavía muestran el mensaje anterior; el código `FORBIDDEN` no cambió.

---

## Rutas y permiso que exige cada una

### Públicas (sin token)

| Ruta |
|---|
| `GET /health` |
| `POST /api/v1/auth/login` |

### Solo sesión (token válido, sin permiso)

| Ruta |
|---|
| `POST /api/v1/auth/logout` |
| `GET /api/v1/auth/me` |
| `GET /api/v1/auth/me/permissions` |

### Protegidas por permiso

| Ruta | Código de permiso |
|---|---|
| `POST /users` | `users.create` |
| `GET /users` | `users.list` |
| `GET /users/by-email`, `GET /users/:id` | `users.read` |
| `PATCH /users/:id` | `users.update` |
| `PUT /users/:id/password` | `users.change_password` |
| `DELETE /users/:id` | `users.delete` |
| `POST /users/:id/activate` | `users.activate` |
| `POST /users/:id/deactivate` | `users.deactivate` |
| `GET /companies` | `users.create` o `users.update` |
| `GET /roles` | `roles.list` |
| `PATCH /roles/:id/status` | `roles.status` |
| `GET /permissions`, `GET /permissions/modules` | `permissions.read` |
| `GET /permissions/export` | `permissions.export` |
| `POST /permissions` | `permissions.create` |
| `PATCH /permissions/:id` | `permissions.update` |
| `DELETE /permissions/:id` | `permissions.delete` |
| `POST /customers` | `customers.create` |
| `GET /customers` | `customers.list` |
| `GET /customers/:id` | `customers.read` |
| `PUT /customers/:id`, `PATCH /customers/:id` | `customers.update` |
| `PATCH /customers/:id/status` | `customers.status` |
| `DELETE /customers/:id` | `customers.delete` |
| `POST /categories` | `categories.create` |
| `GET /products`, `GET /products/:id` | `products.read` |
| `POST /products` | `products.create` |
| `PUT /products/:id`, `PATCH /products/:id/status` | `products.update` |
| `DELETE /products/:id` | `products.delete` |

Un test automático (`cmd/server/authorization_test.go`) arranca la aplicación con todos sus módulos y recorre **todas** las rutas registradas. Falla si alguna ruta que no esté en las dos primeras listas responde algo distinto de 401 sin token, o algo distinto de 403 para un usuario sin permisos. Así, una ruta nueva sin guard no pasa los tests.

`GET /units-of-measure` y `GET /taxes` (HU-03-01, aún sin mergear) están previstas en ese test como rutas de solo sesión.

---

## Qué cambió en el código

| Cambio | Detalle |
|---|---|
| Guard por código de permiso | `middleware.RequireAnyPermission(checker, "users.create", "users.update")`. Acepta uno o varios códigos y deja pasar si el rol tiene cualquiera. Un código mal escrito detiene el arranque del servidor. |
| `RequirePermission(checker, módulo, operación)` | Se mantiene para no tocar las rutas existentes; ahora usa el guard anterior por dentro. |
| `GET /companies` | Tenía su propia copia del guard. Ahora usa el compartido, con el mismo comportamiento. |
| Mensaje del 403 | Unificado para todas las rutas. |
| `GET /auth/me` | Devuelve `permissions: []` para un usuario inactivo, igual que el endpoint nuevo. |

No hay cambios en la base de datos.

## Criterios de aceptación

| # | Criterio | Cómo se cumple |
|---|---|---|
| 1 | Con el permiso requerido, la operación continúa | El guard llama al handler solo si el rol tiene el código. |
| 2 | Sin el permiso, el acceso es rechazado | El guard responde antes del handler. |
| 3 | Una operación rechazada no se ejecuta | Los tests cuentan las ejecuciones del handler; en Bruno, el usuario de la petición rechazada no existe después. |
| 4 | Sin autorización responde 403 | Formato único `FORBIDDEN`. |
| 5 | Validar el permiso no sustituye la autenticación | Sin identidad responde 401 y no se consultan permisos. |

## Pruebas

- Tests de Go: `go test ./cmd/server/... ./internal/shared/middleware/... ./internal/modules/auth/...`
- Bruno: carpeta `bruno/Authorization` (16 peticiones). Es autónoma: inicia sesión con `companyAdminEmail` / `companyAdminPassword` y crea un usuario `employee` (sin permisos) y uno `business_admin` (solo lectura).

## Notas para el frontend

1. Pedir `GET /auth/me/permissions` al iniciar la aplicación y después de un 403 inesperado, por si los permisos cambiaron.
2. Ocultar o deshabilitar acciones con esa lista es solo comodidad para el usuario. Cualquier petición puede responder 403 y hay que manejarlo.
3. Un 401 significa "vuelve a iniciar sesión"; un 403 significa "tu sesión es válida, pero no puedes hacer esto". No redirigir al login en un 403.
