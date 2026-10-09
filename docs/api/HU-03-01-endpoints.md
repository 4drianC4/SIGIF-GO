# Reporte de endpoints — HU-03-01 Registro de producto y categoría

## Información general

- HU: HU-03-01 — Registro de producto y categoría
- Rama: `feature/HU-03-01-product-and-category-registration`
- Prefijo base: `/api/v1`
- Requiere contexto autenticado del backend: `userId` y `companyId`, tomados del access token (`Authorization: Bearer <token>`). No se envían por header ni por body.
- Formato de respuesta exitosa: `{ "success": true, "data": ... }` — en listados paginados se agrega `meta` con `page`, `limit`, `total`, `total_pages`
- Formato de respuesta de error: `{ "success": false, "error": { "code": "CODIGO", "message": "descripción" } }` — en errores de validación se agrega `details` por campo
- Montos de dinero: número JSON con 2 decimales (`4.20`), no string

---

## Resumen rápido

| Método | Ruta | Permiso | Descripción |
|---|---|---|---|
| `GET` | `/api/v1/products` | `products.list` | Lista productos con paginación, búsqueda y filtros. |
| `GET` | `/api/v1/products/summary` | `products.list` | KPIs de las tarjetas superiores. |
| `POST` | `/api/v1/products` | `products.create` | Registra un producto (HU principal). |
| `GET` | `/api/v1/products/validate-duplicate` | `products.create` | Valida en tiempo real si nombre, SKU o código de barras ya existen. |
| `GET` | `/api/v1/products/:id` | `products.read` | Devuelve un producto (HU-03-03/04). |
| `PUT` | `/api/v1/products/:id` | `products.update` | Edita un producto (HU-03-03/04). |
| `PATCH` | `/api/v1/products/:id/status` | `products.update` | Activa o desactiva un producto (HU-03-03/04). |
| `DELETE` | `/api/v1/products/:id` | `products.delete` | Elimina un producto (HU-03-03/04). |
| `GET` | `/api/v1/categories` | `categories.list` | Lista todas las categorías (árbol + select). |
| `POST` | `/api/v1/categories` | `categories.create` | Registra una categoría o subcategoría. |
| `GET` | `/api/v1/units-of-measure` | — (solo sesión) | Lista unidades de medida (semilla). |
| `GET` | `/api/v1/taxes` | — (solo sesión) | Lista impuestos (semilla). |

---

## Autenticación / permisos (todas las rutas)

| Header | Obligatorio | Descripción |
|---|---|---|
| `Authorization` | Sí | `Bearer <access_token>` obtenido con `POST /api/v1/auth/login`. La sesión debe seguir activa: después de `logout` el token responde 401. |
| `Content-Type` | Sí, en `POST` | `application/json` |

- `userId`: claim `user_id` del token. Se usa para verificar el permiso contra `app_user → role → role_permission → permission`.
- `companyId`: claim `company_id` del token. Productos y categorías se leen y guardan solo en esa empresa. Si el usuario no tiene empresa, responde `400 company is required`.

Permisos usados: `products.list`, `products.create`, `products.read`, `products.update`, `products.delete`, `categories.list`, `categories.create`

Todos se crean y se asignan al rol `superadmin` al ejecutar `make migrate-up`, también en bases ya existentes. Los roles `business_admin` y `employee` no los tienen (403). Unidades e impuestos solo requieren sesión.

---

## 1) Listar productos — `GET /api/v1/products`

Lista los productos de la empresa, ordenados por nombre.

### Permiso
`products.list`

### Query params
- `page` (opcional, por defecto `1`): number — página.
- `limit` (opcional, por defecto `20`, máximo `100`): number — elementos por página.
- `search` (opcional): string — busca en nombre, SKU o código de barras, sin distinguir mayúsculas. Máximo 100 caracteres.
- `category_id` (opcional): uuid — filtra por categoría.
- `status` (opcional): enum — `active` | `inactive`.

### Respuesta exitosa

