package shared

import (
	"fmt"
	"log"
	"os"
	"time"
)

var errorLogger *log.Logger

const (
	LogFileName    = "debug.log"
	FilePerms      = 0666
	DateTimeFormat = "2006-01-02 15:04:05"
)

func init() {
	// Clear the log file on startup by truncating it
	file, err := os.OpenFile(LogFileName, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, FilePerms)
	if err != nil {
		log.Printf("Failed to open error log file: %v", err)
		errorLogger = log.New(os.Stderr, "ERROR: ", log.LstdFlags)
		return
	}
	errorLogger = log.New(file, "", log.LstdFlags)
}

func LogError(context string, err error) {
	if errorLogger != nil {
		errorLogger.Printf("[%s] %s: %v", time.Now().Format(DateTimeFormat), context, err)
	}
}

func LogErrorf(context string, format string, args ...interface{}) {
	if errorLogger != nil {
		message := fmt.Sprintf(format, args...)
		errorLogger.Printf("[%s] %s: %s", time.Now().Format(DateTimeFormat), context, message)
	}
}
