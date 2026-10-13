# Reporte de endpoints — HU-03-03 Visualización de producto y categoría

## Información general

- HU: HU-03-03 — Visualización de producto y categoría.
- RF relacionado: RF 03. Épica / módulo: Gestión de producto. Sprint: 1.
- Rama: `feature/HU-03-03-HU-03-04-product-management`.
- Prefijo base: `/api/v1`.
- Requiere contexto autenticado del backend: `company_id`, `user_id`, obtenidos del JWT mediante `AuthRequired`.
- Formato de respuesta exitosa: `{ "success": true, "data": ... }`; en listados se agrega `meta: { "page", "limit", "total", "total_pages" }`. Los datos de paginación no están en la raíz.
- Formato de respuesta de error: `{ "success": false, "error": { "code": "CODIGO", "message": "descripción", "details": { ... } } }`; `details` es opcional.
- Alcance implementado: detalle de productos (BE-12 y BE-13) y consulta del historial. La categoría se representa por `category_id`; no existe un GET de detalle o listado de categorías en el router actual.

---

## Resumen rápido

| Método | Ruta | Permiso | Descripción |
|---|---|---|---|
| `GET` | `/api/v1/products/:id` | `products.read` | Detalle individual; actualizado para incluir `stock_total`. |
| `GET` | `/api/v1/products/:id/price-history` | `products.read` | Historial paginado de costo y precio de venta. |

La edición de precios se documenta en [HU-03-04](HU-03-04-endpoints.md), como acción
relacionada con HU-03-02.

---

## Autenticación / permisos (todas las rutas)

| Header | Obligatorio | Descripción |
|---|---|---|
| `Authorization` | Sí | `Bearer <access_token>`, obtenido con `POST /api/v1/auth/login`; debe corresponder a una sesión activa. |
| `Content-Type` | Sí, si se envía un body JSON | `application/json`. Los GET y DELETE no requieren body. |

Permisos usados en esta HU: `products.read`.

La empresa y el usuario se obtienen de los claims `company_id` y `user_id`.
No se aceptan `x-company-id` ni `x-user-id` como reemplazo del JWT y no se utiliza
una empresa predeterminada. La autorización consulta los permisos del rol mediante
`role_permission → permission`. Cada operación restringe el producto a la empresa
autenticada, incluso cuando el usuario tiene permiso para la ruta.

---

## 1) Ver detalle y stock total — `GET /api/v1/products/:id`

Devuelve la información completa del producto, su estado y la cantidad física total.

### Permiso

`products.read`

### Params de ruta

- `id`: string (UUID) — ID del producto de la empresa autenticada.

### Query params

No aplica.

### Body

No aplica; no enviar body.

### Respuesta exitosa

**`200`**

```json
{
  "success": true,
  "data": {
    "id": "e1844a92-0665-45bc-b7fe-a2516fbc458d",
    "company_id": "aaaaaaaa-aaaa-4aaa-8aaa-aaaaaaaaaaaa",
    "category_id": "b45978d0-37f8-4051-92a8-218441dac948",
    "sku": "ABA-210",
    "barcode": "7772000002103",
    "name": "Aceite Fino 900 ml",
    "description": "Botella de 900 ml",
    "unit_of_measure": "unit",
    "cost_price": "13.30",
    "sale_price": "16.90",
    "status": "active",
    "created_at": "2026-10-02T12:00:00Z",
    "updated_at": "2026-10-12T15:00:00Z",
    "stock_total": "8.000"
  }
}
```

### Códigos de error

| Code | HTTP | Causa |
|---|---|---|
| `BAD_REQUEST` | 400 | El `id` de ruta no es un UUID válido. |
| `NOT_FOUND` | 404 | El producto no existe, fue eliminado o pertenece a otra empresa. |
| `BAD_REQUEST` | 400 | El contexto autenticado no contiene empresa (`company is required`). |
| `UNAUTHORIZED` | 401 | Token ausente, inválido o expirado; sesión cerrada o identidad inválida. |
| `FORBIDDEN` | 403 | Usuario sin el permiso requerido. |
| `INTERNAL_ERROR` | 500 | Fallo no controlado de persistencia o de comprobación de permisos. |

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

## 2) Consultar historial de precios — `GET /api/v1/products/:id/price-history`

Devuelve los cambios de costo y precio de venta del producto, con autor, rol, fecha y motivo.

### Permiso

`products.read`

