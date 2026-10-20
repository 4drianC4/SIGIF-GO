# Reporte de endpoints — HU-03-02 Edición de producto y categoría

> **Nota de integración (`feature/integration-product`).** Este documento se redactó contra el catálogo anterior a HU-03-01. Al integrarlo con `dev`, el contrato del **producto** cambió como indica la tabla. Rutas, permisos, códigos de error, `stock_total`, historial de precios y paginación no cambiaron. La forma completa del producto está en [HU-03-01](HU-03-01-endpoints.md).
>
> | En este documento | Contrato integrado |
> |---|---|
> | `cost_price` (peticiones y respuestas) | `cost` |
> | `sale_price` en **respuestas** de producto | `price` (en **peticiones** sigue siendo `sale_price`) |
> | `unit_of_measure` (texto: `unit`, `kg`…) | `unit_of_measure_id` (uuid; ver `GET /units-of-measure`) |
> | El producto trae `company_id` y `category_id` | Trae `category: { id, name }`, `margin`, `stock`, `min_stock` y `stock_status` |
> | El costo puede ser `0` | El costo debe ser mayor que 0. `PUT /products/:id` y `PATCH /products/:id/price` validan `cost` y `sale_price` con la misma regla |
> | `PUT /products/:id` sin stock mínimo | Acepta `min_stock` opcional |
>
> `stock_total` (suma de `stock_balance`) y `stock` (cantidad guardada en el propio producto) son valores independientes hasta que los movimientos de inventario mantengan ambos.


## Información general

- HU: HU-03-02 — Edición de producto y categoría
- Rama: `feature/HU-03-02-product-category-edition`
- Prefijo base: `/api/v1`
- Autenticación: todas las rutas requieren `Authorization: Bearer <access_token>` obtenido con `POST /api/v1/auth/login`. El middleware global `AuthRequired` valida el token y que la sesión siga activa (tras `logout` el token deja de servir).
- Empresa: los productos pertenecen a la empresa del claim `company_id` del token. Un usuario sin empresa recibe `400 company is required`.
- Formato de respuesta exitosa: `{ "success": true, "data": ... }`.
- Formato de respuesta de error: `{ "success": false, "error": { "code": "CODIGO", "message": "descripción", "details": { "campo": "motivo" } } }`. `details` solo aparece en errores de validación.

---

## Resumen rápido

| Método | Ruta | Permiso | Descripción |
|---|---|---|---|
| `PUT` | `/api/v1/products/:id` | `products.update` | Edita todos los campos de un producto existente. |

---

## Autenticación / permisos (todas las rutas)

| Header | Obligatorio | Descripción |
|---|---|---|
| `Authorization` | Sí | `Bearer <access_token>` de una sesión activa. |
| `Content-Type` | Sí | `application/json` |

Permisos usados en esta HU:

| Permiso | Endpoint | Roles sembrados que lo tienen |
|---|---|---|
| `products.update` | `PUT /api/v1/products/:id` | `superadmin`, `business_admin` (solo su empresa) |

---

## 1) Editar producto — `PUT /api/v1/products/:id`

Reemplaza todos los campos editables de un producto existente. La edición es completa: todos los campos obligatorios deben enviarse aunque no cambien.

### Permiso
`products.update`

### Params de ruta
- `id`: uuid — ID del producto a editar.

### Body

```json
{
  "category_id": "b45978d0-37f8-4051-92a8-218441dac948",
  "sku": "pepsi-500",
  "barcode": "7750182000456",
  "name": "Pepsi 500ml",
  "description": "Botella PET 500ml",
  "unit_of_measure": "unit",
  "cost_price": 2.50,
  "sale_price": "4.00"
}
```

| Campo | Tipo | Obligatorio | Restricciones / notas |
|---|---|---|---|
| `category_id` | uuid | Sí | Debe existir, pertenecer a la empresa y estar activa. |
| `sku` | string | Sí | Máx. 50 caracteres. Solo letras, números, `.`, `-` y `_`; debe empezar por letra o número. **Se guarda en mayúsculas** (`pepsi-500` → `PEPSI-500`). **Único por empresa**, excluyendo el propio producto. |
| `barcode` | string | No | Solo dígitos, 8–14 caracteres (EAN-8, UPC-A, EAN-13, GTIN-14). **Único por empresa** si se envía, excluyendo el propio producto. |
| `name` | string | Sí | 2–150 caracteres tras quitar espacios. Los espacios repetidos se colapsan. |
| `description` | string | No | Máx. 1000 caracteres. |
| `unit_of_measure` | enum | Sí | `unit` \| `kg` \| `g` \| `l` \| `ml` \| `m` \| `box` \| `pack` |
| `cost_price` | number \| string | No | ≥ 0, máx. 2 decimales. Por defecto `0` si se omite. |
| `sale_price` | number \| string | Sí | > 0, máx. 2 decimales. |

### Respuesta exitosa

**`200`**
```json
{
  "success": true,
  "data": {
    "id": "e1844a92-0665-45bc-b7fe-a2516fbc458d",
    "company_id": "aaaaaaaa-aaaa-4aaa-8aaa-aaaaaaaaaaaa",
    "category_id": "b45978d0-37f8-4051-92a8-218441dac948",
    "sku": "PEPSI-500",
    "barcode": "7750182000456",
    "name": "Pepsi 500ml",
    "description": "Botella PET 500ml",
    "unit_of_measure": "unit",
    "cost_price": "2.50",
    "sale_price": "4.00",
    "status": "active",
    "created_at": "2026-10-02T23:40:27Z",
    "updated_at": "2026-10-07T14:00:00Z"
  }
}
```

