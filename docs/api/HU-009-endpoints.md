# Reporte de endpoints — [HU-009] Apertura de sesión de caja y bloqueo de transacciones sin sesión abierta

## Información general

- HU: HU-009 — Apertura de sesión de caja y bloqueo de transacciones sin sesión abierta (RF-09, P0, Sprint 1). Depende de RF-01 y de la asignación de caja de RF-10 (HU-10-01, HU-10-02).
- Rama: `feature/HU-009-cash-session-opening` (parte de `feat/HU-10-cash-register-assignment`).
- Prefijo base: `/api/v1`
- Requiere autenticación (HU-01): en el navegador, la cookie HttpOnly `sigif_access_token` que deja el login (enviar las peticiones con credenciales); en otros clientes, `Authorization: Bearer <access_token>`. Las peticiones `POST` desde el navegador deben venir de un origen permitido (`cors.allow_origins`); si no, `403` (`origin not allowed`).
- Empresa y cajero: salen **siempre del token** (`company_id` y `user_id`); nunca se aceptan en el body. Un usuario sin empresa recibe `400 company is required`.
- La sesión se abre **sobre la caja que el cajero tiene asignada** (HU-10). El frontend no envía la asignación: el backend la obtiene.
- Esta HU **no libera la caja** en ningún caso. La liberación pertenece al fin de turno y al cierre (RF-17, RF-20).
- Formato de error: `{ "success": false, "error": { "code": "CODIGO", "message": "descripción", "details": { "campo": "motivo" } } }`. `details` solo aparece en errores de validación.
- Colección Bruno: `bruno/CashSession/` (001–025).

---

## Resumen rápido

| Método | Ruta | Permiso | Descripción |
|---|---|---|---|
| `POST` | `/api/v1/cash-sessions/` | `cash_sessions.open` | Abre la sesión de caja del cajero con el monto inicial. |
| `GET` | `/api/v1/cash-sessions/current` | `cash_sessions.read` | Devuelve la sesión abierta del cajero, o `404 NO_OPEN_CASH_SESSION`. |

Además, las operaciones de caja de otros módulos (ventas, cobros, movimientos) se protegen con el guard `RequireOpenCashSession`, que responde `409 CASH_SESSION_REQUIRED` cuando no hay sesión abierta (ver sección 3).

---

## Autenticación / permisos

| Permiso | Uso | Roles sembrados que lo tienen |
|---|---|---|
| `cash_sessions.open` | `POST /api/v1/cash-sessions/` | `superadmin`, `employee` |
| `cash_sessions.read` | `GET /api/v1/cash-sessions/current` | `superadmin`, `employee` |

`employee` es el cajero (como en HU-10). `business_admin` no tiene estos permisos.

---

## 1) Abrir sesión de caja — `POST /api/v1/cash-sessions/`

### Body

| Campo | Tipo | Obligatorio | Reglas |
|---|---|---|---|
| `opening_amount` | number o string decimal | Sí | Mayor o igual a `0` (el `0` es válido). Máximo 2 decimales. Máximo `9999999999.99`. |
| `opening_notes` | string | No | Observación de apertura. Máximo 200 caracteres tras recortar espacios. Si se omite o queda vacía se guarda `null`. |
| `cash_register_id` | uuid | No | Normalmente no se envía. Si se envía, debe ser la caja asignada al cajero; si no, se rechaza con el error de HU-10 correspondiente. |

```json
{
  "opening_amount": 150.50,
  "opening_notes": "Billetes de 10 en el cajón"
}
```

### Reglas

1. El cajero debe tener una asignación activa (HU-10). Si no la tiene: `404 NO_ACTIVE_ASSIGNMENT`. No se crea ninguna asignación automáticamente.
2. La caja asignada debe estar activa. Si se puso fuera de servicio después de asignarla: `409 CASH_REGISTER_INACTIVE`.
3. Solo una sesión abierta por cajero y por caja: `409 CASH_SESSION_ALREADY_OPEN`. Lo garantizan dos índices únicos parciales en la base, también con peticiones simultáneas.
4. La sesión queda asociada al cajero, a la caja, a la asignación y a la fecha y hora de apertura, que genera el backend (UTC), con estado `open`.

### Respuesta exitosa — `201 Created`

```json
{
  "success": true,
  "data": {
    "id": "7f9d8c28-5177-4afc-8d46-a4abebc86d7a",
    "status": "open",
    "opening_amount": 150.50,
    "opening_notes": "Billetes de 10 en el cajón",
    "opened_at": "2026-10-16T10:05:31Z",
    "closed_at": null,
    "cashier_id": "15715165-ef65-4e00-aec7-cf1702ce8723",
    "assignment_id": "48dcbb68-9fed-4406-bd61-92b444bbe8f0",
    "cash_register": {
      "id": "f090d2eb-b203-4e08-baea-d322ba034dc2",
      "code": "CAJA-01",
      "name": "Caja 1",
      "status": "active"
    }
  }
}
```

