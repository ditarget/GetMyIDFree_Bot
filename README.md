# Telegram ID Bot / Бот для получения ID в Telegram

<p align="center">
  <img src="https://img.shields.io/badge/Go-1.25+-00ADD8?logo=go&logoColor=white" alt="Go">
  <img src="https://img.shields.io/badge/Docker-2496ED?logo=docker&logoColor=white" alt="Docker">
  <img src="https://img.shields.io/badge/Telegram-B9A3EC?logo=telegram&logoColor=white" alt="Telegram">
  <img src="https://img.shields.io/badge/License-MIT-blue" alt="License">
</p>

Простой и **бесплатный** Telegram-бот на Go, который помогает узнать ваш **User ID** и **Chat ID**.  
Показывает ID автора при пересылке сообщения.  
Никаких подписок, оплаты или рекламы — навсегда.

---

## 📌 Описание

Этот бот полезен, если вы:
- Разрабатываете своего Telegram-бота
- Настраиваете доступ по ID
- Просто хотите узнать свой ID

### 🔹 Функции
- Показывает ваш `User ID` и `Chat ID`
- Определяет `ID автора` при пересылке сообщения
- Сохраняет список пользователей в `users.json` (локально)
- Ведёт логи с ротацией (удаляет старше 7 дней)
- Работает в Docker с автозапуском

### 🔐 Безопасность
- **Токен бота** хранится в `.env` (не попадает в Git)
- Никакие данные не отправляются на сторонние сервисы
- Не требует прав администратора

### 🚫 Нет
- Подписок
- Оплаты
- Рекламы
  
---

## 🚀 Запуск

### 1. Установите зависимости
- [Docker](https://docs.docker.com/get-docker/) и [Docker Compose](https://docs.docker.com/compose/install/)

### 2. Склонируйте репозиторий
```bash
git clone https://github.com/ditarget/GetMyIDFree_Bot.git
cd GetMyIDFree_Bot
```

### 3. Создайте .env файл
Откройте .env и добавьте токен от @BotFather:
```
BOT_TOKEN=123456789:YOUR_TELEGRAM_BOT_TOKEN_HERE
```

### 4. Создайте папки для данных
```bash
mkdir data logs
```
### 5. Запустите бота
```bash
docker-compose up -d --build
```

Бот запустится и будет работать в фоне.
Он автоматически перезапустится при ошибках или перезагрузке сервера.

### 📂 Структура данных
```
data/users.json    # Список пользователей (ID, имя, дата первого входа)
logs/bot-*.log     # Логи по дням (удаляются старше 7 дней)
```

### 🛠 Поддержка
Если найдёте баг или хотите улучшение — создайте Issue.
