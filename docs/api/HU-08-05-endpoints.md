# Reporte de endpoints — [HU-08-05] Desactivación de cliente

## Información general

- HU: HU-08-05 — Desactivación de cliente (RF-08, P0, Sprint 1). Depende de HU-08-03.
- Rama: `feat/HU-08-clients-management`
- Prefijo base: `/api/v1`
- Requiere autenticación: token JWT válido en el esquema de autorización de portador (`Authorization: Bearer <access_token>`). El middleware global `AuthRequired` valida el token y que la sesión siga activa.
- Empresa: el cliente se busca siempre por `id` **y** por el `company_id` del token. Un cliente de otra empresa responde `404`, igual que uno inexistente. Un usuario sin empresa (por ejemplo, el superadmin global `admin@sigif.com`) recibe `400 company is required`.
- Formato de respuesta exitosa: `{ "success": true, "data": { ... } }`.
- Formato de respuesta de error: `{ "success": false, "error": { "code": "CODIGO", "message": "descripción", "details": { "campo": "motivo" } } }`. `details` solo aparece en errores de validación.
- Colección Bruno: `bruno/Customer/Status/` (001–019). Cada petición incluye sus `tests`.

La ruta ya existía desde la HU-08-01. Esta HU mantiene su contrato y corrige su comportamiento (ver «Diferencias respecto a la versión de HU-08-01»).

---

## Resumen rápido

| Método | Ruta | Permiso | Descripción |
|---|---|---|---|
| `PATCH` | `/api/v1/customers/:id/status` | `customers.status` | Cambia el estado del cliente (desactivar, bloquear o reactivar) sin eliminarlo. |

---

## Autenticación / permisos

| Header | Obligatorio | Descripción |
|---|---|---|
| `Authorization` | Sí | `Bearer <access_token>` de una sesión activa. |

| Permiso | Endpoint | Roles sembrados que lo tienen |
|---|---|---|
| `customers.status` | `PATCH /api/v1/customers/:id/status` | `superadmin` |

`customers.update` **no** basta para cambiar el estado. El rol `business_admin` no tiene permisos de clientes y recibe **403**.

---

## 1) Cambiar estado — `PATCH /api/v1/customers/:id/status`

Cambia el estado del cliente. Desactivar **no borra** al cliente ni sus datos: solo deja de estar disponible para nuevas operaciones que requieran un cliente activo. El mismo endpoint lo reactiva.

### Permiso
`customers.status`

### Params de ruta

| Parámetro | Tipo | Descripción |
|---|---|---|
| `id` | UUID | Identificador del cliente. |

### Query params
Ninguno.

### Body

```json
{
  "status": "inactive"
}
```

| Campo | Tipo | Obligatorio | Descripción |
|---|---|---|---|
| `status` | enum | Sí | `active` \| `inactive` \| `blocked`. En minúsculas. |

### Ejemplo de request

```bash
curl -s -X PATCH "http://localhost:4600/api/v1/customers/86d60c05-c7c6-4e4f-89b2-8a308817300e/status" \
  -H "Authorization: Bearer $ACCESS_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"status":"inactive"}'
```

### Respuesta exitosa

**`200 OK`** — cliente actualizado (misma estructura que `GET /api/v1/customers/:id`)

```json
{
  "success": true,
  "data": {
    "id": "86d60c05-c7c6-4e4f-89b2-8a308817300e",
    "company_id": "7dcfcd06-724a-4dd7-a10e-6ba716edf3cc",
    "legal_name": "Ferreteria Central",
    "document_type": "tax_id",
    "document_number": "1020304050",
    "phone": "+59170000001",
    "email": "central@example.com",
    "address": "Av. Blanco Galindo km 4",
    "credit_limit": "0",
    "credit_balance": "0",
    "points_accrued": 0,
    "status": "inactive",
    "created_at": "2026-10-13T14:20:00Z",
    "updated_at": "2026-10-13T15:02:11Z"
  }
}
```

**`200 OK`** — el cliente ya tenía ese estado: se devuelve tal cual, **sin escribir en la base de datos** y sin cambiar `updated_at`.

### Códigos de error

| Code | HTTP | Causa |
|---|---|---|
| `VALIDATION_ERROR` | 400 | `status` falta, está vacío o no es uno de los valores permitidos. |
| `BAD_REQUEST` | 400 | `id` no es un UUID válido (`invalid customer id`), el body no es JSON válido, o el token no tiene empresa (`company is required`). |
| `UNAUTHORIZED` | 401 | Falta el token o es inválido/expiró, o la sesión fue cerrada. |
| `FORBIDDEN` | 403 | El usuario no tiene el permiso `customers.status`. |
| `NOT_FOUND` | 404 | El cliente no existe, pertenece a otra empresa o fue eliminado. |
| `INTERNAL_ERROR` | 500 | Error inesperado de persistencia. |

