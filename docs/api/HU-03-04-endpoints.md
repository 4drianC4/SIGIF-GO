# Reporte de endpoints — HU-03-04 Consulta, desactivación y eliminación de producto y categoría

## Información general

- HU: HU-03-04 — Consulta, desactivación y eliminación de producto y categoría.
- RF relacionado: RF 03. Épica / módulo: Gestión de producto. Sprint: 1. Dependencias declaradas: RF 06 y RF 03-01.
- Rama: `feature/HU-03-03-HU-03-04-product-management`.
- Prefijo base: `/api/v1`.
- Requiere contexto autenticado del backend: `company_id`, `user_id`, obtenidos del JWT.
- Formato de respuesta exitosa: `{ "success": true, "data": ... }`; en listados se agrega `meta: { "page", "limit", "total", "total_pages" }`. El DELETE devuelve 204 sin body.
- Formato de respuesta de error: `{ "success": false, "error": { "code": "CODIGO", "message": "descripción", "details": { ... } } }`; `details` es opcional.
- Alcance implementado: listado con filtros, desactivación/reactivación y borrado lógico de productos. Se incluyen las acciones relacionadas de edición de HU-03-02 para documentar el PATCH nuevo y las modificaciones al PUT en esta rama. La gestión equivalente de categorías sigue pendiente.

---

## Resumen rápido

| Método | Ruta | Permiso | Descripción |
|---|---|---|---|
| `GET` | `/api/v1/products` | `products.read` | Lista productos activos e inactivos con filtros y paginación. |
| `PATCH` | `/api/v1/products/:id/status` | `products.update` | Desactiva o reactiva el producto. |
| `DELETE` | `/api/v1/products/:id` | `products.delete` | Elimina el producto de las consultas mediante soft-delete. |
| `PATCH` | `/api/v1/products/:id/price` | `products.update` | Nuevo: edita solo costo y/o precio de venta; genera historial. |
| `PUT` | `/api/v1/products/:id` | `products.update` | Modificado: edición completa con `reason` e historial automático. |

El `GET /api/v1/products/:id` mencionado en BE-14 y su extensión
`GET /api/v1/products/:id/price-history` están documentados en
[HU-03-03](HU-03-03-endpoints.md), incluyendo el nuevo campo `stock_total`.

---

## Autenticación / permisos (todas las rutas)

| Header | Obligatorio | Descripción |
|---|---|---|
| `Authorization` | Sí | `Bearer <access_token>` de una sesión activa, obtenido con `POST /api/v1/auth/login`. |
| `Content-Type` | Sí, en PATCH y PUT con body | `application/json`. |

Permisos usados en esta HU: `products.read`, `products.update`, `products.delete`.

El backend obtiene `company_id` y `user_id` del JWT. No utiliza
`x-company-id`, `x-user-id` ni una empresa predeterminada.
El permiso se comprueba por el rol mediante `role_permission → permission`.
Tener permiso no permite operar sobre productos de otra empresa.

---

## 1) Listar productos — `GET /api/v1/products`

Devuelve el catálogo paginado de la empresa con filtros opcionales por nombre, categoría y estado.

### Permiso

`products.read`

### Params de ruta

No aplica.

### Query params

