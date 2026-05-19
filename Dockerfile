# Этап 1: сборка бинарника
FROM golang:1.25.1-alpine AS builder

WORKDIR /app

# Копируем файлы модуля
COPY go.mod ./

# Копируем исходный код
COPY main.go ./

# Собираем статический бинарник
RUN CGO_ENABLED=0 GOOS=linux go build -o tasks .

# Этап 2: минимальный итоговый образ
FROM alpine:3.18

WORKDIR /app

# Копируем бинарник из этапа сборки
COPY --from=builder /app/tasks .

# Порт сервиса
EXPOSE 8082

# Запуск
CMD ["./tasks"]