**`400`** — validación

```json
{
  "success": false,
  "error": {
    "code": "VALIDATION_ERROR",
    "message": "validation failed",
    "details": {
      "status": "status must be one of [active inactive blocked]"
    }
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

- **No es una eliminación:** el registro se conserva con todos sus datos y relaciones. Nunca se ejecuta un `DELETE`.
- **Solo cambia el estado:** la escritura afecta únicamente a `status` y `updated_at` (`UpdateStatus`). Una edición simultánea del cliente no se pierde.
- **Atómica:** es una única sentencia `UPDATE`; si falla, el estado anterior se mantiene.
- **Transiciones:** se permite cualquier cambio entre `active`, `inactive` y `blocked`, incluida la reactivación.
- **Idempotente:** pedir el estado actual responde `200` sin escribir (mismo criterio que `PATCH /api/v1/roles/:id/status`).
- **No disponible para nuevas operaciones:**
  - El listado `GET /api/v1/customers/` devuelve solo clientes `active` por defecto (HU-08-02). Un cliente desactivado se consulta con `status=inactive` o `status=all`, y siempre por `GET /api/v1/customers/:id`.
  - El dominio expone `Customer.IsActive()` y `CustomerService.EnsureAvailableForOperations(ctx, companyID, id)`. Este método devuelve el cliente si existe, es de la empresa y está activo; si no, `404 NOT_FOUND` (inexistente, de otra empresa o eliminado) o `409 CONFLICT` (`customer is not active and cannot be used in new operations`). Es el punto de entrada para ventas, créditos (RF-55) y fidelización (RF-57), que aún no existen.
- **Edición:** un cliente inactivo o bloqueado se puede seguir editando (HU-08-04).

---

## Cobertura de criterios de aceptación

| Criterio HU-08-05 | Cómo se cubre |
|---|---|
| Desactivar un cliente | `PATCH /api/v1/customers/:id/status` con `{ "status": "inactive" }`. |
| No eliminar el registro | Solo se actualiza `status`; `TestChangeCustomerStatusEndpointDeactivates` comprueba que el registro y el conteo se conservan. |
| Conservar la información y las relaciones | Solo se escriben `status` y `updated_at`; probado contra PostgreSQL en `TestUpdateStatusWritesOnlyStatusAndUpdatedAt`. |
| No disponible en nuevas operaciones que requieran cliente activo | Listado por defecto solo con activos; `EnsureAvailableForOperations` para los módulos futuros. |
| Recibir el nuevo estado permitido | Body `status` validado contra `active`, `inactive`, `blocked`. |
| Validar existencia y autorización | `404`, `401` y `403`. |
| Impedir el DELETE físico | El servicio no llama a ninguna operación de borrado. |
| Retornar el cliente actualizado | `200` con el cliente completo. |

---

## Pruebas

- Unitarias HTTP: `internal/modules/customer/interfaces/http/handler/status_handler_test.go`.
- Dominio: `internal/modules/customer/domain/service/availability_test.go`.
- Integración con PostgreSQL: `internal/modules/customer/infrastructure/persistence/gorm/status_delete_integration_test.go`:

```bash
SIGIF_TEST_DATABASE_DSN="host=localhost port=5432 user=sigif password=sigif dbname=sigif_test sslmode=disable" \
  go test -tags integration ./internal/modules/customer/...
```

- Bruno (requiere la API levantada y la base sembrada; usa `admin.empresa@sigif.com`, que sí tiene empresa, y `negocio@sigif.com` como usuario sin permisos de clientes):

```bash
cd bruno
bru run Customer/Status --env development \
  --env-var baseUrl=http://localhost:4600 \
  --env-var companyAdminPassword=<contraseña del seed> \
  --env-var businessAdminPassword=<contraseña del seed>
```

La carpeta crea su propio cliente (prefijo único `HU0805-<timestamp>`) y lo elimina al terminar, por lo que puede ejecutarse varias veces. El servidor limita a 100 peticiones por minuto: si se ejecutan varias carpetas seguidas puede responder `429`; basta esperar un minuto.

---

## Diferencias respecto a la versión de HU-08-01

1. La escritura pasa de reescribir toda la fila a actualizar solo `status` y `updated_at`.
2. Pedir el estado que el cliente ya tiene no escribe en la base de datos ni cambia `updated_at` (antes sí).
3. Se agregan `IsActive()` y `EnsureAvailableForOperations` para los módulos que requieran un cliente activo.

El contrato HTTP (ruta, permiso, body y respuestas) no cambia.
