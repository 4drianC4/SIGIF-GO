# Reporte de endpoints — [HU-06-01] Validación de unicidad de SKU y código de barras al registrar producto

## Información general

- HU: HU-06-01 — Validación de unicidad de SKU y código de barras al registrar producto (P0, Sprint 1). Depende de RF-03 (HU-03-01).
- Rama: `feature/HU-06-01-product-sku-barcode-uniqueness` (parte de `feature/integration-product`).
- Prefijo base: `/api/v1`. La HU nombra `/api/v1/productos`; se mantiene la ruta existente `/api/v1/products`.
- Requiere autenticación (HU-01): cookie HttpOnly `sigif_access_token` en el navegador o `Authorization: Bearer <access_token>`.
- La empresa sale del token. La unicidad se evalúa solo dentro de esa empresa.
- No hay endpoints nuevos: cambia la respuesta `409` de `POST /products` y `PUT /products/:id`, y se amplía `GET /products/validate-duplicate`.
- Colección Bruno: `bruno/Products/` (025–032).

---

## Resumen rápido

| Método | Ruta | Permiso | Cambio |
|---|---|---|---|
| `POST` | `/api/v1/products` | `products.create` | El `409` lista **todos** los valores repetidos con el producto que los tiene. |
| `PUT` | `/api/v1/products/:id` | `products.update` | Igual que al crear; el propio producto no cuenta como duplicado. |
| `GET` | `/api/v1/products/validate-duplicate` | `products.create` | Añade `conflicts`, con el mismo formato. |

---

## Reglas

| # | Regla |
|---|---|
| 1 | SKU y código de barras son únicos por empresa. Otra empresa puede usar los mismos valores. |
| 2 | El SKU se compara y se guarda sin espacios al inicio ni al final y en mayúsculas: `abc-01` = ` ABC-01 `. El código de barras se compara sin espacios externos. |
| 3 | Un SKU o código de barras vacío no cuenta: puede haber muchos productos sin ellos. |
| 4 | Un producto **inactivo o eliminado** sigue ocupando su SKU y su código de barras. Su nombre sí queda libre al eliminarlo. |
| 5 | Si hay varios valores repetidos se informan **todos a la vez**, en este orden: `name`, `sku`, `barcode`. |
| 6 | Ante un conflicto no se guarda nada. Si dos peticiones intentan el mismo valor a la vez, gana una y la otra recibe el mismo `409` con el producto ganador. |

El nombre del producto también es único (regla de HU-03-01) y usa el mismo formato de conflicto.

---

## Respuesta `409 CONFLICT`

`error.code` sigue siendo `CONFLICT`. `error.details.conflicts` trae un elemento por valor repetido.

| Campo | Tipo | Descripción |
|---|---|---|
| `field` | `name` \| `sku` \| `barcode` | Campo del formulario con el error. |
| `code` | `NAME_DUPLICATE` \| `SKU_DUPLICATE` \| `BARCODE_DUPLICATE` | Código estable para el mensaje del frontend. |
| `value` | string | Valor repetido, ya normalizado (SKU en mayúsculas). |
| `existing_product.id` | uuid | Producto que ya tiene el valor. |
| `existing_product.name` | string | Su nombre, para el mensaje. |
| `existing_product.sku` | string \| null | Su SKU. |
| `existing_product.barcode` | string \| null | Su código de barras. |
| `existing_product.status` | `active` \| `inactive` | Su estado. |
| `existing_product.deleted` | boolean | `true` si fue eliminado. |

`error.message`: con un solo conflicto, `SKU already exists`, `barcode already exists` o `product name already exists`; con varios, `product has duplicated values`. El frontend debe armar su propio texto con `value` y `existing_product.name`, por ejemplo:

- `SKU_DUPLICATE` → "El SKU {value} ya está asignado al producto {existing_product.name}".
- `BARCODE_DUPLICATE` → "El código de barras {value} ya está asignado al producto {existing_product.name}".

### Ejemplo: SKU y código de barras repetidos

`POST /api/v1/products` con `"sku": "  aba-412 "` y `"barcode": "7750000000001"`:

```json
{
  "success": false,
  "error": {
    "code": "CONFLICT",
    "message": "product has duplicated values",
    "details": {
      "conflicts": [
        {
          "field": "sku",
          "code": "SKU_DUPLICATE",
          "value": "ABA-412",
          "existing_product": {
            "id": "6e436ff1-30d4-4572-b63c-c5ba2f60996b",
            "name": "Galletas María 200g",
            "sku": "ABA-412",
            "barcode": "7750000000001",
            "status": "active",
            "deleted": false
          }
        },
        {
          "field": "barcode",
          "code": "BARCODE_DUPLICATE",
          "value": "7750000000001",
          "existing_product": {
            "id": "6e436ff1-30d4-4572-b63c-c5ba2f60996b",
            "name": "Galletas María 200g",
            "sku": "ABA-412",
            "barcode": "7750000000001",
            "status": "active",
            "deleted": false
          }
        }
      ]
    }
  }
}
```

### Ejemplo: SKU de un producto eliminado

```json
{
  "success": false,
  "error": {
    "code": "CONFLICT",
    "message": "SKU already exists",
    "details": {
      "conflicts": [
        {
          "field": "sku",
          "code": "SKU_DUPLICATE",
          "value": "DEL-1",
          "existing_product": {
            "id": "a9da80fd-6e74-4302-8e80-c6d767ee61bb",
            "name": "Descontinuado",
            "sku": "DEL-1",
            "barcode": null,
            "status": "active",
            "deleted": true
          }
        }
      ]
    }
  }
}
```

