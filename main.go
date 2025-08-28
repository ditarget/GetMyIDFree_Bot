package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/joho/godotenv"
)

// Структура для хранения данных пользователя
type UserRecord struct {
	UserID    int64  `json:"user_id"`
	Username  string `json:"username,omitempty"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name,omitempty"`
	FirstSeen int64  `json:"first_seen"` // Unix timestamp
}

// Константы
const (
	usersFile = "data/users.json"  // Путь к файлу с пользователями
	logsDir   = "logs"             // Папка для логов
	maxAge    = 7 * 24 * time.Hour // Максимальный возраст логов — 7 дней
)

// Загружает пользователей из JSON
func loadUsers() map[int64]UserRecord {
	users := make(map[int64]UserRecord)

	file, err := os.Open(usersFile)
	if err != nil {
		if os.IsNotExist(err) {
			log.Println("users.json not found, will be created.")
			return users
		}
		log.Printf("Error opening users.json: %v", err)
		return users
	}
	defer file.Close()

	decoder := json.NewDecoder(file)
	err = decoder.Decode(&users)
	if err != nil {
		log.Printf("Error decoding users.json: %v", err)
	}
	return users
}

// Сохраняет пользователей в JSON
func saveUsers(users map[int64]UserRecord) {
	log.Printf("💾 Попытка сохранить users.json в: %s", usersFile)

	err := os.MkdirAll(filepath.Dir(usersFile), 0755)
	if err != nil {
		log.Printf("❌ Ошибка создания папки: %v", err)
		return
	}

	file, err := os.Create(usersFile)
	if err != nil {
		log.Printf("❌ Ошибка создания файла users.json: %v", err)
		return
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")
	err = encoder.Encode(users)
	if err != nil {
		log.Printf("❌ Ошибка записи JSON: %v", err)
		return
	}

	log.Printf("✅ Успешно сохранено %d пользователей в %s", len(users), usersFile)
}

// Регистрирует нового пользователя, если его ещё нет
func registerUser(users map[int64]UserRecord, user *tgbotapi.User) {
	if _, exists := users[user.ID]; exists {
		return
	}

	now := time.Now().Unix()
	username := ""
	if user.UserName != "" {
		username = user.UserName
	}

	users[user.ID] = UserRecord{
		UserID:    user.ID,
		Username:  username,
		FirstName: user.FirstName,
		LastName:  user.LastName,
		FirstSeen: now,
	}

	log.Printf("New user registered: ID=%d, Name=%s %s, Username=@%s",
		user.ID, user.FirstName, user.LastName, user.UserName)
}

// Возвращает имя файла лога по дате
func getLogFileName(t time.Time) string {
	return filepath.Join(logsDir, "bot-"+t.Format("2006-01-02")+".log")
}

// Очищает логи старше maxAge
func cleanupOldLogs() {
	now := time.Now()
	err := filepath.Walk(logsDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}
		filename := info.Name()
		if strings.HasPrefix(filename, "bot-") && strings.HasSuffix(filename, ".log") {
			dateStr := strings.TrimSuffix(strings.TrimPrefix(filename, "bot-"), ".log")
			if logDate, err := time.Parse("2006-01-02", dateStr); err == nil {
				if now.Sub(logDate) > maxAge {
					os.Remove(path)
					log.Printf("🧹 Deleted old log: %s", path)
				}
			}
		}
		return nil
	})
	if err != nil {
		log.Printf("⚠️ Error during log cleanup: %v", err)
	}
}

