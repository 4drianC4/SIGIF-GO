# Reporte de endpoints — [HU-10-02] Restricción de uso simultáneo de caja

## Información general

- HU: HU-10-02 — Restricción de uso simultáneo de caja (RF-10, P0, Sprint 1). Depende de RF-01 y de HU-10-01.
- Rama: `feat/HU-10-cash-register-assignment`
- Prefijo base: `/api/v1`
- Requiere autenticación: token JWT válido en el esquema de autorización de portador (`Authorization: Bearer <access_token>`). El middleware global `AuthRequired` valida el token y que la sesión siga activa.
- Empresa y cajero: ambos salen **siempre del token** (`company_id` y `user_id`); nunca se aceptan en el body ni en la query. Un usuario sin empresa (por ejemplo, el superadmin global `admin@sigif.com`) recibe `400 company is required`.
- Formato de respuesta de error: `{ "success": false, "error": { "code": "CODIGO", "message": "descripción", "details": { "campo": "motivo" } } }`. `details` solo aparece en errores de validación.
- Colección Bruno: `bruno/CashRegister/Concurrency/` (001–023). Cada petición incluye sus `tests`.

---

## Resumen rápido

| Método | Ruta | Permiso | Descripción |
|---|---|---|---|
| `GET` | `/api/v1/cash-registers/?availability=in_use` | `cash_registers.list` | Identifica las cajas en uso y su cajero. |
| `POST` | `/api/v1/cash-register-assignments/` | `cash_registers.assign` | Rechaza una caja ocupada (`409 CASH_REGISTER_IN_USE`) o un cajero que ya tiene caja (`409 CASHIER_ALREADY_ASSIGNED`). |
| `POST` | `/api/v1/cash-register-assignments/:id/release` | `cash_registers.release` o `cash_registers.release_any` | Libera la caja al finalizar el turno. |

El listado, la asignación y la caja actual se documentan en `HU-10-01-endpoints.md`.

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

## 1) Caja ocupada — `POST /api/v1/cash-register-assignments/`

Mismo endpoint que en la HU-10-01. Si la caja ya está asignada a otro cajero, se rechaza sin modificar nada:

**`409`**

```json
{
  "success": false,
  "error": {
    "code": "CASH_REGISTER_IN_USE",
    "message": "cash register is already assigned to another cashier"
  }
}
```

Si el cajero ya tiene una caja activa (otra, o la misma pedida de nuevo):

**`409`**

```json
{
  "success": false,
  "error": {
    "code": "CASHIER_ALREADY_ASSIGNED",
    "message": "cashier already has an active cash register; release it first"
  }
}
```

Cuando la caja está ocupada por otro cajero y además el solicitante ya tiene caja, se informa primero `CASH_REGISTER_IN_USE`.

---

## 2) Liberar caja — `POST /api/v1/cash-register-assignments/:id/release`

Finaliza la asignación y deja la caja disponible para otro cajero. **No borra nada:** cambia `status` a `released` y registra `released_at` y `released_by`; la fila se conserva como historial.

### Permiso
`cash_registers.release` (solo la asignación propia) o `cash_registers.release_any` (cualquier asignación de la empresa). Basta con uno de los dos para entrar a la ruta; el servicio decide si el usuario puede liberar esa asignación concreta.

### Params de ruta

| Parámetro | Tipo | Descripción |
|---|---|---|
| `id` | UUID | Id de la asignación (`id` de la respuesta de asignar, `assignment_id` del listado o `id` de `/current`). |

### Body
Ninguno.

### Ejemplo de request

```bash
curl -s -X POST "http://localhost:4600/api/v1/cash-register-assignments/d01bcbb1-49a7-424c-ad20-943519ff5089/release" \
  -H "Authorization: Bearer $ACCESS_TOKEN"
```

### Respuesta exitosa

**`200 OK`**

