package utils

import (
	"fmt"
	"log"
	"os"
	"time"
)

const (
	LvlDebug = 0
	LvlInfo  = 1
	LvlWarn  = 2
	LvlError = 3
)

var (
	chLog    = make(chan string, 1000)
	logLevel = LvlInfo
)

func InitLogs(path string, lvl string, active bool) {
	if !active {
		return
	}
	logLevel = parseLvl(lvl)
	f, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Fatalf("Failed to open log file: %v", err)
	}
	go func() {
		for msg := range chLog {
			f.WriteString(msg)
		}
	}()
}

func parseLvl(lvl string) int {
	switch lvl {
	case "debug":
		return LvlDebug
	case "info":
		return LvlInfo
	case "warn":
		return LvlWarn
	case "error":
		return LvlError
	default:
		return LvlInfo
	}
}

func writeLog(lvl int, tag string, format string, args ...any) {
	if lvl < logLevel {
		return
	}
	msg := fmt.Sprintf("%s [%s] %s\n",
		time.Now().Format("2006-01-02 15:04:05"),
		tag,
		fmt.Sprintf(format, args...),
	)
	select {
	case chLog <- msg:
	default:
	}
}

func LogDebug(format string, args ...any) {
	writeLog(LvlDebug, "DEBUG", format, args...)
}

func LogInfo(format string, args ...any) {
	writeLog(LvlInfo, "INFO", format, args...)
}

func LogWarn(format string, args ...any) {
	writeLog(LvlWarn, "WARN", format, args...)
}

func LogError(format string, args ...any) {
	writeLog(LvlError, "ERROR", format, args...)
}
