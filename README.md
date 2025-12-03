# Практическое занятие №9: регистрация и вход пользователей, bcrypt Бурылин Дмитрий Сергеевич ПИМО-01-25

---

## Обзор проекта

Мини-сервис pz9-auth реализует эндпоинты:
- POST /auth/register — регистрация пользователя с хэшированием пароля через bcrypt и записью в БД.
- POST /auth/login — проверка пары email+password через bcrypt и возврат статуса.

Используется PostgreSQL и GORM (или database/sql), роутер chi. Проект готовит основу для JWT-аутентификации в ПЗ №10.

---

## Структура проекта

```
pz9-auth/
  cmd/api/main.go
  internal/http/handlers/auth.go
  internal/core/user.go
  internal/repo/postgres.go
  internal/repo/user_repo.go
  internal/platform/config/config.go
  go.mod
```

- **cmd/api/main.go:** Точка входа, маршруты, запуск HTTP-сервера.
- **internal/http/handlers/auth.go:** Хэндлеры Register и Login, валидация и bcrypt.
- **internal/core/user.go:** Модель пользователя для БД (email уникален).
- **internal/repo/postgres.go:** Инициализация GORM с DSN.
- **internal/repo/user_repo.go:** Репозиторий пользователей, AutoMigrate, Create, ByEmail.
- **internal/platform/config/config.go:** Конфигурация окружения (DB_DSN, BCRYPT_COST, APP_ADDR).

---

## Настройка окружения и подключение к БД

- **Переменные окружения:**
  - **DB_DSN:** Строка подключения к Postgres (dev-примеры ниже).
  - **BCRYPT_COST:** Стоимость bcrypt (по умолчанию 12).
  - **APP_ADDR:** Адрес сервера (по умолчанию :8080).

```bash
export DB_DSN="postgres://user:pass@localhost:5432/pz9?sslmode=disable"
export BCRYPT_COST=12
export APP_ADDR=":8080"


- **Инициализация зависимостей:**
```bash
mkdir pz9-auth && cd pz9-auth
go mod init example.com/pz9-auth
go get github.com/go-chi/chi/v5
go get gorm.io/gorm gorm.io/driver/postgres
go get golang.org/x/crypto/bcrypt
```
![sc2](image-1.png)

## Команда запуска

```bash
go run cmd/api/main.go
# Ожидаемый лог: listening on :8080
```

---

## Миграции и подтверждение AutoMigrate


- **AutoMigrate (internal/repo/user_repo.go):**
```go
func (r *UserRepo) AutoMigrate() error {
    return r.db.AutoMigrate(&core.User{})
}
```

- **Подтверждение схемы в БД:**
```bash
psql "$DB_DSN"
# Внутри psql:
\d users
```
![sc1](image.png)

---

## Инициализация GORM и запуск сервера

- **internal/repo/postgres.go:**
```go
package repo

import (
    "gorm.io/driver/postgres"
    "gorm.io/gorm"
)

func Open(dsn string) (*gorm.DB, error) {
    return gorm.Open(postgres.Open(dsn), &gorm.Config{
        TranslateError: true,
    })
}
```

- **cmd/api/main.go:**
```go
package main

import (
    "log"
    "net/http"

    "github.com/go-chi/chi/v5"
    "example.com/pz9-auth/internal/platform/config"
    "example.com/pz9-auth/internal/http/handlers"
    "example.com/pz9-auth/internal/repo"
)

func main() {
    cfg := config.Load()

    db, err := repo.Open(cfg.DB_DSN)
    if err != nil { log.Fatal("db connect:", err) }

    users := repo.NewUserRepo(db)
    if err := users.AutoMigrate(); err != nil { log.Fatal("migrate:", err) }

    auth := &handlers.AuthHandler{Users: users, BcryptCost: cfg.BcryptCost}

    r := chi.NewRouter()
    r.Post("/auth/register", auth.Register)
    r.Post("/auth/login", auth.Login)

    log.Println("listening on", cfg.Addr)
    log.Fatal(http.ListenAndServe(cfg.Addr, r))
}
```

---

## Хэндлеры и использование bcrypt

- **internal/http/handlers/auth.go (фрагменты):**
```go
type AuthHandler struct {
    Users      *repo.UserRepo
    BcryptCost int
}

