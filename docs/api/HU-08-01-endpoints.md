# Reporte de endpoints — [HU-08-01] Registro de clientes

## Información general

- HU: HU-08-01 — Registro de clientes
- Rama: `feat/HU-08-01-customer-registration`
- Prefijo base: `/customers` (el router del módulo no agrega `/api/v1`; el prefijo efectivo depende del router padre)
- Requiere autenticación: token JWT válido en el esquema de autorización de portador. El `tenant_id` se obtiene del contexto autenticado; no se recibe en el body ni mediante `x-company-id`.
- Formato de respuesta exitosa: `{ "success": true, "data": ... }`. Las respuestas paginadas incluyen `meta`.
- Formato de respuesta de error: `{ "success": false, "error": { "code": "CODIGO", "message": "descripción" } }`

> El módulo define las rutas en `customer_router.go` relativas al router que recibe. A diferencia de auth y user, no agrega `/api/v1` por sí mismo.

---

## Resumen rápido

| Método | Ruta | Permiso | Descripción |
|---|---|---|---|
| `POST` | `/customers/` | — | Registra un cliente en el tenant autenticado. |
| `GET` | `/customers/` | — | Lista clientes del tenant con filtros y paginación. |
| `GET` | `/customers/:id` | — | Obtiene un cliente por UUID. |
| `PUT` | `/customers/:id` | — | Actualiza los datos editables del cliente. |
| `PATCH` | `/customers/:id/status` | — | Cambia el estado del cliente. |
| `DELETE` | `/customers/:id` | — | Da de baja lógicamente al cliente. |

---

## Autenticación / permisos (todas las rutas)

Todas las rutas requieren el header `Authorization` con un token JWT válido. El middleware de autenticación obtiene el tenant de la identidad autenticada. No se declara un permiso RBAC específico en estas rutas.

| Header / contexto | Obligatorio | Descripción |
|---|---|---|
| `Authorization` | Sí | Token JWT usando el esquema de portador. |
| `tenant_id` | Sí, en el contexto del token | UUID usado para limitar lecturas y escrituras al tenant autenticado. No enviarlo como campo del body. |

Permisos usados en esta HU: ninguno declarado en el router; se exige autenticación.

---

## 1) Registrar cliente — `POST /customers/`

Registra un cliente en el tenant autenticado. El UUID del cliente lo genera el backend.

### Permiso
`—` (requiere autenticación).

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
| `document_number` | string | No | Máximo 30 caracteres. Si se informa, debe ser único por tenant y tipo de documento entre clientes no eliminados. |
| `phone` | string | No | Máximo 30 caracteres. |
| `email` | string | No | Formato email, máximo 160 caracteres. |

### Respuesta exitosa

**`201 Created`**