**`200`**
```json
{
  "success": true,
  "data": [
    {
      "id": "550e8400-e29b-41d4-a716-446655440000",
      "name": "Leche PIL entera 1L",
      "sku": "LAC-001",
      "barcode": null,
      "category": {
        "id": "550e8400-e29b-41d4-a716-446655440003",
        "name": "Lácteos"
      },
      "unit_of_measure_id": "bae16597-2960-406b-8dcd-5acab1260164",
      "cost": 5.40,
      "price": 7.00,
      "margin": 23,
      "stock": 496,
      "min_stock": 10,
      "stock_status": "normal",
      "status": "active",
      "created_at": "2026-10-05T12:00:00Z",
      "updated_at": "2026-10-05T12:00:00Z"
    }
  ],
  "meta": {
    "page": 1,
    "limit": 10,
    "total": 1284,
    "total_pages": 129
  }
}
```

### Códigos de error

| Code | HTTP | Causa |
|---|---|---|
| `VALIDATION_ERROR` | 400 | `category_id` no es UUID, `status` no es `active`/`inactive`, `search` demasiado largo. |
| `BAD_REQUEST` | 400 | El usuario no tiene empresa (`company is required`). |
| `UNAUTHORIZED` | 401 | Falta el token, es inválido, expiró o su sesión se cerró. |
| `FORBIDDEN` | 403 | Usuario sin `products.list`. |

---

## 2) Resumen / KPIs — `GET /api/v1/products/summary`

Devuelve los indicadores de las tarjetas superiores, calculados sobre los productos **activos** de la empresa.

### Permiso
`products.list`

### Respuesta exitosa

**`200`**
```json
{
  "success": true,
  "data": {
    "active_products": 1284,
    "inventory_value": 412880.00,
    "low_stock": 7,
    "no_movement_90_days": 23
  }
}
```

| Campo | Cálculo |
|---|---|
| `active_products` | Productos con `status = active`. |
| `inventory_value` | Suma de `stock × cost`. |
| `low_stock` | Productos con `0 < stock <= min_stock`. No incluye los agotados. |
| `no_movement_90_days` | Productos cuyo último movimiento de stock fue hace más de 90 días. Hoy el único movimiento es el stock inicial al crear. |

### Códigos de error

| Code | HTTP | Causa |
|---|---|---|
| `BAD_REQUEST` | 400 | El usuario no tiene empresa. |
| `UNAUTHORIZED` | 401 | Token ausente, inválido, expirado o sesión cerrada. |
| `FORBIDDEN` | 403 | Usuario sin `products.list`. |

---

## 3) Crear producto — `POST /api/v1/products`

Registra un producto activo en la empresa del usuario.

### Permiso
`products.create`

### Body

```json
{
  "name": "Galletas María 200g",
  "sku": "ABA-412",
  "category_id": "550e8400-e29b-41d4-a716-446655440001",
  "unit_of_measure_id": "550e8400-e29b-41d4-a716-446655440002",
  "cost": 4.20,
  "sale_price": 6.00,
  "initial_stock": 0
}
```

| Campo | Tipo | Obligatorio | Restricciones / notas |
|---|---|---|---|
| `name` | string | Sí | 2–150 caracteres después de quitar espacios. Único por empresa, sin distinguir mayúsculas ni espacios repetidos. |
| `sku` | string | No | Máximo 50 caracteres; letras, números, `.`, `-` y `_`, empezando por letra o número. Se guarda en mayúsculas. Único por empresa si se envía. |
| `category_id` | uuid | Sí | Categoría existente, de la empresa y activa. |
| `unit_of_measure_id` | uuid | Sí | Un `id` de `GET /api/v1/units-of-measure`. |
| `cost` | number | Sí | > 0, máximo 2 decimales. |
| `sale_price` | number | Sí | > 0, máximo 2 decimales. |
| `initial_stock` | number | No | >= 0, máximo 3 decimales. Por defecto `0`. |
| `barcode` | string | No | Extra al contrato. Solo dígitos, 8–14 caracteres. Único por empresa si se envía. |
| `description` | string | No | Extra al contrato. Máximo 1000 caracteres. |
| `min_stock` | number | No | Extra al contrato. >= 0, máximo 3 decimales. Por defecto `0`. Umbral para `low_stock`. |