### Params de ruta

- `id`: string (UUID) — ID del producto de la empresa autenticada.

### Query params

- `page` (opcional, por defecto `1`): integer — página solicitada; mínimo efectivo 1.
- `limit` (opcional, por defecto `20`): integer — cantidad por página; rango válido 1–100.


### Body

No aplica; no enviar body.

### Respuesta exitosa

**`200`**

```json
{
  "success": true,
  "data": [
    {
      "id": "c4e9b3a7-3892-4ca2-9abf-b28f0e519e68",
      "user_full_name": "Luis Fernández",
      "user_role": "superadmin",
      "previous_cost": "13.10",
      "new_cost": "13.30",
      "previous_price": "16.50",
      "new_price": "16.90",
      "margin_pct": "21%",
      "reason": "Ajuste por costo",
      "created_at": "2026-10-12T15:00:00Z"
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
| `BAD_REQUEST` | 400 | El `id` de ruta no es un UUID válido. |
| `NOT_FOUND` | 404 | El producto no existe, fue eliminado o pertenece a otra empresa. |
| `BAD_REQUEST` | 400 | El contexto autenticado no contiene empresa (`company is required`). |
| `UNAUTHORIZED` | 401 | Token ausente, inválido o expirado; sesión cerrada o identidad inválida. |
| `FORBIDDEN` | 403 | Usuario sin el permiso requerido. |
| `INTERNAL_ERROR` | 500 | Fallo no controlado de persistencia o de comprobación de permisos. |

Ejemplo de error:

```json
{
  "success": false,
  "error": {
    "code": "BAD_REQUEST",
    "message": "invalid product id"
  }
}
```

---

## Estructura de respuesta de entidad `ProductDetail`

El detalle agrega `stock_total` a la entidad base `Product`.

```json
{
  "id": "e1844a92-0665-45bc-b7fe-a2516fbc458d",
  "company_id": "aaaaaaaa-aaaa-4aaa-8aaa-aaaaaaaaaaaa",
  "category_id": "b45978d0-37f8-4051-92a8-218441dac948",
  "sku": "ABA-210",
  "barcode": "7772000002103",
  "name": "Aceite Fino 900 ml",
  "description": "Botella de 900 ml",
  "unit_of_measure": "unit",
  "cost_price": "13.30",
  "sale_price": "16.90",
  "status": "active",
  "created_at": "2026-10-02T12:00:00Z",
  "updated_at": "2026-10-12T15:00:00Z",
  "stock_total": "8.000"
}
```

| Campo | Tipo | Notas |
|---|---|---|
| `id`, `company_id`, `category_id` | string (UUID) | Identificadores; la categoría no viene expandida. |
| `sku`, `name` | string | SKU y nombre del producto. |
| `barcode`, `description` | string opcional | Se omiten cuando están vacíos; no se envían como `null`. |
| `unit_of_measure` | enum | `unit`, `kg`, `g`, `l`, `ml`, `m`, `box`, `pack`. |
| `cost_price`, `sale_price` | string decimal | Dos decimales; importes en unidad monetaria completa, no centavos. `sale_price` es el precio al cliente. |
| `status` | enum | `active` o `inactive`. |
| `created_at`, `updated_at` | string | Fecha ISO 8601, UTC en la ejecución actual; `updated_at` también tiene valor al crear. |
| `stock_total` | string decimal | Suma de `stock_balance.quantity` de la empresa y el producto, con tres decimales. Solo lo incluye el GET por ID. |

### Entidad `PriceHistoryEntry`

```json
{
  "id": "c4e9b3a7-3892-4ca2-9abf-b28f0e519e68",
  "user_full_name": "Luis Fernández",
  "user_role": "superadmin",
  "previous_cost": "13.10",
  "new_cost": "13.30",
  "previous_price": "16.50",
  "new_price": "16.90",
  "margin_pct": "21%",
  "reason": "Ajuste por costo",
  "created_at": "2026-10-12T15:00:00Z"
}
```

| Campo | Tipo | Notas |
|---|---|---|
| `id` | string (UUID) | ID del registro de historial. |
| `user_full_name`, `user_role` | string | Nombre y rol consultados desde `app_user` y `role` al leer; no son una copia histórica del nombre o rol. |
| `previous_cost`, `new_cost` | string decimal | Costo anterior y nuevo con dos decimales. |
| `previous_price`, `new_price` | string decimal | Precio de venta al cliente anterior y nuevo con dos decimales. |
| `margin_pct` | string | Porcentaje redondeado al entero, con sufijo `%`, por ejemplo `"21%"`. |
| `reason` | string opcional | Se omite si el motivo está vacío. |
| `created_at` | string | Fecha ISO 8601 del cambio. |

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
| `BAD_REQUEST` | 400 | UUID de ruta inválido o contexto sin empresa. |
| `UNAUTHORIZED` | 401 | Token o sesión inválidos; usuario no autenticado. |
| `FORBIDDEN` | 403 | Usuario sin permiso para la operación. |
| `NOT_FOUND` | 404 | Producto eliminado, inexistente o de otra empresa. |
| `INTERNAL_ERROR` | 500 | Error no controlado de base de datos o de autorización. |

---

## Casos borde / comportamiento no obvio

- `stock_total` suma todas las sucursales y ubicaciones de la empresa. No hay filtro por sucursal ni desglose por ubicación en esta ruta.
- La cantidad es stock físico total: incluye las unidades reservadas. No representa `quantity - reserved_quantity`.
- Sin saldos registrados devuelve `"0.000"`. Un error de consulta devuelve 500; no se sustituye por cero.
- La lectura usa la caché `stock_balance`; los movimientos de inventario deberán mantenerla actualizada. Este endpoint no crea ni modifica movimientos o saldos.
- El detalle incluye productos activos e inactivos, con `status` para que el frontend muestre FE-11. Las fechas permiten mostrar registro y última edición; la API no genera indicadores visuales.
- Las relaciones del producto vienen como IDs. No devuelve nombres de categoría, proveedores, lotes, variantes ni ubicaciones.
- El historial se ordena por `created_at DESC`. Si dos registros tienen la misma fecha, no se garantiza un orden adicional entre ellos.
- `page` menor a 1 o no numérico se interpreta como 1. `limit` menor a 1, mayor a 100 o no numérico se reemplaza por 20.
- Historial vacío o página fuera de rango: `data: []`, 200 y `meta.total` real. Sin registros, `meta.total_pages` es 0.
- El margen se calcula como `(new_price - new_cost) / new_price × 100`, redondeado a cero decimales. Si el precio nuevo no es positivo, el cálculo devuelve 0%.
- Un cambio solo de costo también genera historial; en ese caso el precio anterior y nuevo pueden ser iguales. El PATCH de precios y el PUT de edición generan el historial automáticamente.
- Los productos eliminados no se pueden consultar por estas rutas: devuelven 404 aunque sus filas e historial se conserven en la base.
- Las categorías solo cuentan con `POST /api/v1/categories` en el router actual, correspondiente al registro. La visualización de categorías indicada en el título de la HU sigue pendiente como endpoint independiente.

---

## Diferencias respecto a una versión anterior (si aplica)

Esta rama agrega o modifica, sobre la base funcional de HU-03-02 (edición de producto):

1. `GET /api/v1/products/:id` para consulta individual con permiso `products.read`; ahora agrega `stock_total` al contrato original del detalle.
2. `GET /api/v1/products/:id/price-history` para lectura paginada del historial.
3. Persistencia `product_price_history` con costo/precio anterior y nuevo, usuario, motivo y fecha.
4. Lectura de saldos desde `stock_balance` y registro de su modelo en el migrador. Cantidades `numeric(14,3)`; empresa y producto mantienen UUID conforme al código actual.
5. El GET de listado mantiene su contrato sin `stock_total`; no se creó una ruta separada de stock.

El modelo `StockBalanceModel` contiene `stock_balance_id`, `company_id`,
`product_id`, `branch_id`, `location_id`, `quantity`, `reserved_quantity` y
`updated_at`. Las cantidades no pueden ser negativas y la reserva no puede
superar la cantidad física. Empresa y producto usan UUID; `stock_balance_id`,
`branch_id` y `location_id` usan bigint. Una base creada desde el DBML con empresa
y producto bigint requiere reconciliar esos tipos con el catálogo UUID actual.

Para cargar las tablas y el código actual en Docker:

```bash
docker compose -f deployments/docker/docker-compose.yml up -d --build
```

El servicio `app` espera la finalización correcta de `migrate`.
`--build` recompila la aplicación y el migrador; también lo hace `make docker-up`.
Si falta `product_price_history`, la consulta de historial devuelve 500 hasta
ejecutar las migraciones actuales.

