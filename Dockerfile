# === Stage 1: Сборка бинарника ===
FROM golang:1.25-alpine AS builder

WORKDIR /app

# Копируем go.mod и зависимости
COPY go.mod go.sum ./
RUN go mod download

# 🔹 Копируем ВЕСЬ исходный код
COPY main.go ./
COPY bot/ ./bot/
COPY logger/ ./logger/
COPY storage/ ./storage/

# 🛠 Собираем бинарник с именем, отличным от папки, например: app
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o app main.go


# === Stage 2: Минимальный образ для запуска ===
FROM alpine:latest

# Устанавливаем сертификаты для HTTPS (Telegram API)
RUN apk --no-cache add ca-certificates

# Рабочая директория
WORKDIR /root

# 🔹 Копируем бинарник `app` → переименовываем в `bot` при копировании
COPY --from=builder /app/app ./bot

# 🔐 Делаем исполняемым!
RUN chmod +x ./bot

# Создаём папки для данных и логов
RUN mkdir -p /root/data /root/logs

# Запускаем бота
CMD ["./bot"]