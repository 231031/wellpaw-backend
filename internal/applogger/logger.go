package applogger

import (
	"io"
	"log"
	"time"
)

var bangkok *time.Location
var appLogger *log.Logger

func InitLogtime(multiwriter io.Writer) {
	var err error
	bangkok, err = time.LoadLocation("Asia/Bangkok")
	if err != nil {
		log.Fatalf("Failed to load Asia/Bangkok timezone: %v", err)
	}

	appLogger = log.New(multiwriter, "", 0)
}

func timestamp() string {
	return time.Now().In(bangkok).Format("2006-01-02 15:04:05")
}

func LogInfo(message string, level string) {
	appLogger.Printf("[INFO] %s %s | %s", timestamp(), level, message)
}

func LogError(message string, level string) {
	appLogger.Printf("[ERROR] %s %s | %s", timestamp(), level, message)
}

func LogWithContext(level string, context string, message string) {
	appLogger.Printf("[%s] %s | %s: %s", level, timestamp(), context, message)
}
