# Reporte de endpoints — [HU-08-01] Registro de clientes

## Información general

- HU: HU-08-01 — Registro de clientes
- Rama: `feat/HU-08-01-customer-registration`
- Prefijo base: `/api/v1`
- Requiere autenticación: token JWT válido en el esquema de autorización de portador (`Authorization: Bearer <access_token>`). El middleware global `AuthRequired` valida el token y que la sesión siga activa (tras `logout` el token deja de servir).
- Empresa: los clientes pertenecen a la empresa del usuario, que se toma del claim `company_id` del token. Un usuario sin empresa (por ejemplo, un superadmin global) recibe `400 company is required`.
- Formato de respuesta exitosa: `{ "success": true, "data": ... }`. Las respuestas paginadas incluyen `meta`.
- Formato de respuesta de error: `{ "success": false, "error": { "code": "CODIGO", "message": "descripción", "details": { "campo": "motivo" } } }`. `details` solo aparece en errores de validación.

---

## Resumen rápido

| Método | Ruta | Permiso | Descripción |
|---|---|---|---|
| `POST` | `/api/v1/customers/` | `customers.create` | Registra un cliente en la empresa autenticada. |
| `GET` | `/api/v1/customers/` | `customers.list` | Lista clientes de la empresa con filtros y paginación. |
| `GET` | `/api/v1/customers/:id` | `customers.read` | Obtiene un cliente por UUID. |
| `PUT` | `/api/v1/customers/:id` | `customers.update` | Actualiza los datos editables del cliente. |
| `PATCH` | `/api/v1/customers/:id` | `customers.update` | Actualización parcial (ver HU-08-04). |
| `PATCH` | `/api/v1/customers/:id/status` | `customers.status` | Cambia el estado del cliente. |
| `DELETE` | `/api/v1/customers/:id` | `customers.delete` | Da de baja lógicamente al cliente. |

---

## Autenticación / permisos (todas las rutas)

| Header | Obligatorio | Descripción |
|---|---|---|
| `Authorization` | Sí | `Bearer <access_token>` de una sesión activa. |
| `Content-Type` | Sí (en rutas con body) | `application/json`. |

Los permisos se comprueban con el RBAC de la base de datos (`RequirePermission`):

| Permiso | Endpoint | Roles sembrados que lo tienen |
|---|---|---|
| `customers.create` | `POST /api/v1/customers` | `superadmin` |
| `customers.list` | `GET /api/v1/customers` | `superadmin` |
| `customers.read` | `GET /api/v1/customers/:id` | `superadmin` |
| `customers.update` | `PUT /api/v1/customers/:id` | `superadmin` |
| `customers.status` | `PATCH /api/v1/customers/:id/status` | `superadmin` |
| `customers.delete` | `DELETE /api/v1/customers/:id` | `superadmin` |

Estos permisos se añadieron al catálogo de `AllPermissions()`; `make migrate-up` los crea y se los asigna al superadmin también en bases ya existentes. El rol `soporte` no los tiene y recibe **403**. Para darlos a otro rol, asignarlos en `role_permission`.

---

## 1) Registrar cliente — `POST /api/v1/customers/`

Registra un cliente en la empresa autenticada. El UUID del cliente lo genera el backend.

### Permiso
`customers.create`

### Params de ruta
Ninguno.

### Query params
Ninguno.

### Body

```json
{
  "legal_name": "Comercial Ejemplo",
  "document_type": "tax_id",
  "document_number": "1020304050",
  "phone": "+59170000000",
  "email": "contacto@ejemplo.com"
}
```

| Campo | Tipo | Obligatorio | Restricciones / notas |
|---|---|---|---|
| `legal_name` | string | Sí | Entre 1 y 160 caracteres; se recortan espacios en los extremos. |
| `document_type` | enum | No | `national_id` \| `tax_id` \| `passport` \| `other`. Si se omite, usa `national_id`. |
| `document_number` | string | No | Máximo 30 caracteres. Si se informa, debe ser único por empresa y tipo de documento entre clientes no eliminados. |
| `phone` | string | No | Máximo 30 caracteres. |
| `email` | string | No | Formato email, máximo 160 caracteres. |

### Respuesta exitosa

**`201 Created`**

```json
{
  "success": true,
  "data": {
    "id": "f7e8862e-b6d2-4a08-a3cf-0d2f57f158e2",
    "company_id": "aaaaaaaa-aaaa-4aaa-8aaa-aaaaaaaaaaaa",
    "legal_name": "Comercial Ejemplo",
    "document_type": "tax_id",
    "document_number": "1020304050",
    "phone": "+59170000000",
    "email": "contacto@ejemplo.com",
    "credit_limit": "0",
    "credit_balance": "0",
    "points_accrued": 0,
    "status": "active",
    "created_at": "2026-10-03T12:00:00Z"
  }
}
```

