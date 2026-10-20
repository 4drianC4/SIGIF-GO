# Reporte de endpoints — [HU-08-06] Eliminación de cliente

## Información general

- HU: HU-08-06 — Eliminación de cliente (RF-08, P0, Sprint 1). Depende de HU-08-03.
- Rama: `feat/HU-08-clients-management`
- Prefijo base: `/api/v1`
- Requiere autenticación: token JWT válido en el esquema de autorización de portador (`Authorization: Bearer <access_token>`). El middleware global `AuthRequired` valida el token y que la sesión siga activa.
- Empresa: el cliente se busca siempre por `id` **y** por el `company_id` del token. Un cliente de otra empresa responde `404`, igual que uno inexistente. Un usuario sin empresa (por ejemplo, el superadmin global `admin@sigif.com`) recibe `400 company is required`.
- Formato de respuesta exitosa: `204 No Content`, sin cuerpo.
- Formato de respuesta de error: `{ "success": false, "error": { "code": "CODIGO", "message": "descripción", "details": { "campo": "motivo" } } }`. `details` solo aparece en errores de validación.
- Colección Bruno: `bruno/Customer/Delete/` (001–016). Cada petición incluye sus `tests`.

La ruta ya existía desde la HU-08-01. Esta HU mantiene su contrato y agrega la verificación de dependencias (ver «Diferencias respecto a la versión de HU-08-01»).

---

## Resumen rápido

| Método | Ruta | Permiso | Descripción |
|---|---|---|---|
| `DELETE` | `/api/v1/customers/:id` | `customers.delete` | Elimina lógicamente un cliente, solo si no compromete la integridad de la información. |

---

## Autenticación / permisos

| Header | Obligatorio | Descripción |
|---|---|---|
| `Authorization` | Sí | `Bearer <access_token>` de una sesión activa. |

| Permiso | Endpoint | Roles sembrados que lo tienen |
|---|---|---|
| `customers.delete` | `DELETE /api/v1/customers/:id` | `superadmin` |

`customers.status` **no** basta para eliminar. El rol `business_admin` no tiene permisos de clientes y recibe **403**.

---

## 1) Eliminar cliente — `DELETE /api/v1/customers/:id`

Elimina al cliente cuando nada que deba conservarse depende de él. Si existen relaciones, la eliminación se rechaza con `409` y el cliente queda intacto; en ese caso puede desactivarse con `PATCH /api/v1/customers/:id/status` (HU-08-05).

### Permiso
`customers.delete`

### Params de ruta

| Parámetro | Tipo | Descripción |
|---|---|---|
| `id` | UUID | Identificador del cliente. |

### Query params
Ninguno.

### Body
Ninguno.

### Ejemplo de request

```bash
curl -s -i -X DELETE "http://localhost:4600/api/v1/customers/86d60c05-c7c6-4e4f-89b2-8a308817300e" \
  -H "Authorization: Bearer $ACCESS_TOKEN"
```

### Respuesta exitosa

**`204 No Content`** — sin cuerpo (misma convención que `DELETE /api/v1/users/:id` y `DELETE /api/v1/products/:id`).

### Códigos de error

| Code | HTTP | Causa |
|---|---|---|
| `BAD_REQUEST` | 400 | `id` no es un UUID válido (`invalid customer id`), o el token no tiene empresa (`company is required`). |
| `UNAUTHORIZED` | 401 | Falta el token o es inválido/expiró, o la sesión fue cerrada. |
| `FORBIDDEN` | 403 | El usuario no tiene el permiso `customers.delete`. |
| `NOT_FOUND` | 404 | El cliente no existe, pertenece a otra empresa o ya fue eliminado. |
| `CONFLICT` | 409 | El cliente tiene saldo de crédito, puntos acumulados o registros relacionados que deben conservarse. |
| `INTERNAL_ERROR` | 500 | Error inesperado de persistencia. |

**`409`** — saldo de crédito o puntos

```json
{
  "success": false,
  "error": {
    "code": "CONFLICT",
    "message": "customer has credit balance or accrued points and cannot be deleted; deactivate it instead"
  }
}
```

**`409`** — registros relacionados

```json
{
  "success": false,
  "error": {
    "code": "CONFLICT",
    "message": "customer has related records and cannot be deleted; deactivate it instead"
  }
}
```

**`404`**

```json
{
  "success": false,
  "error": {
    "code": "NOT_FOUND",
    "message": "customer not found"
  }
}
```

---

## Reglas de negocio

