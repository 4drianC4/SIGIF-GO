# Reporte de endpoints — HU-04 Listado de productos y stock

## Información general

- Prefijo base: `/api/v1`.
- La empresa se obtiene del token; solo se devuelven datos de esa empresa.
- El stock efectivo de un producto sale de la tabla `stock_balance` (suma por producto). Mientras un producto no tenga filas en `stock_balance`, se usa `products.stock` como respaldo. Así, cualquier movimiento de inventario se refleja de inmediato en el listado y en el resumen.
- **Contrato decimal:** se conserva el formato numérico JSON vigente para los campos decimales de productos. En las peticiones el backend acepta tanto número como string.
- Colección Bruno: `bruno/Products/014`, `015`, `035`–`040`.

## Listado de productos

- Ruta: `GET /api/v1/products`.
- Permiso: `products.list`.

### Parámetros

| Parámetro | Requerido | Descripción |
|---|---:|---|
| `page` | No | Página (por defecto 1). |
| `limit` | No | Tamaño de página (por defecto 20, máximo 100). |
| `search` | No | Busca en nombre, SKU y código de barras. |
| `category_id` | No | UUID de categoría. |
| `status` | No | `active` o `inactive`. |
| `stock_status` | No | `normal`, `low_stock` o `out_of_stock` (ver más abajo). |

### `stock_status`

Se calcula sobre el stock efectivo:

| Valor | Regla |
|---|---|
| `out_of_stock` | stock ≤ 0 |
| `low_stock` | stock > 0 y stock ≤ `min_stock` |
| `normal` | stock > `min_stock` |

Mandar `stock_status=low_stock` devuelve solo los productos bajo mínimo; `out_of_stock` los agotados; `normal` el resto.

**Criterio 4 (bajo o nulo stock):** el listado identifica ambos casos con dos estados distintos —`low_stock` y `out_of_stock`— y se pueden filtrar por separado. El umbral es `stock <= min_stock` (incluye el caso exactamente igual al mínimo). En el resumen, `low_stock` cuenta únicamente los productos con saldo **positivo** bajo mínimo; los agotados no se suman ahí y se consultan con `stock_status=out_of_stock`.

### Respuesta (fragmento)

```json
{
  "success": true,
  "data": [
    {
      "id": "6e436ff1-30d4-4572-b63c-c5ba2f60996b",
      "name": "Galletas María 200g",
      "sku": "ABA-412",
      "stock": 2,
      "min_stock": 5,
      "stock_status": "low_stock",
      "status": "active"
    }
  ],
  "meta": { "page": 1, "limit": 20, "total": 1, "total_pages": 1 }
}
```

Cuando la lista queda vacía, `data: []` y `total: 0`, y la respuesta incluye un campo `message` explicativo (criterio 5):

- `"No hay productos registrados."` si la empresa no tiene productos.
- `"No se encontraron productos para los filtros aplicados."` si hay filtros activos y ninguno coincide.

El campo `message` se omite cuando hay resultados.

## Resumen de indicadores

- Ruta: `GET /api/v1/products/summary`.
- Permiso: `products.list`.
- Devuelve `active_products`, `inventory_value` (stock efectivo × costo), `low_stock` y `no_movement_90_days`. Solo cuenta productos activos.

## Movimientos de inventario (criterio 6)

Módulo `inventory`. El stock cambia mediante movimientos; el listado lo refleja al instante.

### Registrar movimiento

- Ruta: `POST /api/v1/inventory/movements`.
- Permiso: `inventory.create`.

| Campo | Requerido | Descripción |
|---|---:|---|
| `product_id` | Sí | UUID del producto. |
| `type` | Sí | `in`, `out` o `adjustment`. |
| `quantity` | Sí | Magnitud para `in`/`out`; delta firmado para `adjustment`. |
| `branch_id` | No | Sucursal. Si se omite (o `location_id`) se resuelve una sucursal por defecto. |
| `location_id` | No | Ubicación dentro de la sucursal. |
| `reason` | No | Motivo (máx. 200). |

```json
{
  "success": true,
  "data": {
    "movement": { "id": "...", "type": "in", "quantity": "10", "branch_id": 1, "location_id": 1 },
    "balance": "10"
  }
}
```

Efectos en una única transacción: actualiza (o crea) la fila de `stock_balance`, registra el movimiento en `stock_movements` y actualiza `products.last_movement_at`.

### Consultas

| Ruta | Permiso | Descripción |
|---|---|---|
| `GET /api/v1/inventory/branches` | `inventory.list` | Sucursales con sus ubicaciones. |
| `GET /api/v1/inventory/movements?product_id=...` | `inventory.list` | Historial paginado de movimientos. |
| `GET /api/v1/inventory/stock?product_id=...` | `inventory.read` | Saldo por sucursal/ubicación. |

## Errores

| HTTP | Código | Situación |
|---:|---|---|
| 400 | `BAD_REQUEST` | Parámetros inválidos (tipo de movimiento, UUID, etc.). |
| 401 | `UNAUTHORIZED` | Falta autenticación. |
| 403 | `FORBIDDEN` | Sin el permiso requerido. |
| 404 | `NOT_FOUND` | El producto no existe o es de otra empresa. |
| 409 | `CONFLICT` | La salida dejaría el stock en negativo. |
