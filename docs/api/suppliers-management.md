# Reporte de endpoints — [Flujo] Gestión de proveedores

## Información general

- Flujo: Proveedores — lista, detalle, alta, edición, desactivación y gestión de asociaciones desde proveedor y ficha de producto.
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
| `GET` | `/api/v1/suppliers/:id/product-options` | `suppliers.update` | Productos para el selector de asociación. |
| `POST` | `/api/v1/suppliers/:id/products/batch` | `suppliers.update` | Asociar varios productos. |
| `POST` | `/api/v1/suppliers/:id/products` | `suppliers.update` | Asociar un proveedor a un producto. |
| `PUT` | `/api/v1/suppliers/:id/products/:product_id` | `suppliers.update` | Editar datos de la asociación. |
| `DELETE` | `/api/v1/suppliers/:id/products/:product_id` | `suppliers.update` | Desvincular proveedor y producto. |
| `GET` | `/api/v1/products/:product_id/suppliers` | `suppliers.read` | Listar proveedores de un producto. |
| `GET` | `/api/v1/products/:product_id/supplier-options` | `suppliers.update` | Proveedores para el selector del producto. |

---

## Autenticación / permisos (todas las rutas)

| Header | Obligatorio | Descripción |
|---|---|---|
| `Authorization` | Sí | `Bearer <access_token>`; se validan token y sesión activa. |
| `Content-Type` | En POST y PUT con body | `application/json`. |

No se utilizan `x-company-id`, `x-user-id` ni una empresa por defecto del entorno. El cliente no puede cambiar de empresa enviando IDs en headers o body: el contexto proviene del token.

Permisos usados en este flujo: `suppliers.list`, `suppliers.read`, `suppliers.create`, `suppliers.update`, `suppliers.deactivate`.

El seed incorpora estos permisos al catálogo y los asigna al rol `superadmin`, como hace con los demás módulos. `business_admin` y `employee` no reciben permisos nuevos automáticamente. Para operar hace falta además un usuario vinculado a una empresa; un superadmin sin empresa recibe `400 BAD_REQUEST`.

Las listas de asociados usan `suppliers.read`. Los selectores y todas las modificaciones de asociaciones usan `suppliers.update`. Todas estas rutas se implementan dentro de supplier.

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

Es un reemplazo completo (PUT): hay que reenviar legal_name y todos los opcionales que se desee conservar. Omitir un opcional lo limpia a null. Se puede conservar el propio NIT. No cambia costos ni plazos de las asociaciones; estas se editan por separado.

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
      "supplier_code": "DA-ACE-900",
      "preferred": false,
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


## 9) Productos para el selector de asociación — `GET /api/v1/suppliers/:id/product-options`

Incluye productos ya asociados para mostrarlos deshabilitados. También se muestran inactivos, con selectable=false. El stock es el total actual del producto, sin filtro de sucursal. El catálogo de categorías del dropdown sigue disponible en GET /api/v1/categories.

### Permiso

`suppliers.update`

### Params de ruta

- id: UUID del proveedor.

### Query params

- `search` (opcional, máximo 160): nombre, SKU o código de barras; búsqueda parcial sin distinguir mayúsculas.
- `category_id` (opcional): UUID; incluye la categoría y sus descendientes.
- `page` (opcional, por defecto 1): número de página.
- `limit` (opcional, por defecto 20): tamaño de página, máximo 100.

### Body

No aplica.

### Respuesta exitosa

**`200`**