- `name` (opcional, sin filtro por defecto): string — búsqueda parcial en el nombre, insensible a mayúsculas.
- `category_id` (opcional, sin filtro por defecto): string (UUID) — ID de la categoría a filtrar.
- `status` (opcional, ambos estados por defecto): enum — `active` o `inactive`.
- `page` (opcional, por defecto `1`): integer — número de página; mínimo efectivo 1.
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
      "updated_at": "2026-10-12T15:00:00Z"
    }
  ],
  "meta": {
    "page": 1,
    "limit": 20,
    "total": 25,
    "total_pages": 2
  }
}
```

### Códigos de error

| Code | HTTP | Causa |
|---|---|---|
| `BAD_REQUEST` | 400 | `category_id` no es UUID o `status` no es `active`/`inactive`. |
| `BAD_REQUEST` | 400 | Contexto autenticado sin empresa (`company is required`). |
| `UNAUTHORIZED` | 401 | Token ausente, inválido o expirado; sesión cerrada o usuario no autenticado. |
| `FORBIDDEN` | 403 | Usuario sin el permiso requerido para la acción. |
| `INTERNAL_ERROR` | 500 | Error no controlado de base de datos o comprobación de permisos. |

Ejemplo de error:

```json
{
  "success": false,
  "error": {
    "code": "BAD_REQUEST",
    "message": "status must be 'active' or 'inactive'"
  }
}
```

---

## 2) Desactivar o reactivar producto — `PATCH /api/v1/products/:id/status`

Cambia el estado sin eliminar el producto; la desactivación es reversible.

### Permiso

`products.update`

### Params de ruta

- `id`: string (UUID) — ID del producto de la empresa autenticada.

### Query params

No aplica.

### Body

```json
{
  "status": "inactive"
}
```

| Campo | Tipo | Obligatorio | Restricciones / notas |
|---|---|---|---|
| `status` | enum | Sí | `active` para reactivar; `inactive` para desactivar. |

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
    "status": "inactive",
    "created_at": "2026-10-02T12:00:00Z",
    "updated_at": "2026-10-12T15:00:00Z"
  }
}
```

### Códigos de error

| Code | HTTP | Causa |
|---|---|---|
| `BAD_REQUEST` | 400 | El `id` de ruta no es un UUID válido. |
| `BAD_REQUEST` | 400 | Body malformado, estado ausente/inválido o producto que ya tiene el estado solicitado. |
| `NOT_FOUND` | 404 | Producto inexistente, eliminado o de otra empresa. |
| `BAD_REQUEST` | 400 | Contexto autenticado sin empresa (`company is required`). |
| `UNAUTHORIZED` | 401 | Token ausente, inválido o expirado; sesión cerrada o usuario no autenticado. |
| `FORBIDDEN` | 403 | Usuario sin el permiso requerido para la acción. |
| `INTERNAL_ERROR` | 500 | Error no controlado de base de datos o comprobación de permisos. |

Ejemplo de error:

```json
{
  "success": false,
  "error": {
    "code": "BAD_REQUEST",
    "message": "product is already inactive"
  }
}
```

---

## 3) Eliminar producto — `DELETE /api/v1/products/:id`

Marca el producto como eliminado en la base y deja de exponerlo en las consultas del catálogo.

### Permiso

`products.delete`

### Params de ruta

- `id`: string (UUID) — ID del producto de la empresa autenticada.

### Query params

No aplica.

### Body

No aplica; no enviar body.

### Respuesta exitosa

**`204 No Content`** — sin body; no devuelve un objeto `data`.


### Códigos de error

| Code | HTTP | Causa |
|---|---|---|
| `BAD_REQUEST` | 400 | El `id` de ruta no es un UUID válido. |
| `NOT_FOUND` | 404 | Producto inexistente, eliminado o de otra empresa. |
| `BAD_REQUEST` | 400 | Contexto autenticado sin empresa (`company is required`). |
| `UNAUTHORIZED` | 401 | Token ausente, inválido o expirado; sesión cerrada o usuario no autenticado. |
| `FORBIDDEN` | 403 | Usuario sin el permiso requerido para la acción. |
| `INTERNAL_ERROR` | 500 | Error no controlado de base de datos o comprobación de permisos. |

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

## 4) Actualizar solo precios — `PATCH /api/v1/products/:id/price`

Modifica el costo y/o precio de venta del producto, conservando los demás datos y registrando el cambio en historial.

### Permiso

`products.update`

### Params de ruta

- `id`: string (UUID) — ID del producto de la empresa autenticada.

### Query params

No aplica.

### Body

```json
{
  "cost_price": 13.3,
  "sale_price": 16.9,
  "reason": "Ajuste por costo"
}
```