| Campo | Tipo | Descripción |
|---|---|---|
| `id` | uuid | Sesión de caja. |
| `status` | `open` \| `closed` | Siempre `open` al abrir. `closed` lo pondrá RF-17. |
| `opening_amount` | number | Monto inicial, siempre con 2 decimales. |
| `opening_notes` | string \| null | Observación de apertura. |
| `opened_at` | string RFC 3339 (UTC) | Fecha y hora de apertura. |
| `closed_at` | string \| null | Siempre `null` en esta HU. |
| `cashier_id` | uuid | Usuario que abrió la sesión. |
| `assignment_id` | uuid | Asignación de HU-10 sobre la que se abrió. |
| `cash_register` | objeto | Caja: `id`, `code`, `name`, `status`. |

### Códigos de error

| Code | HTTP | Causa | Toast de Frontend |
|---|---|---|---|
| `VALIDATION_ERROR` | 400 | Monto negativo, con más de 2 decimales, demasiado grande o ausente; observación de más de 200 caracteres; `cash_register_id` que no es uuid. `details` indica el campo. | `t-monto-invalido` |
| `BAD_REQUEST` | 400 | JSON mal formado o monto que no es un número (`"abc"`). | — |
| `BAD_REQUEST` | 400 | El token no tiene empresa (`company is required`). | — |
| `UNAUTHORIZED` | 401 | Sin token, token inválido o sesión de login cerrada. | — |
| `FORBIDDEN` | 403 | Sin el permiso `cash_sessions.open`. | — |
| `FORBIDDEN` | 403 | Se envió `cash_register_id` de una caja libre, no asignada al cajero. | — |
| `NO_ACTIVE_ASSIGNMENT` | 404 | El cajero no tiene caja asignada (código de HU-10). | `t-caja-no-asignada` |
| `NOT_FOUND` | 404 | Se envió `cash_register_id` de una caja que no existe o es de otra empresa. | — |
| `CASH_SESSION_ALREADY_OPEN` | 409 | El cajero ya tiene una sesión abierta, o la caja ya tiene una sesión abierta de otro cajero. | `t-caja-abierta` |
| `CASH_REGISTER_IN_USE` | 409 | Se envió `cash_register_id` de una caja asignada a otro cajero (código de HU-10). | `t-caja-ocupada` |
| `CASH_REGISTER_INACTIVE` | 409 | La caja asignada está fuera de servicio. | `t-caja-fuera-servicio` |

Ejemplo de validación:

```json
{
  "success": false,
  "error": {
    "code": "VALIDATION_ERROR",
    "message": "validation failed",
    "details": { "opening_amount": "no puede ser negativo" }
  }
}
```

Ejemplo de sesión ya abierta:

```json
{
  "success": false,
  "error": {
    "code": "CASH_SESSION_ALREADY_OPEN",
    "message": "cashier already has an open cash session"
  }
}
```

---

## 2) Sesión actual — `GET /api/v1/cash-sessions/current`

Devuelve la sesión abierta del cajero autenticado. Sin body ni parámetros.

| Situación | Respuesta | Qué hace Frontend |
|---|---|---|
| Tiene sesión abierta | `200` con la sesión (mismo formato que la apertura) | Mostrar su caja y "Continuar al punto de venta". |
| No tiene sesión abierta | `404 NO_OPEN_CASH_SESSION` | Mostrar la apertura. |

```json
{
  "success": false,
  "error": {
    "code": "NO_OPEN_CASH_SESSION",
    "message": "cashier has no open cash session"
  }
}
```

Otros errores: `401 UNAUTHORIZED`, `403 FORBIDDEN` (sin `cash_sessions.read`), `400 BAD_REQUEST` (token sin empresa).

---

## 3) Bloqueo de transacciones sin sesión abierta

No es un endpoint. Toda ruta que mueva dinero (ventas RF-11, cobros, movimientos de caja) debe montar el guard **después** del permiso:

```go
sales.Post("/",
    middleware.RequirePermission(checker, "sales", "create"),
    middleware.RequireOpenCashSession(cashSessions), // middleware.CashSessionChecker, lo provee el módulo cashsession
    h.CreateSale,
)
```

El handler obtiene la sesión con `middleware.CashSessionIDFromContext(c.UserContext())`.

| Situación | Respuesta |
|---|---|
| El cajero tiene sesión abierta y la caja sigue asignada a él | Se ejecuta la operación. |
| No tiene sesión abierta | `409 CASH_SESSION_REQUIRED` — `"an open cash session is required: open the cash register first"`. La operación no se ejecuta. |
| Tiene sesión abierta pero liberó la caja | `403 FORBIDDEN` — `"cash register is not assigned to this cashier"` (HU-10). |
| No se pudo consultar | `500`. La operación no se ejecuta. |

Frontend debe decidir por el código `CASH_SESSION_REQUIRED` y mostrar su propio texto, por ejemplo "Primero debes abrir una caja para realizar esta operación". Se usa `409` y no `403` porque `403` significa "no tienes permiso" (HU-082-06).

---

## Modelo de datos

El módulo `cashsession` crea su tabla con `AutoMigrate`, igual que `cashregister`. La HU la llama `sesion_caja`.

**`cash_session`**

