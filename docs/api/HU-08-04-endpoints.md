# Reporte de endpoints — [HU-08-04] Edición de cliente

## Información general

- HU: HU-08-04 — Edición de cliente (RF-08, P0, Sprint 1). Depende de HU-08-03 (`GET /api/v1/customers/:id`, ya disponible en `dev`).
- Rama: `feat/HU-08-04-clients-edit` (incluye el merge de `feat/HU-08-02-clients-search`).
- Prefijo base: `/api/v1`
- Requiere autenticación: token JWT válido en el esquema de autorización de portador (`Authorization: Bearer <access_token>`). El middleware global `AuthRequired` valida el token y que la sesión siga activa.
- Empresa: el cliente se busca siempre por `id` **y** por el `company_id` del token. Un cliente de otra empresa responde `404`, igual que uno inexistente. Un usuario sin empresa recibe `400 company is required`.
- Formato de respuesta exitosa: `{ "success": true, "data": { ... } }`.
- Formato de respuesta de error: `{ "success": false, "error": { "code": "CODIGO", "message": "descripción", "details": { "campo": "motivo" } } }`.
- Colección Bruno: `bruno/Customer/Update/` (001–019). Cada petición incluye sus `tests`.

---

## Resumen rápido

| Método | Ruta | Permiso | Descripción |
|---|---|---|---|
| `PATCH` | `/api/v1/customers/:id` | `customers.update` | Actualiza **solo** los campos enviados. **Nuevo en HU-08-04.** |
| `PUT` | `/api/v1/customers/:id` | `customers.update` | Reemplaza todos los campos editables (HU-08-01, sin cambios de contrato). |

Los dos endpoints comparten la misma lógica de servicio: búsqueda por empresa, validación de unicidad del documento y actualización dentro de una transacción.

---

## Autenticación / permisos

| Header | Obligatorio | Descripción |
|---|---|---|
| `Authorization` | Sí | `Bearer <access_token>` de una sesión activa. |
| `Content-Type` | Sí | `application/json`. |

| Permiso | Endpoint | Roles sembrados que lo tienen |
|---|---|---|
| `customers.update` | `PATCH /api/v1/customers/:id` | `superadmin` |

El permiso ya existía (HU-08-01) y es el mismo del `PUT`; esta HU no agrega permisos. `customers.read`, `customers.list` o `customers.status` **no** permiten editar.

---

## 1) Editar cliente — `PATCH /api/v1/customers/:id`

Modifica los campos enviados de un cliente existente de la empresa autenticada y devuelve el cliente actualizado. Nunca crea un cliente nuevo.

### Permiso
`customers.update`

### Params de ruta
- `id`: UUID — identificador del cliente (el que devuelve el listado de HU-08-02). Un valor que no es UUID responde `400 invalid customer id`.

### Body

JSON **parcial**: solo se envían los campos a cambiar. Los campos que no se envían conservan su valor.

```json
{
  "phone": "+59171111111",
  "email": "ventas@central.com"
}
```

#### Campos editables

| Campo | Tipo | Reglas (las mismas del `PUT`) | ¿Se puede borrar? |
|---|---|---|---|
| `legal_name` | string | 1–160 caracteres; se recortan espacios. No puede ser vacío, en blanco ni `null`. | No |
| `document_type` | enum | `national_id` \| `tax_id` \| `passport` \| `other`. No puede ser `null`. | No |
| `document_number` | string | Máximo 30 caracteres. Único por empresa y tipo de documento entre clientes no eliminados. | Sí |
| `phone` | string | Máximo 30 caracteres. | Sí |
| `email` | string | Formato email, máximo 160 caracteres. **No es único.** | Sí |
| `address` | string | Máximo 200 caracteres. | Sí |

Para borrar un campo opcional se envía `null` o una cadena vacía (`""`): el valor queda en `NULL` y se omite en la respuesta.

#### Campos no editables

Si llegan en el body responden `400 VALIDATION_ERROR` con un detalle por campo (`"<campo> is not editable"`):

`id`, `customer_id`, `company_id`, `status` (se cambia con `PATCH /:id/status`, HU-08-05), `credit_limit`, `credit_balance`, `points_accrued`, `created_at`, `updated_at`, `deleted_at`.

Cualquier otra clave desconocida también responde `400` (`"<campo> is not a known field"`).

