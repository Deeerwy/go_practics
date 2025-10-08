# Практическое задание 3 Бурылин Дмитрий Сергеевич ПИМО-01-25  
## Реализация простого HTTP‑сервера на Go (net/http)

### 🎯 Цели
- Освоить работу со стандартной библиотекой `net/http` без сторонних фреймворков.  
- Научиться поднимать HTTP‑сервер и настраивать маршрутизацию через `http.ServeMux`.  
- Реализовать обработку GET/POST запросов, чтение query‑параметров и тела запроса (JSON).  
- Формировать корректные ответы (статус, заголовки, JSON).  
- Добавить middleware для логирования.  
- Реализовать CRUD‑операции для сущности **Task** в памяти.  

---

##  Запуск проекта


   ```
   git clone https://github.com/Deeerwy/pz3-http.git
   cd pz3-http
   go run ./cmd/server
   ```

## Структура проекта

pz3-http/
├─ cmd/server/main.go        
├─ internal/api/handlers.go  
├─ internal/api/middleware.go
├─ internal/api/responses.go 
├─ internal/storage/memory.go
└─ go.mod

## Тестовые запросы для ПЗ №3

Ниже приведены примеры команд `curl` для проверки работы сервера.  
Все запросы выполнены после запуска сервера на `http://localhost:8080`.

### 1. Проверка здоровья сервера
```
curl -i http://localhost:8080/health
```
![Response](image.png)

### 2. Создание задачи (POST /tasks)
```
curl -i -X POST http://localhost:8080/tasks \
  -H "Content-Type: application/json" \
  -d '{"title":"Купить молоко"}'
```
![Reponse](image-1.png)

### 3. Получение списка задач (GET /tasks) + фильтрация
```
curl -i http://localhost:8080/tasks
curl -i "http://localhost:8080/tasks?q=молоко"
```
![Response](image-2.png)

### 4. Получение задачи по ID (GET /tasks/{id})
```
curl -i http://localhost:8080/tasks/1
```
![Response](image-3.png)

### 5. Ошибки при вводе некорректных запросов
```
curl -i -X POST http://localhost:8080/tasks \
  -H "Content-Type: application/json" \
  -d '{}'
```
![Response](image-4.png)

```
curl -i http://localhost:8080/tasks/abc
```
![alt text](image-5.png)