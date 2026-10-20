# SIGIF-GO - Inventory & Sales Management System

Sistema de gestión de inventarios y ventas multi-empresa construido en Go, Fiber y GORM. Arquitectura **Vertical Slice + Hexagonal** (puertos y adaptadores) por módulo.

## Estado actual

Módulos implementados:

- **Auth** (`internal/modules/auth`): inicio de sesión, cierre de sesión, sesiones persistentes (`user_session`) y auditoría de intentos (`login_attempt`).
- **User** (`internal/modules/user`): registro/gestión de usuarios (`app_user`), roles (`role`), permisos (`permission`, `role_permission`) y RBAC.

## Funcionalidades

- **Autenticación por tokens (JWT)** con sesiones persistidas en base de datos (`user_session`): solo se almacena el hash del token.
- **RBAC por permisos de módulo/operación** (`users.create`, `users.list`, ...) vía middleware `RequirePermission`.
- **Roles semilla**: `superadmin` (acceso total) y `business_admin` (solo lectura de usuarios).
- **Hashing de contraseñas** con Argon2id (`password_algorithm` almacenado para migración futura).
- **Respuestas de error genéricas** en login para evitar enumeración de cuentas.
- **Trazabilidad**: cada login intentado se registra en `login_attempt`; cada sesión en `user_session` con IP y dispositivo.

## Tech Stack

- **Lenguaje**: Go 1.26+
- **Framework**: Fiber v2
- **ORM**: GORM con PostgreSQL (identificadores UUID en todas las tablas)
- **Auth**: JWT (HS256) + sesiones en `user_session`
- **DI**: Uber FX
- **Config**: Viper (YAML + variables de entorno con prefijo `SIGIF_` + `.env` local)
- **Logging**: Zap
- **Validación**: Go Playground Validator v10
- **API Client**: Bruno

## Estructura del proyecto

```
.
├── cmd/
│   ├── server/          # Servidor HTTP
│   └── migrate/         # AutoMigrate + migraciones SQL versionadas + seed (roles, permisos, admin)
├── configs/
│   └── config.yaml      # Configuración por defecto
├── deployments/docker/  # Docker Compose + Dockerfiles
├── docs/api/            # Documentación de endpoints
├── internal/
│   ├── modules/         # Vertical slices
│   │   ├── user/        # app_user, role, permission, role_permission
│   │   └── auth/        # login, logout, user_session, login_attempt
│   └── shared/          # Shared kernel (config, database, errors, jwt, middleware, ...)
└── bruno/               # Colecciones Bruno
```

Cada módulo sigue la estructura hexagonal:

```
internal/modules/<modulo>/
├── domain/          # entity, repository (puertos), service
├── application/     # command, query, dto, handler
├── infrastructure/  # persistence (model, mapper, gorm), seed
├── interfaces/http/ # handler, router, dtos
└── module.go        # wiring FX
```

## Configuración

Copia `.env.example` a `.env` y ajusta los valores (todas con prefijo `SIGIF_`):

```bash
cp .env.example .env
```

| Variable | Descripción | Default |
|----------|-------------|---------|
| `SIGIF_APP_PORT` | Puerto HTTP | 4600 |
| `SIGIF_DATABASE_HOST` | Host PostgreSQL | localhost |
| `SIGIF_DATABASE_USER` | Usuario PostgreSQL | sigif |
| `SIGIF_DATABASE_PASSWORD` | Contraseña | sigif |
| `SIGIF_DATABASE_NAME` | Base de datos | sigif |
| `SIGIF_JWT_SECRET` | Secreto JWT | (requerido en producción) |
| `SIGIF_JWT_ACCESS_TOKEN_EXPIRY` | Expiración del token (min) | 15 |
| `SIGIF_AUTH_COOKIE_SECURE` | Cookie de sesión solo por HTTPS (`true` en producción) | false |
| `SIGIF_AUTH_COOKIE_SAME_SITE` | `Lax`, `Strict` o `None` (`None` exige `Secure`) | Lax |
| `SIGIF_AUTH_COOKIE_DOMAIN` | Dominio de la cookie de sesión | (host del backend) |
| `SIGIF_SEED_ADMIN_EMAIL` | Email del admin inicial | admin@sigif.com |
| `SIGIF_SEED_ADMIN_PASSWORD` | Contraseña del admin inicial | admin123 |
| `SIGIF_SEED_BUSINESS_ADMIN_EMAIL` | Email del `business_admin` de prueba | negocio@sigif.com |
| `SIGIF_SEED_BUSINESS_ADMIN_PASSWORD` | Contraseña del `business_admin` de prueba | negocio123 |

