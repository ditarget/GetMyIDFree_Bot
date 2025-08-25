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
	// Проверим, куда мы пытаемся сохранить
	log.Printf("💾 Попытка сохранить users.json в: %s", usersFile)

	// Создаём папку, если её нет
	err := os.MkdirAll(filepath.Dir(usersFile), 0755)
	if err != nil {
		log.Printf("❌ Ошибка создания папки: %v", err)
		return
	}

	// Создаём файл
	file, err := os.Create(usersFile)
	if err != nil {
		log.Printf("❌ Ошибка создания файла users.json: %v", err)
		return
	}
	defer file.Close()

	// Записываем JSON
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

// Настраивает логирование в файл + stdout
func setupLogger() *os.File {
	// Создаём папку для логов
	if err := os.MkdirAll(logsDir, 0755); err != nil {
		log.Fatalf("❌ Failed to create logs directory: %v", err)
	}

	// Открываем файл лога за сегодня
	logFile, err := os.OpenFile(getLogFileName(time.Now()), os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		log.Fatalf("❌ Failed to open log file: %v", err)
	}

	// Логируем и в файл, и в консоль
	multiWriter := io.MultiWriter(logFile, os.Stdout)
	log.SetOutput(multiWriter)
	log.SetFlags(log.Ldate | log.Ltime | log.Lshortfile)

	return logFile
}

func main() {
	// Настройка логгера
	logFile := setupLogger()
	defer logFile.Close()

	// Удаляем старые логи (старше 7 дней)
	cleanupOldLogs()

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

		response := fmt.Sprintf("Your user ID: %d\nCurrent chat ID: %d", userID, chatID)

		// Обработка пересланных сообщений
		if update.Message.ForwardSenderName != "" {
			response += fmt.Sprintf("\nForwarded from: [hidden name] %s", update.Message.ForwardSenderName)
		} else if update.Message.ForwardFrom != nil {
			response += fmt.Sprintf("\nForwarded from: %d", update.Message.ForwardFrom.ID)
		} else if update.Message.ForwardFromChat != nil {
			response += fmt.Sprintf("\nForwarded from chat: %d", update.Message.ForwardFromChat.ID)
		}

		// Отправляем ответ
		msg := tgbotapi.NewMessage(chatID, response)
		_, err := bot.Send(msg)
		if err != nil {
			log.Printf("❌ Failed to send message to %d: %v", chatID, err)
		}
	}
}
