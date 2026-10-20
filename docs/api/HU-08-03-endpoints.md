# Reporte de endpoints — [HU-08-03] Visualización de cliente

## Información general

- HU: HU-08-03 — Visualización de cliente (RF-08, P0, Sprint 1). Depende de HU-08-02.
- Rama: `feat/HU-08-clients-management`
- Prefijo base: `/api/v1`
- Requiere autenticación: token JWT válido en el esquema de autorización de portador (`Authorization: Bearer <access_token>`). El middleware global `AuthRequired` valida el token y que la sesión siga activa.
- Empresa: el cliente se busca siempre por `id` **y** por el `company_id` del token. Un cliente de otra empresa responde `404`, igual que uno inexistente. Un usuario sin empresa (por ejemplo, el superadmin global `admin@sigif.com`) recibe `400 company is required`.
- Formato de respuesta exitosa: `{ "success": true, "data": { ... } }`.
- Formato de respuesta de error: `{ "success": false, "error": { "code": "CODIGO", "message": "descripción", "details": { "campo": "motivo" } } }`. `details` solo aparece en errores de validación.
- Colección Bruno: `bruno/Customer/Detail/` (001–012). Cada petición incluye sus `tests`.

El endpoint ya existía desde la HU-08-01. Esta HU **no cambia su contrato ni su código**: lo verifica contra los criterios de aceptación y agrega las pruebas, este documento y la colección Bruno.

---

## Resumen rápido

| Método | Ruta | Permiso | Descripción |
|---|---|---|---|
| `GET` | `/api/v1/customers/:id` | `customers.read` | Devuelve la información registrada de un cliente de la empresa, con su estado actual. |

---

## Autenticación / permisos

| Header | Obligatorio | Descripción |
|---|---|---|
| `Authorization` | Sí | `Bearer <access_token>` de una sesión activa. |

| Permiso | Endpoint | Roles sembrados que lo tienen |
|---|---|---|
| `customers.read` | `GET /api/v1/customers/:id` | `superadmin` |

`customers.list` **no** basta para ver el detalle. El rol `business_admin` no tiene permisos de clientes y recibe **403**.

---

## 1) Ver cliente — `GET /api/v1/customers/:id`

Devuelve el cliente seleccionado. La operación es de **solo lectura**: no modifica ningún registro.

### Permiso
`customers.read`

### Params de ruta

| Parámetro | Tipo | Descripción |
|---|---|---|
| `id` | UUID | Identificador del cliente (el `id` de cada ítem del listado de la HU-08-02). |

### Query params
Ninguno.

### Body
Ninguno.

### Ejemplo de request

```bash
curl -s "http://localhost:4600/api/v1/customers/86d60c05-c7c6-4e4f-89b2-8a308817300e" \
  -H "Authorization: Bearer $ACCESS_TOKEN"
```

### Respuesta exitosa

**`200 OK`**

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
    "status": "active",
    "created_at": "2026-10-13T14:20:00Z",
    "updated_at": "2026-10-13T15:02:11Z"
  }
}
```

| Campo | Tipo | Descripción |
|---|---|---|
| `id` | UUID | Identificador del cliente. |
| `company_id` | UUID | Empresa dueña del cliente (siempre la del token). |
| `legal_name` | string | Nombre o razón social. |
| `document_type` | enum | `national_id` \| `tax_id` \| `passport` \| `other`. |
| `document_number` | string | Opcional. |
| `phone`, `email`, `address` | string | Opcionales. |
| `credit_limit`, `credit_balance` | string decimal | Solo lectura; los gestionará el módulo de créditos (RF-55). |
| `points_accrued` | integer | Solo lectura; lo gestionará el módulo de fidelización (RF-57). |
| `status` | enum | Estado actual: `active` \| `inactive` \| `blocked`. |
| `created_at` | fecha RFC 3339 | Fecha de registro. |
| `updated_at` | fecha RFC 3339 | Última modificación. |

Los campos opcionales sin valor (`document_number`, `phone`, `email`, `address`, `updated_at`) se omiten. `deleted_at` nunca se expone.

### Códigos de error

| Code | HTTP | Causa |
|---|---|---|
| `BAD_REQUEST` | 400 | `id` no es un UUID válido (`invalid customer id`), o el token no tiene empresa (`company is required`). |
| `UNAUTHORIZED` | 401 | Falta el token o es inválido/expiró, o la sesión fue cerrada. |
| `FORBIDDEN` | 403 | El usuario no tiene el permiso `customers.read`. |
| `NOT_FOUND` | 404 | El cliente no existe, pertenece a otra empresa o fue eliminado. |
| `INTERNAL_ERROR` | 500 | Error inesperado de persistencia. |

**`400`**

```json
{
  "success": false,
  "error": {
    "code": "BAD_REQUEST",
    "message": "invalid customer id"
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

- **Empresa:** el cliente se busca por `id` y `company_id` del token. El mensaje de un cliente de otra empresa es el mismo que el de uno inexistente: no se revela que existe.
- **Estado:** se devuelven clientes en cualquier estado (`active`, `inactive`, `blocked`); el detalle siempre informa el estado actual.
- **Eliminados:** un cliente eliminado (HU-08-06) responde `404`.
- **Solo lectura:** la consulta no escribe en la base de datos.

---

## Cobertura de criterios de aceptación

| Criterio HU-08-03 | Cómo se cubre |
|---|---|
| Ver la información registrada del cliente seleccionado | `GET /api/v1/customers/:id` devuelve el cliente completo. |
| Mostrar el estado actual | Campo `status` en la respuesta. |
| No modificar la información | El servicio solo lee; probado en `TestGetCustomerEndpointDoesNotModifyData`. |
| Informar si el cliente no existe | `404 NOT_FOUND` con `customer not found`. |
| Validar el id | `400 BAD_REQUEST` si no es un UUID. |
| Verificar la autorización | `401` sin token y `403` sin `customers.read`. |
| Retornar solo los datos permitidos | Solo clientes de la empresa del token; `deleted_at` no se expone. |

---

## Pruebas

- Unitarias HTTP: `internal/modules/customer/interfaces/http/handler/get_handler_test.go` (`go test ./...`).
- Bruno (requiere la API levantada y la base sembrada; usa `admin.empresa@sigif.com`, que sí tiene empresa, y `negocio@sigif.com` como usuario sin permisos de clientes):

```bash
cd bruno
bru run Customer/Detail --env development \
  --env-var baseUrl=http://localhost:4600 \
  --env-var companyAdminPassword=<contraseña del seed> \
  --env-var businessAdminPassword=<contraseña del seed>
```

La carpeta crea su propio cliente (prefijo único `HU0803-<timestamp>`) y lo elimina al terminar, por lo que puede ejecutarse varias veces. El servidor limita a 100 peticiones por minuto: si se ejecutan varias carpetas seguidas puede responder `429`; basta esperar un minuto.

---

## Comportamiento no obvio

- Las fechas se devuelven en RFC 3339, pero la zona horaria depende del origen del dato: las respuestas de una escritura (`POST`, `PUT`, `PATCH`) salen en UTC (`Z`) y las de una lectura salen con el desfase del servidor (por ejemplo `-04:00`). Es el mismo instante; el frontend debe interpretar la fecha, no compararla como texto.
