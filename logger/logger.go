// logger/logger.go
package logger

import (
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"
)

const logsDir = "logs"

func getLogFileName(t time.Time) string {
	return filepath.Join(logsDir, "bot-"+t.Format("2006-01-02")+".log")
}

// SetupLogger настраивает логирование в файл + stdout и запускает ротацию
func SetupLogger() *os.File {
	if err := os.MkdirAll(logsDir, 0755); err != nil {
		log.Fatalf("❌ Failed to create logs directory: %v", err)
	}

	logFile := openLogFile()
	multiWriter := io.MultiWriter(logFile, os.Stdout)
	log.SetOutput(multiWriter)
	log.SetFlags(log.Ldate | log.Ltime | log.Lshortfile)

	go func() {
		ticker := time.NewTicker(10 * time.Minute)
		defer ticker.Stop()

		for range ticker.C {
			newFile := reopenLogFileIfNewDay(logFile)
			if newFile != nil {
				logFile.Close()
				multiWriter := io.MultiWriter(newFile, os.Stdout)
				log.SetOutput(multiWriter)
				logFile = newFile
			}

			// Очистка старых логов в 00:00–00:14
			now := time.Now()
			if now.Hour() == 0 && now.Minute() < 15 {
				cleanupOldLogs()
			}
		}
	}()

	return logFile
}

func openLogFile() *os.File {
	filename := getLogFileName(time.Now())
	file, err := os.OpenFile(filename, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		log.Printf("❌ Failed to open log file %s: %v", filename, err)
		log.Fatal(err)
	}
	return file
}

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
		return openLogFile()
	}
	return nil
}

func cleanupOldLogs() {
	const maxAge = 7 * 24 * time.Hour
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
