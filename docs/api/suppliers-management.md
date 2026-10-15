# Reporte de endpoints — [Flujo] Gestión de proveedores

## Información general

- Flujo: Proveedores — lista, detalle, alta, edición, desactivación y consulta de productos asociados desde el detalle.
- Rama: `feat/suppliers-management`.
- Prefijo base: `/api/v1`.
- Requiere contexto autenticado del backend: `user_id` y `company_id`, obtenidos de un JWT válido con sesión activa.
- Respuesta exitosa real del proyecto: `{ "success": true, "data": ... }`. Los listados agregan `meta: { "page", "limit", "total", "total_pages" }`.
- Respuesta de error real del proyecto: `{ "success": false, "error": { "code": "CODIGO", "message": "descripción", "details": ... } }`. `details` es opcional.
- Persistencia: `AutoMigrate` para `suppliers` y `product_suppliers`, ejecutado por `go run ./cmd/migrate` después de las migraciones existentes del catálogo. Este flujo no agrega archivos SQL.

El backend devuelve únicamente datos implementados. El frontend debe mockear compras de 12 meses, entrega a tiempo, saldo por pagar, devoluciones y última compra (general y por producto). Esos campos no aparecen en las respuestas; tampoco hay tablas, consultas ni endpoints de órdenes de compra, cuentas por pagar o devoluciones en este módulo.

---

## Resumen rápido

| Método | Ruta | Permiso | Descripción |
|---|---|---|---|
| `GET` | `/api/v1/suppliers` | `suppliers.list` | Listar proveedores. |
| `GET` | `/api/v1/suppliers/summary` | `suppliers.list` | Consultar resumen. |
| `GET` | `/api/v1/suppliers/form-options` | `suppliers.list` | Consultar opciones de formulario. |
| `POST` | `/api/v1/suppliers` | `suppliers.create` | Crear proveedor. |
| `GET` | `/api/v1/suppliers/:id` | `suppliers.read` | Consultar detalle. |
| `PUT` | `/api/v1/suppliers/:id` | `suppliers.update` | Editar proveedor. |
| `PATCH` | `/api/v1/suppliers/:id/deactivate` | `suppliers.deactivate` | Desactivar proveedor. |
| `GET` | `/api/v1/suppliers/:id/products` | `suppliers.read` | Listar productos asociados. |

---

## Autenticación / permisos (todas las rutas)

| Header | Obligatorio | Descripción |
|---|---|---|
| `Authorization` | Sí | `Bearer <access_token>`; se validan token y sesión activa. |
| `Content-Type` | En POST y PUT con body | `application/json`. |

No se utilizan `x-company-id`, `x-user-id` ni una empresa por defecto del entorno. El cliente no puede cambiar de empresa enviando IDs en headers o body: el contexto proviene del token.

Permisos usados en este flujo: `suppliers.list`, `suppliers.read`, `suppliers.create`, `suppliers.update`, `suppliers.deactivate`.

El seed incorpora estos permisos al catálogo y los asigna al rol `superadmin`, como hace con los demás módulos. `business_admin` y `employee` no reciben permisos nuevos automáticamente. Para operar hace falta además un usuario vinculado a una empresa; un superadmin sin empresa recibe `400 BAD_REQUEST`.

La consulta de productos asociados usa `suppliers.read`. Agregar, editar o desvincular asociaciones y el selector de productos quedan fuera de este alcance.

---

## 1) Listar proveedores — `GET /api/v1/suppliers`

Devuelve los proveedores de la empresa con sus datos fiscales, de contacto, comerciales y cantidad de productos asociados.

### Permiso

`suppliers.list`

### Params de ruta

No aplica.

### Query params

- `search` (opcional): string, máximo 160 — búsqueda parcial por nombre comercial, razón social o NIT; ignora mayúsculas y minúsculas.
- `status` (opcional): `active | inactive`; si se omite, incluye ambos.
- `main_industry` (opcional): string, máximo 80 — coincidencia exacta del rubro.
- `sort_by` (opcional, por defecto `name`): `name | legal_name | created_at | lead_time_days`.
- `sort_order` (opcional, por defecto `asc`): `asc | desc`.
- `page` (opcional, por defecto `1`): entero — número de página.
- `limit` (opcional, por defecto `20`): entero — tamaño de página, máximo 100.

