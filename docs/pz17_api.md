# PZ17 Auth & Tasks API (Go)

## Сервисы и порты

- Auth service: `AUTH_PORT` (по умолчанию `8081`)
- Tasks service: `TASKS_PORT` (по умолчанию `8082`)
- Tasks обращается к Auth по `AUTH_BASE_URL` (по умолчанию `http://localhost:8081`)

## Общие заголовки

- `Authorization: Bearer <token>` — обязателен для всех эндпоинтов Tasks
- `X-Request-ID: <uuid>` — опционально (если не задан, сервисы сгенерируют сами и будут прокидывать дальше)

## Auth service

### POST /v1/auth/login

Request:

```json
{
  "username": "student",
  "password": "student"
}
```

Response 200:

```json
{
  "access_token": "demo-token",
  "token_type": "Bearer"
}
```

Ошибки:

- 400 — неверный формат JSON
- 401 — неверные учётные данные

### GET /v1/auth/verify

Headers:

- `Authorization: Bearer demo-token`
- `X-Request-ID: <uuid>` (опционально)

Успех 200:

```json
{
  "valid": true,
  "subject": "student"
}
```

Ошибка 401:

```json
{
  "valid": false,
  "error": "unauthorized"
}
```

## Tasks service

### POST /v1/tasks

Headers:

- `Authorization: Bearer <token>`
- `X-Request-ID: <uuid>` (опционально)

Request:

```json
{
  "title": "Read lecture",
  "description": "Prepare notes",
  "due_date": "2026-01-10"
}
```

Response 201:

```json
{
  "id": "t_1",
  "title": "Read lecture",
  "description": "Prepare notes",
  "due_date": "2026-01-10",
  "done": false
}
```

Ошибки: 400, 401, 500.

### GET /v1/tasks

Response 200:

```json
[
  {"id":"t_1","title":"Read lecture","done":false},
  {"id":"t_2","title":"Do practice","done":true}
]
```

Ошибки: 401, 500.

### GET /v1/tasks/{id}

Response 200:

```json
{
  "id":"t_1",
  "title":"Read lecture",
  "description":"Prepare notes",
  "due_date": "2026-01-10",
  "done": false
}
```

Ошибки: 401, 404, 500.

### PATCH /v1/tasks/{id}

Request:

```json
{
  "title": "Read lecture (updated)",
  "done": true
}
```

Response 200 — обновлённый объект `Task`.

Ошибки: 400, 401, 404, 500.

### DELETE /v1/tasks/{id}

Response 204 — без тела.

Ошибки: 401, 404, 500.

## Запуск сервисов

В одном терминале:

```bash
cd services/auth
export AUTH_PORT=8081
go run ./cmd/auth
```

В другом терминале:

```bash
cd services/tasks
export TASKS_PORT=8082
export AUTH_BASE_URL=http://localhost:8081
go run ./cmd/tasks
```

## Примеры curl

Получить токен:

```bash
curl -s -X POST http://localhost:8081/v1/auth/login \
  -H "Content-Type: application/json" \
  -H "X-Request-ID: req-001" \
  -d '{"username":"student","password":"student"}'
```

Создать задачу:

```bash
curl -i -X POST http://localhost:8082/v1/tasks \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer demo-token" \
  -H "X-Request-ID: req-003" \
  -d '{"title":"Do PZ17","description":"split services","due_date":"2026-01-10"}'
```

Список задач:

```bash
curl -i http://localhost:8082/v1/tasks \
  -H "Authorization: Bearer demo-token" \
  -H "X-Request-ID: req-004"
```