### Respuesta exitosa

**`201`** — mismo shape que un elemento de `GET /api/v1/products`:
```json
{
  "success": true,
  "data": {
    "id": "2a69907f-e4ac-414e-b547-31a499b256cb",
    "name": "Galletas María 200g",
    "sku": "ABA-412",
    "barcode": null,
    "category": { "id": "550e8400-e29b-41d4-a716-446655440001", "name": "Abarrotes" },
    "unit_of_measure_id": "550e8400-e29b-41d4-a716-446655440002",
    "cost": 4.20,
    "price": 6.00,
    "margin": 30,
    "stock": 0,
    "min_stock": 0,
    "stock_status": "out_of_stock",
    "status": "active",
    "created_at": "2026-10-05T12:00:00Z",
    "updated_at": "2026-10-05T12:00:00Z"
  }
}
```

### Códigos de error

| Code | HTTP | Causa |
|---|---|---|
| `BAD_REQUEST` | 400 | Body no es JSON válido o un número no es numérico; usuario sin empresa; categoría inactiva. |
| `VALIDATION_ERROR` | 400 | Campo obligatorio faltante o con formato inválido; `cost`/`sale_price` <= 0; más decimales de los permitidos; stock negativo; SKU con caracteres no permitidos. Ver `details`. |
| `UNAUTHORIZED` | 401 | Token ausente, inválido, expirado o sesión cerrada. |
| `FORBIDDEN` | 403 | Usuario sin `products.create`. |
| `NOT_FOUND` | 404 | `category not found` (no existe o es de otra empresa) / `unit of measure not found`. |
| `CONFLICT` | 409 | `product name already exists` / `SKU already exists` / `barcode already exists`. |

Ejemplo de error:
```json
{
  "success": false,
  "error": {
    "code": "CONFLICT",
    "message": "product name already exists"
  }
}
```

---

## 4) Validar duplicados en tiempo real — `GET /api/v1/products/validate-duplicate`

Indica si un nombre, SKU o código de barras ya existe en la empresa, para avisar en el formulario antes de enviar.

### Permiso
`products.create`

### Query params
- `name` (opcional): string — se compara sin distinguir mayúsculas ni espacios repetidos.
- `sku` (opcional): string — se compara en mayúsculas.
- `barcode` (opcional): string.
- `exclude_id` (opcional): uuid — producto que se está editando; se ignora en la comparación para que no se detecte como duplicado de sí mismo.

Al menos uno de `name`, `sku` o `barcode` es obligatorio. Si hay varios repetidos, `field` indica el primero en este orden: `name`, `sku`, `barcode`.

### Respuesta exitosa

**`200`**
```json
{
  "success": true,
  "data": {
    "exists": true,
    "field": "name"
  }
}
```

Sin duplicados: `{ "exists": false, "field": null }`.

### Códigos de error

| Code | HTTP | Causa |
|---|---|---|
| `BAD_REQUEST` | 400 | No se envió `name`, `sku` ni `barcode` (`name, sku or barcode is required`) o el usuario no tiene empresa. |
| `VALIDATION_ERROR` | 400 | `exclude_id` no es UUID o un parámetro supera su longitud máxima. |
| `UNAUTHORIZED` | 401 | Token ausente, inválido, expirado o sesión cerrada. |
| `FORBIDDEN` | 403 | Usuario sin `products.create`. |

---

## 5) Listar categorías — `GET /api/v1/categories`

Lista todas las categorías de la empresa, ordenadas por nombre. No está paginado.

### Permiso
`categories.list`

### Respuesta exitosa

