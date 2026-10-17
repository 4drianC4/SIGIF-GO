# Reporte de endpoints — [HU-10-01] Asignar caja

## Información general

- HU: HU-10-01 — Asignar caja (RF-10, P0, Sprint 1). Depende de RF-01.
- Rama: `feat/HU-10-cash-register-assignment`
- Prefijo base: `/api/v1`
- Requiere autenticación: token JWT válido en el esquema de autorización de portador (`Authorization: Bearer <access_token>`). El middleware global `AuthRequired` valida el token y que la sesión siga activa.
- Empresa y cajero: ambos salen **siempre del token** (`company_id` y `user_id`); nunca se aceptan en el body ni en la query. Un usuario sin empresa (por ejemplo, el superadmin global `admin@sigif.com`) recibe `400 company is required`.
- Formato de respuesta de error: `{ "success": false, "error": { "code": "CODIGO", "message": "descripción", "details": { "campo": "motivo" } } }`. `details` solo aparece en errores de validación.
- Colección Bruno: `bruno/CashRegister/Assign/` (001–021). Cada petición incluye sus `tests`.

---

## Resumen rápido

| Método | Ruta | Permiso | Descripción |
|---|---|---|---|
| `GET` | `/api/v1/cash-registers/` | `cash_registers.list` | Lista las cajas activas de la empresa con su disponibilidad. |
| `POST` | `/api/v1/cash-register-assignments/` | `cash_registers.assign` | Asigna una caja disponible al cajero del token. |
| `GET` | `/api/v1/cash-register-assignments/current` | `cash_registers.assign` | Devuelve la caja asignada al cajero durante su turno. |

La liberación (`POST /api/v1/cash-register-assignments/:id/release`) y las reglas de uso simultáneo están en `HU-10-02-endpoints.md`.

---

## Autenticación / permisos

| Header | Obligatorio | Descripción |
|---|---|---|
| `Authorization` | Sí | `Bearer <access_token>` de una sesión activa. |

| Permiso | Uso | Roles sembrados que lo tienen |
|---|---|---|
| `cash_registers.list` | `GET /api/v1/cash-registers/` | `superadmin`, `business_admin`, `employee` |
| `cash_registers.assign` | `POST /api/v1/cash-register-assignments/` y `GET /api/v1/cash-register-assignments/current` | `superadmin`, `employee` |
| `cash_registers.release` | Liberar la asignación propia | `superadmin`, `employee` |
| `cash_registers.release_any` | Liberar la asignación de cualquier cajero de la empresa | `superadmin`, `business_admin` |

No existe un rol «cajero»: el rol `employee` trabaja como cajero (hasta ahora no tenía permisos). `business_admin` supervisa (ve las cajas en uso y puede liberarlas) pero **no** puede tomar una caja. El seed crea dos cajeros de prueba en la empresa demo: `cajero.a@sigif.com` y `cajero.b@sigif.com` (contraseñas en `SIGIF_SEED_CASHIER_A_PASSWORD` y `SIGIF_SEED_CASHIER_B_PASSWORD`).

---

## 1) Listar cajas — `GET /api/v1/cash-registers/`

Devuelve las cajas **activas y no eliminadas** de la empresa del token, ordenadas por código, indicando si cada una está disponible o en uso. Es de solo lectura.

### Permiso
`cash_registers.list`

### Query params

| Parámetro | Obligatorio / por defecto | Tipo | Descripción |
|---|---|---|---|
| `availability` | Opcional, por defecto `all` | enum | `available` \| `in_use` \| `all`. |
| `page` | Opcional, por defecto `1` | integer ≥ 1 | Página. |
| `limit` | Opcional, por defecto `20` | integer 1–100 | Tamaño de página (`pagination.max_limit`). |

### Ejemplo de request

```bash
curl -s "http://localhost:4600/api/v1/cash-registers/?availability=all&page=1&limit=20" \
  -H "Authorization: Bearer $ACCESS_TOKEN"
```

### Respuesta exitosa

**`200 OK`**

```json
{
  "success": true,
  "data": [
    {
      "id": "cf6b6bb9-38f6-4050-90ed-1dcc92da9875",
      "code": "CAJA-01",
      "name": "Caja 1",
      "status": "active",
      "availability": "in_use",
      "assignment_id": "d01bcbb1-49a7-424c-ad20-943519ff5089",
      "assigned_to": {
        "id": "6b7a2a40-1e87-4676-a536-2308e6015dd5",
        "name": "Cajero A"
      },
      "assigned_at": "2026-10-15T12:00:00Z"
    },
    {
      "id": "0f3c6a51-1d2b-4c4e-9a43-7b2a51c0e9d1",
      "code": "CAJA-02",
      "name": "Caja 2",
      "status": "active",
      "availability": "available"
    }
  ],
  "meta": {
    "page": 1,
    "limit": 20,
    "total": 2,
    "total_pages": 1
  }
}
```