### Ejemplo de error de validación

**`400`**
```json
{
  "success": false,
  "error": {
    "code": "VALIDATION_ERROR",
    "message": "validation failed",
    "details": {
      "name": "name is a required field",
      "sale_price": "sale_price is a required field",
      "unit_of_measure": "unit_of_measure must be one of [unit kg g l ml m box pack]"
    }
  }
}
```

### Códigos de error

| Code | HTTP | Causa |
|---|---|---|
| `BAD_REQUEST` | 400 | El `id` de ruta no es un UUID válido. |
| `BAD_REQUEST` | 400 | El body no es JSON válido o un precio no es numérico (`"abc"`). |
| `VALIDATION_ERROR` | 400 | Campos obligatorios vacíos o con formato inválido; `sale_price` ≤ 0; `cost_price` negativo; precio con más de 2 decimales; SKU con caracteres no permitidos. Ver `details`. |
| `BAD_REQUEST` | 400 | La categoría existe pero está inactiva (`category is inactive`). |
| `BAD_REQUEST` | 400 | El token no tiene empresa (`company is required`). |
| `UNAUTHORIZED` | 401 | Falta el token, es inválido, expiró o su sesión se cerró. |
| `FORBIDDEN` | 403 | El usuario no tiene el permiso `products.update`. |
| `NOT_FOUND` | 404 | El producto no existe o no pertenece a la empresa del usuario. |
| `NOT_FOUND` | 404 | La categoría indicada no existe en la empresa. |
| `CONFLICT` | 409 | Otro producto de la empresa ya usa ese SKU o ese código de barras. |
| `INTERNAL_ERROR` | 500 | Error no controlado. |

---

## Estructura de respuesta de entidad `Product`

Referencia del shape completo para tipar en el frontend (TS/interfaces).

```json
{
  "id": "uuid",
  "company_id": "uuid",
  "category_id": "uuid",
  "sku": "string",
  "barcode": "string | omitido si vacío",
  "name": "string",
  "description": "string | omitido si vacío",
  "unit_of_measure": "unit | kg | g | l | ml | m | box | pack",
  "cost_price": "string (decimal con 2 cifras)",
  "sale_price": "string (decimal con 2 cifras)",
  "status": "active | inactive",
  "created_at": "ISO 8601",
  "updated_at": "ISO 8601"
}
```

---

## Tabla de errores consolidada

| Code | HTTP | Origen |
|---|---|---|
| `BAD_REQUEST` | 400 | UUID de ruta inválido, body malformado, precio no numérico, categoría inactiva, token sin empresa. |
| `VALIDATION_ERROR` | 400 | Campos obligatorios faltantes o con formato inválido. Incluye `details` con el campo afectado. |
| `UNAUTHORIZED` | 401 | Token ausente, inválido, expirado o sesión cerrada. |
| `FORBIDDEN` | 403 | Usuario sin el permiso `products.update`. |
| `NOT_FOUND` | 404 | Producto o categoría inexistente o de otra empresa. |
| `CONFLICT` | 409 | SKU o código de barras ya usado por otro producto de la empresa. |
| `INTERNAL_ERROR` | 500 | Error no controlado en el servidor. |

---

## Casos borde / comportamiento no obvio

- **Edición con el mismo SKU**: editar un producto enviando su propio SKU actual no genera conflicto `409`. La validación de unicidad excluye explícitamente el propio producto.
- **Edición con el mismo barcode**: igual que el SKU, enviar el código de barras propio no genera conflicto.
- **SKU normalizado**: el SKU siempre se guarda en mayúsculas. `pepsi-500` y `PEPSI-500` son el mismo SKU. La respuesta siempre devuelve la versión en mayúsculas.
- **`cost_price` omitido**: si no se envía `cost_price`, se interpreta como `0.00`. No es un error.
- **Precios como string o number**: `cost_price` y `sale_price` aceptan número (`4.0`) o string (`"4.00"`). En la respuesta siempre se devuelven como **string con 2 decimales** (`"4.00"`).
- **`status` no es editable por este endpoint**: el campo `status` del producto no se puede cambiar con `PUT`. Sigue siendo `active` tras cualquier edición. Para cambiar el estado usar `PATCH /products/:id/status` (HU-03-04).
- **`created_at` no cambia**: la edición solo actualiza `updated_at`. El `created_at` original se preserva siempre.
- **`barcode` y `description` opcionales**: si se omiten o se envían vacíos, el campo queda vacío en BD y se omite del JSON de respuesta (`omitempty`).
- **Categoría inactiva**: si la categoría a la que se reasigna el producto está inactiva, el backend devuelve `400 category is inactive`, no `404`.

---

## Diferencias respecto a una versión anterior

Esta rama agrega, sobre la base de `feature/HU-03-01-product-and-category-registration`:
1. `PUT /api/v1/products/:id` — edición completa de un producto existente.
2. Permiso nuevo `products.update` en `AllPermissions()`, asignado automáticamente al `superadmin` al correr `make migrate-up`.
3. Métodos nuevos en `ProductRepository`: `Update`, `GetByID`, `ExistsBySKUExcluding`, `ExistsByBarcodeExcluding`.