> El admin inicial se crea en la primera migración si no existe ningún usuario. Cambia su contraseña en producción. El usuario `business_admin` de prueba (`negocio@sigif.com`) permite verificar los accesos de solo lectura.

## Quick Start

```bash
# Infraestructura (PostgreSQL)
make docker-up

# Migraciones + seed
make migrate-up

# Servidor
make run
```

Servidor disponible en `http://localhost:4600`.

## API Endpoints

Prefijo base: `/api/v1`.

### Autenticación

```
POST /api/v1/auth/login    # Iniciar sesión (email + contraseña)
POST /api/v1/auth/logout   # Cerrar sesión (revoca la sesión actual)
GET  /api/v1/auth/me       # Usuario autenticado
GET  /api/v1/auth/me/permissions   # Códigos de permiso del usuario autenticado
```

### Usuarios (requieren permiso)

```
POST   /api/v1/users                  # Registrar usuario (permiso users.create)
GET    /api/v1/users                  # Listar usuarios (users.list)
GET    /api/v1/users/:id              # Ver usuario (users.read)
GET    /api/v1/users/:id/history      # Historial del usuario (users.read)
GET    /api/v1/users/by-email         # Buscar por email (users.read)
PATCH  /api/v1/users/:id              # Editar usuario (users.update) — mismos campos que el registro, todos opcionales
PUT    /api/v1/users/:id/password     # Cambiar contraseña (users.change_password)
DELETE /api/v1/users/:id              # Eliminar usuario (users.delete)
POST   /api/v1/users/:id/activate     # Activar (users.activate)
POST   /api/v1/users/:id/deactivate   # Desactivar (users.deactivate)
```

### Cajas (RF-10)

```
GET  /api/v1/cash-registers/                             # Cajas activas y su disponibilidad (cash_registers.list) — HU-10-01
POST /api/v1/cash-register-assignments/                  # Tomar una caja disponible (cash_registers.assign) — HU-10-01
GET  /api/v1/cash-register-assignments/current           # Caja asignada al cajero (cash_registers.assign) — HU-10-01
POST /api/v1/cash-register-assignments/:id/release       # Liberar la caja (cash_registers.release o release_any) — HU-10-02
```

El cajero y la empresa salen del token. Una caja tiene un solo cajero y un cajero una sola caja a la vez; la regla la garantizan índices únicos parciales, también ante peticiones simultáneas (`409 CASH_REGISTER_IN_USE` / `409 CASHIER_ALREADY_ASSIGNED`). Liberar conserva el historial. El rol `employee` trabaja como cajero; `business_admin` ve las cajas en uso y puede liberarlas. Detalle en `docs/api/HU-10-01-endpoints.md` y `docs/api/HU-10-02-endpoints.md`.

### Sesión de caja (RF-09)

```
POST /api/v1/cash-sessions/          # Abrir la sesión de caja con el monto inicial (cash_sessions.open) — HU-009
GET  /api/v1/cash-sessions/current   # Sesión abierta del cajero (cash_sessions.read) — HU-009
```

La sesión se abre sobre la caja asignada al cajero (HU-10), con un monto inicial de 0 o más. Solo hay una sesión abierta por cajero y por caja (`409 CASH_SESSION_ALREADY_OPEN`). Las operaciones de caja de otros módulos se protegen con `middleware.RequireOpenCashSession`, que responde `409 CASH_SESSION_REQUIRED` sin sesión abierta. Detalle en `docs/api/HU-009-endpoints.md`.

### Permisos (catálogo, HU-082-01)

```
GET  /api/v1/permissions/modules   # Catálogo de módulos y operaciones (permiso permissions.read)
GET  /api/v1/permissions           # Listar permisos; opcional ?module=<módulo> (permissions.read)
POST /api/v1/permissions           # Registrar un permiso (permissions.create)
```

```json
POST /api/v1/permissions
{
  "module": "products",
  "operation": "export",
  "description": "Exportar productos"
}
```

`module` debe pertenecer al catálogo y `operation` a las operaciones soportadas del módulo. El código (`code`) se deriva del backend como `module.operation`; si se envía debe coincidir. Los pares repetidos responden `409 CONFLICT` (índice único `(module, operation)`).

### Roles (listado, HU-082-02)