`assignment_id`, `assigned_to` y `assigned_at` solo aparecen en las cajas en uso. Sin coincidencias responde `200` con `data: []`.

### Códigos de error

| Code | HTTP | Causa |
|---|---|---|
| `VALIDATION_ERROR` | 400 | `availability` fuera de los valores permitidos, `page` < 1, `limit` < 1 o mayor al máximo. |
| `BAD_REQUEST` | 400 | `page` o `limit` no numéricos, o el token no tiene empresa (`company is required`). |
| `UNAUTHORIZED` | 401 | Falta el token o es inválido/expiró, o la sesión fue cerrada. |
| `FORBIDDEN` | 403 | Sin el permiso `cash_registers.list`. |

```json
{
  "success": false,
  "error": {
    "code": "VALIDATION_ERROR",
    "message": "validation failed",
    "details": {
      "availability": "availability must be one of [available in_use all]"
    }
  }
}
```

---

## 2) Asignar caja — `POST /api/v1/cash-register-assignments/`

Registra la asignación de una caja disponible al cajero del token, al iniciar su turno.

### Permiso
`cash_registers.assign`

### Body

```json
{
  "cash_register_id": "cf6b6bb9-38f6-4050-90ed-1dcc92da9875"
}
```

| Campo | Tipo | Obligatorio | Descripción |
|---|---|---|---|
| `cash_register_id` | UUID | Sí | Caja elegida (el `id` del listado). |

Cualquier otro campo (por ejemplo `user_id` o `company_id`) se ignora: el cajero y la empresa salen del token.

### Ejemplo de request

```bash
curl -s -X POST "http://localhost:4600/api/v1/cash-register-assignments/" \
  -H "Authorization: Bearer $ACCESS_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"cash_register_id":"cf6b6bb9-38f6-4050-90ed-1dcc92da9875"}'
```

### Respuesta exitosa

**`201 Created`**

```json
{
  "success": true,
  "data": {
    "id": "d01bcbb1-49a7-424c-ad20-943519ff5089",
    "status": "active",
    "cash_register": {
      "id": "cf6b6bb9-38f6-4050-90ed-1dcc92da9875",
      "code": "CAJA-01",
      "name": "Caja 1",
      "status": "active"
    },
    "cashier": {
      "id": "6b7a2a40-1e87-4676-a536-2308e6015dd5",
      "name": "Cajero A"
    },
    "assigned_at": "2026-10-15T12:00:00Z"
  }
}
```

### Códigos de error

| Code | HTTP | Causa |
|---|---|---|
| `VALIDATION_ERROR` | 400 | Falta `cash_register_id` o no es un UUID. |
| `BAD_REQUEST` | 400 | Body que no es JSON válido (`invalid request body`), o el token no tiene empresa. |
| `UNAUTHORIZED` | 401 | Falta el token o es inválido/expiró. |
| `FORBIDDEN` | 403 | Sin el permiso `cash_registers.assign` (por ejemplo, `business_admin`). |
| `NOT_FOUND` | 404 | La caja no existe, es de otra empresa, está inactiva o eliminada (`cash register not found`). |
| `CASH_REGISTER_IN_USE` | 409 | La caja ya está asignada a otro cajero (ver HU-10-02). |
| `CASHIER_ALREADY_ASSIGNED` | 409 | El cajero ya tiene una caja activa (ver HU-10-02). |

```json
{
  "success": false,
  "error": {
    "code": "NOT_FOUND",
    "message": "cash register not found"
  }
}
```

---

## 3) Caja asignada — `GET /api/v1/cash-register-assignments/current`

Devuelve la asignación activa del cajero del token, para mostrar la caja durante el turno.

### Permiso
`cash_registers.assign`

### Ejemplo de request

```bash
curl -s "http://localhost:4600/api/v1/cash-register-assignments/current" \
  -H "Authorization: Bearer $ACCESS_TOKEN"
```

### Respuestas

**`200 OK`** — misma estructura que la respuesta de asignar.

**`404`** — el cajero no tiene una caja asignada:

```json
{
  "success": false,
  "error": {
    "code": "NO_ACTIVE_ASSIGNMENT",
    "message": "cashier has no active cash register assignment"
  }
}
```

Además: `400` (token sin empresa), `401` y `403` como en los demás endpoints.

---

## Reglas de negocio

- **Cajas visibles:** solo las activas y no eliminadas de la empresa del token. Una caja inactiva no se lista ni se puede asignar (`404`).
- **Disponibilidad:** una caja está `in_use` mientras tiene una asignación `active`; si no, está `available`.
- **Cajero:** siempre el usuario del token. En esta HU no se puede asignar una caja a otra persona.
- **Uso simultáneo, concurrencia y liberación:** ver `HU-10-02-endpoints.md`.
- **Fechas:** se devuelven en RFC 3339 y en UTC.

