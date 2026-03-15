Практические занятия №2 — замена проверки доступа с HTTP на gRPC.

Проект на Go 1.22, использует HTTP/REST, gRPC, env‑конфигурацию, таймауты, request‑id и базовое логирование.

---

### Структура репозитория

```
tech-ip-sem2/
  services/
    auth/
      cmd/auth/main.go
      internal/
        http/...
        grpc/...
        service/...
    tasks/
      cmd/tasks/main.go
      internal/
        http/...
        service/...
        client/authclient/...
  shared/
    middleware/
      requestid.go
      logging.go
    httpx/
      client.go
  proto/
    auth.proto
    authpb/
      auth.pb.go
      auth_grpc.pb.go
  docs/
    pz17_api.md
  README.md
```

---

### Границы сервисов

- **Auth service**
  - Выдаёт “учебный” токен через HTTP `POST /v1/auth/login`.
  - Проверяет токен:
    - ПЗ1: HTTP `GET /v1/auth/verify`.
    - ПЗ2: gRPC метод `AuthService.Verify`.
  - Возвращает: валиден/не валиден и `subject`.

- **Tasks service**
  - CRUD задач (in‑memory `map[id]Task`).
  - Перед каждой операцией проверяет токен через Auth:
    - ПЗ1: HTTP‑запрос к `GET /v1/auth/verify`.
    - ПЗ2: gRPC‑вызов `AuthService.Verify`.

---

### Переменные окружения

- `AUTH_PORT` — порт HTTP Auth (по умолчанию `8081`).
- `TASKS_PORT` — порт HTTP Tasks (по умолчанию `8082`).
- `AUTH_BASE_URL` — базовый URL Auth для HTTP‑клиента Tasks (используется в ПЗ1, по умолчанию `http://localhost:8081`).
- `AUTH_GRPC_PORT` — порт gRPC‑сервера Auth (ПЗ2, по умолчанию `50051`).
- `AUTH_GRPC_ADDR` — адрес gRPC для Tasks (ПЗ2, по умолчанию `localhost:50051`).

---


### ПЗ2: gRPC‑проверка Verify

#### .proto и генерация

**Файл контракта:** `proto/auth.proto`

Содержит:

- сервис `AuthService` с методом `Verify(VerifyRequest) returns (VerifyResponse)`;
- `VerifyRequest` с полем `token`;
- `VerifyResponse` с полями `valid`, `subject`;
- опцию `go_package = "tech-ip-sem2/proto/authpb;authpb"`.

**Команда генерации (из корня проекта):**

```bash
cd tech-ip-sem2

protoc \
  --go_out=. --go_opt=paths=source_relative \
  --go-grpc_out=. --go-grpc_opt=paths=source_relative \
  proto/auth.proto
```

Сгенерированные файлы (`auth.pb.go`, `auth_grpc.pb.go`) располагаются в `proto/authpb/` и используются как пакет `tech-ip-sem2/proto/authpb`.

#### Запуск сервисов с gRPC Verify

1. Auth (HTTP + gRPC):

```bash
cd services/auth
export AUTH_PORT=8081
export AUTH_GRPC_PORT=50051
go run ./cmd/auth
```

2. Tasks (HTTP API + gRPC‑клиент):

```bash
cd services/tasks
export TASKS_PORT=8082
export AUTH_GRPC_ADDR=localhost:50051
go run ./cmd/tasks
```

#### Сценарий проверки

1. Получить токен через HTTP `POST /v1/auth/login` (как в ПЗ1).  
2. Сделать HTTP‑запрос к Tasks с заголовком `Authorization: Bearer <token>`.  
3. В логах Tasks увидеть запись `calling grpc verify` и успешное выполнение запроса.  
4. Остановить Auth и повторить запрос к Tasks — получим `503 Service Unavailable`, при этом запрос не “зависает”.

#### Маппинг ошибок

- На стороне Auth (gRPC):
  - невалидный/отсутствующий токен → статус gRPC `Unauthenticated`;
  - внутренние ошибки → `Internal`.

- На стороне Tasks (HTTP API):
  - `Unauthenticated` → HTTP `401 Unauthorized`;
  - `PermissionDenied` (если бы использовался) → HTTP `403 Forbidden`;
  - сетевые ошибки, таймауты, `Internal`, `Unavailable` для Auth → HTTP `503 Service Unavailable`;
  - задача не найдена → `404 Not Found`;
  - некорректное тело запроса → `400 Bad Request`.

Пример логов (успех):

![success_call](<./screens/Screenshot 2026-03-15 at 6.06.51 PM.png>)

Auth недоступен:

![auth_closed](<./screens/Screenshot 2026-03-15 at 6.07.52 PM.png>)



### Контрольные вопросы 

1. **Что такое .proto и почему он считается контрактом?**  
   `.proto` — это декларативное описание сообщений и сервисов (RPC‑методов) в Protocol Buffers. На его основе для разных языков генерируется код клиента и сервера, поэтому именно `.proto` определяет формат данных и интерфейс взаимодействия между сервисами — то есть является формальным “контрактом”.

2. **Что такое deadline в gRPC и чем он полезен?**  
   Deadline — это момент времени, после которого RPC‑вызов должен быть отменён. Клиент передаёт дедлайн на сервер, и оба понимают, сколько максимально можно ждать: это защищает от “висящих” запросов, освобождает ресурсы и делает поведение распределённой системы более предсказуемым.

3. **Почему “exactly-once” не даётся просто так даже в RPC?**  
   Из‑за сбоев сети, потерь ответов и возможных перезапусков ни клиент, ни сервер не могут гарантировать, что операция выполнится ровно один раз: клиент может не получить ответ и повторить запрос, а исходный уже был выполнен. Для приближения к “exactly-once” нужны идемпотентные операции, уникальные идентификаторы запросов и дополнительное хранение состояния на стороне сервера.

4. **Как обеспечивать совместимость при расширении .proto?**  
   Нельзя переиспользовать и удалять уже занятые номера полей. Новые поля добавляются с новыми номерами и как опциональные. Старые клиенты будут игнорировать незнакомые поля, а новые — корректно работать и с обновлёнными, и со старыми сообщениями, обеспечивая обратную и частично прямую совместимость.

