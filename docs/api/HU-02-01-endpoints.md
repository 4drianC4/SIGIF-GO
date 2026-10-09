# HU-02-01 — Registrar usuario

## POST /api/v1/users

Requiere `Authorization: Bearer <access_token>`, sesión activa y permiso `users.create`. Body JSON:

```json
{
  "first_name": "Juan",
  "last_name": "Pérez",
  "email": "juan@sigif.com",
  "company_id": "11111111-1111-4111-8111-111111111111",
  "role": "employee"
}
```

| Campo | Requerido | Validación |
|---|---|---|
| first_name | Sí | 1–80 caracteres. |
| last_name | Sí | 1–80 caracteres. |
| email | Sí | Email válido y único. |
| company_id | Sí | UUID no nulo de una compañía existente y no eliminada. |
| role | Sí | `business_admin` o `employee`; el rol debe existir. |

El servidor genera una contraseña aleatoria criptográficamente segura (24 bytes, codificados como 32 caracteres base64url) y almacena únicamente el hash Argon2id. El request ya no acepta una contraseña inicial; si un cliente antiguo envía `password`, se ignora igual que otros campos fuera del DTO. `area` y `full_name` también se ignoran y no forman parte del contrato.

Respuesta **201**, con `Cache-Control: no-store`:

```json
{
  "success": true,
  "data": {
    "id": "uuid",
    "company_id": "11111111-1111-4111-8111-111111111111",
    "role_id": "uuid",
    "role": "employee",
    "first_name": "Juan",
    "last_name": "Pérez",
    "username": "juan@sigif.com",
    "email": "juan@sigif.com",
    "status": "active",
    "created_at": "2026-10-08T12:00:00Z",
    "updated_at": "2026-10-08T12:00:00Z",
    "generated_password": "contraseña-generada-por-el-servidor"
  }
}
```

`generated_password` se devuelve **solo al crear**, para que el administrador la entregue al usuario. No se guarda en texto plano, no aparece en GET/PATCH/login y no se envía por correo. No hay expiración ni cambio obligatorio de contraseña en este flujo.

## GET /api/v1/companies — Opciones del selector

Requiere sesión activa y al menos uno de los permisos `users.create` o `users.update`. No requiere parámetros ni body. Devuelve **todas** las compañías no eliminadas, incluidas las inactivas, sin paginación, ordenadas por razón social y luego ID. No limita por la compañía del administrador, pues este formulario administra usuarios globalmente.

Respuesta **200**:

```json
{
  "success": true,
  "data": [
    {
      "id": "11111111-1111-4111-8111-111111111111",
      "legal_name": "Empresa Ejemplo SRL",
      "trade_name": "Ejemplo"
    }
  ]
}
```

Si no hay compañías, `data` es `[]`. Errores: 401 sin autenticación, 403 sin permisos y 500 por persistencia.

En el frontend, usar `trade_name || legal_name` como etiqueta y `id` como valor de cada opción. Al seleccionar, enviar ese valor en `company_id` del POST o PATCH. El backend valida su existencia nuevamente al guardar.

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

## Compatibilidad y persistencia

Este cambio requiere actualizar los clientes que enviaban contraseña o creaban usuarios sin compañía. Los usuarios anteriores sin compañía siguen siendo legibles/editables. Se eliminó `area` del modelo y del contrato; GORM AutoMigrate no elimina la columna histórica en bases existentes. No se ejecuta un DROP destructivo. `full_name` era un valor calculado, sin columna propia.
