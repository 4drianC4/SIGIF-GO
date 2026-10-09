# Reporte de endpoints — HU-02-03 Visualizar y preservar historial del usuario

## Información general

- HU: HU-02-03 — Visualizar y preservar historial del usuario
- Rama: `feature/HU-02-03-view-and-preserve-user-history`
- Prefijo base: `/api/v1`
- Requiere contexto autenticado del backend: `userId`, tomado del access token (`Authorization: Bearer <token>`). No se envía por header ni por body.
- Formato de respuesta exitosa: `{ "success": true, "data": [...], "meta": { "page", "limit", "total", "total_pages" } }`
- Formato de respuesta de error: `{ "success": false, "error": { "code": "CODIGO", "message": "descripción" } }`
- Fechas: RFC 3339 en UTC (`2026-10-08T18:20:00Z`)

---

## Resumen rápido

| Método | Ruta | Permiso | Descripción |
|---|---|---|---|
| `GET` | `/api/v1/users/:id/history` | `users.read` | Lista el historial de un usuario, del evento más reciente al más antiguo. |

Es el único endpoint nuevo. El frontend no crea registros de historial: el backend los genera al ejecutar las operaciones que ya existían (ver [Registro automático](#registro-automático-del-historial)).

---

## Autenticación / permisos

| Header | Obligatorio | Descripción |
|---|---|---|
| `Authorization` | Sí | `Bearer <access_token>` obtenido con `POST /api/v1/auth/login`. La sesión debe seguir activa. |

- `userId`: claim `user_id` del token. Se usa para verificar el permiso y queda guardado como autor (`performed_by`) de cada evento.

Permiso usado: `users.read`. Se reutiliza el permiso existente; no se creó `users.history`. Lo tienen `superadmin` y `business_admin`; `employee` no (403).

---

## 1) Historial de un usuario — `GET /api/v1/users/:id/history`

Devuelve los eventos registrados para el usuario, ordenados del más reciente al más antiguo.

### Permiso
`users.read`

### Path params
- `id` (obligatorio): uuid — usuario cuyo historial se consulta.

### Query params
- `page` (opcional, por defecto `1`): number — página.
- `limit` (opcional, por defecto `20`, máximo `100`): number — eventos por página. Un valor fuera de rango usa el valor por defecto.

### Respuesta exitosa

**`200`**
```json
{
  "success": true,
  "data": [
    {
      "id": "ad15c1fd-b5f2-4fa2-a6c4-eff52738c32d",
      "user_id": "de8cb7e2-692c-424a-880f-dc80eb48191b",
      "action": "updated",
      "description": "User updated",
      "changes": {
        "first_name": { "from": "Juan", "to": "Juan Carlos" },
        "email": { "from": "juan@sigif.com", "to": "juan.carlos@sigif.com" }
      },
      "performed_by": {
        "id": "f9f82645-bbac-40d0-8671-698d6ecc335a",
        "full_name": "Admin Empresa",
        "email": "admin.empresa@sigif.com"
      },
      "created_at": "2026-10-08T18:20:00Z"
    },
    {
      "id": "1e6ff5b3-e22c-4289-8e2b-20ba9c187957",
      "user_id": "de8cb7e2-692c-424a-880f-dc80eb48191b",
      "action": "created",
      "description": "User created",
      "changes": {},
      "performed_by": {
        "id": "f9f82645-bbac-40d0-8671-698d6ecc335a",
        "full_name": "Admin Empresa",
        "email": "admin.empresa@sigif.com"
      },
      "created_at": "2026-10-08T18:00:00Z"
    }
  ],
  "meta": {
    "page": 1,
    "limit": 20,
    "total": 2,
    "total_pages": 1
  }
}
```

Un usuario sin eventos responde `200` con `"data": []` y `total: 0`.

### Campos de cada evento

| Campo | Tipo | Descripción |
|---|---|---|
| `id` | uuid | Identificador del evento. |
| `user_id` | uuid | Usuario al que pertenece el evento. |
| `action` | enum | Acción realizada (ver tabla siguiente). |
| `description` | string | Resumen legible, en inglés. Para mostrar texto en español, traducir según `action`. |
| `changes` | objeto | Campos modificados, cada uno con `from` y `to`. Siempre es un objeto; está vacío (`{}`) cuando la acción no modifica campos. |
| `performed_by` | objeto o `null` | Quién ejecutó la operación: `id`, `full_name` y `email` tal como eran en ese momento. `null` si no hubo un usuario autenticado. |
| `created_at` | fecha | Momento del evento, en UTC. |

### Valores de `action`

| `action` | `description` | Se registra al ejecutar | `changes` |
|---|---|---|---|
| `created` | `User created` | `POST /users` | `{}` |
| `updated` | `User updated` | `PATCH /users/:id` cuando cambia algún dato | Campos que cambiaron |
| `password_changed` | `Password changed` | `PATCH /users/:id` con `password`, o `PUT /users/:id/password` | `{}` |
| `deactivated` | `User deactivated` | `POST /users/:id/deactivate` | `{}` |
| `activated` | `User activated` | `POST /users/:id/activate` | `{}` |
| `deleted` | `User deleted` | `DELETE /users/:id` | `{}` |

Campos que pueden aparecer en `changes`: `first_name`, `last_name`, `email`, `company_id` y `role` (nombre del rol). `from` y `to` son strings, o `null` cuando el dato no existía.

### Códigos de error

| Code | HTTP | Causa |
|---|---|---|
| `BAD_REQUEST` | 400 | `id` no es un UUID válido. |
| `UNAUTHORIZED` | 401 | Falta el token, es inválido, expiró o su sesión se cerró. |
| `FORBIDDEN` | 403 | Usuario sin `users.read`. |
| `NOT_FOUND` | 404 | El usuario no existe. |
| `INTERNAL_ERROR` | 500 | Error de persistencia. |

---

## Registro automático del historial

Los endpoints existentes de usuarios no cambiaron su contrato (rutas, bodies, respuestas y códigos son los mismos). Lo único nuevo es que cada operación exitosa guarda un evento:

- **La operación y su evento se guardan en la misma transacción.** Si el evento no se puede guardar, la operación se deshace y responde 500; si la operación falla (400, 404, 409...), no se guarda ningún evento.
- **Un `PATCH` que no cambia ningún dato no genera evento** (por ejemplo `{}` o enviar los mismos valores).
- **Un `PATCH` que cambia datos y además envía `password`** genera dos eventos: `updated` y `password_changed`.
- **Las contraseñas nunca se guardan en el historial**, ni la anterior ni la nueva, ni su hash. Solo queda el evento `password_changed`.
- **El autor se guarda como copia** (`id`, nombre y email del momento). Si después se edita o desactiva a ese administrador, los eventos antiguos no cambian.

## Preservación del historial

- La tabla `user_history` solo recibe inserciones: no hay endpoint ni código que edite o borre eventos.
- No tiene borrado lógico ni clave foránea en cascada hacia `app_user`: editar, desactivar o eliminar al usuario no toca sus eventos.
- El historial sigue disponible después de editar al usuario, cambiar su email, su compañía o su estado.

## Base de datos

Tabla nueva `user_history`, creada por `make migrate-up` (también en bases ya existentes; no modifica tablas previas):

| Columna | Tipo | Notas |
|---|---|---|
| `id` | uuid | PK |
| `user_id` | uuid | Índice `(user_id, created_at DESC)` |
| `action` | varchar(40) | |
| `description` | varchar(255) | |
| `changes` | jsonb | Por defecto `{}` |
| `performed_by_id` | uuid | Nulo si no hubo autor |
| `performed_by_name` | varchar(161) | |
| `performed_by_email` | varchar(160) | |
| `created_at` | timestamptz | |

## Pruebas

- Tests de Go: `go test ./internal/modules/user/...`
- Tests de integración con Postgres: `SIGIF_TEST_DATABASE_DSN='host=localhost port=5432 user=sigif password=sigif dbname=sigif sslmode=disable' go test -tags=integration ./internal/modules/user/infrastructure/...` (usan un esquema temporal; no tocan las tablas de `public`).
- Bruno: carpeta `bruno/Users/History` (15 peticiones). Es autónoma: inicia sesión con `companyAdminEmail` / `companyAdminPassword` y crea su propio usuario.

## Notas para el frontend

1. Los usuarios creados antes de esta HU no tienen evento `created`: su historial empieza en la primera operación posterior.
2. `description` llega en inglés; el texto en pantalla conviene decidirlo según `action`.
3. La desactivación (`POST /users/:id/deactivate`) ya queda registrada como `deactivated`. Su respuesta y su comportamiento de borrado lógico se definen en HU-02-04.
