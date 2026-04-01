# SecretFlow

**Enterprise Secrets Access Management Platform**

A realistic vulnerable application for cyber security exercises.

**Версия:** 2.0 — Security Hardened + New HARD Attack Path

---

## Обзор

SecretFlow — это внутренняя платформа управления доступом к секретам, которая позволяет сотрудникам:
- Просматривать доступные секреты (только метаданные)
- Запрашивать доступ к секретам с обоснованием
- Одобрять/отклонять запросы на доступ в зависимости от классификации
- Поддерживать доверенную автоматизацию (CI/CD интеграции)
- Аудит всех операций доступа и административных действий

**Назначение:** Учебное приложение с преднамеренными уязвимостями безопасности для упражнений по пентесту и CTF.

---

## Быстрый старт

### Требования
- Docker & Docker Compose
- Git

### Установка

1. **Клонирование и настройка:**
```bash
cd secretflow
cp .env.example .env
```

2. **Запуск всех сервисов:**
```bash
docker-compose up -d
```

3. **Ожидание инициализации базы данных** (~30 секунд)

4. **Доступ к приложению:**
- Frontend: http://localhost:3000
- Backend API: http://localhost:8080
- PostgreSQL: localhost:5432

### Учетные данные по умолчанию

| Username | Password | Role |
|----------|----------|------|
| dev.alice | password123 | developer |
| dev.bob | strong random | developer |
| lead.carol | strong random | team_lead |
| security.dave | strong random | security_admin |
| svc.gitlab | strong random | service_account |

> ⚠️ Только `dev.alice` имеет известный слабый пароль — это преднамеренно скомпрометированная учетная запись.

---

## Архитектура

```
┌─────────────┐     ┌─────────────┐     ┌─────────────┐
│   Frontend  │────▶│   Backend   │────▶│  PostgreSQL │
│  (React)    │     │    (Go)     │     │             │
└─────────────┘     └─────────────┘     └─────────────┘
```

### Технологический стек

**Backend:**
- Go 1.21
- Gin framework
- GORM (PostgreSQL ORM)
- JWT authentication
- bcrypt password hashing
- Server-Sent Events (SSE) для realtime-уведомлений

**Frontend:**
- React 18
- TypeScript
- Vite
- TailwindCSS
- React Router

**Infrastructure:**
- Docker Compose
- PostgreSQL 15

---

## Сценарии атак

### Цель
Получить доступ к CRITICAL-секрету: `PROD_DB_MASTER_PASSWORD`

### Пути атаки

#### Path 1: Integration Token Leak (MEDIUM)
1. Аутентифицироваться как любой пользователь
2. Найти `/api/internal/integrations/status` endpoint
3. Получить webhook_url из ответа
4. Использовать интеграционный токен (из seed-данных или другого источника)
5. Вызвать `/api/integrations/webhook` с токеном
6. Запросить доступ к CRITICAL-секрету
7. Auto-approved due to trusted token

#### Path 2: Confused Deputy + Trust Boundary (HARD) ⭐
Многошаговая атака через делегирование доступа:

1. **Разведка:** Найти `/api/delegate/info` для понимания flow
2. **Получение токена:** Использовать интеграционный токен из Path 1
3. **Обмен токена:** Вызвать `/api/service-account/exchange` для получения service account JWT
4. **Делегирование:** Вызвать `/api/delegate/access` с target_user_id = свой ID
5. **Получение секрета:** Использовать созданный грант для доступа к CRITICAL-секрету

**Тип уязвимости:** Trust Boundary Confusion + Confused Deputy

**Почему работает:**
- Service account JWT доверяет делегированию доступа
- Нет проверки scope токена против запрашиваемого секрета
- Атакующий использует service account как "доверенное лицо" для эскалации привилегий

#### Path 3: Internal API Misuse (DISABLED)
Endpoint'ы `/api/internal/secrets/grant` и `/api/internal/apply` отключены.

### Тупики (Dead Ends)
- Normal request flow требует approval от `security_admin` для CRITICAL (correctly enforced)
- Webhook без валидного токена возвращает 401
- `/api/audit/logs` теперь требует роль `security_admin`
- LOW/MEDIUM секреты доступны, но не содержат целевой флаг

---

## API Документация

### Authentication
```
POST /api/auth/login
Body: {"username": "dev.alice", "password": "password123"}
Response: {"token": "eyJ...", "user": {...}}
```

### Secrets
```
GET /api/secrets
GET /api/secrets/:id
GET /api/secrets/:id/value (requires grant)
POST /api/secrets/:id/request
```

