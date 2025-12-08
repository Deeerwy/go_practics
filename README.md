### Практика 11. Бурылин Дмитрий ПИМО-01-25. Проектирование REST API (CRUD для заметок). Разработка структуры


### Задача на практику:

Освоить принципы проектирования REST API. Спроектировать и реализовать CRUD-интерфейс (Create, Read, Update, Delete) для сущности «Заметка». Подготовить основу для интеграции с базой данных и JWT-аутентификацией

### Структура проекта

```
notes-api/
 ├─ cmd/api/main.go
 ├─ internal/
 │   ├─ http/
 │   │   ├─ router.go
 │   │   └─ handlers/notes.go
 │   ├─ core/
 │   │   ├─ note.go
 │   │   └─ service/note_service.go
 │   └─ repo/
 │       └─ note_mem.go
 ├─ api/openapi.yaml
 └─ go.mod
```



### Подготовка проекта

```
mkdir notes-api
cd notes-api
go mod init example.com/notes-api
go get github.com/go-chi/chi/v5
```

### Фрагменты кода из основных файлов 

**internal/core/note.go**
```
package core


import "time"


type Note struct {
  ID        int64
  Title     string
  Content   string
  CreatedAt time.Time
  UpdatedAt *time.Time
}
```

**internal/repo/note_mem.go**
```
package repo
import (
  "sync"
  "example.com/notes-api/internal/core"
)
type NoteRepoMem struct {
  mu    sync.Mutex
  notes map[int64]*core.Note
  next  int64
}


func NewNoteRepoMem() *NoteRepoMem {
  return &NoteRepoMem{notes: make(map[int64]*core.Note)}
}


func (r *NoteRepoMem) Create(n core.Note) (int64, error) {
  r.mu.Lock(); defer r.mu.Unlock()
  r.next++
  n.ID = r.next
  r.notes[n.ID] = &n
  return n.ID, nil
}
```

**internal/http/handlers/notes.go**
```
package handlers


import (
  "encoding/json"
  "net/http"
  "example.com/notes-api/internal/core"
  "example.com/notes-api/internal/repo"
)
type Handler struct {
  Repo *repo.NoteRepoMem
}
func (h *Handler) CreateNote(w http.ResponseWriter, r *http.Request) {
  var n core.Note
  if err := json.NewDecoder(r.Body).Decode(&n); err != nil {
    http.Error(w, "Invalid input", http.StatusBadRequest)
    return
  }
  id, _ := h.Repo.Create(n)
  n.ID = id
  w.Header().Set("Content-Type", "application/json")
  w.WriteHeader(http.StatusCreated)
  json.NewEncoder(w).Encode(n)
}
```
**cmd/api/main.go**
```
package main


import (
  "log"
  "net/http"
  "example.com/notes-api/internal/http"
  "example.com/notes-api/internal/http/handlers"
  "example.com/notes-api/internal/repo"
)


func main() {
  repo := repo.NewNoteRepoMem()
  h := &handlers.Handler{Repo: repo}
  r := httpx.NewRouter(h)


  log.Println("Server started at :8080")
  log.Fatal(http.ListenAndServe(":8080", r))
}
```

### Запуск проекта
```
go run ./cmd/api
```
![screen1](image.png)

### Создание заметки
```
curl -X POST http://localhost:8080/api/v1/notes \
-H "Content-Type: application/json" \
-d '{"title":"Первая заметка", "content":"Это тест"}'
```
![screen2](image-2.png)

### Ответы на контрольные вопросы

**Что означает аббревиатура REST и в чём её суть?**

Аббревиатура: REST расшифровывается как REpresentational State Transfer (Передача репрезентативного состояния).
Суть: Это архитектурный стиль для создания распределённых систем, таких как веб-сервисы. Суть REST заключается в том, что взаимодействие между клиентом и сервером происходит вокруг ресурсов (например, notes, users). Клиент взаимодействует с ресурсом, используя стандартные методы HTTP, и получает его представление (обычно в формате JSON или XML), после чего переходит в новое состояние.
Как связаны CRUD-операции и методы HTTP? CRUD-операции напрямую сопоставляются с HTTP-методами: Create → `POST`, Read → `GET`, Update → `PUT` / `PATCH`, Delete → `DELETE`.

**Для чего нужна слоистая архитектура (handler → service → repository)?**

Она нужна для разделения ответственности, что упрощает тестирование и повышает гибкость и поддерживаемость кода.

**Что означает принцип «stateless» в REST API?**

 Stateless (без состояния) означает, что сервер не хранит информацию о сессии клиента между запросами. Каждый запрос должен содержать всю необходимую информацию для своей полной обработки.

**Почему важно использовать стандартные коды ответов HTTP?**

 Стандартные коды (2xx, 4xx, 5xx) обеспечивают единообразие и предсказуемость. Клиент может однозначно определить результат: успех (2xx), ошибка клиента (4xx) или ошибка сервера (5xx).

**Как можно добавить аутентификацию в REST API?**

 Наиболее популярный способ — использование Bearer Токенов (часто JSON Web Tokens, JWT). Токен передаётся в заголовке `Authorization: Bearer `.

**В чём преимущество версионирования API (например, `/api/v1/`)?**

 Версионирование позволяет развивать API и вносить несовместимые изменения (в новой версии, `/v2/`) без нарушения работы старых клиентов, которые продолжают использовать предыдущую версию (`/v1/`).