### Códigos de error

| Code | HTTP | Causa |
|---|---|---|
| `BAD_REQUEST` | 400 | JSON inválido o el token no tiene empresa (`company is required`). |
| `VALIDATION_ERROR` | 400 | Nombre, tipo de documento, email u otro campo no cumple validación. |
| `UNAUTHORIZED` | 401 | Falta el token o es inválido/expiró. |
| `FORBIDDEN` | 403 | El usuario no tiene el permiso `customers.create`. |
| `CONFLICT` | 409 | Ya existe en la empresa un cliente no eliminado con el mismo tipo y número de documento. |
| `INTERNAL_ERROR` | 500 | Error inesperado o de persistencia. |

---

## 2) Listar clientes — `GET /api/v1/customers/`

Lista únicamente clientes no eliminados de la empresa autenticada.

> El contrato vigente de este endpoint (búsqueda, filtros, orden, paginación, validación y forma de los ítems) se define en [HU-08-02 — Búsqueda de clientes](HU-08-02-endpoints.md).

### Permiso
`customers.list`

---

## 3) Obtener cliente — `GET /api/v1/customers/:id`

Obtiene un cliente de la empresa autenticada por su UUID.

### Permiso
`customers.read`

### Params de ruta
- `id`: UUID — identificador del cliente.

### Query params
Ninguno.

### Respuesta exitosa

**`200 OK`**

```json
{
  "success": true,
  "data": {
    "id": "f7e8862e-b6d2-4a08-a3cf-0d2f57f158e2",
    "company_id": "aaaaaaaa-aaaa-4aaa-8aaa-aaaaaaaaaaaa",
    "legal_name": "Comercial Ejemplo",
    "document_type": "tax_id",
    "document_number": "1020304050",
    "phone": "+59170000000",
    "email": "contacto@ejemplo.com",
    "address": "Av. Ejemplo 123",
    "credit_limit": "0",
    "credit_balance": "0",
    "points_accrued": 0,
    "status": "active",
    "created_at": "2026-10-03T12:00:00Z"
  }
}
```

### Códigos de error

| Code | HTTP | Causa |
|---|---|---|
| `BAD_REQUEST` | 400 | `id` no es un UUID válido, o el token no tiene empresa. |
| `UNAUTHORIZED` | 401 | Falta el token o es inválido/expiró. |
| `FORBIDDEN` | 403 | El usuario no tiene el permiso `customers.read`. |
| `NOT_FOUND` | 404 | El cliente no existe, está eliminado o pertenece a otra empresa. |
| `INTERNAL_ERROR` | 500 | Error inesperado de persistencia. |

---

## 4) Actualizar cliente — `PUT /api/v1/customers/:id`

Actualiza los campos editables del cliente. La operación trata los campos omitidos como vacíos/nulos; enviar los datos que se desean conservar.

> Para cambiar solo algunos campos, usar `PATCH /api/v1/customers/:id` ([HU-08-04 — Edición de cliente](HU-08-04-endpoints.md)).

### Permiso
`customers.update`

### Params de ruta
- `id`: UUID — identificador del cliente.

### Query params
Ninguno.

### Body

```json
{
  "legal_name": "Comercial Ejemplo Actualizado",
  "document_type": "tax_id",
  "document_number": "1020304050",
  "phone": "+59170000000",
  "email": "contacto@ejemplo.com",
  "address": "Av. Ejemplo 123"
}
```

| Campo | Tipo | Obligatorio | Restricciones / notas |
|---|---|---|---|
| `legal_name` | string | Sí | Entre 1 y 160 caracteres. |
| `document_type` | enum | No | `national_id` \| `tax_id` \| `passport` \| `other`. Si se omite, se conserva el tipo actual del cliente. |
| `document_number` | string | No | Máximo 30 caracteres. Se comprueba unicidad si cambia. |
| `phone` | string | No | Máximo 30 caracteres. |
| `email` | string | No | Formato email, máximo 160 caracteres. |
| `address` | string | No | Máximo 200 caracteres. |

### Respuesta exitosa

**`200 OK`** — devuelve el cliente actualizado con la misma estructura de `GET /api/v1/customers/:id`.

### Códigos de error

| Code | HTTP | Causa |
|---|---|---|
| `BAD_REQUEST` | 400 | UUID, JSON o validación inválidos, o falta empresa. |
| `UNAUTHORIZED` | 401 | Falta el token o es inválido/expiró. |
| `FORBIDDEN` | 403 | El usuario no tiene el permiso `customers.update`. |
| `NOT_FOUND` | 404 | El cliente no existe o no pertenece a la empresa. |
| `CONFLICT` | 409 | El documento ya está asociado a otro cliente de la empresa. |
| `INTERNAL_ERROR` | 500 | Error inesperado o de persistencia. |

---

