# SIGIF-GO - Inventory & Sales Management System

Sistema de gestión de inventarios y ventas multi-empresa construido en Go, Fiber y GORM. Arquitectura **Vertical Slice + Hexagonal** (puertos y adaptadores) por módulo.

## Estado actual

Módulos implementados:

- **Auth** (`internal/modules/auth`): inicio de sesión, cierre de sesión, sesiones persistentes (`user_session`) y auditoría de intentos (`login_attempt`).
- **User** (`internal/modules/user`): registro/gestión de usuarios (`app_user`), roles (`role`), permisos (`permission`, `role_permission`) y RBAC.

## Funcionalidades

- **Autenticación por tokens (JWT)** con sesiones persistidas en base de datos (`user_session`): solo se almacena el hash del token.
- **RBAC por permisos de módulo/operación** (`users.create`, `users.list`, ...) vía middleware `RequirePermission`.
- **Roles semilla**: `superadmin` (acceso total) y `soporte` (solo lectura de usuarios).
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
│   └── migrate/         # AutoMigrate + seed (roles, permisos, admin)
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
| `SIGIF_APP_PORT` | Puerto HTTP | 8080 |
| `SIGIF_DATABASE_HOST` | Host PostgreSQL | localhost |
| `SIGIF_DATABASE_USER` | Usuario PostgreSQL | sigif |
| `SIGIF_DATABASE_PASSWORD` | Contraseña | sigif |
| `SIGIF_DATABASE_NAME` | Base de datos | sigif |
| `SIGIF_JWT_SECRET` | Secreto JWT | (requerido en producción) |
| `SIGIF_JWT_ACCESS_TOKEN_EXPIRY` | Expiración del token (min) | 15 |
| `SIGIF_SEED_ADMIN_EMAIL` | Email del admin inicial | admin@sigif.com |
| `SIGIF_SEED_ADMIN_PASSWORD` | Contraseña del admin inicial | admin123 |

> El admin inicial se crea en la primera migración si no existe ningún usuario. Cambia su contraseña en producción.

## Quick Start

```bash
# Infraestructura (PostgreSQL)
make docker-up

# Migraciones + seed
make migrate-up

# Servidor
make run
```

Servidor disponible en `http://localhost:8080`.

## API Endpoints

Prefijo base: `/api/v1`.

### Autenticación

```
POST /api/v1/auth/login    # Iniciar sesión (email + contraseña)
POST /api/v1/auth/logout   # Cerrar sesión (revoca la sesión actual)
GET  /api/v1/auth/me       # Usuario autenticado
```

### Usuarios (requieren permiso)

```
POST   /api/v1/users                  # Registrar usuario (permiso users.create)
GET    /api/v1/users                  # Listar usuarios (users.list)
GET    /api/v1/users/:id              # Ver usuario (users.read)
GET    /api/v1/users/by-email         # Buscar por email (users.read)
PUT    /api/v1/users/:id              # Actualizar usuario (users.update)
PUT    /api/v1/users/:id/password     # Cambiar contraseña (users.change_password)
DELETE /api/v1/users/:id              # Eliminar usuario (users.delete)
POST   /api/v1/users/:id/activate     # Activar (users.activate)
POST   /api/v1/users/:id/deactivate   # Desactivar (users.deactivate)
```

### Registro de usuario

```json
POST /api/v1/users
{
  "first_name": "Juan",
  "last_name": "Pérez",
  "email": "soporte@sigif.com",
  "password": "password123",
  "role": "soporte",
  "area": "Soporte técnico"
}
```

`role` acepta `superadmin` o `soporte`. `area` es opcional.

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
    "token": { "access_token": "...", "expires_in": 900, "token_type": "Bearer" }
  }
}
```

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

Importa la carpeta `bruno/` en Bruno. La colección usa la variable `baseUrl` (`http://localhost:8080`) y los scripts de login persisten `accessToken` y `userId` automáticamente.

## Health Check

```bash
curl http://localhost:8080/health
```

## License

Propietario - Uso interno SIGIF.