| Columna | Tipo | Notas |
|---|---|---|
| `cash_session_id` | uuid | PK. |
| `company_id` | uuid | Empresa (del token). |
| `cash_register_id` | uuid | FK a `cash_register` (`ON DELETE RESTRICT`). |
| `assignment_id` | uuid | FK a `cash_register_assignment` (`ON DELETE RESTRICT`). |
| `user_id` | uuid | Cajero (del token). Sin FK a `app_user`, como en HU-10. |
| `opening_amount` | numeric(12,2) | `CHECK (opening_amount >= 0)`. |
| `opening_notes` | varchar(200) | Nulo si no hay observación. |
| `status` | varchar(20) | `open` \| `closed` (`CHECK`). |
| `opened_at` | timestamptz | Fecha y hora de apertura. |
| `closed_at` | timestamptz | Nulo hasta el cierre (RF-17). |
| `created_at`, `updated_at` | timestamptz | |

| Índice | Definición | Regla |
|---|---|---|
| `idx_cash_session_open_register` | `UNIQUE (cash_register_id) WHERE status = 'open'` | Una caja, una sola sesión abierta. |
| `idx_cash_session_open_user` | `UNIQUE (user_id) WHERE status = 'open'` | Un cajero, una sola sesión abierta. |
| `idx_cash_session_company_status` | `(company_id, status)` | Filtros por empresa. |

---

## Qué cambió en el código

| Cambio | Detalle |
|---|---|
| Módulo nuevo `internal/modules/cashsession` | Dominio, aplicación, persistencia, HTTP y tests, con la misma estructura que `cashregister`. |
| `cashregister`: `GetRegister` | Archivo nuevo `domain/service/lookup.go`. Consulta una caja por id para mostrar su código y nombre. No cambia ninguna regla de HU-10. |
| `shared/middleware/cash_session.go` | `CashSessionChecker`, `RequireOpenCashSession` y `CashSessionIDFromContext`. |
| Permisos | Módulo `cash_sessions` con `open` y `read`; operación `open` añadida al catálogo. El seed se los da a `employee`. |
| `cmd/migrate`, `cmd/server` | Registran el modelo y el módulo. |

## Criterios de aceptación

| # | Criterio | Cómo se cumple | Prueba |
|---|---|---|---|
| 1 | Abrir indicando la caja y el monto inicial | `POST /cash-sessions/` sobre la caja asignada (o la indicada en `cash_register_id`). | `TestOpenCashSessionFollowsContract`, Bruno 016 |
| 2 | Monto no negativo | Validación en el dominio y `CHECK` en la base. | `TestOpenAcceptsZeroAndRejectsInvalidAmounts`, Bruno 012–014, 022 |
| 3 | Sin segunda sesión en una caja con sesión abierta | Comprobación previa e índices únicos parciales. | `TestOnlyOneOpenSessionPerCashierAndPerRegister`, `TestConcurrentOpeningsWithPostgres`, Bruno 017 |
| 4 | Asociada al usuario, fecha/hora y estado abierta | `cashier_id`, `opened_at`, `status: open`. | `TestOpenRecordsCashierRegisterAmountAndTime` |
| 5 | Sin sesión no hay transacciones | `RequireOpenCashSession`. | `TestCashTransactionsRequireAnOpenSession`, `TestRequireOpenCashSession` |
| 6 | Indicar que primero se debe abrir la caja | `409 CASH_SESSION_REQUIRED`. | `TestCashTransactionsRequireAnOpenSession` |

## Pruebas

- Unitarias y HTTP: `go test ./internal/modules/cashsession/... ./internal/shared/middleware/...`
- Integración con Postgres: `SIGIF_TEST_DATABASE_DSN=... go test -p 1 -tags integration ./internal/modules/cashsession/... ./internal/modules/cashregister/...`. El `-p 1` es necesario: si los paquetes corren en paralelo, ambos hacen `AutoMigrate` de las mismas tablas a la vez y Postgres rechaza uno (`pg_type_typname_nsp_index`).
- Bruno: carpeta `bruno/CashSession` (25 peticiones). Usa `cajero.a@sigif.com`, `cajero.b@sigif.com` y `negocio@sigif.com` del seed. Los login leen el token de la cookie `sigif_access_token`, y la petición sin token (023) borra esa cookie antes de enviarse, porque Bruno reenvía automáticamente la del último login.

**La carpeta Bruno solo se puede ejecutar una vez por base de datos**: todavía no existe el cierre de caja (RF-17), así que las sesiones abiertas quedan abiertas. Para volver a ejecutarla en desarrollo:

```sql
UPDATE cash_session SET status = 'closed', closed_at = now(), updated_at = now() WHERE status = 'open';
```

## Pendiente para otras HU

- Cierre de sesión (RF-16/17): debe poner `status = 'closed'` y `closed_at`, y decidir si liberar la caja se bloquea mientras haya sesión abierta. Hoy HU-10 permite liberarla; si se libera, el guard ya impide transaccionar sobre esa caja.
- Historial de sesiones (RF-19), entrega de turno (RF-20), supervisión (RF-21).
