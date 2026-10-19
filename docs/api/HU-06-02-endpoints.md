# Reporte de endpoints — HU-06-02 Verificación de disponibilidad de SKU y código de barras

## Información general

- Prefijo base: `/api/v1`.
- Ruta: `GET /api/v1/products/verificar-identificador`.
- Requiere autenticación y permiso `products.create`.
- La empresa se obtiene del token; la respuesta solo considera productos de esa empresa.
- Límite: 30 consultas por minuto por usuario autenticado. Al excederlo responde `429 TOO_MANY_REQUESTS`. Es configurable en `rate_limit.identifier_check` (`max_requests`, `window_seconds`; env `SIGIF_RATE_LIMIT_IDENTIFIER_CHECK_*`) y es el mismo cupo que comparte `GET /products/validate-duplicate`: agotarlo en uno bloquea el otro durante la ventana.
- Colección Bruno: `bruno/Products/033-034`.

## Parámetros

| Parámetro | Requerido | Descripción |
|---|---:|---|
| `sku` | Uno de los dos | SKU que se desea consultar. Se normaliza como en el alta de producto. |
| `barcode` | Uno de los dos | Código de barras que se desea consultar. |
| `codigo_barras` | Uno de los dos | Alias en español de `barcode`. |

Debe enviarse al menos uno de `sku`, `barcode` o `codigo_barras`.

## Respuesta exitosa

### Identificador disponible

```json
{
  "success": true,
  "data": {
    "available": true,
    "existing_product": null,
    "conflicts": []
  }
}
```

### Identificador ocupado

```json
{
  "success": true,
  "data": {
    "available": false,
    "existing_product": {
      "id": "6e436ff1-30d4-4572-b63c-c5ba2f60996b",
      "name": "Galletas María 200g",
      "sku": "ABA-412",
      "barcode": "7750000000001",
      "status": "active",
      "deleted": false
    },
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
      }
    ]
  }
}
```

`existing_product` es el producto del primer conflicto. `conflicts` conserva todos los conflictos cuando se envían SKU y código de barras juntos.

## Errores

| HTTP | Código | Situación |
|---:|---|---|
| 400 | `BAD_REQUEST` | No se envió ningún identificador o el formato es inválido. |
| 401 | `UNAUTHORIZED` | Falta autenticación o la sesión no es válida. |
| 403 | `FORBIDDEN` | El usuario no tiene `products.create`. |
| 429 | `TOO_MANY_REQUESTS` | Superó 30 consultas en un minuto. |

## Uso en frontend

Aplicar debounce de aproximadamente 400 ms al salir del campo o mientras se escribe. Un fallo de red o un `429` no debe bloquear el guardado: `POST /products` conserva la validación definitiva de HU-06-01.
