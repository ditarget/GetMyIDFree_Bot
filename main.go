// main.go
package main

import (
	"log"
	"os"

	"GteMyID/bot"
	"GteMyID/logger"
	"GteMyID/storage"

	"github.com/joho/godotenv"
)

func main() {
	// Настройка логгера
	logFile := logger.SetupLogger()
	defer logFile.Close()

	// Загружаем .env
	err := godotenv.Load("/root/.env")
	if err != nil {
		log.Println("⚠️ .env file not found, using environment variables")
	}

	botToken := os.Getenv("BOT_TOKEN")
	if botToken == "" {
		log.Fatal("❌ BOT_TOKEN is not set in environment")
	}

	// Загружаем пользователей
	users := storage.LoadUsers()
	log.Printf("📁 Loaded %d users from storage", len(users))

	// Запускаем бота
	bot.Start(botToken, users)
}