**`200`**
```json
{
  "success": true,
  "data": [
    {
      "id": "550e8400-e29b-41d4-a716-446655440001",
      "name": "Abarrotes",
      "parent_id": null,
      "products_count": 412,
      "default_tax": "IVA general 13%",
      "target_margin": 25,
      "status": "active"
    },
    {
      "id": "550e8400-e29b-41d4-a716-446655440003",
      "name": "Aceites",
      "parent_id": "550e8400-e29b-41d4-a716-446655440001",
      "products_count": 0,
      "default_tax": null,
      "target_margin": null,
      "status": "active"
    }
  ]
}
```

> Si una subcategoría tiene `default_tax` o `target_margin` en `null`, el frontend muestra el valor heredado del padre.

### Códigos de error

| Code | HTTP | Causa |
|---|---|---|
| `BAD_REQUEST` | 400 | El usuario no tiene empresa. |
| `UNAUTHORIZED` | 401 | Token ausente, inválido, expirado o sesión cerrada. |
| `FORBIDDEN` | 403 | Usuario sin `categories.list`. |

---

## 6) Crear categoría — `POST /api/v1/categories`

Registra una categoría raíz o una subcategoría.

### Permiso
`categories.create`

### Body

```json
{
  "name": "Congelados",
  "parent_id": null,
  "default_tax": "IVA general 13%",
  "target_margin": 25
}
```

| Campo | Tipo | Obligatorio | Restricciones / notas |
|---|---|---|---|
| `name` | string | Sí | 2–100 caracteres. Único dentro del mismo nivel (mismo `parent_id`) en la empresa, sin distinguir mayúsculas. |
| `parent_id` | uuid \| null | No | Categoría padre existente, de la empresa y activa. `null` = categoría raíz. |
| `default_tax` | string \| null | No | Nombre exacto de un impuesto de `GET /api/v1/taxes` (sin distinguir mayúsculas). `null` = hereda del padre. |
| `target_margin` | number \| null | No | 0–100, máximo 2 decimales. `null` = hereda del padre. |
| `description` | string | No | Extra al contrato. Máximo 500 caracteres. |

### Respuesta exitosa

**`201`** — mismo shape que un elemento de `GET /api/v1/categories`:
```json
{
  "success": true,
  "data": {
    "id": "550e8400-e29b-41d4-a716-446655440015",
    "name": "Congelados",
    "parent_id": null,
    "products_count": 0,
    "default_tax": "IVA general 13%",
    "target_margin": 25,
    "status": "active"
  }
}
```

### Códigos de error

| Code | HTTP | Causa |
|---|---|---|
| `BAD_REQUEST` | 400 | Body no es JSON válido; usuario sin empresa; categoría padre inactiva. |
| `VALIDATION_ERROR` | 400 | `name` vacío o fuera de rango, `parent_id` no es UUID, `target_margin` fuera de 0–100. |
| `UNAUTHORIZED` | 401 | Token ausente, inválido, expirado o sesión cerrada. |
| `FORBIDDEN` | 403 | Usuario sin `categories.create`. |
| `NOT_FOUND` | 404 | `parent category not found` / `tax not found`. |
| `CONFLICT` | 409 | `category name already exists` (mismo nombre en el mismo nivel). |

---

## 7) Listar unidades de medida — `GET /api/v1/units-of-measure`

Devuelve la lista sembrada por la migración. No hay CRUD en el Sprint 1.

### Permiso
`—` (solo sesión activa)

### Respuesta exitosa

**`200`**
```json
{
  "success": true,
  "data": [
    { "id": "uuid", "name": "Unit", "abbreviation": "UNIT" },
    { "id": "uuid", "name": "Box", "abbreviation": "BOX" },
    { "id": "uuid", "name": "Pack", "abbreviation": "PACK" },
    { "id": "uuid", "name": "Kilogram", "abbreviation": "KG" },
    { "id": "uuid", "name": "Liter", "abbreviation": "L" },
    { "id": "uuid", "name": "Dozen", "abbreviation": "DOZ" }
  ]
}
```

Los `id` se generan al sembrar; el frontend debe tomarlos de esta respuesta.

### Códigos de error

| Code | HTTP | Causa |
|---|---|---|
| `UNAUTHORIZED` | 401 | Token ausente, inválido, expirado o sesión cerrada. |

