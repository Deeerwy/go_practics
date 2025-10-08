# Практическое занятие №4  Бурылин Дмитрий Сергеевич ПИМО-01-25
**Тема:** Маршрутизация с chi. Создание CRUD-сервиса «Список задач»  

## 🎯 Цели работы
- Освоить базовую маршрутизацию HTTP-запросов в Go с использованием роутера **chi**.  
- Научиться строить REST-маршруты и обрабатывать методы **GET/POST/PUT/DELETE**.  
- Реализовать CRUD-сервис «ToDo» с хранением данных в памяти.  
- Добавить простое middleware (логирование, CORS).  
- Научиться тестировать API через **curl/Postman/HTTPie**.  


## 📂 Структура проекта
```
pz4-todo/ 
├── go.mod 
├── go.sum 
├── main.go 
├── internal/ 
    └── task/ 
        ├── model.go 
        ├── repo.go 
        └── handler.go 
└── pkg/ 
    └── middleware/ 
        ├── logger.go 
        └── cors.go
```

##  Ход работы

### Роутер
```
r := chi.NewRouter()
r.Use(chimw.RequestID, chimw.Recoverer, myMW.Logger, myMW.SimpleCORS)

r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
    w.Write([]byte("OK"))
})

r.Route("/api", func(api chi.Router) {
    api.Mount("/tasks", h.Routes())
})
```
### Middleware
### Логирование
```
func Logger(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        start := time.Now()
        next.ServeHTTP(w, r)
        log.Printf("%s %s %s", r.Method, r.URL.Path, time.Since(start))
    })
}
```
### CORS
```
func Logger(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        start := time.Now()
        next.ServeHTTP(w, r)
        log.Printf("%s %s %s", r.Method, r.URL.Path, time.Since(start))
    })
}

```
### Обработчик
```
func (h *Handler) get(w http.ResponseWriter, r *http.Request) {
	id, bad := parseID(w, r)
	if bad {
		return
	}
	t, err := h.repo.Get(id)
	if err != nil {
		httpError(w, http.StatusNotFound, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, t)
}
```

### Примеры запросов и ответов
### Cоздание элемента
```
curl -X POST http://localhost:8080/api/tasks \
-H "Content-Type: application/json" \
-d '{"title":"Выучить chi"}'
```
![Response](image.png)

### Список
```
curl http://localhost:8080/api/tasks
```
![Response](image-1.png)

### Получение по ID
```
curl http://localhost:8080/api/tasks/1
```
![Resp](image-2.png)

### Обновление элемента
```
curl -X PUT http://localhost:8080/api/tasks/1 \
    -H "Content-Type: application/json" \
    -d '{"title":"Выучить chi глубже","done":true}'
```
![CRUD](image-3.png)

### Удаление элемента
```
curl -X DELETE http://localhost:8080/api/tasks/1
```
![DELETE](image-4.png)

### Таблица с результатами

| Маршрут              | Запрос (пример)                                                                 | Ожидаемый ответ         | Фактический ответ       |
|----------------------|----------------------------------------------------------------------------------|-------------------------|-------------------------|
| `POST /api/tasks`    | `{"title":"Test"}`                                                               | 201 Created + JSON      | 201 Created + JSON      |
| `GET /api/tasks`     | —                                                                                | 200 OK + список задач   | 200 OK + список задач   |
| `GET /api/tasks/1`   | —                                                                                | 200 OK + JSON задачи    | 200 OK + JSON задачи    |
| `PUT /api/tasks/1`   | `{"title":"Upd","done":true}`                                                    | 200 OK + обновлённая задача | 200 OK + обновлённая задача |
| `DELETE /api/tasks/1`| —                                                                                | 204 No Content          | 204 No Content          |
| `GET /api/tasks/99`  | —                                                                                | 404 Not Found + JSON    | 404 Not Found + JSON    |

##  Выводы

-  В ходе работы был реализован полноценный CRUD-сервис «ToDo» на языке Go с использованием роутера **chi**.  
-  Все основные маршруты (GET, POST, PUT, DELETE) протестированы и возвращают корректные коды статусов и ответы в формате JSON.  
-  Подключены middleware для логирования и CORS, что позволило сделать сервис более удобным и приближенным к промышленным стандартам.  

**Что получилось:**  
- Удалось построить модульную структуру проекта с разделением на `internal` и `pkg`.  
- Реализована корректная обработка ошибок и возврат статусов HTTP.  
- Сервис успешно протестирован через `curl` и может быть расширен.  

**Что было сложным:**  
- Правильная обработка ошибок при неверных входных данных (например, пустой `title` или некорректный `id`).  
- Экранирование кавычек в PowerShell при тестировании запросов.  
- Настройка CORS и понимание, как правильно обрабатывать `OPTIONS`-запросы.  