### Body

No aplica.

### Respuesta exitosa

**`200`**

```json
{
  "success": true,
  "data": [
    {
      "id": "a1111111-1111-4111-8111-111111111111",
      "name": "Distribuidora Andina",
      "trade_name": "Distribuidora Andina",
      "legal_name": "Distribuidora Andina S.R.L.",
      "tax_id": "1044556677",
      "contact": "Marcela Quispe",
      "phone": "4 4521890",
      "email": "pedidos@distandina.bo",
      "address": null,
      "main_industry": "Abarrotes",
      "status": "active",
      "lead_time_days": 2,
      "payment_terms": "Crédito 30 días",
      "minimum_order": 1500,
      "created_at": "2026-10-15T14:00:00Z",
      "updated_at": null,
      "products_count": 5
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

Ordena por el campo solicitado y desempata por ID ascendente. Los valores nulos se ubican al final. El conteo de productos excluye productos eliminados lógicamente.

### Códigos de error

| Code | HTTP | Causa |
|---|---|---|
| `VALIDATION_ERROR` | 400 | Filtro inválido o texto demasiado largo. |
| `BAD_REQUEST` | 400 | El token no contiene una empresa válida. |
| `UNAUTHORIZED` | 401 | Falta el token, es inválido o su sesión no está activa. |
| `FORBIDDEN` | 403 | El usuario no tiene el permiso requerido. |
| `INTERNAL_ERROR` | 500 | Fallo interno o de persistencia. |

Ejemplo de error:

```json
{
  "success": false,
  "error": {
    "code": "VALIDATION_ERROR",
    "message": "validation failed"
  }
}
```

---

## 2) Consultar resumen — `GET /api/v1/suppliers/summary`

Devuelve únicamente el conteo real de proveedores activos de la empresa.

### Permiso

`suppliers.list`

### Params de ruta

No aplica.

### Query params

No aplica.

### Body

No aplica.

### Respuesta exitosa

**`200`**

```json
{
  "success": true,
  "data": {
    "active_suppliers": 24
  }
}
```

No depende de los filtros de la lista ni de una sucursal. Los indicadores de compras, entregas y devoluciones deben mockearse en el frontend.

### Códigos de error

| Code | HTTP | Causa |
|---|---|---|
| `BAD_REQUEST` | 400 | El token no contiene una empresa válida. |
| `UNAUTHORIZED` | 401 | Falta el token, es inválido o su sesión no está activa. |
| `FORBIDDEN` | 403 | El usuario no tiene el permiso requerido. |
| `INTERNAL_ERROR` | 500 | Fallo interno o de persistencia. |

Ejemplo de error:

```json
{
  "success": false,
  "error": {
    "code": "FORBIDDEN",
    "message": "forbidden"
  }
}
```

---

## 3) Consultar opciones de formulario — `GET /api/v1/suppliers/form-options`

Devuelve rubros y condiciones de pago ya registrados por la empresa, sin valores de ejemplo sembrados.

### Permiso

`suppliers.list`

### Params de ruta

No aplica.

### Query params

No aplica.

### Body

No aplica.

### Respuesta exitosa

**`200`**

```json
{
  "success": true,
  "data": {
    "industries": [
      "Abarrotes",
      "Lácteos"
    ],
    "payment_terms": [
      "Contado",
      "Crédito 30 días"
    ]
  }
}
```

Las listas son únicas, ordenadas y pueden estar vacías. Son sugerencias: crear y editar permiten otros textos dentro de los límites del formulario. Incluyen proveedores activos e inactivos, excluyen los eliminados lógicamente.

### Códigos de error

| Code | HTTP | Causa |
|---|---|---|
| `BAD_REQUEST` | 400 | El token no contiene una empresa válida. |
| `UNAUTHORIZED` | 401 | Falta el token, es inválido o su sesión no está activa. |
| `FORBIDDEN` | 403 | El usuario no tiene el permiso requerido. |
| `INTERNAL_ERROR` | 500 | Fallo interno o de persistencia. |

Ejemplo de error:

```json
{
  "success": false,
  "error": {
    "code": "FORBIDDEN",
    "message": "forbidden"
  }
}
```

---

## 4) Crear proveedor — `POST /api/v1/suppliers`

Registra un proveedor activo sin asociarlo automáticamente a ningún producto.

### Permiso

`suppliers.create`

### Params de ruta

No aplica.

### Query params

No aplica.

### Body

```json
{
  "trade_name": "Distribuidora Andina",
  "legal_name": "Distribuidora Andina S.R.L.",
  "tax_id": "1044556677",
  "phone": "4 4521890",
  "contact": "Marcela Quispe",
  "email": "pedidos@distandina.bo",
  "address": null,
  "main_industry": "Abarrotes",
  "payment_terms": "Crédito 30 días",
  "lead_time_days": 2,
  "minimum_order": 1500
}
```

| Campo | Tipo | Obligatorio | Restricciones / notas |
|---|---|---|---|
| `legal_name` | string | Sí | Razón social, 1–160 caracteres después de recortar espacios. |
| `trade_name` | string o null | No | Nombre comercial, máximo 160. |
| `tax_id` | string o null | No | NIT, máximo 20 dígitos; único dentro de la empresa. Conserva ceros iniciales. |
| `contact` | string o null | No | Persona de contacto, máximo 120. |
| `phone` | string o null | No | Teléfono, máximo 30; admite espacios y prefijos. |
| `email` | string o null | No | Correo válido, máximo 160; sin nombre de presentación. |
| `address` | string o null | No | Dirección, máximo 200. |
| `main_industry` | string o null | No | Rubro principal, máximo 80; texto libre. |
| `payment_terms` | string o null | No | Condición de pago, máximo 60; texto libre. |
| `lead_time_days` | integer o null | No | Plazo general, entre 0 y 32767 días. |
| `minimum_order` | number o null | No | Pedido mínimo en la unidad monetaria de la empresa; 0–9999999999.99 y hasta 2 decimales. |

### Respuesta exitosa

**`201`**

```json
{
  "success": true,
  "data": {
    "id": "a1111111-1111-4111-8111-111111111111",
    "name": "Distribuidora Andina",
    "trade_name": "Distribuidora Andina",
    "legal_name": "Distribuidora Andina S.R.L.",
    "tax_id": "1044556677",
    "contact": "Marcela Quispe",
    "phone": "4 4521890",
    "email": "pedidos@distandina.bo",
    "address": null,
    "main_industry": "Abarrotes",
    "status": "active",
    "lead_time_days": 2,
    "payment_terms": "Crédito 30 días",
    "minimum_order": 1500,
    "created_at": "2026-10-15T14:00:00Z",
    "updated_at": null
  }
}
```

El backend genera ID, estado y fechas. La empresa se obtiene del token. Los campos opcionales omitidos o nulos se guardan como null; los textos vacíos también se normalizan a null.

### Códigos de error

| Code | HTTP | Causa |
|---|---|---|
| `BAD_REQUEST` | 400 | Body JSON mal formado o tipos incompatibles. |
| `VALIDATION_ERROR` | 400 | Campo requerido, formato, longitud, rango o precisión inválidos. |
| `CONFLICT` | 409 | El NIT ya está registrado por otro proveedor de la empresa. |
| `BAD_REQUEST` | 400 | El token no contiene una empresa válida. |
| `UNAUTHORIZED` | 401 | Falta el token, es inválido o su sesión no está activa. |
| `FORBIDDEN` | 403 | El usuario no tiene el permiso requerido. |
| `INTERNAL_ERROR` | 500 | Fallo interno o de persistencia. |

Ejemplo de error:

```json
{
  "success": false,
  "error": {
    "code": "VALIDATION_ERROR",
    "message": "validation failed"
  }
}
```

---

## 5) Consultar detalle — `GET /api/v1/suppliers/:id`

Devuelve los datos completos del proveedor para la pantalla de detalle y para precargar el formulario de edición.

### Permiso

`suppliers.read`

### Params de ruta

- `id`: UUID — identificador del proveedor de la empresa autenticada.

### Query params

No aplica.

### Body

No aplica.

### Respuesta exitosa

**`200`**

```json
{
  "success": true,
  "data": {
    "id": "a1111111-1111-4111-8111-111111111111",
    "name": "Distribuidora Andina",
    "trade_name": "Distribuidora Andina",
    "legal_name": "Distribuidora Andina S.R.L.",
    "tax_id": "1044556677",
    "contact": "Marcela Quispe",
    "phone": "4 4521890",
    "email": "pedidos@distandina.bo",
    "address": null,
    "main_industry": "Abarrotes",
    "status": "active",
    "lead_time_days": 2,
    "payment_terms": "Crédito 30 días",
    "minimum_order": 1500,
    "created_at": "2026-10-15T14:00:00Z",
    "updated_at": null
  }
}
```

Los productos se consultan por separado con GET /suppliers/:id/products. created_at permite mostrar “Proveedor desde”. El endpoint no devuelve indicadores financieros ni fecha de última compra.

### Códigos de error

| Code | HTTP | Causa |
|---|---|---|
| `BAD_REQUEST` | 400 | UUID de ruta inválido o vacío. |
| `NOT_FOUND` | 404 | El proveedor no existe, está eliminado o pertenece a otra empresa. |
| `BAD_REQUEST` | 400 | El token no contiene una empresa válida. |
| `UNAUTHORIZED` | 401 | Falta el token, es inválido o su sesión no está activa. |
| `FORBIDDEN` | 403 | El usuario no tiene el permiso requerido. |
| `INTERNAL_ERROR` | 500 | Fallo interno o de persistencia. |

Ejemplo de error:

```json
{
  "success": false,
  "error": {
    "code": "NOT_FOUND",
    "message": "supplier not found"
  }
}
```

---

## 6) Editar proveedor — `PUT /api/v1/suppliers/:id`

Reemplaza los campos editables del proveedor sin modificar su estado ni sus asociaciones.

### Permiso

`suppliers.update`

### Params de ruta

- `id`: UUID — identificador del proveedor de la empresa autenticada.

### Query params

No aplica.

### Body

```json
{
  "trade_name": "Distribuidora Andina",
  "legal_name": "Distribuidora Andina S.R.L.",
  "tax_id": "1044556677",
  "phone": "4 4521890",
  "contact": "Marcela Quispe",
  "email": "pedidos@distandina.bo",
  "address": null,
  "main_industry": "Abarrotes",
  "payment_terms": "Crédito 30 días",
  "lead_time_days": 2,
  "minimum_order": 1500
}
```

| Campo | Tipo | Obligatorio | Restricciones / notas |
|---|---|---|---|
| `legal_name` | string | Sí | Razón social, 1–160 caracteres después de recortar espacios. |
| `trade_name` | string o null | No | Nombre comercial, máximo 160. |
| `tax_id` | string o null | No | NIT, máximo 20 dígitos; único dentro de la empresa. Conserva ceros iniciales. |
| `contact` | string o null | No | Persona de contacto, máximo 120. |
| `phone` | string o null | No | Teléfono, máximo 30; admite espacios y prefijos. |
| `email` | string o null | No | Correo válido, máximo 160; sin nombre de presentación. |
| `address` | string o null | No | Dirección, máximo 200. |
| `main_industry` | string o null | No | Rubro principal, máximo 80; texto libre. |
| `payment_terms` | string o null | No | Condición de pago, máximo 60; texto libre. |
| `lead_time_days` | integer o null | No | Plazo general, entre 0 y 32767 días. |
| `minimum_order` | number o null | No | Pedido mínimo en la unidad monetaria de la empresa; 0–9999999999.99 y hasta 2 decimales. |

### Respuesta exitosa

**`200`**

```json
{
  "success": true,
  "data": {
    "id": "a1111111-1111-4111-8111-111111111111",
    "name": "Distribuidora Andina",
    "trade_name": "Distribuidora Andina",
    "legal_name": "Distribuidora Andina S.R.L.",
    "tax_id": "1044556677",
    "contact": "Marcela Quispe",
    "phone": "4 4521890",
    "email": "pedidos@distandina.bo",
    "address": null,
    "main_industry": "Abarrotes",
    "status": "active",
    "lead_time_days": 2,
    "payment_terms": "Crédito 30 días",
    "minimum_order": 1500,
    "created_at": "2026-10-15T14:00:00Z",
    "updated_at": "2026-10-15T15:00:00Z"
  }
}
```

Es un reemplazo completo (PUT): hay que reenviar legal_name y todos los opcionales que se desee conservar. Omitir un opcional lo limpia a null. Se puede conservar el propio NIT. No cambia costos ni plazos de las asociaciones; su modificación queda fuera de alcance.

### Códigos de error

| Code | HTTP | Causa |
|---|---|---|
| `BAD_REQUEST` | 400 | UUID de ruta inválido o vacío. |
| `BAD_REQUEST` | 400 | Body JSON mal formado o tipos incompatibles. |
| `VALIDATION_ERROR` | 400 | Campo requerido, formato, longitud, rango o precisión inválidos. |
| `NOT_FOUND` | 404 | El proveedor no existe o no pertenece a la empresa. |
| `CONFLICT` | 409 | El nuevo NIT pertenece a otro proveedor de la empresa. |
| `BAD_REQUEST` | 400 | El token no contiene una empresa válida. |
| `UNAUTHORIZED` | 401 | Falta el token, es inválido o su sesión no está activa. |
| `FORBIDDEN` | 403 | El usuario no tiene el permiso requerido. |
| `INTERNAL_ERROR` | 500 | Fallo interno o de persistencia. |

Ejemplo de error:

```json
{
  "success": false,
  "error": {
    "code": "VALIDATION_ERROR",
    "message": "validation failed"
  }
}
```

---

## 7) Desactivar proveedor — `PATCH /api/v1/suppliers/:id/deactivate`

Cambia el estado a inactive y conserva el registro y sus productos asociados.

### Permiso

`suppliers.deactivate`

### Params de ruta

- `id`: UUID — identificador del proveedor de la empresa autenticada.

### Query params

No aplica.

### Body

No aplica.

### Respuesta exitosa

**`200`**

```json
{
  "success": true,
  "data": {
    "id": "a1111111-1111-4111-8111-111111111111",
    "name": "Distribuidora Andina",
    "trade_name": "Distribuidora Andina",
    "legal_name": "Distribuidora Andina S.R.L.",
    "tax_id": "1044556677",
    "contact": "Marcela Quispe",
    "phone": "4 4521890",
    "email": "pedidos@distandina.bo",
    "address": null,
    "main_industry": "Abarrotes",
    "status": "inactive",
    "lead_time_days": 2,
    "payment_terms": "Crédito 30 días",
    "minimum_order": 1500,
    "created_at": "2026-10-15T14:00:00Z",
    "updated_at": "2026-10-15T15:10:00Z"
  }
}
```

Sin body. Es idempotente: repetirla responde 200 y no vuelve a cambiar updated_at. La confirmación visual pertenece al frontend. El backend no consulta saldos ni valida órdenes de compra; esos módulos todavía no están implementados.

### Códigos de error

| Code | HTTP | Causa |
|---|---|---|
| `BAD_REQUEST` | 400 | UUID de ruta inválido o vacío. |
| `NOT_FOUND` | 404 | El proveedor no existe o no pertenece a la empresa. |
| `BAD_REQUEST` | 400 | El token no contiene una empresa válida. |
| `UNAUTHORIZED` | 401 | Falta el token, es inválido o su sesión no está activa. |
| `FORBIDDEN` | 403 | El usuario no tiene el permiso requerido. |
| `INTERNAL_ERROR` | 500 | Fallo interno o de persistencia. |

Ejemplo de error:

```json
{
  "success": false,
  "error": {
    "code": "NOT_FOUND",
    "message": "supplier not found"
  }
}
```

---

## 8) Listar productos asociados — `GET /api/v1/suppliers/:id/products`

Lista las asociaciones reales del proveedor junto con el nombre y SKU de cada producto.

### Permiso

`suppliers.read`

### Params de ruta

- `id`: UUID — identificador del proveedor de la empresa autenticada.

### Query params

- `search` (opcional): string, máximo 160 — búsqueda parcial por nombre o SKU, sin distinguir mayúsculas.
- `page` (opcional, por defecto `1`): entero — número de página.
- `limit` (opcional, por defecto `20`): entero — tamaño de página, máximo 100.

### Body

No aplica.

### Respuesta exitosa

**`200`**

```json
{
  "success": true,
  "data": [
    {
      "id": "b2222222-2222-4222-8222-222222222222",
      "supplier_id": "a1111111-1111-4111-8111-111111111111",
      "product_id": "c3333333-3333-4333-8333-333333333333",
      "lead_time_days": 5,
      "current_cost": 13.3,
      "created_at": "2026-10-15T14:05:00Z",
      "updated_at": null,
      "name": "Aceite Fino 900 ml",
      "sku": "ABA-210"
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

Ordena por nombre del producto e ID ascendente. Conserva los productos inactivos asociados, pero excluye productos eliminados lógicamente. No devuelve última compra.

### Códigos de error

| Code | HTTP | Causa |
|---|---|---|
| `BAD_REQUEST` | 400 | UUID de ruta inválido o vacío. |
| `VALIDATION_ERROR` | 400 | Búsqueda demasiado larga. |
| `NOT_FOUND` | 404 | El proveedor no existe o no pertenece a la empresa. |
| `BAD_REQUEST` | 400 | El token no contiene una empresa válida. |
| `UNAUTHORIZED` | 401 | Falta el token, es inválido o su sesión no está activa. |
| `FORBIDDEN` | 403 | El usuario no tiene el permiso requerido. |
| `INTERNAL_ERROR` | 500 | Fallo interno o de persistencia. |

Ejemplo de error:

```json
{
  "success": false,
  "error": {
    "code": "VALIDATION_ERROR",
    "message": "validation failed"
  }
}
```

---

## Estructura de respuesta de entidad `Supplier`

```json
{
  "id": "a1111111-1111-4111-8111-111111111111",
  "name": "Distribuidora Andina",
  "trade_name": "Distribuidora Andina",
  "legal_name": "Distribuidora Andina S.R.L.",
  "tax_id": "1044556677",
  "contact": "Marcela Quispe",
  "phone": "4 4521890",
  "email": "pedidos@distandina.bo",
  "address": null,
  "main_industry": "Abarrotes",
  "status": "active",
  "lead_time_days": 2,
  "payment_terms": "Crédito 30 días",
  "minimum_order": 1500,
  "created_at": "2026-10-15T14:00:00Z",
  "updated_at": null
}
```

- `name`: valor derivado de `trade_name`; si no hay nombre comercial, usa `legal_name`.
- `status`: `active | inactive`.
- Los campos opcionales se devuelven como `null`, no como texto vacío.
- Los importes son números JSON en la unidad completa de la moneda de la empresa, sin prefijos ni separadores de miles. Los requests también admiten strings decimales válidos.
- Las fechas se devuelven en UTC, formato RFC 3339. `updated_at` es `null` al crear.
- En la lista se agrega `products_count`: entero con asociaciones a productos no eliminados.
- El detalle devuelve el mismo shape de `Supplier`; sus productos tienen una consulta separada.

## Estructura de respuesta de entidad `ProductSupplier`

```json
{
  "id": "b2222222-2222-4222-8222-222222222222",
  "supplier_id": "a1111111-1111-4111-8111-111111111111",
  "product_id": "c3333333-3333-4333-8333-333333333333",
  "lead_time_days": 5,
  "current_cost": 13.3,
  "created_at": "2026-10-15T14:05:00Z",
  "updated_at": null
}
```

En el listado de productos asociados se agregan `name` y `sku` (string o null). El costo es específico del proveedor y `lead_time_days` puede ser null. El backend no copia automáticamente el plazo general del proveedor a la asociación.

## Estructura de respuesta de entidad `SupplierSummary`

```json
{
  "active_suppliers": 24
}
```

Solo ese contador está implementado. No hay valores simulados ni ceros sustituyendo indicadores pendientes.

---

## Tabla de errores consolidada

```json
{
  "success": false,
  "error": {
    "code": "VALIDATION_ERROR",
    "message": "validation failed",
    "details": {
      "minimum_order": "must be between 0 and 9999999999.99 with at most two decimal places"
    }
  }
}
```

| Code | HTTP | Origen |
|---|---|---|
| `BAD_REQUEST` | 400 | JSON o tipos incompatibles, UUID de ruta inválido o ausencia de empresa en el contexto. |
| `VALIDATION_ERROR` | 400 | Campos requeridos, filtros, longitudes, NIT, correo, días o importes inválidos. |
| `UNAUTHORIZED` | 401 | Token ausente/inválido/vencido o sesión inactiva. |
| `FORBIDDEN` | 403 | Falta el permiso de la ruta. |
| `NOT_FOUND` | 404 | Proveedor inexistente, eliminado o fuera de la empresa. |
| `CONFLICT` | 409 | NIT repetido dentro de la empresa. |
| `INTERNAL_ERROR` | 500 | Error interno; no se exponen detalles SQL. |

Los mensajes actuales del backend están en inglés. El frontend puede traducir por código y usar `details` para marcar campos.

---

## Casos borde / comportamiento no obvio

- En Bruno, definir manualmente `supplierId` en las variables de la carpeta Supplier con el ID obtenido al crear o listar. Las rutas correctas son `PUT /api/v1/suppliers/:id` y `PATCH /api/v1/suppliers/:id/deactivate`. Una variable vacía produce `/suppliers/` o `/suppliers//deactivate`, que no son rutas válidas. No hay scripts para seleccionar o guardar IDs.
- Crear un proveedor no lo asocia a productos. Las asociaciones existentes se muestran solo para consulta.

- La lista entrega suficientes datos para agrupar columnas en este orden sugerido: Proveedor (`name`, `legal_name`, `tax_id`); Rubro; Persona de contacto; Teléfono/correo; Plazo de entrega; Estado. La condición de pago puede ser una columna opcional.
- Todas las consultas y cambios se limitan a la empresa del token. Un ID de otra empresa responde 404 en las rutas de detalle y modificación.
- El NIT es único por empresa, incluso si el proveedor está inactivo. La razón social puede repetirse. Varios proveedores sin NIT son válidos.
- Los textos se recortan. En campos opcionales, cadena vacía o espacios se convierten en null.
- PUT reemplaza todos los campos editables. El frontend debe enviar los datos que quiera conservar; omitir un opcional lo limpia.
- La edición no permite cambiar IDs, empresa, estado ni fecha de creación. El estado solo cambia mediante la ruta de desactivación.
- Desactivar no elimina el proveedor ni sus asociaciones y no lo oculta por defecto de la lista. Permanece consultable y editable.
- En una lista vacía se devuelve `data: []` y `meta.total_pages: 0`. Una página fuera de rango devuelve un array vacío conservando la página solicitada y el total real.
- Si `page` no es un entero válido o es menor que 1, se usa 1. Si `limit` no es un entero válido, es menor que 1 o supera 100, se usa 20, sin responder 400. Filtros y ordenamientos no admitidos sí responden 400.
- Los caracteres %, _ y barra inversa se buscan literalmente; no actúan como comodines proporcionados por el usuario.
- Proveedores pertenecen a la empresa, sin filtro por sucursal en estos endpoints.
- Compras, entrega a tiempo, saldo por pagar, devoluciones y última compra se mockean exclusivamente en el frontend. El aviso de saldo pendiente del modal de desactivación también queda del lado del frontend hasta implementar cuentas por pagar.
- El frontend muestra confirmaciones, detecta cambios sin guardar y presenta notificaciones según el resultado HTTP. No hay endpoints para confirmar/cancelar ni para descartar cambios.
- No se incluyen agregar, editar o desvincular asociaciones, el selector de productos, la pantalla de devoluciones, los flujos de la ficha de producto ni módulos de compras o cuentas por pagar.

---

## Diferencias respecto a una versión anterior (si aplica)

No aplica: primera versión del contrato del flujo de proveedores en esta rama.