## 5) Cambiar estado — `PATCH /api/v1/customers/:id/status`

Cambia el estado operativo del cliente.

### Permiso
`customers.status`

### Params de ruta
- `id`: UUID — identificador del cliente.

### Query params
Ninguno.

### Body

```json
{
  "status": "blocked"
}
```

| Campo | Tipo | Obligatorio | Restricciones / notas |
|---|---|---|---|
| `status` | enum | Sí | `active` \| `inactive` \| `blocked`. |

### Respuesta exitosa

**`200 OK`** — devuelve el cliente actualizado con la misma estructura de `GET /api/v1/customers/:id`.

### Códigos de error

| Code | HTTP | Causa |
|---|---|---|
| `BAD_REQUEST` | 400 | UUID o body inválido, o falta empresa. |
| `VALIDATION_ERROR` | 400 | Estado fuera del enum permitido. |
| `UNAUTHORIZED` | 401 | Falta el token o es inválido/expiró. |
| `FORBIDDEN` | 403 | El usuario no tiene el permiso `customers.status`. |
| `NOT_FOUND` | 404 | El cliente no existe o no pertenece a la empresa. |
| `INTERNAL_ERROR` | 500 | Error inesperado o de persistencia. |

---

## 6) Eliminar cliente — `DELETE /api/v1/customers/:id`

Marca al cliente como eliminado (soft delete); no elimina físicamente el registro.

### Permiso
`customers.delete`

### Params de ruta
- `id`: UUID — identificador del cliente.

### Query params
Ninguno.

### Body
Sin body.

### Respuesta exitosa

**`204 No Content`** — cuerpo vacío.

### Códigos de error

| Code | HTTP | Causa |
|---|---|---|
| `BAD_REQUEST` | 400 | `id` no es un UUID válido, o falta empresa. |
| `UNAUTHORIZED` | 401 | Falta el token o es inválido/expiró. |
| `FORBIDDEN` | 403 | El usuario no tiene el permiso `customers.delete`. |
| `NOT_FOUND` | 404 | El cliente no existe o no pertenece a la empresa. |
| `INTERNAL_ERROR` | 500 | Error inesperado o de persistencia. |

---

## Estructura de respuesta de entidad `Customer`

```json
{
  "id": "f7e8862e-b6d2-4a08-a3cf-0d2f57f158e2",
  "company_id": "aaaaaaaa-aaaa-4aaa-8aaa-aaaaaaaaaaaa",
  "legal_name": "Comercial Ejemplo",
  "document_type": "national_id | tax_id | passport | other",
  "document_number": "string | null",
  "phone": "string | null",
  "email": "string | null",
  "address": "string | null",
  "credit_limit": "0",
  "credit_balance": "0",
  "points_accrued": 0,
  "status": "active | inactive | blocked",
  "created_at": "2026-10-03T12:00:00Z",
  "updated_at": "2026-10-03T12:00:00Z"
}
```

Los campos opcionales nulos se omiten cuando están etiquetados con `omitempty`. Los importes se serializan como valores decimales (en la creación, su valor inicial es cero).

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
| `BAD_REQUEST` | 400 | UUID inválido, JSON inválido o falta la empresa requerida. |
| `VALIDATION_ERROR` | 400 | Falla de validación de campos o estado. |
| `UNAUTHORIZED` | 401 | Token JWT ausente, inválido o expirado. |
| `FORBIDDEN` | 403 | El usuario no tiene el permiso requerido. |
| `NOT_FOUND` | 404 | Cliente inexistente, eliminado o de otra empresa. |
| `CONFLICT` | 409 | Documento duplicado dentro de la empresa. |
| `INTERNAL_ERROR` | 500 | Error inesperado o de persistencia. |

---

## Casos borde / comportamiento no obvio

- Los IDs de cliente son UUID; los endpoints de detalle y comandos reciben el UUID en la ruta.
- La lista siempre restringe resultados a la empresa autenticada y excluye clientes con soft delete.
- El soft delete establece `deleted_at` y cambia el estado a `inactive`.
- Al omitir `document_type` al crear, se usa `national_id`. Al omitirlo al actualizar (`PUT`), se conserva el tipo actual del cliente (corregido en HU-08-04; antes se forzaba `national_id`).
- En la lista, los parámetros inválidos (`status`, `page`, `limit`, etc.) responden 400; ver HU-08-02.
- Un usuario sin empresa (claim `company_id` ausente) recibe `400 company is required` en todas las rutas.

---

## Diferencias respecto a una versión anterior (si aplica)

Esta rama agrega, sobre la base de la rama definida para HU-08-01:
1. Registro y consulta de clientes con datos de identificación y contacto.
2. Actualización, cambio de estado y baja lógica del cliente.
3. IDs UUID para los clientes y aislamiento de consultas por empresa (`company_id`).
4. Control de acceso por rol con permisos `customers.*`.
