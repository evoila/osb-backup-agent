package errorlog

import (
	"errors"
	"log"
	"strings"
)

func LogError(params ...string) error {
	if len(params) > 0 {
		var message = Concat(params, "")
		log.Println("[ERROR]", message)
		return errors.New(message)
	}
	return nil
}

func Concat(strs []string, seperator string) string {
	return strings.Join(strs, seperator)
}