| Campo | Tipo | Obligatorio | Restricciones / notas |
|---|---|---|---|
| `cost_price` | number o string decimal | Al menos uno de los dos precios | Mayor o igual a 0; máximo dos decimales. Si se omite o es `null`, conserva el costo actual. |
| `sale_price` | number o string decimal | Al menos uno de los dos precios | Mayor que 0; máximo dos decimales; no puede ser menor que el costo resultante. Si se omite o es `null`, conserva el precio actual. |
| `reason` | string | No | Máximo 160 caracteres tras recortar espacios externos. Motivo del cambio, no dato permanente del producto. |

Los importes se envían en unidad monetaria completa (por ejemplo, 13.30), no en centavos.
`sale_price` es el precio que la tienda cobra a sus clientes.

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
    "updated_at": "2026-10-12T15:00:00Z"
  }
}
```

### Códigos de error

| Code | HTTP | Causa |
|---|---|---|
| `BAD_REQUEST` | 400 | El `id` de ruta no es un UUID válido. |
| `BAD_REQUEST` | 400 | Body malformado o importe que no se puede interpretar como decimal. |
| `VALIDATION_ERROR` | 400 | Ningún precio no nulo; importes inválidos, precisión excesiva, motivo demasiado largo o precio menor que el costo. |
| `NOT_FOUND` | 404 | Producto inexistente, eliminado o de otra empresa. |
| `BAD_REQUEST` | 400 | Contexto autenticado sin empresa (`company is required`). |
| `UNAUTHORIZED` | 401 | Token ausente, inválido o expirado; sesión cerrada o usuario no autenticado. |
| `FORBIDDEN` | 403 | Usuario sin el permiso requerido para la acción. |
| `INTERNAL_ERROR` | 500 | Error no controlado de base de datos o comprobación de permisos. |

Ejemplo de error:

```json
{
  "success": false,
  "error": {
    "code": "VALIDATION_ERROR",
    "message": "validation failed",
    "details": {
      "sale_price": "no puede ser menor que el costo"
    }
  }
}
```

---

## 5) Editar producto completo — `PUT /api/v1/products/:id`

Actualiza los datos del producto; esta acción de HU-03-02 ahora acepta motivo y registra los cambios de precios en historial.

### Permiso

`products.update`

### Params de ruta

- `id`: string (UUID) — ID del producto de la empresa autenticada.

### Query params

No aplica.

### Body

```json
{
  "category_id": "b45978d0-37f8-4051-92a8-218441dac948",
  "sku": "ABA-210",
  "barcode": "7772000002103",
  "name": "Aceite Fino 900 ml",
  "description": "Botella de 900 ml",
  "unit_of_measure": "unit",
  "cost_price": 13.3,
  "sale_price": 16.9,
  "reason": "Ajuste por costo"
}
```

| Campo | Tipo | Obligatorio | Restricciones / notas |
|---|---|---|---|
| `category_id` | string (UUID) | Sí | Categoría existente, activa y de la empresa autenticada. |
| `sku` | string | Sí | Máximo 50 caracteres; se convierte a mayúsculas. Letras A–Z, números, `.`, `-` y `_`; comienza por letra o número. Único dentro de la empresa entre productos no eliminados. |
| `barcode` | string | No | Si no está vacío, solo números y entre 8 y 14 caracteres. Único dentro de la empresa entre productos no eliminados. |
| `name` | string | Sí | Entre 2 y 150 caracteres; se recortan espacios y se unifican espacios internos. |
| `description` | string | No | Máximo 1000 caracteres. |
| `unit_of_measure` | enum | Sí | `unit`, `kg`, `g`, `l`, `ml`, `m`, `box`, `pack`. |
| `cost_price` | number o string decimal | No | Mayor o igual a 0; máximo dos decimales. Omitido o `null` se reemplaza por cero. |
| `sale_price` | number o string decimal | Sí | Mayor que 0; máximo dos decimales. |
| `reason` | string | No | Nuevo: máximo 160 caracteres, tras recortar espacios externos; se guarda con el historial si cambia costo o precio. |

Este PUT requiere los campos obligatorios de edición completa. Para el modal
de precio, utilizar el PATCH de la sección anterior.

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
    "updated_at": "2026-10-12T15:00:00Z"
  }
}
```