> **Patrón nuevo en el repo:** es el primer endpoint que valida el body de forma estricta (rechaza claves desconocidas y distingue "no enviado" de "enviado vacío" con punteros). Los demás endpoints (`PUT`, `POST`, usuarios, productos) usan `BodyParser`, que ignora en silencio los campos desconocidos.

### Ejemplo de request

```bash
curl -s -X PATCH "http://localhost:4600/api/v1/customers/487b129b-e91b-45b5-8d41-6e5d4da4d27a" \
  -H "Authorization: Bearer $ACCESS_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"phone": "+59171111111"}'
```

### Respuesta exitosa

**`200 OK`** — el cliente actualizado, con la misma estructura de `GET /api/v1/customers/:id`:

```json
{
  "success": true,
  "data": {
    "id": "487b129b-e91b-45b5-8d41-6e5d4da4d27a",
    "company_id": "f72b06af-6c88-4559-be2d-024cb28d8f86",
    "legal_name": "Ferreteria Central",
    "document_type": "tax_id",
    "document_number": "1020304050",
    "phone": "+59171111111",
    "email": "central@example.com",
    "credit_limit": "0",
    "credit_balance": "0",
    "points_accrued": 0,
    "status": "active",
    "created_at": "2026-10-05T17:54:51Z",
    "updated_at": "2026-10-05T17:54:54Z"
  }
}
```

### Códigos de error

| Code | HTTP | Causa |
|---|---|---|
| `VALIDATION_ERROR` | 400 | Body vacío o `{}`; campo no editable o desconocido; formato o longitud inválidos; `legal_name`/`document_type` vacíos o `null`; valor que no es string. Todos los errores de campo se devuelven juntos en `details`. |
| `BAD_REQUEST` | 400 | `id` no es UUID; el body no es un objeto JSON (malformado, array, `null`); el token no tiene empresa. |
| `UNAUTHORIZED` | 401 | Falta el token o es inválido/expiró, o la sesión fue cerrada. |
| `FORBIDDEN` | 403 | El usuario no tiene el permiso `customers.update`. |
| `NOT_FOUND` | 404 | El cliente no existe, está eliminado o pertenece a otra empresa. |
| `CONFLICT` | 409 | El tipo y número de documento ya pertenecen a **otro** cliente no eliminado de la misma empresa. |
| `INTERNAL_ERROR` | 500 | Error inesperado de persistencia. |

**`400`** — campos no editables o desconocidos

```json
{
  "success": false,
  "error": {
    "code": "VALIDATION_ERROR",
    "message": "validation failed",
    "details": {
      "company_id": "company_id is not editable",
      "nickname": "nickname is not a known field",
      "status": "status is not editable"
    }
  }
}
```

**`400`** — body vacío

```json
{
  "success": false,
  "error": {
    "code": "VALIDATION_ERROR",
    "message": "at least one editable field is required"
  }
}
```

**`400`** — formato inválido

```json
{
  "success": false,
  "error": {
    "code": "VALIDATION_ERROR",
    "message": "validation failed",
    "details": {
      "email": "email must be a valid email address",
      "legal_name": "legal_name cannot be empty"
    }
  }
}
```

**`401`**

```json
{ "success": false, "error": { "code": "UNAUTHORIZED", "message": "unauthorized" } }
```

**`403`**

```json
{ "success": false, "error": { "code": "FORBIDDEN", "message": "forbidden" } }
```

**`404`**

```json
{ "success": false, "error": { "code": "NOT_FOUND", "message": "customer not found" } }
```

**`409`**

```json
{
  "success": false,
  "error": {
    "code": "CONFLICT",
    "message": "a customer with this document already exists in this company",
    "details": {
      "document_number": "already registered for another customer"
    }
  }
}
```

---

## Reglas de negocio