- **Eliminación lógica:** se informa `deleted_at` (soft delete de GORM); la fila se conserva en la base de datos para preservar el historial. No hay eliminación física.
- **Verificación de dependencias**, en este orden:
  1. `credit_balance` distinto de `0` o `points_accrued` mayor que `0` → `409`. Esos datos no deben perderse.
  2. El puerto de dominio `repository.DependencyChecker` informa registros relacionados → `409`.
- **Puerto `DependencyChecker`:** hoy lo implementa `NoDependencyChecker`, que no encuentra relaciones porque todavía no hay tablas que referencien a `customer`. Los módulos de ventas, créditos (RF-55) y fidelización (RF-57) deberán reemplazar esa implementación en el wiring de FX por una que consulte sus propios registros.
- **Transacción:** la lectura, las verificaciones y la eliminación se ejecutan en una sola transacción. Una eliminación rechazada no modifica nada.
- **Estado:** se puede eliminar un cliente en cualquier estado (`active`, `inactive`, `blocked`).
- **Después de eliminar:** el cliente no aparece en `GET /api/v1/customers/` (ni con `status=all`), y `GET`, `PUT`, `PATCH`, `PATCH /status` y `DELETE` sobre su `id` responden `404`.
- **Documento reutilizable:** el índice único `idx_customer_company_document_unique` es parcial (`deleted_at IS NULL`), así que el mismo documento puede registrarse de nuevo en la empresa.
- **Efecto en la fila eliminada:** además de `deleted_at`, la eliminación deja `status = inactive` y actualiza `updated_at` (comportamiento heredado de la HU-08-01, sin cambios en esta HU). No es visible por la API porque la fila ya no se consulta, pero hay que tenerlo en cuenta si algún día se restauran clientes: volverían como inactivos.

---

## Cobertura de criterios de aceptación

| Criterio HU-08-06 | Cómo se cubre |
|---|---|
| Eliminar un cliente | `DELETE /api/v1/customers/:id` → `204`. |
| Solo cuando no compromete la integridad | Verificación de saldo, puntos y `DependencyChecker` antes de eliminar. |
| Rechazar si hay relaciones que deban conservarse | `409 CONFLICT` con un mensaje que sugiere desactivar. |
| El cliente rechazado se mantiene | Transacción única; `TestDeleteCustomerEndpointBlockedByDependencies` y `TestDeleteRejectedChangesNothing` comprueban que nada cambia. |
| Puede desactivarse en su lugar | `PATCH /api/v1/customers/:id/status` (HU-08-05). |
| Validar existencia y autorización | `404`, `401` y `403`. |
| El cliente eliminado deja de estar disponible | No aparece en el listado ni en `GET /:id`. |

---

## Pruebas

- Unitarias HTTP: `internal/modules/customer/interfaces/http/handler/delete_handler_test.go`. La eliminación bloqueada se prueba con `testutil.StubDependencyChecker` y con saldo o puntos cargados directamente en el repositorio.
- Integración con PostgreSQL: `internal/modules/customer/infrastructure/persistence/gorm/status_delete_integration_test.go`:

```bash
SIGIF_TEST_DATABASE_DSN="host=localhost port=5432 user=sigif password=sigif dbname=sigif_test sslmode=disable" \
  go test -tags integration ./internal/modules/customer/...
```

- Bruno (requiere la API levantada y la base sembrada; usa `admin.empresa@sigif.com`, que sí tiene empresa, y `negocio@sigif.com` como usuario sin permisos de clientes):

```bash
cd bruno
bru run Customer/Delete --env development \
  --env-var baseUrl=http://localhost:4600 \
  --env-var companyAdminPassword=<contraseña del seed> \
  --env-var businessAdminPassword=<contraseña del seed>
```

La carpeta crea su propio cliente (prefijo único `HU0806-<timestamp>`) y lo elimina al terminar, por lo que puede ejecutarse varias veces. El servidor limita a 100 peticiones por minuto: si se ejecutan varias carpetas seguidas puede responder `429`; basta esperar un minuto.

La eliminación bloqueada (`409`) **no se puede provocar por HTTP** todavía: ningún endpoint modifica `credit_balance` ni `points_accrued`, y no hay módulos con registros relacionados. Queda cubierta por las pruebas Go.

---

## Diferencias respecto a la versión de HU-08-01

1. Se agrega la verificación de dependencias y la respuesta `409`.
2. La lectura, la verificación y la eliminación se ejecutan en una transacción.

El contrato HTTP (ruta, permiso y `204`) no cambia.
