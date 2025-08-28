// storage/users.go
package storage

import (
	"encoding/json"
	"log"
	"os"
	"path/filepath"
)

const usersFile = "data/users.json"

type UserRecord struct {
	UserID    int64  `json:"user_id"`
	Username  string `json:"username,omitempty"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name,omitempty"`
	FirstSeen int64  `json:"first_seen"`
}

func LoadUsers() map[int64]UserRecord {
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

func SaveUsers(users map[int64]UserRecord) {
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