```json
{
  "success": true,
  "data": {
    "id": "d01bcbb1-49a7-424c-ad20-943519ff5089",
    "status": "released",
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
    "assigned_at": "2026-10-15T12:00:00Z",
    "released_at": "2026-10-15T20:00:00Z",
    "released_by": "6b7a2a40-1e87-4676-a536-2308e6015dd5"
  }
}
```

**`200 OK`** — liberar una asignación ya liberada es **idempotente**: devuelve la liberación original sin escribir en la base de datos (`released_at` y `released_by` no cambian).

### Códigos de error

| Code | HTTP | Causa |
|---|---|---|
| `BAD_REQUEST` | 400 | `id` no es un UUID (`invalid assignment id`), o el token no tiene empresa. |
| `UNAUTHORIZED` | 401 | Falta el token o es inválido/expiró. |
| `FORBIDDEN` | 403 | Sin `cash_registers.release` ni `cash_registers.release_any`, o es la asignación de otro cajero y el usuario no tiene `release_any`. |
| `NOT_FOUND` | 404 | La asignación no existe o es de otra empresa. |

```json
{
  "success": false,
  "error": {
    "code": "FORBIDDEN",
    "message": "only the assigned cashier or a supervisor can release this cash register"
  }
}
```

---

## Reglas de negocio

- **Una caja, un cajero:** una caja con una asignación `active` no puede asignarse a otro cajero (`409 CASH_REGISTER_IN_USE`).
- **Un cajero, una caja:** un cajero con una asignación `active` no puede tomar otra caja hasta liberar la suya (`409 CASHIER_ALREADY_ASSIGNED`).
- **Concurrencia:**
  1. La verificación (caja activa y libre, cajero sin caja) y la inserción se ejecutan en **una sola transacción**.
  2. Si dos peticiones pasan la verificación a la vez, los **índices únicos parciales** `... WHERE status = 'active'` rechazan la segunda en la base de datos.
  3. El repositorio traduce esa violación (`23505`) al mismo `409`, según el índice violado.

  El proyecto no usa `SELECT ... FOR UPDATE` en ningún módulo, así que no se introdujo; la garantía la dan los índices. El test de integración `TestConcurrentAssignmentOfSameRegister` lanza 20 cajeros a la vez sobre la misma caja: exactamente uno obtiene la asignación y queda una sola fila activa.
- **Liberar:** solo el cajero asignado o un usuario con `release_any` (supervisor). Otro cajero recibe `403`. La liberación es idempotente, y dos liberaciones simultáneas escriben una sola vez (el `UPDATE` exige `status = 'active'`).
- **Historial:** las asignaciones liberadas nunca se borran; sirven para el historial de sesiones (RF-19).
- **Pertenencia caja ↔ cajero (BE-3):** el servicio expone `CashRegisterService.EnsureCashierOwnsRegister(ctx, companyID, userID, cashRegisterID)`:

  | Situación | Resultado |
  |---|---|
  | La caja está asignada a ese cajero | `nil` |
  | Asignada a otro cajero | `409 CASH_REGISTER_IN_USE` |
  | Libre o liberada | `403 FORBIDDEN` (`cash register is not assigned to this cashier`) |
  | No existe o es de otra empresa | `404 NOT_FOUND` |

  Es lo que ventas (RF-11) y la sesión de caja (RF-09) deben llamar para cumplir el criterio 3. Los otros módulos lo consumen con un adaptador propio, igual que `auth` consume `user` (`auth/application/adapter`).
- **Aislamiento multiempresa:** todas las consultas filtran por el `company_id` del token. Las cajas y asignaciones de otra empresa responden `404`.

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

## Relación con RF-09, RF-17 y RF-20 (propuesta para el equipo)

La asignación representa **el turno**: qué cajero tiene qué caja. Para no duplicar datos se propone:

