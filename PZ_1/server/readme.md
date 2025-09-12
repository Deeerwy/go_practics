# HTTP Server

Простой HTTP-сервер на Go с REST API endpoints.

## Требования
- Go 1.16+
- Переменная окружения `APP_PORT` 

## Запуск и сборка

### Создание проекта и инициализация модуля
```bash
mkdir helloapi
cd helloapi
go mod init example.com/helloapi
```
# Подключение внешней зависимости
```bash
go get github.com/google/uuid@latest
go mod tidy (очистка)
```

## Команда запуска и сборки
```bash
go build -o helloapi.exe ./cmd/server.\helloapi.exe

go run ./cmd/server
```
# Проверки кода
```bash
go fmt ./...
go vet ./...
```

## Проверка эндпоинтов в другом Powershell
```bash
curl http://localhost:8080/hello
curl http://localhost:8080/user
```

# Ответы: 
```bash
Hello, world!
{"id":"a1b2c3d4-e5f6-7890-abcd-ef1234567890","name":"Gopher"}
```

# Структура проекта:
```bash
.
├── main.go          # Основной файл приложения
├── go.mod          # Модуль Go и зависимости
├── go.sum          # Хеши зависимостей для безопасности
└── README.md       # Документация
```

# Команда для добавления окружения
```bash
export APP_PORT=8081
```

# Возможные проблемы
```bash
go: command not found — переоткройте PowerShell после установки Go, проверьте PATH, переустановите Go-installer.
no Go files in ... при go run — запускайте из корня helloapi командой go run ./cmd/server.
Порт уже занят (listen tcp :8080: bind: ...) — запустите на другом порту ($env:APP_PORT="8081"), либо завершите процесс, который держит порт.
Неверный JSON/заголовки — проверяйте Content-Type: application/json и используйте json.NewEncoder(w).Encode(...).
go: module github.com/google/uuid found (vX.Y.Z), but does not contain package ... — проверьте правильность import и повторите go get/go mod tidy.
```
