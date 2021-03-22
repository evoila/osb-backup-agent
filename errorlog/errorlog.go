package errorlog

import (
	"errors"
	"log"
	"os"
	"strings"
)

var errorLogger *log.Logger

func InitErrorLog() {
	errorLogger = log.New(os.Stderr, log.Default().Prefix(), log.Default().Flags())
}

func LogError(params ...string) error {
	if len(params) > 0 {
		var message = Concat(params, "")
		errorLogger.Println("[ERROR]", message)
		return errors.New(message)
	}
	return nil
}

func Concat(strs []string, seperator string) string {
	return strings.Join(strs, seperator)
}