---

## 8) Listar impuestos — `GET /api/v1/taxes`

Devuelve la lista sembrada por la migración.

### Permiso
`—` (solo sesión activa)

### Respuesta exitosa

**`200`**
```json
{
  "success": true,
  "data": [
    { "id": "uuid", "name": "IVA general 13%", "percentage": 13 },
    { "id": "uuid", "name": "IVA reducido 5%", "percentage": 5 },
    { "id": "uuid", "name": "Exempt", "percentage": 0 }
  ]
}
```

### Códigos de error

| Code | HTTP | Causa |
|---|---|---|
| `UNAUTHORIZED` | 401 | Token ausente, inválido, expirado o sesión cerrada. |

---

## 9) Obtener producto — `GET /api/v1/products/:id`

Devuelve un producto de la empresa del usuario. Endpoint de HU-03-03/04, adaptado al modelo de esta HU.

### Permiso
`products.read`

### Params de ruta
- `id`: uuid — id del producto.

### Respuesta exitosa

**`200`** — `data` con el shape de `Product` (igual que un elemento de `GET /api/v1/products`).

### Códigos de error

| Code | HTTP | Causa |
|---|---|---|
| `BAD_REQUEST` | 400 | `id` no es UUID (`invalid product id`) o el usuario no tiene empresa. |
| `UNAUTHORIZED` | 401 | Token ausente, inválido, expirado o sesión cerrada. |
| `FORBIDDEN` | 403 | Usuario sin `products.read`. |
| `NOT_FOUND` | 404 | `product not found` (no existe, fue eliminado o es de otra empresa). |

---

## 10) Editar producto — `PUT /api/v1/products/:id`

Reemplaza los datos editables de un producto. Endpoint de HU-03-03/04, adaptado al modelo de esta HU.

### Permiso
`products.update`

### Params de ruta
- `id`: uuid — id del producto.

### Body

```json
{
  "name": "Galletas María 200g",
  "sku": "ABA-412",
  "category_id": "550e8400-e29b-41d4-a716-446655440001",
  "unit_of_measure_id": "550e8400-e29b-41d4-a716-446655440002",
  "cost": 4.50,
  "sale_price": 7.50,
  "min_stock": 5
}
```

| Campo | Tipo | Obligatorio | Restricciones / notas |
|---|---|---|---|
| `name` | string | Sí | Mismas reglas que al crear. Único por empresa, sin contar el propio producto. |
| `sku` | string | No | Mismas reglas que al crear. Si se omite, el producto queda sin SKU. |
| `category_id` | uuid | Sí | Categoría existente, de la empresa y activa. |
| `unit_of_measure_id` | uuid | Sí | Un `id` de `GET /api/v1/units-of-measure`. |
| `cost` | number | Sí | > 0, máximo 2 decimales. |
| `sale_price` | number | Sí | > 0, máximo 2 decimales. |
| `barcode` | string | No | Solo dígitos, 8–14 caracteres. Si se omite, el producto queda sin código de barras. |
| `description` | string | No | Máximo 1000 caracteres. Si se omite, queda vacía. |
| `min_stock` | number | No | >= 0, máximo 3 decimales. Si se omite, conserva el valor actual. |

`stock` no se edita por este endpoint: solo cambia con movimientos de inventario.

### Respuesta exitosa

**`200`** — `data` con el shape de `Product` ya actualizado (`margin` y `stock_status` recalculados).

### Códigos de error

| Code | HTTP | Causa |
|---|---|---|
| `BAD_REQUEST` | 400 | Body no es JSON válido; `id` no es UUID; usuario sin empresa; categoría inactiva. |
| `VALIDATION_ERROR` | 400 | Mismas validaciones que al crear. Ver `details`. |
| `UNAUTHORIZED` | 401 | Token ausente, inválido, expirado o sesión cerrada. |
| `FORBIDDEN` | 403 | Usuario sin `products.update`. |
| `NOT_FOUND` | 404 | `product not found` / `category not found` / `unit of measure not found`. |
| `CONFLICT` | 409 | `product name already exists` / `SKU already exists` / `barcode already exists` en otro producto. |