func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
    var in struct {
        Email    string `json:"email"`
        Password string `json:"password"`
    }
    if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
        writeErr(w, http.StatusBadRequest, "invalid_json"); return
    }
    in.Email = strings.TrimSpace(strings.ToLower(in.Email))
    if in.Email == "" || len(in.Password) < 8 {
        writeErr(w, http.StatusBadRequest, "email_required_and_password_min_8"); return
    }

    hash, err := bcrypt.GenerateFromPassword([]byte(in.Password), h.BcryptCost)
    if err != nil { writeErr(w, http.StatusInternalServerError, "hash_failed"); return }

    u := core.User{Email: in.Email, PasswordHash: string(hash)}
    if err := h.Users.Create(r.Context(), &u); err != nil {
        if err == repo.ErrEmailTaken {
            writeErr(w, http.StatusConflict, "email_taken"); return
        }
        writeErr(w, http.StatusInternalServerError, "db_error"); return
    }

    writeJSON(w, http.StatusCreated, map[string]any{
        "status": "ok",
        "user":   map[string]any{"id": u.ID, "email": u.Email},
    })
}

func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
    var in struct {
        Email    string `json:"email"`
        Password string `json:"password"`
    }
    if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
        writeErr(w, http.StatusBadRequest, "invalid_json"); return
    }
    in.Email = strings.TrimSpace(strings.ToLower(in.Email))
    if in.Email == "" || in.Password == "" {
        writeErr(w, http.StatusBadRequest, "email_and_password_required"); return
    }

    u, err := h.Users.ByEmail(r.Context(), in.Email)
    if err != nil {
        writeErr(w, http.StatusUnauthorized, "invalid_credentials"); return
    }
    if bcrypt.CompareHashAndPassword([]byte(u.PasswordHash), []byte(in.Password)) != nil {
        writeErr(w, http.StatusUnauthorized, "invalid_credentials"); return
    }

    writeJSON(w, http.StatusOK, map[string]any{
        "status": "ok",
        "user":   map[string]any{"id": u.ID, "email": u.Email},
    })
}
```

- **internal/repo/user_repo.go (обработка уникальности):**
```go
package repo

import (
    "context"
    "errors"

    "gorm.io/gorm"
    "example.com/pz9-auth/internal/core"
)

var ErrUserNotFound = errors.New("user not found")
var ErrEmailTaken   = errors.New("email already in use")

type UserRepo struct{ db *gorm.DB }

func NewUserRepo(db *gorm.DB) *UserRepo { return &UserRepo{db: db} }

func (r *UserRepo) AutoMigrate() error {
    return r.db.AutoMigrate(&core.User{})
}

func (r *UserRepo) Create(ctx context.Context, u *core.User) error {
    if err := r.db.WithContext(ctx).Create(u).Error; err != nil {
        // Требует GORM >= 1.25.2 и TranslateError: true
        if errors.Is(err, gorm.ErrDuplicatedKey) {
            return ErrEmailTaken
        }
        return err
    }
    return nil
}