### Códigos de error

| Code | HTTP | Causa |
|---|---|---|
| `BAD_REQUEST` | 400 | El `id` de ruta no es un UUID válido. |
| `BAD_REQUEST` | 400 | Body malformado, importes no interpretables o categoría inactiva. |
| `VALIDATION_ERROR` | 400 | Campo obligatorio ausente o inválido; SKU, código de barras, unidad, importes, precisión o longitud inválidos. |
| `NOT_FOUND` | 404 | Producto o categoría no existen en la empresa; el producto eliminado tampoco se puede editar. |
| `CONFLICT` | 409 | SKU o código de barras usados por otro producto no eliminado de la misma empresa. |
| `BAD_REQUEST` | 400 | Contexto autenticado sin empresa (`company is required`). |
| `UNAUTHORIZED` | 401 | Token ausente, inválido o expirado; sesión cerrada o usuario no autenticado. |
| `FORBIDDEN` | 403 | Usuario sin el permiso requerido para la acción. |
| `INTERNAL_ERROR` | 500 | Error no controlado de base de datos o comprobación de permisos. |

Ejemplo de error:

```json
{
  "success": false,
  "error": {
    "code": "CONFLICT",
    "message": "a product with this sku already exists"
  }
}
```

---

## Estructura de respuesta de entidad `Product`

El listado, los PATCH y el PUT devuelven este contrato base:

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
  "updated_at": "2026-10-12T15:00:00Z"
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

`stock_total` solo se incluye en `GET /api/v1/products/:id`, cuyo contrato
`ProductDetail` se documenta en [HU-03-03](HU-03-03-endpoints.md).
El historial usa `PriceHistoryEntry`, descrito en ese mismo reporte.
`reason` no se agrega a la entidad `Product` devuelta por el PATCH o PUT.

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
| `BAD_REQUEST` | 400 | UUID inválido, contexto sin empresa, body malformado, filtro/estado inválido, estado ya aplicado o categoría de edición inactiva. |
| `VALIDATION_ERROR` | 400 | Campos inválidos en edición; ningún precio en PATCH; margen negativo en PATCH; motivo o precisión fuera de límites. `details` identifica campos. |
| `UNAUTHORIZED` | 401 | Token o sesión inválidos; usuario no autenticado. |
| `FORBIDDEN` | 403 | Usuario sin permiso para la operación. |
| `NOT_FOUND` | 404 | Producto eliminado, inexistente o de otra empresa; categoría de edición inexistente o de otra empresa. |
| `CONFLICT` | 409 | SKU o código de barras duplicado en el PUT. |
| `INTERNAL_ERROR` | 500 | Error no controlado de persistencia, historial o autorización. |

---

## Casos borde / comportamiento no obvio

