FROM golang:1.25 AS builder

WORKDIR /app

COPY internal/go.mod internal/go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /app/bin/tasks ./cmd/tasks

FROM alpine:3.20

WORKDIR /app

COPY --from=builder /app/bin/tasks /app/tasks

EXPOSE 8082

CMD ["/app/tasks"]
