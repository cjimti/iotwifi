package bunyan

import (
	"fmt"
	"regexp"
	"strings"
	"time"
)

func stringDefault(value string, defaultValue string) string {
	if value == "" {
		return defaultValue
	}
	return value
}

func intDefault(value int, defaultValue int) int {
	if value == 0 {
		return defaultValue
	}
	return value
}

func nowTimestamp() string {
	return time.Now().UTC().Format("2006-01-02T15:04:05.999Z07:00")
}

func typeName(value interface{}) string {
	return fmt.Sprintf("%T", value)
}

func isHashMap(value interface{}) bool {
	r := regexp.MustCompile(`^*?map\[string\]`)
	return r.MatchString(typeName(value))
}

func isError(value interface{}) bool {
	r := regexp.MustCompile(`^*?errors.errorString`)
	return r.MatchString(typeName(value))
}

func canSetField(key interface{}) bool {
	switch strings.ToLower(key.(string)) {
	case "v":
		return false
	case "level":
		return false
	default:
		return true
	}
}

func toLogLevelInt(level string) int {
	switch strings.ToLower(level) {
	case LogLevelFatal:
		return 60
	case LogLevelError:
		return 50
	case LogLevelWarn:
		return 40
	case LogLevelInfo:
		return 30
	case LogLevelDebug:
		return 20
	case LogLevelTrace:
		return 10
	default:
		return 0
	}
}
