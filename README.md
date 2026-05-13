# Практическое занятие №10 — Горизонтальное масштабирование: Load Balancer (NGINX). Бурылин Дмитрий ПИМО-01-25

**Дисциплина:** Технологии создания программного обеспечения  


## Цель занятия

Освоить базовый подход к горизонтальному масштабированию backend-приложения за счёт запуска нескольких экземпляров одного сервиса и распределения входящих HTTP-запросов через NGINX в роли балансировщика нагрузки.


## Структура проекта

```
pz10-load-balancer/
  services/
    tasks/
      cmd/
        server/
          main.go
      go.mod
      go.sum
      Dockerfile
  deploy/
    lb/
      docker-compose.yml
      nginx.conf
```


## Шаг 1. Подготовка сервиса tasks

Файл `services/tasks/cmd/server/main.go`:

```go
package main

import (
    "encoding/json"
    "log"
    "net/http"
    "os"
)

type Task struct {
    ID    int64  `json:"id"`
    Title string `json:"title"`
}

func main() {
    instanceID := os.Getenv("INSTANCE_ID")
    if instanceID == "" {
        instanceID = "tasks-unknown"
    }

    port := os.Getenv("APP_PORT")
    if port == "" {
        port = "8082"
    }

    tasks := []Task{
        {ID: 1, Title: "Изучить NGINX"},
        {ID: 2, Title: "Понять load balancing"},
    }

    mux := http.NewServeMux()

    mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
        w.Header().Set("Content-Type", "application/json; charset=utf-8")
        w.Header().Set("X-Instance-ID", instanceID)
        _ = json.NewEncoder(w).Encode(map[string]string{
            "status":   "ok",
            "instance": instanceID,
        })
    })

    mux.HandleFunc("/v1/tasks", func(w http.ResponseWriter, r *http.Request) {
        w.Header().Set("Content-Type", "application/json; charset=utf-8")
        w.Header().Set("X-Instance-ID", instanceID)
        _ = json.NewEncoder(w).Encode(tasks)
    })

    addr := ":" + port
    log.Println("tasks service started on", addr, "instance =", instanceID)
    if err := http.ListenAndServe(addr, mux); err != nil {
        log.Fatal(err)
    }
}
```

## Шаг 2. Инициализация Go-модуля

```bash
cd services/tasks
go mod init example.com/tasks-lb
go mod tidy
touch go.sum  # если go.sum не был создан автоматически (нет внешних зависимостей)
```


## Шаг 3. Dockerfile

Файл `services/tasks/Dockerfile`:

```dockerfile
FROM golang:1.25 AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download || true

COPY . .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /app/bin/tasks ./cmd/server

FROM alpine:3.20

WORKDIR /app

COPY --from=builder /app/bin/tasks /app/tasks

EXPOSE 8082

CMD ["/app/tasks"]
```

## Шаг 4. Конфигурация NGINX

Файл `deploy/lb/nginx.conf`:

```nginx
events {}

http {
    upstream tasks_backend {
        server tasks_1:8082;
        server tasks_2:8082;
        server tasks_3:8082;
    }

    server {
        listen 8080;

        location / {
            proxy_pass http://tasks_backend;
            proxy_set_header Host $host;
            proxy_set_header X-Request-ID $request_id;
            proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
            proxy_set_header Authorization $http_authorization;
        }
    }
}
```

## Шаг 5. docker-compose.yml

Файл `deploy/lb/docker-compose.yml`:

```yaml
services:
  tasks_1:
    build:
      context: ../../services/tasks
    container_name: tasks_1
    environment:
      APP_PORT: "8082"
      INSTANCE_ID: "tasks-1"

  tasks_2:
    build:
      context: ../../services/tasks
    container_name: tasks_2
    environment:
      APP_PORT: "8082"
      INSTANCE_ID: "tasks-2"

  tasks_3:
    build:
      context: ../../services/tasks
    container_name: tasks_3
    environment:
      APP_PORT: "8082"
      INSTANCE_ID: "tasks-3"

  nginx:
    image: nginx:1.27-alpine
    container_name: nginx_lb
    ports:
      - "8080:8080"
    volumes:
      - ./nginx.conf:/etc/nginx/nginx.conf:ro
    depends_on:
      - tasks_1
      - tasks_2
      - tasks_3
```

## Шаг 6. Запуск стенда

```bash
cd deploy/lb
docker compose up -d --build
```
![1](image.png)

Проверка состояния контейнеров:

```bash
docker compose ps
```
![2](image-1.png)


## Шаг 7. Проверка health endpoint