```json
{
  "success": true,
  "data": {
    "id": "f7e8862e-b6d2-4a08-a3cf-0d2f57f158e2",
    "tenant_id": "11111111-1111-4111-8111-111111111111",
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
| `BAD_REQUEST` | 400 | JSON inválido o falta el tenant requerido. |
| `VALIDATION_ERROR` | 400 | Nombre, tipo de documento, email u otro campo no cumple validación. |
| `CONFLICT` | 409 | Ya existe en el tenant un cliente no eliminado con el mismo tipo y número de documento. |
| `UNAUTHORIZED` | 401 | Falta el token o es inválido/expiró. |
| `INTERNAL_ERROR` | 500 | Error inesperado o de persistencia. |

---

## 2) Listar clientes — `GET /customers/`

Lista únicamente clientes no eliminados del tenant autenticado.

### Permiso
`—` (requiere autenticación).

### Params de ruta
Ninguno.

### Query params

| Parámetro | Obligatorio / por defecto | Tipo | Descripción |
|---|---|---|---|
| `page` | Opcional, por defecto `1` | integer | Página solicitada. Valores menores que 1 se ajustan a 1. |
| `limit` | Opcional, por defecto `20` | integer | Tamaño de página; valores menores que 1 o mayores que 100 se ajustan a 20. Máximo: 100. |
| `q` | Opcional | string | Búsqueda parcial en nombre legal, documento, teléfono o email. |
| `status` | Opcional | enum | `active` \| `inactive` \| `blocked`. Un valor inválido se ignora actualmente. |

### Respuesta exitosa

**`200 OK`**

```json
{
  "success": true,
  "data": [
    {
      "id": "f7e8862e-b6d2-4a08-a3cf-0d2f57f158e2",
      "tenant_id": "11111111-1111-4111-8111-111111111111",
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
| `UNAUTHORIZED` | 401 | Falta el token o es inválido/expiró. |
| `INTERNAL_ERROR` | 500 | Error inesperado de persistencia. |

---

## 3) Obtener cliente — `GET /customers/:id`

Obtiene un cliente del tenant autenticado por su UUID.

### Permiso
`—` (requiere autenticación).

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
    "tenant_id": "11111111-1111-4111-8111-111111111111",
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
| `BAD_REQUEST` | 400 | `id` no es un UUID válido. |
| `NOT_FOUND` | 404 | El cliente no existe, está eliminado o pertenece a otro tenant. |
| `UNAUTHORIZED` | 401 | Falta el token o es inválido/expiró. |
| `INTERNAL_ERROR` | 500 | Error inesperado de persistencia. |

---

## 4) Actualizar cliente — `PUT /customers/:id`

Actualiza los campos editables del cliente. La operación trata los campos omitidos como vacíos/nulos; enviar los datos que se desean conservar.

### Permiso
`—` (requiere autenticación).

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
| `document_type` | enum | No | `national_id` \| `tax_id` \| `passport` \| `other`. Si se omite, el handler lo convierte en `national_id`. |
| `document_number` | string | No | Máximo 30 caracteres. Se comprueba unicidad si cambia. |
| `phone` | string | No | Máximo 30 caracteres. |
| `email` | string | No | Formato email, máximo 160 caracteres. |
| `address` | string | No | Máximo 200 caracteres. |

### Respuesta exitosa

**`200 OK`** — devuelve el cliente actualizado con la misma estructura de `GET /customers/:id`.

### Códigos de error

| Code | HTTP | Causa |
|---|---|---|
| `BAD_REQUEST` | 400 | UUID, JSON o validación inválidos. |
| `NOT_FOUND` | 404 | El cliente no existe o no pertenece al tenant. |
| `CONFLICT` | 409 | El documento ya está asociado a otro cliente del tenant. |
| `UNAUTHORIZED` | 401 | Falta el token o es inválido/expiró. |
| `INTERNAL_ERROR` | 500 | Error inesperado o de persistencia. |

---

## 5) Cambiar estado — `PATCH /customers/:id/status`

Cambia el estado operativo del cliente.

### Permiso
`—` (requiere autenticación).

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

**`200 OK`** — devuelve el cliente actualizado con la misma estructura de `GET /customers/:id`.

### Códigos de error

| Code | HTTP | Causa |
|---|---|---|
| `BAD_REQUEST` | 400 | UUID o body inválido. |
| `VALIDATION_ERROR` | 400 | Estado fuera del enum permitido. |
| `NOT_FOUND` | 404 | El cliente no existe o no pertenece al tenant. |
| `UNAUTHORIZED` | 401 | Falta el token o es inválido/expiró. |
| `INTERNAL_ERROR` | 500 | Error inesperado o de persistencia. |

---

## 6) Eliminar cliente — `DELETE /customers/:id`

Marca al cliente como eliminado (soft delete); no elimina físicamente el registro.

### Permiso
`—` (requiere autenticación).

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
| `BAD_REQUEST` | 400 | `id` no es un UUID válido. |
| `NOT_FOUND` | 404 | El cliente no existe o no pertenece al tenant. |
| `UNAUTHORIZED` | 401 | Falta el token o es inválido/expiró. |
| `INTERNAL_ERROR` | 500 | Error inesperado o de persistencia. |

---

## Estructura de respuesta de entidad `Customer`

```json
{
  "id": "f7e8862e-b6d2-4a08-a3cf-0d2f57f158e2",
  "tenant_id": "11111111-1111-4111-8111-111111111111",
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
| `BAD_REQUEST` | 400 | UUID inválido, JSON inválido o falta el tenant requerido. |
| `VALIDATION_ERROR` | 400 | Falla de validación de campos o estado. |
| `UNAUTHORIZED` | 401 | Token JWT ausente, inválido o expirado. |
| `NOT_FOUND` | 404 | Cliente inexistente, eliminado o de otro tenant. |
| `CONFLICT` | 409 | Documento duplicado dentro del tenant. |
| `INTERNAL_ERROR` | 500 | Error inesperado o de persistencia. |

---

## Casos borde / comportamiento no obvio

- Los IDs de cliente son UUID; los endpoints de detalle y comandos reciben el UUID en la ruta.
- La lista siempre restringe resultados al tenant autenticado y excluye clientes con soft delete.
- El soft delete establece `deleted_at` y cambia el estado a `inactive`.
- Al omitir `document_type` al crear, se usa `national_id`. Al omitirlo al actualizar, también se convierte a `national_id`; por tanto, puede cambiar el tipo previamente almacenado.
- En la lista, un valor inválido de `status` se ignora en lugar de devolver 400.
- `page < 1` se ajusta a 1. `limit < 1` o `limit > 100` se ajusta al valor predeterminado 20.
- El router del módulo registra `/customers` relativo al router padre; el propio módulo no agrega `/api/v1`.

---

## Diferencias respecto a una versión anterior (si aplica)

Esta rama agrega, sobre la base de la rama definida para HU-08-01:
1. Registro y consulta de clientes con datos de identificación y contacto.
2. Actualización, cambio de estado y baja lógica del cliente.
3. IDs UUID para los clientes y aislamiento de consultas por tenant.