---

## 11) Activar o desactivar producto — `PATCH /api/v1/products/:id/status`

Cambia el estado de un producto. Endpoint de HU-03-03/04.

### Permiso
`products.update`

### Params de ruta
- `id`: uuid — id del producto.

### Body

```json
{
  "status": "inactive"
}
```

| Campo | Tipo | Obligatorio | Restricciones / notas |
|---|---|---|---|
| `status` | enum | Sí | `active` \| `inactive` |

### Respuesta exitosa

**`200`** — `data` con el shape de `Product` y el nuevo `status`.

### Códigos de error

| Code | HTTP | Causa |
|---|---|---|
| `BAD_REQUEST` | 400 | `status` distinto de `active`/`inactive`; el producto ya está en ese estado (`product is already active` / `product is already inactive`); `id` no es UUID; usuario sin empresa. |
| `UNAUTHORIZED` | 401 | Token ausente, inválido, expirado o sesión cerrada. |
| `FORBIDDEN` | 403 | Usuario sin `products.update`. |
| `NOT_FOUND` | 404 | `product not found`. |

---

## 12) Eliminar producto — `DELETE /api/v1/products/:id`

Elimina un producto con borrado lógico. Endpoint de HU-03-03/04.

### Permiso
`products.delete`

### Params de ruta
- `id`: uuid — id del producto.

### Respuesta exitosa

**`204`** — sin cuerpo.

### Códigos de error

| Code | HTTP | Causa |
|---|---|---|
| `BAD_REQUEST` | 400 | `id` no es UUID o el usuario no tiene empresa. |
| `UNAUTHORIZED` | 401 | Token ausente, inválido, expirado o sesión cerrada. |
| `FORBIDDEN` | 403 | Usuario sin `products.delete`. |
| `NOT_FOUND` | 404 | `product not found` (incluye un producto ya eliminado). |

---

## Estructura de respuesta de entidad `Product`

```json
{
  "id": "uuid",
  "name": "string",
  "sku": "string | null",
  "barcode": "string | null",
  "description": "string (se omite si está vacío)",
  "category": { "id": "uuid", "name": "string" },
  "unit_of_measure_id": "uuid",
  "cost": 0.00,
  "price": 0.00,
  "margin": 0,
  "stock": 0,
  "min_stock": 0,
  "stock_status": "normal | low_stock | out_of_stock",
  "status": "active | inactive",
  "created_at": "date (RFC3339, UTC)",
  "updated_at": "date (RFC3339, UTC)"
}
```

## Estructura de respuesta de entidad `Category`

```json
{
  "id": "uuid",
  "name": "string",
  "parent_id": "uuid | null",
  "description": "string (se omite si está vacío)",
  "products_count": 0,
  "default_tax": "string | null",
  "target_margin": "number | null",
  "status": "active | inactive"
}
```

---

## Tabla de errores consolidada

```json
{ "success": false, "error": { "code": "CODIGO", "message": "descripción" } }
```

| Code | HTTP | Origen |
|---|---|---|
| `BAD_REQUEST` | 400 | JSON inválido, usuario sin empresa, categoría inactiva, `validate-duplicate` sin parámetros, id de producto inválido, estado inválido o repetido |
| `VALIDATION_ERROR` | 400 | Falla de validación de body o query (incluye `details` por campo) |
| `UNAUTHORIZED` | 401 | Token ausente, inválido, expirado o sesión cerrada |
| `FORBIDDEN` | 403 | Usuario sin el permiso del endpoint |
| `NOT_FOUND` | 404 | Producto, categoría, categoría padre, unidad de medida o impuesto inexistente |
| `CONFLICT` | 409 | Nombre de producto, SKU, código de barras o nombre de categoría repetido |
| `INTERNAL_ERROR` | 500 | Error no controlado |

---

## Casos borde / comportamiento no obvio

