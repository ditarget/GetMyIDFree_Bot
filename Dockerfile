FROM golang:1.22-alpine AS builder

WORKDIR /app

# Установим зависимости
COPY go.mod ./
RUN go mod download

# Скопируем исходники
COPY main.go ./

# Соберём бинарник
RUN go build -o bot main.go

# Финальный образ (маленький)
FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /root/

# Копируем бинарник
COPY --from=builder /app/bot .
COPY --from=builder /app/data data
COPY --from=builder /app/logs logs

# Запускаем бота
CMD ["./bot"]