- **Empresa:** el cliente se busca por `id` y `company_id` del token; nunca se revela si existe en otra empresa (responde 404).
- **Unicidad:** solo el documento es único, por empresa y **tipo** de documento (`company_id`, `document_type`, `document_number`), entre clientes no eliminados. La validación excluye al propio cliente: reenviar su mismo documento responde `200`. Mismo número con otro tipo de documento, o en otra empresa, está permitido. El email **no** es único.
- **Sin creación:** el repositorio actualiza con un `UPDATE ... WHERE customer_id = ? AND company_id = ? AND deleted_at IS NULL`. Si no afecta ninguna fila responde `404`; nunca inserta (antes se usaba `Save` de GORM, que hace `INSERT` cuando el `UPDATE` no afecta filas). Se conservan `customer_id`, `company_id` y `created_at`.
- **Atomicidad:** lectura, chequeo de unicidad y actualización se ejecutan en una sola transacción; si algo falla no queda ningún cambio parcial.
- **Concurrencia:** dos peticiones simultáneas que intentan usar el mismo documento las resuelve el índice único parcial `idx_customer_company_document_unique`; la violación (`SQLSTATE 23505`) se traduce al mismo `409`, tanto en `PATCH` como en `PUT` y `POST`.
- **Estado:** se pueden editar clientes `active`, `inactive` y `blocked`. El `PATCH` no cambia el estado (eso es de HU-08-05). Los clientes eliminados no se pueden editar (404).
- **Cancelar la edición** (criterio 8) es responsabilidad del front: si no se envía el `PATCH`, no se modifica nada.

---

## Índice único (commit separado)

```sql
CREATE UNIQUE INDEX idx_customer_company_document_unique
  ON customer (company_id, document_type, document_number)
  WHERE deleted_at IS NULL AND document_number IS NOT NULL;
```

Se declara en `CustomerModel` y lo crea `AutoMigrate` (`make migrate-up`). **Si en la base ya hay documentos duplicados, la migración falla.** Antes de aplicarla, ejecutar:

```sql
SELECT company_id,
       document_type,
       document_number,
       COUNT(*)                                   AS total,
       array_agg(customer_id ORDER BY created_at) AS customer_ids
FROM customer
WHERE deleted_at IS NULL
  AND document_number IS NOT NULL
GROUP BY company_id, document_type, document_number
HAVING COUNT(*) > 1
ORDER BY total DESC;
```

Si devuelve filas, hay que corregir esos clientes (cambiar el documento o darlos de baja) antes de migrar.

---

## Cobertura de criterios de aceptación

| Criterio HU-08-04 | Cómo se cubre |
|---|---|
| 1. Iniciar la edición desde el detalle | El front usa el `id` del detalle (`GET /:id`) en `PATCH /:id`. |
| 2. Mostrar los valores actuales | `GET /api/v1/customers/:id` (HU-08-03, ya en `dev`). |
| 3. Solo campos editables | Lista blanca de 6 campos; cualquier otro responde 400 con detalle. |
| 4. Validar los nuevos valores | Mismas reglas de formato y longitud del `PUT`; todos los errores en `details`. |
| 5. Integridad y unicidad | 409 por documento duplicado (servicio + índice único); 404 para otra empresa. |
| 6. Actualizar el cliente existente | `UPDATE` del registro identificado por `id` y `company_id`. |
| 7. No crear un cliente nuevo | Update estricto sin `INSERT`; probado con el conteo de clientes. |
| 8. Cancelar sin modificar | Sin petición no hay cambios; los PATCH rechazados no modifican nada (probado). |
| 9. Mostrar la información actualizada | La respuesta 200 trae el cliente completo, igual a `GET /:id`. |

---

## Pruebas

- Unitarias HTTP: `internal/modules/customer/interfaces/http/handler/update_handler_test.go` (`go test ./...`).
- Integración con PostgreSQL (`-tags integration`): `repository_integration_test.go` (update estricto, rollback) y `document_index_integration_test.go` (índice único).
- Bruno:

```bash
cd bruno
bru run Customer/Update --env development \
  --env-var baseUrl=http://localhost:4600 \
  --env-var companyAdminPassword=<contraseña del seed>
```

La carpeta crea sus propios datos (prefijo `HU0804-<timestamp>`), termina con un `GET /:id` que verifica que los cambios persistieron y un `PATCH` de restauración, por lo que puede ejecutarse varias veces. Los casos "cliente de otra empresa" y "documento usado en otra empresa" no están en Bruno porque el seed tiene una sola empresa; están cubiertos en las pruebas de Go.

---

## Diferencias respecto a HU-08-01

1. Nuevo `PATCH /api/v1/customers/:id` (parcial y estricto). El `PUT` mantiene su contrato.
2. `PUT` y `PATCH` comparten el flujo del servicio, que ahora corre en una transacción.
3. El repositorio ya no usa `Save` (que podía insertar); `Update` es estricto y responde 404 si no hay fila.
4. El `409` de documento duplicado incluye `details.document_number` (también en `PUT` y en la violación del índice en `POST`).
5. Índice único parcial sobre el documento (commit separado).
