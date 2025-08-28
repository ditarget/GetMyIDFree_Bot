// bot/bot.go
package bot

import (
	"fmt"
	"log"
	"time"

	"GteMyID/storage"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

func Start(token string, users map[int64]storage.UserRecord) {
	bot, err := tgbotapi.NewBotAPI(token)
	if err != nil {
		log.Fatalf("❌ Failed to create Telegram bot: %v", err)
	}

	log.Printf("✅ Bot is running as @%s", bot.Self.UserName)

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
		storage.SaveUsers(users)

		// Формируем ответ
		userID := update.Message.From.ID
		chatID := update.Message.Chat.ID

		response := fmt.Sprintf("<b>Your user ID:</b> <code>%d</code>\n<b>Current chat ID:</b> <code>%d</code>", userID, chatID)

		if update.Message.ForwardSenderName != "" {
			response += fmt.Sprintf("\n<b>Forwarded from:</b> [hidden name] %s", update.Message.ForwardSenderName)
		} else if update.Message.ForwardFrom != nil {
			response += fmt.Sprintf("\n<b>Forwarded from:</b> <code>%d</code>", update.Message.ForwardFrom.ID)
		} else if update.Message.ForwardFromChat != nil {
			response += fmt.Sprintf("\n<b>Forwarded from chat:</b> <code>%d</code>", update.Message.ForwardFromChat.ID)
		}

		msg := tgbotapi.NewMessage(chatID, response)
		msg.ParseMode = "HTML"
		_, err := bot.Send(msg)
		if err != nil {
			log.Printf("❌ Failed to send message to %d: %v", chatID, err)
		}
	}
}

func registerUser(users map[int64]storage.UserRecord, user *tgbotapi.User) {
	if _, exists := users[user.ID]; exists {
		return
	}

	users[user.ID] = storage.UserRecord{
		UserID:    user.ID,
		Username:  user.UserName,
		FirstName: user.FirstName,
		LastName:  user.LastName,
		FirstSeen: time.Now().Unix(),
	}

	log.Printf("New user registered: ID=%d, Name=%s %s, Username=@%s",
		user.ID, user.FirstName, user.LastName, user.UserName)
}