func (r *UserRepo) ByEmail(ctx context.Context, email string) (core.User, error) {
    var u core.User
    err := r.db.WithContext(ctx).Where("email = ?", email).First(&u).Error
    if errors.Is(err, gorm.ErrRecordNotFound) {
        return core.User{}, ErrUserNotFound
    }
    return u, err
}
```

---

## Тестирование через curl/Postman

```
# Регистрация 
curl -i -X POST http://localhost:8080/auth/register \
  -H "Content-Type: application/json" \
  -d '{"email":"user@example.com","password":"Secret123!"}'
  ```
![sc3](image-2.png)
```
# Повторная регистрация 
curl -i -X POST http://localhost:8080/auth/register \
  -H "Content-Type: application/json" \
  -d '{"email":"user@example.com","password":"AnotherPass"}'
  ```
  ![sc4](image-3.png)
```
# Вход с верными данными 
curl -i -X POST http://localhost:8080/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email":"user@example.com","password":"Secret123!"}'
  ```
  ![sc4](image-4.png)
```
# Вход с неверным паролем 
curl -i -X POST http://localhost:8080/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email":"user@example.com","password":"wrong"}'
  ```
  ![sc5](image-5.png)


---

## Минимальные замечания по безопасности

- **Общие ошибки:** Не раскрывать детали при логине — только "invalid_credentials".
- **Лимиты:** Ввести rate limiting для попыток входа.
- **Логи:** Никогда не логировать пароли или их хэши.
- **HTTPS:** Все продовые запросы должны идти по HTTPS.
- **bcrypt cost:** Подбирать разумный cost (для учебных машин 12–14).

---

## Ответы на контрольные вопросы

- **В чём разница между хранением пароля и хранением его хэша?**
  - **Пароль:** Исходная секретная строка пользователя; хранение в открытом виде опасно — утечка даёт мгновенный доступ.
  - **Хэш:** Одностороннее преобразование пароля; восстановить пароль из хэша практически невозможно, проверка идёт сравнением хэшей.
  - **Соль:** Случайная добавка, защищает от предвычисленных атак и одинаковых хэшей для одинаковых паролей.

- **Зачем соль?**
  - **Индивидуализация:** Даже одинаковые пароли дают разные хэши.
  - **Защита:** Против "rainbow tables" и массового перебора заранее посчитанных хэшей.
  - **Практика:** В bcrypt соль встроена в формат хэша и хранится вместе с ним.

- **Почему bcrypt, а не SHA-256?**
  - **Замедление перебора:** bcrypt специально "медленный", регулируемый cost усложняет brute force.
  - **Соль встроена:** Автоматически включается и хранится внутри хэша.
  - **Паролезависимый алгоритм:** Адаптирован для хранения паролей, в отличие от общих хэш-функций (SHA-256).

- **Что произойдёт при снижении/повышении cost у bcrypt? Как подобрать значение?**
  - **Снижение cost:** Быстрее хэширование, но слабее защита от перебора.
  - **Повышение cost:** Сильнее защита, но медленнее регистрация/логин.
  - **Подбор:** Измерить среднее время операции на целевом окружении и выбрать порог, который приемлем по UX и безопасности (в учебных целях — 12; в проде — 12–14+ с учётом инфраструктуры).

- **Какие статусы и ответы должны возвращать POST /auth/register и POST /auth/login?**
  - **/auth/register:**
    - **201 Created:** Успешная регистрация, вернуть id и email.
    - **400 Bad Request:** Некорректный ввод (пустой email, пароль < 8 символов, неверный JSON).
    - **409 Conflict:** Email уже используется (нарушение уникального индекса).
    - **500 Internal Server Error:** Непредвиденная ошибка БД/хэширования.
  - **/auth/login:**
    - **200 OK:** Верные учётные данные.
    - **400 Bad Request:** Пустой email/пароль, неверный JSON.
    - **401 Unauthorized:** Неверные учётные данные (не раскрывать деталей).
    - **500 Internal Server Error:** Непредвиденная ошибка.

- **Какие риски несут подробные сообщения об ошибках при логине?**
  - **Информационная утечка:** Подтверждение существования email облегчает сбор валидных адресов.
  - **Таргетинг атак:** Злоумышленник отдельно атакует пароли для существующих email.
  - **Практика:** Всегда отвечать обобщённо — "invalid_credentials".

- **Почему в этом ПЗ не выдаём токен, и что изменится в ПЗ10 (JWT)?**
  - **Сейчас:** Фокус на безопасном хранении паролей и корректной проверке входа.
  - **Далее:** В ПЗ10 добавится выдача JWT при успешном логине, настройка сроков жизни, refresh-токены, мидлвары проверки авторизации.