```
GET /api/v1/roles/   # Listar roles (permiso roles.list) — solo lectura
```

Params opcionales: `q` (búsqueda parcial e insensible a mayúsculas sobre el nombre, máx. 100), `type` (`system` | `custom`), `status` (`active` | `inactive` | `all`, por defecto `all`), `page` (≥1) y `limit` (1–100, por defecto 20).

```json
GET /api/v1/roles/?q=vendedor&type=custom&status=active&page=1&limit=20
```

Respuesta:

```json
{
  "success": true,
  "data": [
    {
      "id": "...",
      "company_id": null,
      "name": "superadmin",
      "type": "system",
      "status": "active",
      "permissions_count": 22,
      "description": "Acceso total al sistema",
      "created_at": "2026-10-05T12:00:00Z"
    }
  ],
  "meta": { "page": 1, "limit": 20, "total": 4, "total_pages": 1 }
}
```

Los roles visibles son los del sistema (`company_id` nulo) más los de la empresa del token; un usuario sin empresa solo ve los del sistema. `permissions_count` cuenta los permisos asignados a cada rol. Una lista vacía responde `200` con `data: []` y `total: 0` (nunca `404`).

### Clientes (RF-08)

```
POST   /api/v1/customers/            # Registrar cliente (permiso customers.create) — HU-08-01
GET    /api/v1/customers/            # Buscar y listar clientes (customers.list) — HU-08-02
GET    /api/v1/customers/:id         # Ver cliente (customers.read) — HU-08-03
PUT    /api/v1/customers/:id         # Reemplazar los datos editables (customers.update) — HU-08-01
PATCH  /api/v1/customers/:id         # Editar parcialmente (customers.update) — HU-08-04
PATCH  /api/v1/customers/:id/status  # Desactivar, bloquear o reactivar (customers.status) — HU-08-05
DELETE /api/v1/customers/:id         # Eliminar lógicamente (customers.delete) — HU-08-06
```

Todas las rutas trabajan sobre la empresa del token; un cliente de otra empresa responde `404`. El listado devuelve solo clientes `active` por defecto (`status=inactive|blocked|all` para el resto). Desactivar conserva el registro; eliminar es un borrado lógico y responde `409` si el cliente tiene saldo de crédito, puntos o registros relacionados. El detalle de cada endpoint está en `docs/api/HU-08-0X-endpoints.md`.

### Registro de usuario

```json
POST /api/v1/users
{
  "first_name": "Juan",
  "last_name": "Pérez",
  "email": "business_admin@sigif.com",
  "password": "password123",
  "role": "business_admin",
  "area": "Administración del negocio"
}
```

`role` acepta `business_admin` o `employee`. `superadmin` se crea únicamente
desde el seed/backend y no puede asignarse desde los endpoints de usuarios.
`area` es opcional.

### Login

```json
POST /api/v1/auth/login
{
  "email": "admin@sigif.com",
  "password": "admin123"
}
```

Respuesta:

```json
{
  "success": true,
  "data": {
    "user": { "id": "...", "role": "superadmin", "email": "admin@sigif.com", "..." : "..." },
    "expires_in": 900
  }
}
```

El access token se entrega en la cookie HttpOnly `sigif_access_token` (`Set-Cookie`), no en el JSON. El navegador la envía sola con `withCredentials: true`; los clientes no navegador pueden seguir usando `Authorization: Bearer <token>`. Detalle en [docs/api/HU-01-endpoints.md](docs/api/HU-01-endpoints.md).

## Testing

```bash
make test        # go test -v -race ./...
make lint        # golangci-lint
make check       # fmt + vet + lint + test
```

## Arquitectura y reglas de dependencia

- **Domain**: sin dependencias externas (solo `shared`).
- **Application**: depende de Domain + Shared.
- **Infrastructure**: implementa los puertos de Domain.
- **Interfaces**: depende de Application + Shared.
- **Shared**: no depende de módulos.

### Comunicación entre módulos

El módulo `auth` consume el módulo `user` mediante un **adaptador** (`auth/application/adapter`) que implementa el puerto `auth/domain/repository.UserRepo`. No hay acoplamiento directo a la infraestructura del otro módulo.

## Bruno Collections

Importa la carpeta `bruno/` en Bruno. La colección usa la variable `baseUrl` (`http://localhost:4600`) y los scripts de login persisten `accessToken` y `userId` automáticamente.

## Health Check

```bash
curl http://localhost:4600/health
```

## License

Propietario - Uso interno SIGIF.
