package logx

import (
	"log"
	"strings"
	"sync/atomic"
)

type Level int32

const (
	LevelError Level = iota
	LevelInfo
	LevelDebug
)

var currentLevel int32 = int32(LevelInfo)

func SetLevelFromString(raw string) Level {
	level := ParseLevel(raw)
	atomic.StoreInt32(&currentLevel, int32(level))
	return level
}

func ParseLevel(raw string) Level {
	switch strings.ToLower(strings.TrimSpace(raw)) {
	case "error":
		return LevelError
	case "debug":
		return LevelDebug
	default:
		return LevelInfo
	}
}

func LevelString(level Level) string {
	switch level {
	case LevelError:
		return "error"
	case LevelDebug:
		return "debug"
	default:
		return "info"
	}
}

func Enabled(level Level) bool {
	return level <= Level(atomic.LoadInt32(&currentLevel))
}

func Errorf(format string, args ...any) {
	if Enabled(LevelError) {
		log.Printf("[ERROR] "+format, args...)
	}
}

func Infof(format string, args ...any) {
	if Enabled(LevelInfo) {
		log.Printf("[INFO] "+format, args...)
	}
}

func Debugf(format string, args ...any) {
	if Enabled(LevelDebug) {
		log.Printf("[DEBUG] "+format, args...)
	}
}