- El listado incluye productos activos e inactivos por defecto y excluye productos eliminados. Se ordena por nombre ascendente.
- El filtro `name` busca coincidencias parciales sin distinguir mayúsculas; `category_id` filtra por ID sin expandir la categoría. Un UUID de categoría válido sin coincidencias devuelve un listado vacío, no 404.
- `page` menor a 1 o no numérico se interpreta como 1. `limit` menor a 1, mayor a 100 o no numérico se reemplaza por 20; no se recorta a 100.
- Página fuera de rango: 200, `data: []` y total real. Sin productos, `meta.total_pages` es 0.
- El listado no incluye stock. Para consultar la cantidad física de un producto se usa el GET por ID con `stock_total`; no existe una ruta independiente de stock.
- Enviar el estado actual devuelve 400. Un producto inactivo sigue disponible en listado, detalle e historial y puede reactivarse.
- El DELETE es un **borrado lógico**: GORM marca `deleted_at`. No ejecuta un borrado físico de la fila ni de su historial. Desde la API deja de estar disponible y no existe un endpoint para restaurarlo.
- La doble confirmación de FE-15/FE-16 corresponde al frontend. El DELETE no exige nombre, checkbox ni token de confirmación en el body.
- Los índices de SKU y código de barras permiten reutilizarlos tras el soft-delete; un SKU idéntico en otra empresa no produce conflicto.
- **BE-17 sigue pendiente:** el servicio de eliminación no consulta ventas, movimientos ni stock antes de borrar. No existe un error de negocio de eliminación bloqueada por dependencias que pueda documentarse como implementado.
- El bloqueo de productos inactivos en ventas/compras debe implementarse en los módulos de esas operaciones; estas rutas de catálogo no prueban ni aplican ese criterio.
- El PATCH de precios conserva el importe omitido y el resto del producto. Enviar solo `reason`, `{}` o ambos precios `null` devuelve `VALIDATION_ERROR`.
- El PATCH valida el margen con los importes finales, incluidos los valores conservados. Costo igual a precio permite margen cero; precio de venta cero sigue siendo inválido.
- **La regla de margen negativo se aplica al PATCH de precios.** El PUT actual valida costo no negativo y precio positivo, pero no rechaza `sale_price < cost_price`. No asumir que ambos contratos tienen esa misma validación.
- El PUT es edición completa: omitir `barcode` o `description` los vacía; omitir `cost_price` lo fija en cero. El estado y la fecha de registro no se cambian por ese PUT.
- PATCH y PUT guardan precio e historial en la misma transacción. Si falla el historial, se revierte la actualización del producto. El autor procede del JWT, no del body.
- La lectura del producto para editarlo utiliza `SELECT ... FOR UPDATE` dentro de la transacción para serializar los cambios concurrentes de precio.
- Se registra un historial si cambia costo o precio; ambos cambios en una petición generan una sola entrada. Sin cambios de precios no se agrega historial. El PATCH sin cambios devuelve 200 y conserva `updated_at`; el PUT sigue siendo una edición completa.
- El PATCH no cambia precios especiales por sucursal (`branch_price`) ni modifica ventas pasadas. Actualiza el precio general del producto.
- No existen GET, PUT, PATCH de estado ni DELETE de categorías en el router actual. La parte de categorías indicada en el título de la HU sigue pendiente.

---

## Diferencias respecto a una versión anterior (si aplica)

Esta rama agrega o modifica, sobre la base funcional de HU-03-02 (edición de producto):

1. `GET /api/v1/products` con filtros por nombre, categoría y estado.
2. `GET /api/v1/products/:id` para consulta individual, documentado en HU-03-03; ahora incluye `stock_total`.
3. `PATCH /api/v1/products/:id/status` para desactivar y reactivar.
4. `DELETE /api/v1/products/:id` con soft-delete; se precisa esta decisión respecto a la eliminación física mencionada en la HU.
5. `PATCH /api/v1/products/:id/price` como ruta nueva para edición independiente de costo/precio, motivo opcional y rechazo de margen negativo.
6. Modificación al PUT existente: `reason`, historial automático y guardado transaccional de producto e historial.
7. `GET /api/v1/products/:id/price-history`, documentado en HU-03-03.
8. Modelos `product_price_history` y `stock_balance` registrados en el migrador; el catálogo mantiene UUID.
9. Arranque de Docker: `app` espera el éxito de `migrate`, y `make docker-up` reconstruye las imágenes.

Para incorporar código y migraciones actuales:

```bash
docker compose -f deployments/docker/docker-compose.yml up -d --build
```

Referencia de edición original: [HU-03-02](HU-03-02-endpoints.md).

Ejemplo de petición en Bruno: `bruno/Products/017-update-product-price.bru`.
Un 404 `Cannot PATCH .../price` indica que el servidor no tiene registrada esa
ruta; reconstruir y recrear el contenedor carga el código actual.

