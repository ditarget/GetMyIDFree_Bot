# === Stage 1: Сборка бинарника ===
FROM golang:1.25-alpine AS builder

WORKDIR /app

# Копируем go.mod и загружаем зависимости
COPY go.mod ./
COPY go.sum ./
RUN go mod download

# 🔻 Копируем ВЕСЬ исходный код, включая папки
COPY main.go ./
COPY bot/ ./bot/
COPY logger/ ./logger/
COPY storage/ ./storage/

# Собираем статически скомпилированный бинарник (чтобы работал в alpine)
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o bot main.go


# === Stage 2: Минимальный образ для запуска ===
FROM alpine:latest

# Установим сертификаты для HTTPS (Telegram API)
RUN apk --no-cache add ca-certificates

# Рабочая директория
WORKDIR /root

# 🔻 Копируем бинарник из builder
COPY --from=builder /app/bot .

# Создаём папки (опционально, но безопасно)
RUN mkdir -p /root/data /root/logs

# Запускаем бота
CMD ["./bot"]