// setupLogger запускает логгер и фоновую ротацию
func setupLogger() *os.File {
	// Создаём папку
	if err := os.MkdirAll(logsDir, 0755); err != nil {
		log.Fatalf("❌ Failed to create logs directory: %v", err)
	}

	// Открываем текущий файл
	logFile := openLogFile()
	multiWriter := io.MultiWriter(logFile, os.Stdout)
	log.SetOutput(multiWriter)
	log.SetFlags(log.Ldate | log.Ltime | log.Lshortfile)

	// Фоновая горутина: ротация логов + очистка старых
	go func() {
		ticker := time.NewTicker(10 * time.Minute) // Проверяем каждые 10 минут
		defer ticker.Stop()

		for range ticker.C {
			// 1. Проверяем, нужно ли сменить файл лога
			newFile := reopenLogFileIfNewDay(logFile)
			if newFile != nil {
				logFile.Close()
				multiWriter := io.MultiWriter(newFile, os.Stdout)
				log.SetOutput(multiWriter)
				logFile = newFile
			}

			// 2. Раз в сутки (или при старте нового дня) чистим старые логи
			now := time.Now()
			// Если время близко к 00:00–00:10 — чистим (защита от многократного вызова)
			if now.Hour() == 0 && now.Minute() < 15 {
				cleanupOldLogs()
			}
		}
	}()

	return logFile
}

// openLogFile открывает файл по текущей дате
func openLogFile() *os.File {
	filename := getLogFileName(time.Now())
	file, err := os.OpenFile(filename, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		log.Printf("❌ Failed to open log file %s: %v", filename, err)
		log.Fatal(err)
	}
	return file
}

// reopenLogFileIfNewDay проверяет, нужно ли сменить файл
// Возвращает *os.File, если дата изменилась, иначе nil
func reopenLogFileIfNewDay(currentFile *os.File) *os.File {
	currentDate := time.Now().Format("2006-01-02")
	fileInfo, err := currentFile.Stat()
	if err != nil {
		log.Printf("⚠️ Cannot stat current log file: %v", err)
		return nil
	}
	fileDate := fileInfo.ModTime().Format("2006-01-02")

	if currentDate != fileDate {
		log.Println("🔄 Date changed, rotating log file...")
		newFile := openLogFile()
		return newFile
	}
	return nil
}

func main() {
	// Настройка логгера (с ротацией)
	logFile := setupLogger()
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

	// Создаём бота
	bot, err := tgbotapi.NewBotAPI(botToken)
	if err != nil {
		log.Fatalf("❌ Failed to create Telegram bot: %v", err)
	}

	log.Printf("✅ Bot is running as @%s", bot.Self.UserName)

	// Загружаем пользователей
	users := loadUsers()
	log.Printf("📁 Loaded %d users from storage", len(users))

	// Настройка получения обновлений
	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates := bot.GetUpdatesChan(u)

	log.Println("📨 Bot is listening for messages...")

	for update := range updates {
		if update.Message == nil {
			continue
		}

		// Регистрируем пользователя
		registerUser(users, update.Message.From)
		saveUsers(users) // Сохраняем сразу

		// Формируем ответ
		userID := update.Message.From.ID
		chatID := update.Message.Chat.ID

		response := fmt.Sprintf("<b>Your user ID:</b> <code>%d</code>\n<b>Current chat ID:</b> <code>%d</code>", userID, chatID)

		// Обработка пересланных сообщений
		if update.Message.ForwardSenderName != "" {
			response += fmt.Sprintf("\n<b>Forwarded from:</b> [hidden name] %s", update.Message.ForwardSenderName)
		} else if update.Message.ForwardFrom != nil {
			response += fmt.Sprintf("\n<b>Forwarded from:</b> <code>%d</code>", update.Message.ForwardFrom.ID)
		} else if update.Message.ForwardFromChat != nil {
			response += fmt.Sprintf("\n<b>Forwarded from chat:</b> <code>%d</code>", update.Message.ForwardFromChat.ID)
		}

		// Отправляем ответ
		msg := tgbotapi.NewMessage(chatID, response)
		msg.ParseMode = "HTML"
		_, err := bot.Send(msg)
		if err != nil {
			log.Printf("❌ Failed to send message to %d: %v", chatID, err)
		}
	}
}