- **Margen:** `margin = round((sale_price − cost) / sale_price × 100)`, entero. Con costo 5.40 y precio 7.00 da `23`. Puede ser negativo si el costo supera el precio.
- **`stock_status`:** `out_of_stock` si `stock <= 0`; `low_stock` si `stock <= min_stock`; si no, `normal`. Como `min_stock` vale `0` por defecto, un producto sin umbral nunca está en `low_stock`.
- **Nombres en entrada y salida:** al crear se envía `sale_price`, pero en las respuestas el campo se llama `price`, como en el contrato.
- **Normalización:** el SKU se guarda en mayúsculas (`aba-412` → `ABA-412`); en los nombres se quitan los espacios sobrantes. La respuesta devuelve el valor ya normalizado.
- **Unicidad por empresa:** nombre de producto, SKU y código de barras se pueden repetir en otra empresa. El nombre de categoría solo es único dentro del mismo nivel: "Aceites" puede existir como raíz y como hija de "Abarrotes".
- **Borrado lógico:** un producto o categoría borrado libera su nombre, SKU o código de barras.
- **Peticiones simultáneas:** dos altas iguales a la vez crean una sola; la otra recibe `409`, no `500` (índices únicos parciales en la base de datos).
- **Dinero:** se guarda como `numeric(12,2)` y se devuelve como número con 2 decimales exactos (`4.20`). Cantidades (`stock`, `percentage`, `target_margin`) se devuelven sin ceros de relleno (`496`, `13`, `25`).
- **Paginación:** si `limit` es menor que 1 o mayor que 100, se usa `20`; si `page` es menor que 1, se usa `1`. Una página fuera de rango devuelve `data: []` con el `meta` correcto.
- **Relaciones:** el producto trae la categoría expandida (`id` + `name`); la unidad viene solo como `unit_of_measure_id`.
- **Herencia de categorías:** el backend guarda `null` cuando una subcategoría no define impuesto o margen; el frontend resuelve el valor heredado del padre.
- **Idioma de mensajes:** los `message` y los `details` del validador están en inglés, igual que en el resto de la API.

---

## Diferencias respecto a una versión anterior

Sobre la primera versión de la HU que estaba en `dev` (solo `POST /products` y `POST /categories`):
1. Endpoints nuevos: `GET /products`, `GET /products/summary`, `GET /products/validate-duplicate`, `GET /categories`, `GET /units-of-measure`, `GET /taxes`.
2. `POST /products` adaptado al contrato: `cost` (antes `cost_price`), `unit_of_measure_id` (antes un enum `unit_of_measure`), `initial_stock`, `sku` opcional y `name` único. Las respuestas incluyen `price`, `margin`, `stock` y `stock_status`.
3. `POST /categories` acepta `parent_id`, `default_tax` y `target_margin`; el nombre ahora es único por nivel y no por empresa.
4. El dinero se devuelve como número (`4.20`) en lugar de string (`"4.20"`).
5. Tablas nuevas `units_of_measure` y `taxes`, sembradas por `make migrate-up`; permisos nuevos `products.list` y `categories.list`.
6. El esquema del catálogo se cambia con migraciones SQL versionadas, que convierten las bases existentes sin perder datos.

Sobre los endpoints de HU-03-03/04 que ya estaban en `dev` (`GET`, `PUT`, `PATCH .../status` y `DELETE` de `/products/:id`):
1. `PUT /products/:id` usa los mismos campos que `POST /products`: `cost` y `unit_of_measure_id` en lugar de `cost_price` y `unit_of_measure`; `sku` es opcional; el nombre debe ser único; acepta `min_stock`.
2. Todas las respuestas usan el shape de `Product` de este documento (`price`, `margin`, `stock`, `stock_status`, `category` expandida).
3. El listado `GET /products` filtra con `search` (nombre, SKU o código de barras) en lugar de `name`, y requiere `products.list` en lugar de `products.read`.
4. Colección Bruno `bruno/Products`: 21 peticiones con tests (001–016 de esta HU, 017–021 de HU-03-03/04).
