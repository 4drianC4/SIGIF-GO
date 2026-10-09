# HU-02-02 — Editar usuario

## PATCH /api/v1/users/:id

Requiere `Authorization: Bearer <access_token>`, sesión activa y permiso `users.update`. `id` debe ser un UUID de usuario existente.

```json
{
  "first_name": "Juan Carlos",
  "last_name": "Pérez Gómez",
  "email": "nuevo@sigif.com",
  "company_id": "11111111-1111-4111-8111-111111111111",
  "role": "business_admin",
  "password": "nuevaPassword123"
}
```

Todos los campos son opcionales. Solo se modifican los enviados; los omitidos o `null` conservan su valor. `{}` es válido. No se puede desvincular una compañía mediante `null`.

| Campo | Validación cuando se envía |
|---|---|
| first_name | 1–80 caracteres. |
| last_name | 1–80 caracteres. |
| email | Email válido y único; actualiza también `username`. |
| company_id | UUID no nulo de una compañía existente y no eliminada. |
| role | `business_admin` o `employee`; no permite asignar `superadmin`. |
| password | Mínimo 8 caracteres; no acepta cadena vacía. Se almacena hasheada con Argon2id. |

Respuesta **200**: `{ "success": true, "data": { ...User } }`, con los valores actualizados, incluido el nombre del rol y `company_id`. No devuelve contraseña ni `generated_password`.

Para cargar la empresa actual, seleccionar la opción cuyo `id` coincida con el `company_id` del usuario. Las opciones se obtienen con [GET /api/v1/companies](HU-02-01-endpoints.md#get-apiv1companies--opciones-del-selector).

`area` y `full_name` ya no forman parte del contrato; si se envían, se ignoran. `phone` se conserva, pero no se edita mediante este formulario. El estado se cambia por los endpoints de activación/desactivación.

El cambio administrativo de contraseña no exige la anterior y no invalida sesiones existentes automáticamente. El cambio de compañía tampoco renueva tokens emitidos: el usuario debe volver a iniciar sesión para obtener el nuevo contexto de compañía.

## Respuesta User

Incluye `id`, `company_id`, `role_id`, `role`, `first_name`, `last_name`, `username`, `email`, `phone`, `status`, `last_access`, `created_at` y `updated_at`. Los valores opcionales vacíos (`company_id`, `phone`, `last_access`) se omiten. `username` sigue sincronizado con el email; `phone` se conserva pero no se edita en estos formularios. Ya no se exponen `area` ni `full_name`. Nunca se devuelve `password_hash`.

## Errores

| HTTP | Código | Causa |
|---|---|---|
| 400 | BAD_REQUEST | JSON ilegible o UUID mal formado. |
| 400 | VALIDATION_ERROR | Campos inválidos, rol no permitido o compañía inexistente/eliminada. |
| 401 | UNAUTHORIZED | Token ausente, inválido o sesión terminada. |
| 403 | FORBIDDEN | Sin permiso para la operación. |
| 409 | CONFLICT | Email ya registrado. |
| 500 | INTERNAL_ERROR | Error de persistencia. |

Formato: `{ "success": false, "error": { "code": "...", "message": "..." } }`. Los errores del validador pueden incluir `details`.

También responde **404 NOT_FOUND** si el usuario no existe.