---

## `GET /api/v1/products/validate-duplicate`

Mismos parámetros que antes (`name`, `sku`, `barcode`, `exclude_id`). Añade `conflicts`, con el mismo formato que el `409`, para validar al salir de cada campo.

```json
{
  "success": true,
  "data": {
    "exists": true,
    "field": "sku",
    "conflicts": [
      {
        "field": "sku",
        "code": "SKU_DUPLICATE",
        "value": "ABA-412",
        "existing_product": { "id": "…", "name": "Galletas María 200g", "sku": "ABA-412", "barcode": null, "status": "active", "deleted": false }
      }
    ]
  }
}
```

Sin duplicados: `{ "exists": false, "field": null, "conflicts": [] }`.

---

## Base de datos

Migración `000004_product_reserved_codes`:

1. Convierte en `NULL` los SKU y códigos de barras vacíos.
2. **Comprueba** que ningún SKU (sin espacios y en mayúsculas) ni código de barras esté repetido en la empresa contando los productos eliminados. Si lo está, **falla sin cambiar nada** y lista los valores, por ejemplo: `SKUs shared by several products, counting deleted ones: DUP-1 (2 products)`. Hay que cambiar el valor de todos menos uno y volver a ejecutar la migración.
3. Normaliza los SKU guardados (`upper(btrim(sku))`) y los códigos de barras (`btrim(barcode)`).
4. Recrea los índices únicos sin excluir los eliminados:

| Índice | Antes | Ahora |
|---|---|---|
| `idx_products_company_sku` | `(company_id, sku) WHERE deleted_at IS NULL AND sku IS NOT NULL` | `(company_id, sku) WHERE sku IS NOT NULL` |
| `idx_products_company_barcode` | `(company_id, barcode) WHERE deleted_at IS NULL AND barcode IS NOT NULL` | `(company_id, barcode) WHERE barcode IS NOT NULL` |

La migración se puede revertir (`go run ./cmd/migrate down`): vuelve a los índices anteriores y mantiene los valores normalizados.

## Qué cambió en el código

| Cambio | Detalle |
|---|---|
| `domain/service/duplicate_error.go` | Error `409` con `details.conflicts`. Su causa es `ErrSKUTaken`, `ErrBarcodeTaken` o `ErrProductNameTaken`, así que `errors.Is` sigue funcionando. |
| `domain/service/duplicates.go` | Busca todos los conflictos; traduce una violación del índice único (concurrencia) al mismo `409`. |
| Repositorio | `ExistsByName/BySKU/ByBarcode` pasan a `FindByName/BySKU/ByBarcode`, que devuelven el producto. SKU y código de barras incluyen eliminados. |
| `validate-duplicate` | Añade `conflicts`. |
| Migración `000004` | Ver arriba. |

## Criterios de aceptación

| # | Criterio | Cómo se cumple | Prueba |
|---|---|---|---|
| 1 | SKU nuevo se registra | Sin cambios respecto a HU-03-01. | `TestCreateProductFollowsContract`, Bruno 025 |
| 2 | SKU repetido: no se crea y se informa el producto | `409` con `SKU_DUPLICATE` y `existing_product`. | `TestCreateProductReportsTheProductHoldingTheSKU`, Bruno 026 |
| 3 | Lo mismo para el código de barras | `BARCODE_DUPLICATE`. | `TestCreateProductReportsSKUAndBarcodeTogether`, Bruno 027 |
| 4 | Ambos repetidos: los dos errores a la vez | `details.conflicts` con dos elementos. | `TestCreateProductDuplicateContract`, Bruno 028 |
| 5 | Ignora mayúsculas y espacios | Normalización al guardar y en la migración. | `TestSKUComparisonIgnoresCaseAndOuterSpaces`, `TestReservedCodesNormalizeStoredValues`, Bruno 026 |
| 6 | Solo dentro de la empresa | Filtro e índices por `company_id`. | `TestUniquenessIsPerCompany` |
| 7 | Varios productos sin código de barras | Los vacíos se guardan como `NULL`. | `TestProductsWithoutBarcodeOrSKUAreNotDuplicates`, Bruno 031–032 |
| 8 | Nada persistido ante error | Comprobación previa e índice único como red de seguridad. | `TestConcurrentCreationsWithTheSameSKU` (Postgres, 20 altas simultáneas) |
| 9 | Inactivo o eliminado sigue ocupado | Búsqueda que incluye eliminados e índices de la migración `000004`. | `TestInactiveAndDeletedProductsKeepTheirSKUAndBarcode`, `TestDeletedProductsKeepTheirSKUAndBarcodeInPostgres`, Bruno 029 |

## Pruebas

- Unitarias y HTTP: `go test ./internal/modules/product/...`
- Integración con Postgres: `SIGIF_TEST_DATABASE_DSN=... go test -p 1 -tags integration ./internal/modules/product/... ./internal/shared/database/...`
- Bruno: carpeta `bruno/Products` (32 peticiones). Se puede ejecutar varias veces: SKU y códigos de barras se generan con la hora.

## Notas para el frontend

1. Mostrar cada elemento de `details.conflicts` en su campo (`field`), no un toast genérico, y enfocar el primero.
2. Armar el texto con `value` y `existing_product.name`; si `existing_product.deleted` es `true`, puede indicarse que el producto fue eliminado.
3. Conservar lo ingresado en el formulario: el backend no guarda nada ante un `409`.
