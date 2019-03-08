package configuration

import (
	"errors"
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

func GetScriptsPath() string {
	return getStringEnvVariableWithDefault("scripts_path", "/var/vcap/jobs/backup-agent/backup")
}

func GetPort() int {
	stringedValue := getStringEnvVariableWithDefault("client_port", "8000")
	value := parseInt(stringedValue)
	if value < 0 {
		log.Println("[ERROR]", "Could not parse '", stringedValue, "' or the value is smaller than 0 -> setting to default '8000'")
		value = 8000
	}
	return value
}

func GetBackupDirectory() string {
	return getStringEnvVariable("directory_backup")
}

func GetRestoreDirectory() string {
	return getStringEnvVariable("directory_restore")
}

func IsAllowedToDeleteFiles() bool {
	stringedValue := getStringEnvVariableWithDefault("allowed_to_delete_files", "false")
	value, err := parseBool(stringedValue)
	if err != nil {
		log.Println("[ERROR]", "Could not parse '", stringedValue, "' -> setting to default 'false'")
		value = false
	}
	return value
}

func IsInstructedToSkipStorage() bool {
	stringedValue := getStringEnvVariableWithDefault("skip_storage", "false")
	value, err := parseBool(stringedValue)
	if err != nil {
		log.Println("[ERROR]", "Could not parse '", stringedValue, "' -> setting to default 'false'")
		value = false
	}
	return value
}

func GetMaxJobNumber() int {
	stringedValue := getStringEnvVariableWithDefault("max_job_number", "10")
	value := parseInt(stringedValue)
	if value < 1 {
		log.Println("[ERROR]", "Could not parse '", stringedValue, "' or the value is smaller than 1 -> setting to default '10'")
		value = 10
	}
	return value
}

func getStringEnvVariable(variable string) string {
	var output = os.Getenv(variable)
	if output == "" {
		log.Println("[ERROR]", variable, "is not set.")
	}
	return output
}

func getStringEnvVariableWithDefault(variable string, defaultVariable string) string {
	output := getStringEnvVariable(variable)
	if output == "" {
		log.Println("[ERROR] -> using default:", defaultVariable)
		output = defaultVariable
	}
	return output
}

func parseInt(number string) int {
	i, err := strconv.Atoi(number)
	if err != nil {
		return -1
	}
	return i
}

func parseBool(boolean string) (bool, error) {
	b, err := strconv.ParseBool(boolean)
	if err != nil {
		return false, errors.New("")
	}
	return b, nil
}