- **RF-09 (apertura, `sesion_caja`):**
  - La sesión guarda `cash_register_assignment_id` (FK a `cash_register_assignment`) más sus datos monetarios: monto inicial, fecha de apertura y de cierre, estado.
  - **No** repite caja, cajero ni estado de ocupación.
  - Antes de abrir, llama a `EnsureCashierOwnsRegister(ctx, companyID, userID, cashRegisterID)`. Así se cumple «no se puede operar sin caja asignada» y no pueden existir dos sesiones abiertas en la misma caja: la caja solo tiene una asignación activa.
- **RF-17 (cierre de caja) y RF-20 (entrega de turno):** al terminar, llaman a `Release` (o al endpoint de liberación) para dejar la caja disponible. RF-20 puede liberar y que el siguiente cajero asigne la caja en la misma operación.
- **RF-11 (ventas):** antes de registrar una venta, `EnsureCashierOwnsRegister`.
- **RF-19 (historial):** las asignaciones liberadas quedan como historial (`assigned_at`, `released_at`, `released_by`).

Nada de esto está implementado aquí: solo la asignación, la liberación y el método de verificación.

---

## Cobertura de criterios de aceptación

| Criterio HU-10-02 | Cómo se cubre | Prueba |
|---|---|---|
| 1. Identificar las cajas en uso | `GET /api/v1/cash-registers/?availability=in_use` con `assigned_to`. | `TestListCashRegistersEndpointAvailabilityFilter`, Bruno 013 |
| 2. Una caja en uso no puede asignarse a otro cajero | `409 CASH_REGISTER_IN_USE`; índice parcial por caja. | `TestAssignCashRegisterEndpointRejectsRegisterInUse`, `TestConcurrentAssignmentOfSameRegister` |
| 3. Rechazar el uso de una caja ocupada por otro cajero | `409` al asignar; `EnsureCashierOwnsRegister` para los módulos que operan. | `TestEnsureCashierOwnsRegister` |
| 4. Informar que la caja ya está asignada | Mensaje `cash register is already assigned to another cashier`. | Bruno 010 |
| 5. Reasignar la caja al finalizar el turno | `POST .../:id/release` → caja `available`; otro cajero la toma (`201`). | `TestReleaseAssignmentEndpointFreesRegister`, Bruno 015–018 |
| BE: impedir asignaciones simultáneas | Transacción + índices parciales + traducción a `409`. | `TestConcurrentAssignmentOfSameRegister`, `TestConcurrentAssignmentOfSameCashier` |
| BE: validar que la caja pertenezca al cajero | `EnsureCashierOwnsRegister`. | `TestEnsureCashierOwnsRegister` |
| BE: liberar la caja | Liberación sin borrado, idempotente. | `TestReleaseKeepsHistoryAndIsIdempotent` |

---

## Pruebas

- Unitarias HTTP: `internal/modules/cashregister/interfaces/http/handler/concurrency_handler_test.go`.
- Dominio: `internal/modules/cashregister/domain/service/service_test.go`.
- Integración con PostgreSQL (concurrencia real con goroutines): `internal/modules/cashregister/infrastructure/persistence/gorm/repository_integration_test.go`.
- Bruno:

```bash
cd bruno
bru run CashRegister/Concurrency --env development \
  --env-var baseUrl=http://localhost:4600 \
  --env-var businessAdminPassword=<contraseña del seed> \
  --env-var cashierAPassword=<contraseña del seed> \
  --env-var cashierBPassword=<contraseña del seed>
```

La carpeta inicia sesión con `cajero.a`, `cajero.b` y `negocio@sigif.com` (variables `cashierAEmail`, `cashierBEmail` y `businessAdminEmail`; contraseñas como secreto), libera al inicio cualquier asignación previa de los cajeros y libera al final las que crea, por lo que puede ejecutarse varias veces.

---

## Fuera de alcance

- Apertura con monto inicial (RF-09).
- Cierre y arqueo (RF-16/17).
- Entrega de turno (RF-20).
- Supervisión remota (RF-21/22).
- Registro, edición y baja de cajas.