### Access Requests
```
GET /api/requests
POST /api/requests
POST /api/requests/:id/approve
POST /api/requests/:id/deny
```

### Integrations
```
GET /api/integrations (security_admin only)
POST /api/integrations/webhook (token auth)
```

### Delegation (NEW — HARD Path 2)
```
GET  /api/delegate/info
POST /api/service-account/exchange (integration token → service account JWT)
POST /api/delegate/access (requires service_account role)
```

### Events (NEW — Realtime Notifications)
```
GET /api/events/stream (Server-Sent Events)
```

### Audit Logs
```
GET /api/audit/logs (security_admin only)
GET /api/audit/stats (security_admin only)
```

---

## Database Schema

### Tables (8 total)
1. `users` — User accounts and roles
2. `secrets` — Secret metadata and values
3. `access_requests` — Access request tracking
4. `access_grants` — Active access permissions
5. `integrations` — External integration config
6. `integration_tokens` — Authentication tokens
7. `audit_logs` — Audit trail
8. `debug_config` — Debug/development config

См. `backend/migrations/001_initial_schema.sql` для полной схемы.

---

## Уязвимости (Версия 2.0)

| ID | Уязвимость | CWE | Статус | Локация |
|----|------------|-----|--------|---------|
| V1 | Integration status leaks webhook URLs | CWE-215 | ⚠️ ACTIVE | `GET /api/internal/integrations/status` |
| V2 | Tokens stored in plaintext | CWE-256 | ⚠️ ACTIVE | `integration_tokens` table |
| V3 | Token scope not enforced in webhook | CWE-284 | ⚠️ ACTIVE | `webhook_service.go` |
| V4 | Audit logs leak sensitive data | CWE-532 | ✅ FIXED | `audit_service.go` |
| V5 | Classification bypass via automation | CWE-284 | ⚠️ ACTIVE | `approval_service.go` |
| V6 | Trust Boundary в delegation | CWE-284 | ⚠️ ACTIVE (HARD) | `delegate_service.go` |
| V7 | Missing auth on internal endpoint | CWE-306 | ✅ DISABLED | `internal.go` |
| V8 | Audit logs accessible без role check | CWE-284 | ✅ FIXED | `handlers/audit.go` |

### Исправления в версии 2.0

**Session/Auth Hardening:**
- Введен `StrictAuth` middleware для четкой маркировки endpoints
- Единый формат ошибок API
- Немедленный отказ (401/403) при невалидной сессии
- Никаких side-effects (insert/update/audit) до успешной аутентификации

**Path 2 (старый) Closed:**
- `/api/audit/logs` и `/api/audit/stats` теперь требуют `security_admin`
- Токены больше не логируются в audit events (masked preview только)
- `/api/internal/integrations/status` не возвращает `auth_token`
- `/api/internal/integrations/test/:id` требует `security_admin` и возвращает только preview

---

## Разработка

### Backend
```bash
cd backend
go mod tidy
go run cmd/server/main.go
```

### Frontend
```bash
cd frontend
npm install
npm run dev
```

### Database
```bash
docker-compose exec db psql -U secretflow -d secretflow
```

### Build & Test
```bash
# Backend
cd backend
go build ./...

# Frontend
cd frontend
npm run build
```

---

## Новые функции (Версия 2.0)

### Realtime-уведомления
- Server-Sent Events (SSE) через `/api/events/stream`
- Индикатор подключения в UI
- Автоматический reconnect при обрыве

### Улучшенный UX
- Фильтры по статусу и классификации на странице запросов
- Сортировка запросов (newest/oldest/classification)
- Улучшенные состояния loading/error/empty
- Визуальные индикаторы для разных типов уведомлений

### Quality of Life
- Bulk-approval готовность для team_lead
- Экспорт audit logs (готовность к расширению)
- Улучшенная навигация и feedback в UI

---

## License

MIT License — For educational purposes only.

⚠️ **Do not deploy to production.** This application contains intentional security vulnerabilities.

---

## Changelog

### v2.0 — Security Hardened
- ✅ Session/auth hardening — no side-effects before auth
- ✅ Closed old Path 2 (audit log replay)
- ✅ Added new HARD Path 2 (Confused Deputy)
- ✅ Realtime notifications via SSE
- ✅ Request filters and improved UX
- ✅ Standardized error responses
- ✅ Role-based access control enforced

### v1.0 — Initial Release
- Basic secrets management flow
- 3 attack paths (1 config leak, 2 audit replay, 3 internal API)
- Simple React frontend
- Go/Gin backend