```json
{
  "success": true,
  "data": [
    {
      "id": "c3333333-3333-4333-8333-333333333333",
      "name": "Aceite Fino 900 ml",
      "sku": "ABA-210",
      "barcode": "7772000002103",
      "status": "active",
      "category": {
        "id": "d4444444-4444-4444-8444-444444444444",
        "name": "Aceites",
        "path": "Abarrotes / Aceites"
      },
      "stock": 8,
      "unit_abbreviation": "UNIT",
      "is_associated": true,
      "selectable": false
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
| `BAD_REQUEST` | 400 | UUID de ruta inválido, JSON incompatible o falta empresa en el contexto. |
| `VALIDATION_ERROR` | 400 | IDs, filtros, código, costo, precisión o plazo inválidos. |
| `NOT_FOUND` | 404 | Proveedor, producto o relación inexistente, eliminado o ajeno a la empresa. |
| `UNAUTHORIZED` | 401 | Token ausente/inválido o sesión inactiva. |
| `FORBIDDEN` | 403 | Falta el permiso de la ruta. |
| `INTERNAL_ERROR` | 500 | Error interno de persistencia. |

Ejemplo de error:

```json
{
  "success": false,
  "error": {
    "code": "NOT_FOUND",
    "message": "product not found"
  }
}
```

---

## 10) Asociar varios productos — `POST /api/v1/suppliers/:id/products/batch`

Operación atómica: un producto inexistente, ajeno, inactivo o ya asociado rechaza la selección completa. Guarda código, costo y plazo específico como null, y preferred=false. Se completan después desde la edición de la relación. La respuesta respeta el orden de product_ids.

### Permiso

`suppliers.update`

### Params de ruta

- id: UUID del proveedor.

### Query params

No aplica.

### Body

```json
{
  "product_ids": [
    "c3333333-3333-4333-8333-333333333333",
    "c5555555-5555-4555-8555-555555555555"
  ]
}
```

| Campo | Tipo | Obligatorio | Restricciones / notas |
|---|---|---|---|
| product_ids | string[] | Sí | Entre 1 y 100 UUID distintos de productos activos de la empresa. |

### Respuesta exitosa

**`201`**

```json
{
  "success": true,
  "data": [
    {
      "id": "b2222222-2222-4222-8222-222222222222",
      "supplier_id": "a1111111-1111-4111-8111-111111111111",
      "product_id": "c3333333-3333-4333-8333-333333333333",
      "supplier_code": null,
      "lead_time_days": null,
      "current_cost": null,
      "preferred": false,
      "created_at": "2026-10-16T14:00:00Z",
      "updated_at": null
    },
    {
      "id": "b6666666-6666-4666-8666-666666666666",
      "supplier_id": "a1111111-1111-4111-8111-111111111111",
      "product_id": "c5555555-5555-4555-8555-555555555555",
      "supplier_code": null,
      "lead_time_days": null,
      "current_cost": null,
      "preferred": false,
      "created_at": "2026-10-16T14:00:00Z",
      "updated_at": null
    }
  ]
}
```

### Códigos de error

| Code | HTTP | Causa |
|---|---|---|
| `BAD_REQUEST` | 400 | UUID de ruta inválido, JSON incompatible o falta empresa en el contexto. |
| `VALIDATION_ERROR` | 400 | IDs, filtros, código, costo, precisión o plazo inválidos. |
| `NOT_FOUND` | 404 | Proveedor, producto o relación inexistente, eliminado o ajeno a la empresa. |
| `CONFLICT` | 409 | Relación duplicada, recurso inactivo al asociar o conflicto de preferencia. |
| `UNAUTHORIZED` | 401 | Token ausente/inválido o sesión inactiva. |
| `FORBIDDEN` | 403 | Falta el permiso de la ruta. |
| `INTERNAL_ERROR` | 500 | Error interno de persistencia. |

Ejemplo de error:

```json
{
  "success": false,
  "error": {
    "code": "NOT_FOUND",
    "message": "product not found"
  }
}
```

---

## 11) Asociar un proveedor a un producto — `POST /api/v1/suppliers/:id/products`

Es la confirmación del segundo modal, después de seleccionar el proveedor existente. Requiere proveedor activo. Crear un proveedor mediante POST /suppliers sigue siendo independiente: no lo asocia automáticamente. No existe un endpoint de escritura duplicado en /products; ambos lados usan esta misma relación.

### Permiso

`suppliers.update`

### Params de ruta

- id: UUID del proveedor seleccionado.

### Query params

No aplica.

### Body

```json
{
  "product_id": "c3333333-3333-4333-8333-333333333333",
  "lead_time_days": 3,
  "current_cost": 13.1
}
```

| Campo | Tipo | Obligatorio | Restricciones / notas |
|---|---|---|---|
| product_id | string UUID | Sí | Producto activo y no eliminado de la misma empresa. |
| lead_time_days | integer o null | No | 0–32767 días; null permite usar el plazo general en la vista. |
| current_cost | number | Sí | 0–9999999999.99, máximo 2 decimales; cero es válido. |
| supplier_code | string o null | No | Máximo 60 caracteres; vacío se convierte en null. |
| preferred | boolean | No | Por defecto false. true cambia el preferido de este producto de forma atómica. |

### Respuesta exitosa

**`201`**

```json
{
  "success": true,
  "data": {
    "id": "b2222222-2222-4222-8222-222222222222",
    "supplier_id": "a1111111-1111-4111-8111-111111111111",
    "product_id": "c3333333-3333-4333-8333-333333333333",
    "supplier_code": "DA-ACE-900",
    "lead_time_days": 3,
    "current_cost": 13.1,
    "preferred": false,
    "created_at": "2026-10-16T14:00:00Z",
    "updated_at": null
  }
}
```

### Códigos de error

| Code | HTTP | Causa |
|---|---|---|
| `BAD_REQUEST` | 400 | UUID de ruta inválido, JSON incompatible o falta empresa en el contexto. |
| `VALIDATION_ERROR` | 400 | IDs, filtros, código, costo, precisión o plazo inválidos. |
| `NOT_FOUND` | 404 | Proveedor, producto o relación inexistente, eliminado o ajeno a la empresa. |
| `CONFLICT` | 409 | Relación duplicada, recurso inactivo al asociar o conflicto de preferencia. |
| `UNAUTHORIZED` | 401 | Token ausente/inválido o sesión inactiva. |
| `FORBIDDEN` | 403 | Falta el permiso de la ruta. |
| `INTERNAL_ERROR` | 500 | Error interno de persistencia. |

Ejemplo de error:

```json
{
  "success": false,
  "error": {
    "code": "NOT_FOUND",
    "message": "product not found"
  }
}
```

---

## 12) Editar datos de la asociación — `PUT /api/v1/suppliers/:id/products/:product_id`

No crea una relación inexistente. PUT reemplaza código, costo y plazo: reenviar los valores que se quieran conservar. No cambia el producto general. Permite editar una relación existente de un proveedor inactivo, pero no marcarlo como nuevo preferido.

### Permiso

`suppliers.update`

### Params de ruta

- id: UUID del proveedor.
- product_id: UUID del producto; no es el ID de la asociación.

### Query params

No aplica.

### Body

```json
{
  "supplier_code": "DA-ACE-900",
  "lead_time_days": 3,
  "current_cost": 13.1,
  "preferred": true
}
```

| Campo | Tipo | Obligatorio | Restricciones / notas |
|---|---|---|---|
| supplier_code | string o null | No | Máximo 60; omitido o vacío limpia el código. |
| lead_time_days | integer o null | No | 0–32767; omitido o null limpia el plazo específico. |
| current_cost | number o null | No | Mismo rango y precisión que al asociar; omitido o null limpia el costo pendiente de completar. |
| preferred | boolean | No | Si se omite conserva el valor actual. true cambia el preferido; false lo desmarca. |

### Respuesta exitosa

**`200`**

```json
{
  "success": true,
  "data": {
    "id": "b2222222-2222-4222-8222-222222222222",
    "supplier_id": "a1111111-1111-4111-8111-111111111111",
    "product_id": "c3333333-3333-4333-8333-333333333333",
    "supplier_code": "DA-ACE-900",
    "lead_time_days": 3,
    "current_cost": 13.1,
    "preferred": true,
    "created_at": "2026-10-16T14:00:00Z",
    "updated_at": "2026-10-16T15:00:00Z"
  }
}
```

### Códigos de error

| Code | HTTP | Causa |
|---|---|---|
| `BAD_REQUEST` | 400 | UUID de ruta inválido, JSON incompatible o falta empresa en el contexto. |
| `VALIDATION_ERROR` | 400 | IDs, filtros, código, costo, precisión o plazo inválidos. |
| `NOT_FOUND` | 404 | Proveedor, producto o relación inexistente, eliminado o ajeno a la empresa. |
| `CONFLICT` | 409 | Relación duplicada, recurso inactivo al asociar o conflicto de preferencia. |
| `UNAUTHORIZED` | 401 | Token ausente/inválido o sesión inactiva. |
| `FORBIDDEN` | 403 | Falta el permiso de la ruta. |
| `INTERNAL_ERROR` | 500 | Error interno de persistencia. |

Ejemplo de error:

```json
{
  "success": false,
  "error": {
    "code": "NOT_FOUND",
    "message": "product not found"
  }
}
```

---

## 13) Desvincular proveedor y producto — `DELETE /api/v1/suppliers/:id/products/:product_id`

Usado por ambos modales de confirmación. Elimina solo product_suppliers; no elimina ni altera proveedor, producto, stock ni otros productos asociados. Si era preferred=true, el frontend puede mostrar el aviso usando el dato de la lista. No elige un sustituto automáticamente. Repetir la solicitud responde 404.

### Permiso

`suppliers.update`

### Params de ruta

- id: UUID del proveedor.
- product_id: UUID del producto.

### Query params

No aplica.

### Body

No aplica.

### Respuesta exitosa

**`204`**

Sin contenido.

### Códigos de error

| Code | HTTP | Causa |
|---|---|---|
| `BAD_REQUEST` | 400 | UUID de ruta inválido, JSON incompatible o falta empresa en el contexto. |
| `VALIDATION_ERROR` | 400 | IDs, filtros, código, costo, precisión o plazo inválidos. |
| `NOT_FOUND` | 404 | Proveedor, producto o relación inexistente, eliminado o ajeno a la empresa. |
| `UNAUTHORIZED` | 401 | Token ausente/inválido o sesión inactiva. |
| `FORBIDDEN` | 403 | Falta el permiso de la ruta. |
| `INTERNAL_ERROR` | 500 | Error interno de persistencia. |

Ejemplo de error:

```json
{
  "success": false,
  "error": {
    "code": "NOT_FOUND",
    "message": "product not found"
  }
}
```

---

## 14) Listar proveedores de un producto — `GET /api/v1/products/:product_id/suppliers`

El endpoint se implementa íntegramente en el módulo supplier, aunque la URL parte del producto. Incluye proveedores inactivos asociados, excluye eliminados lógicamente. Ordena primero el preferido y luego por nombre e ID.

### Permiso

`suppliers.read`

### Params de ruta

- product_id: UUID del producto de la empresa autenticada.

### Query params

- `search` (opcional, máximo 160): nombre comercial, razón social, NIT o código del proveedor.
- `page` (opcional, por defecto 1): número de página.
- `limit` (opcional, por defecto 20): tamaño de página, máximo 100.

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
      "supplier_code": "DA-ACE-900",
      "lead_time_days": 3,
      "current_cost": 13.1,
      "preferred": false,
      "created_at": "2026-10-16T14:00:00Z",
      "updated_at": null,
      "supplier": {
        "id": "a1111111-1111-4111-8111-111111111111",
        "name": "Distribuidora Andina",
        "legal_name": "Distribuidora Andina S.R.L.",
        "tax_id": "1044556677",
        "main_industry": "Abarrotes",
        "status": "active",
        "lead_time_days": 2
      },
      "effective_lead_time_days": 3
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
| `BAD_REQUEST` | 400 | UUID de ruta inválido, JSON incompatible o falta empresa en el contexto. |
| `VALIDATION_ERROR` | 400 | IDs, filtros, código, costo, precisión o plazo inválidos. |
| `NOT_FOUND` | 404 | Proveedor, producto o relación inexistente, eliminado o ajeno a la empresa. |
| `UNAUTHORIZED` | 401 | Token ausente/inválido o sesión inactiva. |
| `FORBIDDEN` | 403 | Falta el permiso de la ruta. |
| `INTERNAL_ERROR` | 500 | Error interno de persistencia. |

Ejemplo de error:

```json
{
  "success": false,
  "error": {
    "code": "NOT_FOUND",
    "message": "product not found"
  }
}
```

---

## 15) Proveedores para el selector del producto — `GET /api/v1/products/:product_id/supplier-options`

Los ya asociados permanecen en la lista con is_associated=true y selectable=false. Los proveedores inactivos también son no seleccionables. Ordena por nombre e ID. El proveedor recién creado aparece aquí sin asociación; el frontend conoce su ID por la respuesta de creación y puede marcarlo como “Nuevo” en ese modal sin un indicador persistido en el backend.

### Permiso

`suppliers.update`

### Params de ruta

- product_id: UUID del producto de la empresa autenticada.

### Query params

- `search` (opcional, máximo 160): nombre comercial, razón social o NIT.
- `page` (opcional, por defecto 1): número de página.
- `limit` (opcional, por defecto 20): tamaño de página, máximo 100.

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
      "legal_name": "Distribuidora Andina S.R.L.",
      "tax_id": "1044556677",
      "main_industry": "Abarrotes",
      "status": "active",
      "lead_time_days": 2,
      "is_associated": true,
      "selectable": false
    },
    {
      "id": "a7777777-7777-4777-8777-777777777777",
      "name": "Importadora del Valle",
      "legal_name": "Importadora del Valle SRL",
      "tax_id": "1042387654",
      "main_industry": "Abarrotes",
      "status": "active",
      "lead_time_days": 2,
      "is_associated": false,
      "selectable": true
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

### Códigos de error

| Code | HTTP | Causa |
|---|---|---|
| `BAD_REQUEST` | 400 | UUID de ruta inválido, JSON incompatible o falta empresa en el contexto. |
| `VALIDATION_ERROR` | 400 | IDs, filtros, código, costo, precisión o plazo inválidos. |
| `NOT_FOUND` | 404 | Proveedor, producto o relación inexistente, eliminado o ajeno a la empresa. |
| `UNAUTHORIZED` | 401 | Token ausente/inválido o sesión inactiva. |
| `FORBIDDEN` | 403 | Falta el permiso de la ruta. |
| `INTERNAL_ERROR` | 500 | Error interno de persistencia. |

Ejemplo de error:

```json
{
  "success": false,
  "error": {
    "code": "NOT_FOUND",
    "message": "product not found"
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
  "supplier_code": "DA-ACE-900",
  "preferred": false,
  "lead_time_days": 5,
  "current_cost": 13.3,
  "created_at": "2026-10-15T14:05:00Z",
  "updated_at": null
}
```

En el listado de productos asociados se agregan `name`, `sku` (string o null) y `effective_lead_time_days`. El costo es específico del proveedor y puede ser null al asociar en lote. `effective_lead_time_days` usa el plazo de la asociación y, si falta, el plazo general del proveedor; respeta el valor cero. No copia ese valor a la relación.

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
| `NOT_FOUND` | 404 | Proveedor, producto o asociación inexistente, eliminado o fuera de la empresa. |
| `CONFLICT` | 409 | NIT repetido, asociación duplicada, recursos inactivos al asociar o conflicto de preferencia. |
| `INTERNAL_ERROR` | 500 | Error interno; no se exponen detalles SQL. |

Los mensajes actuales del backend están en inglés. El frontend puede traducir por código y usar `details` para marcar campos.

---

## Casos borde / comportamiento no obvio

- Esta etapa agrega siete rutas y reutiliza GET /suppliers/:id/products. No se agregan solicitudes de login ni scripts de Bruno.
- AutoMigrate incorpora supplier_code y preferred a product_suppliers y permite current_cost=null para el alta múltiple. Conserva los datos previos; ejecutar la migración habitual antes de usar estas rutas. No hay archivos SQL nuevos.
- Un producto tiene como máximo un proveedor preferido dentro de su empresa. preferred=true cambia esa preferencia dentro de la misma transacción; si falla la operación se conserva la anterior. El índice parcial garantiza la unicidad.
- Desvincular el preferido deja el producto sin preferido; el frontend solicita al usuario elegir otro cuando corresponda. No se generan órdenes automáticas.
- Los selectores muestran is_associated y selectable. Crear un proveedor no lo asocia: hay que volver al selector y confirmar la asociación.
- La asociación múltiple admite hasta 100 productos distintos y se aplica completa o no se aplica. Código y costo pendientes son null, no cero.
- La edición y desvinculación solo cambian la relación. No cambian el costo/stock del producto ni el resto del catálogo del proveedor.

- En Bruno, definir manualmente `supplierId` en las variables de la carpeta Supplier con el ID obtenido al crear o listar. Las rutas correctas son `PUT /api/v1/suppliers/:id` y `PATCH /api/v1/suppliers/:id/deactivate`. Una variable vacía produce `/suppliers/` o `/suppliers//deactivate`, que no son rutas válidas. No hay scripts para seleccionar o guardar IDs.
- Crear un proveedor no lo asocia a productos. Las asociaciones se crean y modifican exclusivamente mediante sus rutas, nunca al guardar el proveedor.

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
- No se incluyen la pantalla de devoluciones ni módulos de órdenes de compra o cuentas por pagar.

---

## Diferencias respecto a una versión anterior (si aplica)

No aplica: primera versión del contrato del flujo de proveedores en esta rama.
