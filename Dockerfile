FROM golang:1.25-alpine AS builder

WORKDIR /app

# Копируем go.mod и загружаем зависимости
COPY go.mod ./
COPY go.sum ./
RUN go mod download

# Копируем исходный код
COPY main.go ./

# Собираем бинарник
RUN go build -o bot main.go

# Финальный образ (маленький)
FROM alpine:latest

# Установим сертификаты для HTTPS (Telegram API)
RUN apk --no-cache add ca-certificates

# Рабочая директория
WORKDIR /root

# Копируем бинарник из builder
COPY --from=builder /app/bot .

# Создаём папки (опционально, но безопасно)
# Они будут перекрыты volume, но на случай отсутствия — пусть будут
RUN mkdir -p /root/data /root/logs

# Запускаем бота
CMD ["./bot"]