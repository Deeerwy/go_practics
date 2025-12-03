### README ПЗ №10, Бурылин Дмитрий Сергеевич, ПИМО-01-25.

#### Суть практической 
   JWT‑аутентификация HS256, выдача access и refresh токенов, middleware AuthN и AuthZ, RBAC и ABAC правило для `/api/v1/users/{id}`, in‑memory репозиторий с bcrypt‑хэшами, endpoint `/api/v1/refresh` с blacklist для отозванных refresh.


#### Структура проекта

```
pz10-auth/
  go.mod
  cmd/server/main.go                # запуск сервера
  internal/http/router.go           # chi роутер + DI
  internal/http/middleware/
    authn.go                         # извлечение и валидация Bearer access
    authz.go                         # RBAC по ролям
  internal/core/
    user.go                          # доменная модель User
    service.go                       # логика логина, refresh, handlers
  internal/repo/
    user_mem.go                      # in-memory users с bcrypt-хэшами
  internal/platform/jwt/
    jwt.go                           # генерация и парсинг access/refresh
  internal/platform/config/
    config.go                        # чтение JWT_SECRET, JWT_TTL, PORT
README.md
```

**Ключевые файлы**  
- **cmd/server/main.go** — точка входа, загружает конфиг и запускает HTTP сервер.  
- **internal/core/service.go** — `LoginHandler`, `RefreshHandler`, `MeHandler`, `UserHandler`, blacklist для refresh.  
- **internal/http/middleware/authn.go** — кладёт клеймы в `context` под единым ключом.  
- **internal/http/middleware/authz.go** — проверяет роль из клеймов.  
- **internal/repo/user_mem.go** — хранит пользователей и проверяет bcrypt.  
- **internal/platform/jwt/jwt.go** — `Sign` для access, `SignRefresh` с `jti` для refresh, `Parse`.

---

#### Инструкция запуска

**Переменные окружения**  
- `JWT_SECRET` — секрет для HS256 (обязательно).  
- `APP_PORT` — порт сервера (по умолчанию `8080`).  
- (опционально) `JWT_TTL` — общий TTL если используется; в реализации используются явные TTL для access и refresh.


```bash
export JWT_SECRET=dev-secret
export APP_PORT=8080
go run ./cmd/server
```


**Скриншоты**

```
# export JWT_SECRET=dev-secret; export JWT_TTL=2h; export APP_PORT=8080; go run ./cmd/server
```
![sc1](image.png)
```
curl -s -X POST http://localhost:8080/api/v1/login \
 -H "Content-Type: application/json" \
 -d '{"Email":"admin@example.com","Password":"secret"}'
```
![sc2](image-1.png)
```
TOKEN= eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJhdWQiOiJwejEwLWNsaWVudHMiLCJlbWFpbCI6ImFkbWluQGV4YW1wbGUuY29tIiwiZXhwIjoxNzY0Nzc2MDQxLCJpYXQiOjE3NjQ3Njg4NDEsImlzcyI6InB6MTAtYXV0aCIsInJvbGUiOiJhZG1pbiIsInN1YiI6MX0.8tVOk3NRp8SHGUVOYgdewQBdrYkZZTwEfw-iBxcGnog
curl -s http://localhost:8080/api/v1/me -H "Authorization: Bearer $TOKEN"
curl -s http://localhost:8080/api/v1/admin/stats -H "Authorization: Bearer $TOKEN"
```
![sc3](image-2.png)
```
TOKEN_USER=$(curl -s -X POST http://localhost:8080/api/v1/login \
 -H "Content-Type: application/json" -d '{"Email":"user@example.com","Password":"secret"}' | jq -r .token)
curl -i http://localhost:8080/api/v1/admin/stats -H "Authorization: Bearer $TOKEN_USER"
```
![sc4](image-3.png)


#### Контрольные вопросы и ответы

1. **Что такое клеймы JWT и чем отличаются registered public private Почему важно exp**  
   **Клеймы** — это утверждения (claims) в payload JWT: стандартные (registered), публичные и приватные. **Registered** — заранее определённые имена (например, `iss`, `sub`, `aud`, `exp`, `iat`) с рекомендованным смыслом. **Public** — общедоступные имена, которые можно зарегистрировать в IANA или договориться о них между сервисами. **Private** — произвольные пользовательские клеймы для конкретного приложения (например, `role`). Поле **exp** важно потому, что ограничивает срок действия токена; без него токен действовал бы бесконечно, что повышает риск компрометации. Проверка `exp` предотвращает использование старых токенов.

2. **Чем stateless-аутентификация на JWT отличается от сессионных cookie на сервере Плюсы минусы**  
   **JWT (stateless)**: сервер не хранит сессии — валидируется подпись и клеймы. **Плюсы**: простота масштабирования, меньше нагрузки на сервер, удобство для микросервисов. **Минусы**: сложнее отзыв токенов (нужен blacklist/refresh), токен может содержать чувствительные данные, риск хранения на клиенте.  
   **Сессионные cookie (stateful)**: сервер хранит сессии (в памяти/БД/Redis). **Плюсы**: лёгкий отзыв и управление сессиями, меньший риск утечки данных в токене. **Минусы**: требуется хранение состояния и синхронизация между инстансами, дополнительная нагрузка.

3. **Как устроена цепочка middleware и почему AuthZ должна идти после AuthN**  
   Middleware — цепочка функций, каждая получает `request` и `next`. **AuthN** (аутентификация) должна идти первой, потому что она определяет **кто** делает запрос и кладёт клеймы в `context`. **AuthZ** (авторизация) использует эти клеймы (роль, sub и т.д.) для принятия решения о доступе. Если AuthZ выполняется до AuthN, у неё нет данных о пользователе и она не сможет корректно проверить права.

4. **RBAC vs ABAC когда что выбирать Примеры**  
   **RBAC (Role Based Access Control)** — права назначаются ролям (admin, user). Подходит для простых систем с фиксированными ролями. Пример: только `admin` может просматривать статистику.  
   **ABAC (Attribute Based Access Control)** — решение на основе атрибутов (роль, owner id, время, IP). Подходит для гибких политик. Пример: пользователь с ролью `user` может читать `/users/{id}` только если `{id} == sub` (владелец ресурса). Часто комбинируют RBAC и ABAC.

5. **Как безопасно хранить пароль и почему нужен bcrypt argon2 вместо SHA-256 соль pepper**  
   Пароли нужно хранить в виде адаптивных хешей (bcrypt, Argon2). Эти алгоритмы медленнее и устойчивы к брутфорсу, поддерживают настройку стоимости (work factor). **SHA‑256** — быстрый хеш, не предназначен для паролей: атаки перебором выполняются очень быстро. **Соль** — уникальная случайная строка для каждого пароля, предотвращает использование радужных таблиц. **Pepper** — секрет, общий для всех паролей и хранящийся отдельно (например, в конфиге/секретном хранилище), добавляет дополнительный уровень защиты при компрометации БД. Всегда используйте проверенные библиотеки и не храните пароли в открытом виде.