---

## Modelo de datos e índices

El módulo `cashregister` crea sus tablas con `AutoMigrate` (`make migrate-up`); el modelo y los índices están en `internal/modules/cashregister/infrastructure/persistence/model/cash_register_model.go`.

**`cash_register`** — caja física

| Columna | Tipo | Notas |
|---|---|---|
| `cash_register_id` | uuid | PK. |
| `company_id` | uuid | Empresa dueña. |
| `code` | varchar(20) | Único por empresa entre cajas no eliminadas. |
| `name` | varchar(80) | Nombre visible. |
| `status` | varchar(20) | `active` \| `inactive` (check). |
| `created_at`, `updated_at`, `deleted_at` | timestamptz | Soft delete. |

**`cash_register_assignment`** — asignación caja ↔ cajero (el turno)

| Columna | Tipo | Notas |
|---|---|---|
| `assignment_id` | uuid | PK. |
| `company_id` | uuid | Empresa (del token). |
| `cash_register_id` | uuid | FK a `cash_register` (`ON DELETE RESTRICT`). |
| `user_id` | uuid | Cajero (del token). Sin FK a `app_user`, para no acoplar módulos en la base. |
| `status` | varchar(20) | `active` \| `released` (check). |
| `assigned_at` | timestamptz | Inicio de la asignación. |
| `released_at`, `released_by` | timestamptz, uuid | Se informan al liberar. |
| `created_at`, `updated_at` | timestamptz | |

| Índice | Definición | Regla |
|---|---|---|
| `idx_cash_register_assignment_active_register` | `UNIQUE (cash_register_id) WHERE status = 'active'` | Una caja, un solo cajero a la vez. |
| `idx_cash_register_assignment_active_user` | `UNIQUE (user_id) WHERE status = 'active'` | Un cajero, una sola caja a la vez. |
| `idx_cash_register_company_code` | `UNIQUE (company_id, code) WHERE deleted_at IS NULL` | Código de caja único por empresa. |
| `idx_cash_register_company_status`, `idx_cash_register_assignment_company_status` | `(company_id, status)` | Filtros por empresa. |

**Cajas de prueba:** no hay endpoint para registrar cajas. El seed (idempotente) crea en la empresa demo `CAJA-01` y `CAJA-02` activas y `CAJA-03` inactiva.

---

## Cobertura de criterios de aceptación

| Criterio HU-10-01 | Cómo se cubre | Prueba |
|---|---|---|
| 1. Mostrar las cajas disponibles (BE-01) | `GET /api/v1/cash-registers/?availability=available`. | `TestListCashRegistersEndpointAvailabilityFilter`, Bruno 010–011 |
| 2. Seleccionar una caja disponible | El `id` de cada ítem se envía en `cash_register_id`. | `TestAssignCashRegisterEndpointAssignsAvailableRegister` |
| 3. Registrar la asignación (BE-02) | `POST /api/v1/cash-register-assignments/` → `201`. | Bruno 014 |
| BE-03. Validar la disponibilidad | Caja activa, de la empresa, sin otro cajero; el cajero sin otra caja. | `TestAssignCashRegisterEndpointNotFound`, `TestAssignCashRegisterEndpointRejectsRegisterInUse` |
| 4. Mostrar la caja asignada durante el turno | `GET /api/v1/cash-register-assignments/current`. | `TestCurrentAssignmentEndpoint`, Bruno 016 |

---

## Pruebas

- Unitarias HTTP: `internal/modules/cashregister/interfaces/http/handler/assign_handler_test.go`.
- Integración con PostgreSQL: `internal/modules/cashregister/infrastructure/persistence/gorm/repository_integration_test.go`:

```bash
SIGIF_TEST_DATABASE_DSN="host=localhost port=5432 user=sigif password=sigif dbname=sigif_test sslmode=disable" \
  go test -tags integration ./internal/modules/cashregister/...
```

- Bruno:

```bash
cd bruno
bru run CashRegister/Assign --env development \
  --env-var baseUrl=http://localhost:4600 \
  --env-var businessAdminPassword=<contraseña del seed> \
  --env-var cashierAPassword=<contraseña del seed> \
  --env-var cashierBPassword=<contraseña del seed>
```

La carpeta inicia sesión con `cajero.a`, `cajero.b` y `negocio@sigif.com` (variables `cashierAEmail`, `cashierBEmail` y `businessAdminEmail`; contraseñas como secreto), libera al inicio cualquier asignación previa de los cajeros y libera al final las que crea, por lo que puede ejecutarse varias veces.

---

## Relación con otros RF

Ver «Relación con RF-09, RF-17 y RF-20» en `HU-10-02-endpoints.md`.