```bash
curl -i http://localhost:8080/health
```
![3](image-2.png)
Примечание: При повторных запросах значение `X-Instance-ID` будет меняться между репликами.


## Шаг 8. Проверка балансировки

Для macOS:

```bash
for i in {1..10}; do
  curl -s -I http://localhost:8080/v1/tasks | grep X-Instance-ID
done
```
![4](image-3.png)


## Шаг 9. Проверка передачи заголовков через NGINX

```bash
curl -i http://localhost:8080/v1/tasks -H "Authorization: Bearer demo-token"
```
![5](image-4.png)


## Шаг 10. Проверка отказоустойчивости

Остановка одной реплики:

```bash
docker compose stop tasks_1
```

Повторная серия запросов:

```bash
for i in {1..5}; do
  curl -s -I http://localhost:8080/v1/tasks | grep X-Instance-ID
done
```
![6](image-5.png)


## Шаг 11. Возврат реплики в работу

```bash
docker compose start tasks_1
```
![7](image-6.png)
После старта балансировка снова распределяется между всеми репликами.


## Выполненное дополнительное задание

Выполнен **Вариант 1: добавлена третья реплика**.

В `docker-compose.yml` добавлен сервис `tasks_3` с переменной `INSTANCE_ID: "tasks-3"`.  
В `nginx.conf` в блок `upstream tasks_backend` добавлена строка `server tasks_3:8082`.

После перезапуска стенда:
![8](image-7.png)


## Ответы на контрольные вопросы

**1. Что такое горизонтальное масштабирование?**  
Горизонтальное масштабирование — это увеличение пропускной способности системы за счёт запуска нескольких одинаковых экземпляров сервиса, а не за счёт усиления одного сервера.

**2. Чем оно отличается от вертикального масштабирования?**  
Вертикальное масштабирование увеличивает ресурсы одного экземпляра: CPU, RAM, дисковую подсистему. Горизонтальное — увеличивает количество экземпляров. Вертикальное проще на старте, но ограничено физическими пределами машины. Горизонтальное требует балансировщика и stateless-архитектуры, но лучше масштабируется при росте нагрузки.

**3. Зачем нужен load balancer?**  
Load balancer принимает все входящие запросы от клиентов и распределяет их между доступными репликами сервиса. Без него клиент должен был бы знать адреса всех реплик и самостоятельно выбирать, куда обращаться, что неудобно и ненадёжно.

**4. Какую роль в этой работе выполняет NGINX?**  
NGINX выступает в роли reverse proxy и балансировщика нагрузки. Он слушает порт 8080, принимает запросы клиентов, пересылает их в одну из реплик сервиса tasks и возвращает ответ обратно клиенту.

**5. Что такое upstream в NGINX?**  
Upstream — это именованная группа backend-серверов, между которыми NGINX распределяет трафик. В данной работе upstream `tasks_backend` включает три реплики. По умолчанию используется алгоритм round-robin — запросы распределяются последовательно.

**6. Почему для горизонтального масштабирования желательно делать сервис stateless?**  
Если сервис хранит состояние в локальной памяти процесса, разные реплики будут иметь разные данные. Запрос на запись может попасть в одну реплику, а запрос на чтение — в другую, которая этих данных не знает. Stateless-сервис хранит всё важное состояние во внешних системах: базе данных, Redis и т.д. Тогда любая реплика может обработать любой запрос корректно.

**7. Зачем нужен health endpoint?**  
Health endpoint позволяет быстро проверить, что экземпляр сервиса запущен и отвечает на запросы. Он используется для ручной диагностики, систем мониторинга, а также балансировщиками нагрузки для исключения недоступных реплик из ротации.

**8. Почему полезно добавлять X-Instance-ID?**  
Заголовок X-Instance-ID позволяет клиенту видеть, какая именно реплика обработала запрос. Это делает балансировку наблюдаемой: можно отправить серию запросов и убедиться, что они реально распределяются между разными экземплярами.

**9. Что происходит при остановке одной из реплик?**  
NGINX перестаёт направлять запросы в недоступную реплику и продолжает обслуживать трафик через оставшиеся. Система деградирует частично, но не выходит из строя полностью. При возврате реплики в работу она снова включается в ротацию.

**10. Почему клиенту удобнее работать с одной точкой входа, а не с несколькими адресами реплик?**  
Единая точка входа скрывает от клиента внутреннюю структуру системы. Клиенту не нужно знать, сколько реплик запущено, какие из них доступны и как между ними распределять запросы. Балансировщик берёт эту ответственность на себя, что упрощает клиентский код и повышает надёжность системы в целом.
```