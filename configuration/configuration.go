package configuration

import (
	"log"
	"os"
	"strconv"
)

func GetUsername() string {
	return getStringEnvVariable("client_username")
}

func GetPassword() string {
	return getStringEnvVariable("client_password")
}

func GetPort() int {
	output := getStringEnvVariable("client_port")
	if parseInt(output) < 0 {
		log.Println("[ERROR]", "Could not parse '", output, "' or the value is smaller than 0")
		return -1
	}
	return parseInt(output)
}

func GetBackupDirectory() string {
	return getStringEnvVariable("directory_backup")
}

func GetRestoreDirectory() string {
	return getStringEnvVariable("directory_restore")
}

func getStringEnvVariable(variable string) string {
	var output = os.Getenv(variable)
	if output == "" {
		log.Println("[ERROR]", variable, "is not set.")
	}
	return os.Getenv(variable)
}

func parseInt(number string) int {
	i, err := strconv.Atoi(number)
	if err != nil {
		return -1
	}
	return i
}
