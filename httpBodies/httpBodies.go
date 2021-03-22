package httpBodies

import (
	"fmt"
	"log"
	"strconv"
	"strings"

	"github.com/evoila/osb-backup-agent/errorlog"
)

const Status_running = "RUNNING"
const Status_success = "SUCCEEDED"
const Status_failed = "FAILED"

type BackupResponse struct {
	Status                   string   `json:"status"`
	Message                  string   `json:"message"`
	State                    string   `json:"state"`
	ErrorMessage             string   `json:"error_message,omitempty"`
	Type                     string   `json:"type"`
	Compression              bool     `json:"compression"`
	SkipStorage              bool     `json:"skip_storage"`
	SkipSSL                  bool     `json:"skipSSL"`
	Region                   string   `json:"region,omitempty"`
	Endpoint                 string   `json:"endpoint,omitempty"`
	Bucket                   string   `json:"bucket,omitempty"`
	AuthUrl                  string   `json:"authUrl,omitempty"`
	Domain                   string   `json:"domain,omitempty"`
	ContainerName            string   `json:"container_name,omitempty"`
	ProjectName              string   `json:"project_name,omitempty"`
	FileName                 string   `json:"filename"`
	FilenamePrefix           string   `json:"filenamePrefix"`
	FileSize                 FileSize `json:"filesize"`
	StartTime                string   `json:"start_time"`
	EndTime                  string   `json:"end_time"`
	ExecutionTime            int64    `json:"execution_time_ms"`
	PreBackupLockLog         string   `json:"pre_backup_lock_log"`
	PreBackupLockErrorLog    string   `json:"pre_backup_lock_errorlog"`
	PreBackupCheckLog        string   `json:"pre_backup_check_log"`
	PreBackupCheckErrorLog   string   `json:"pre_backup_check_errorlog"`
	BackupLog                string   `json:"backup_log"`
	BackupErrorLog           string   `json:"backup_errorlog"`
	BackupCleanupLog         string   `json:"backup_cleanup_log"`
	BackupCleanupErrorLog    string   `json:"backup_cleanup_errorlog"`
	PostBackupUnlockLog      string   `json:"post_backup_unlock_log"`
	PostBackupUnlockErrorLog string   `json:"post_backup_unlock_errorlog"`
}

type FileSize struct {
	Size int64  `json:"size"`
	Unit string `json:"unit"`
}

type RestoreResponse struct {
	Status                    string `json:"status"`
	Message                   string `json:"message"`
	State                     string `json:"state"`
	ErrorMessage              string `json:"error_message,omitempty"`
	Type                      string `json:"type"`
	Compression               bool   `json:"compression"`
	SkipSSL                   bool   `json:"skipSSL"`
	SkipStorage               bool   `json:"skip_storage"`
	StartTime                 string `json:"start_time"`
	EndTime                   string `json:"end_time"`
	ExecutionTime             int64  `json:"execution_time_ms"`
	PreRestoreLockLog         string `json:"pre_restore_lock_log"`
	PreRestoreLockErrorLog    string `json:"pre_restore_lock_errorlog"`
	RestoreLog                string `json:"restore_log"`
	RestoreErrorLog           string `json:"restore_errorlog"`
	RestoreCleanupLog         string `json:"restore_cleanup_log"`
	RestoreCleanupErrorLog    string `json:"restore_cleanup_errorlog"`
	PostRestoreUnlockLog      string `json:"post_restore_unlock_log"`
	PostRestoreUnlockErrorLog string `json:"post_restore_unlock_errorlog"`
}

type ErrorResponse struct {
	Message      string `json:"message"`
	State        string `json:"state"`
	ErrorMessage string `json:"error_message"`
}

type BackupBody struct {
	Id             string
	Compression    bool
	Encryption_key string
	Destination    DestinationInformation
	Backup         DbInformation
}

type RestoreBody struct {
	Id             string
	Compression    bool
	Encryption_key string
	Destination    DestinationInformation
	Restore        DbInformation
}

type DestinationInformation struct {
	Type        string
	SkipStorage bool

	Bucket     string
	Region     string
	Endpoint   string
	AuthKey    string
	AuthSecret string
	SkipSSL    bool

	FilenamePrefix string
	Filename       string

	AuthUrl        string
	Domain         string
	Container_name string
	Project_name   string
	Username       string
	Password       string
}

type DbInformation struct {
	Host       string
	Username   string
	Password   string
	Database   string
	Parameters []map[string]interface{}
}

func PrintOutBackupBody(body BackupBody) {
	authSecret := GetRedactedOrEmptyPasswordString(body.Destination.AuthSecret)
	swiftPassword := GetRedactedOrEmptyPasswordString(body.Destination.Password)
	dbPassword := GetRedactedOrEmptyPasswordString(body.Backup.Password)

	log.Println("Backup Request Body: {\n",
		errorlog.Concat([]string{"    \"id\" : \"", body.Id, "\",\n"}, ""),
		errorlog.Concat([]string{"    \"compression\" : \"", strconv.FormatBool(body.Compression), "\",\n"}, ""),
		errorlog.Concat([]string{"    \"encryption_key\" : \"", body.Encryption_key, "\",\n"}, ""),
		"    \"destination\" : {\n",
		errorlog.Concat([]string{"        \"type\" : \"", body.Destination.Type, "\",\n"}, ""),
		errorlog.Concat([]string{"        \"skipStorage\" : \"", strconv.FormatBool(body.Destination.SkipStorage), "\",\n"}, ""),
		errorlog.Concat([]string{"        \"bucket\" : \"", body.Destination.Bucket, "\",\n"}, ""),
		errorlog.Concat([]string{"        \"endpoint\" : \"", body.Destination.Endpoint, "\",\n"}, ""),
		errorlog.Concat([]string{"        \"region\" : \"", body.Destination.Region, "\",\n"}, ""),
		errorlog.Concat([]string{"        \"authKey\" : \"", body.Destination.AuthKey, "\",\n"}, ""),
		errorlog.Concat([]string{"        \"authSecret\" : \"", authSecret, "\",\n"}, ""),
		errorlog.Concat([]string{"        \"skipSSL\" : \"", strconv.FormatBool(body.Destination.SkipSSL), "\",\n"}, ""),
		errorlog.Concat([]string{"        \"filenamePrefix\" : \"", body.Destination.FilenamePrefix, "\",\n"}, ""),
		"\n",
		errorlog.Concat([]string{"        \"authUrl\" : \"", body.Destination.AuthUrl, "\",\n"}, ""),
		errorlog.Concat([]string{"        \"domain\" : \"", body.Destination.Domain, "\",\n"}, ""),
		errorlog.Concat([]string{"        \"container_name\" : \"", body.Destination.Container_name, "\",\n"}, ""),
		errorlog.Concat([]string{"        \"project_name\" : \"", body.Destination.Project_name, "\",\n"}, ""),
		errorlog.Concat([]string{"        \"username\" : \"", body.Destination.Username, "\",\n"}, ""),
		errorlog.Concat([]string{"        \"password\" : \"", swiftPassword, "\",\n"}, ""),
		"    },\n",
		"    \"backup\" : {\n",
		errorlog.Concat([]string{"        \"host\" : \"", body.Backup.Host, "\",\n"}, ""),
		errorlog.Concat([]string{"        \"user\" : \"", body.Backup.Username, "\",\n"}, ""),
		errorlog.Concat([]string{"        \"password\" : \"", dbPassword, "\",\n"}, ""),
		errorlog.Concat([]string{"        \"database\" : \"", body.Backup.Database, "\",\n"}, ""),
		"        \"parameters\" : ",
		getParametersAsLogStringSlice(body.Backup.Parameters),
		"    }\n",
		"}")

}

// Returns true if no fields are missing
func CheckForMissingFieldsInRestoreBody(body RestoreBody) (bool, string) {
	missingFields := ""
	if body.Id == "" {
		missingFields += " id"
	}
	if body.Encryption_key == "" {
		missingFields += " encryption_key"
	}

	valid, fields := CheckForMissingFieldDestinationInformation(body.Destination, false)
	if !valid {
		missingFields += " destination(" + fields + ")"
	}
	valid, fields = CheckForMissingFieldsInDbInformation(body.Restore)
	if !valid {
		missingFields += " restore(" + fields + ")"
	}
	return missingFields == "", missingFields
}

// Returns true if no fields are missing
func CheckForMissingFieldsInBackupBody(body BackupBody) (bool, string) {
	missingFields := ""
	if body.Id == "" {
		missingFields += " id"
	}
	if body.Encryption_key == "" {
		missingFields += " encryption_key"
	}
	valid, fields := CheckForMissingFieldDestinationInformation(body.Destination, true)
	if !valid {
		missingFields += " destination(" + fields + ")"
	}
	valid, fields = CheckForMissingFieldsInDbInformation(body.Backup)
	if !valid {
		missingFields += " backup(" + fields + ")"
	}
	return missingFields == "", missingFields
}

func CheckForMissingFieldDestinationInformation(body DestinationInformation, fileCanBeMissing bool) (bool, string) {
	missingFields := ""
	if body.Type == "S3" {
		if body.AuthKey == "" {
			missingFields += " authKey"
		}
		if body.AuthSecret == "" {
			missingFields += " authSecret"
		}
		if body.Bucket == "" {
			missingFields += " bucket"
		}
		if body.SkipStorage && body.Endpoint == "" {
			missingFields += " endpoint"
		}
		if body.Filename == "" && !fileCanBeMissing {
			missingFields += " filename"
		}
		if body.FilenamePrefix == "" {
			missingFields += " filenamePrefix"
		}
		return missingFields == "", missingFields
	} else if body.Type == "SWIFT" {
		if body.AuthUrl == "" {
			missingFields += " authUrl"
		}
		if body.Domain == "" {
			missingFields += " domain"
		}
		if body.Container_name == "" {
			missingFields += " container_name"
		}
		if body.Project_name == "" {
			missingFields += " project_name"
		}
		if body.Username == "" {
			missingFields += " username"
		}
		if body.Password == "" {
			missingFields += " password"
		}
		if body.Filename == "" && !fileCanBeMissing {
			missingFields += " filename"
		}
		return missingFields == "", missingFields
	} else if body.Type == "" {
		return false, " type"
	}
	return false, " supported type"
}

func CheckForMissingFieldsInDbInformation(body DbInformation) (bool, string) {
	missingFields := ""
	if body.Database == "" {
		missingFields += " database"
	}
	if body.Host == "" {
		missingFields += " host"
	}
	if body.Password == "" {
		missingFields += " password"
	}
	if body.Username == "" {
		missingFields += " username"
	}
	return missingFields == "", missingFields
}

func PrintOutRestoreBody(body RestoreBody) {
	authSecret := GetRedactedOrEmptyPasswordString(body.Destination.AuthSecret)
	swiftPassword := GetRedactedOrEmptyPasswordString(body.Destination.Password)
	dbPassword := GetRedactedOrEmptyPasswordString(body.Restore.Password)
	privateEncryptionKey := GetRedactedOrEmptyPasswordString(body.Encryption_key)

	log.Println("Restore Request Body: {\n",
		errorlog.Concat([]string{"    \"id\" : \"", body.Id, "\",\n"}, ""),
		errorlog.Concat([]string{"    \"compression\" : \"", strconv.FormatBool(body.Compression), "\",\n"}, ""),
		errorlog.Concat([]string{"    \"encryption_key\" : \"", privateEncryptionKey, "\",\n"}, ""),
		"    \"destination\" : {\n",
		errorlog.Concat([]string{"        \"type\" : \"", body.Destination.Type, "\",\n"}, ""),
		errorlog.Concat([]string{"        \"skipStorage\" : \"", strconv.FormatBool(body.Destination.SkipStorage), "\",\n"}, ""),
		errorlog.Concat([]string{"        \"bucket\" : \"", body.Destination.Bucket, "\",\n"}, ""),
		errorlog.Concat([]string{"        \"endpoint\" : \"", body.Destination.Endpoint, "\",\n"}, ""),
		errorlog.Concat([]string{"        \"region\" : \"", body.Destination.Region, "\",\n"}, ""),
		errorlog.Concat([]string{"        \"authKey\" : \"", body.Destination.AuthKey, "\",\n"}, ""),
		errorlog.Concat([]string{"        \"authSecret\" : \"", authSecret, "\",\n"}, ""),
		errorlog.Concat([]string{"        \"skipSSL\" : \"", strconv.FormatBool(body.Destination.SkipSSL), "\",\n"}, ""),
		errorlog.Concat([]string{"        \"filename\" : \"", body.Destination.Filename, "\",\n"}, ""),
		errorlog.Concat([]string{"        \"filenamePrefix\" : \"", body.Destination.FilenamePrefix, "\",\n"}, ""),
		"\n",
		errorlog.Concat([]string{"        \"authUrl\" : \"", body.Destination.AuthUrl, "\",\n"}, ""),
		errorlog.Concat([]string{"        \"domain\" : \"", body.Destination.Domain, "\",\n"}, ""),
		errorlog.Concat([]string{"        \"container_name\" : \"", body.Destination.Container_name, "\",\n"}, ""),
		errorlog.Concat([]string{"        \"project_name\" : \"", body.Destination.Project_name, "\",\n"}, ""),
		errorlog.Concat([]string{"        \"username\" : \"", body.Destination.Username, "\",\n"}, ""),
		errorlog.Concat([]string{"        \"password\" : \"", swiftPassword, "\",\n"}, ""),
		"    },\n",
		"    \"backup\" : {\n",
		errorlog.Concat([]string{"        \"host\" : \"", body.Restore.Host, "\",\n"}, ""),
		errorlog.Concat([]string{"        \"user\" : \"", body.Restore.Username, "\",\n"}, ""),
		errorlog.Concat([]string{"        \"password\" : \"", dbPassword, "\",\n"}, ""),
		errorlog.Concat([]string{"        \"database\" : \"", body.Restore.Database, "\",\n"}, ""),
		"        \"parameters\" : ",
		getParametersAsLogStringSlice(body.Restore.Parameters),
		"    }\n",
		"}")

}

func getParametersAsLogStringSlice(parameters []map[string]interface{}) []string {
	// Non string simple types will still be returned surrounded by "" !!!
	var strs []string
	for entry, entryValue := range parameters {
		for key, value := range entryValue {
			parsedValue := fmt.Sprintf("%v", value)
			if entry != len(parameters)-1 {
				strs = append(strs, errorlog.Concat([]string{"\n            {\"", key, "\": \"", parsedValue, "\" },"}, ""))
			} else {
				strs = append(strs, errorlog.Concat([]string{"\n            {\"", key, "\": \"", parsedValue, "\" }\n"}, ""))
			}

		}
	}
	return strs
}

func GetParametersAsEnvVarStringSlice(parameters []map[string]interface{}) (envParameters []string) {
	for _, entryValue := range parameters {
		for key, value := range entryValue {
			parsedValue := getAsEnvVar(key, value)
			log.Println("Parsed additional parameter:", parsedValue)
			if strings.Contains(strings.ToLower(key), "secret") || strings.Contains(strings.ToLower(key), "password") {
				log.Println("Parsed additional parameter:", key+"=[redacted]")
			} else {
				log.Println("Parsed additional parameter:", parsedValue)
			}
			envParameters = append(envParameters, parsedValue)
		}
	}
	return
}

func getAsEnvVar(key string, value interface{}) string {
	return fmt.Sprintf("%s=%v", key, value)
}

func GetDestinationInformationAsEnvVarStringSlice(skipStorage bool, body DestinationInformation) (envParameters []string) {
	if !skipStorage {
		return make([]string, 0)
	}

	if body.Type == "S3" {
		list := make([]string, 7)
		list[0] = getAsEnvVar(body.Type+"_BUCKET", body.Bucket)
		list[1] = getAsEnvVar(body.Type+"_ENDPOINT", body.Region)
		list[2] = getAsEnvVar(body.Type+"_REGION", body.Endpoint)
		list[3] = getAsEnvVar(body.Type+"_AUTHKEY", body.AuthKey)
		list[4] = getAsEnvVar(body.Type+"_SKIPSSL", body.SkipSSL)
		list[5] = getAsEnvVar(body.Type+"_AUTHSECRET", body.AuthSecret)
		list[6] = getAsEnvVar(body.Type+"_FILENAMEPREFIX", body.FilenamePrefix)

		log.Println("Adding as destination information:", list[0], list[1], list[2], list[3], list[4], body.Type+"_AUTHSECRET=[redacted]", list[6])

		return list
	} else if body.Type == "SWIFT" {
		list := make([]string, 6)
		list[0] = getAsEnvVar(body.Type+"_AUTHURL", body.AuthUrl)
		list[1] = getAsEnvVar(body.Type+"_DOMAIN", body.Domain)
		list[2] = getAsEnvVar(body.Type+"_CONTAINERNAME", body.Container_name)
		list[3] = getAsEnvVar(body.Type+"_PROJECTNAME", body.Project_name)
		list[4] = getAsEnvVar(body.Type+"_USERNAME", body.Username)
		list[5] = getAsEnvVar(body.Type+"_PASSWORD", body.Password)

		log.Println("Adding as destination information:", list[0], list[1], list[2], list[3], list[4], body.Type+"_PASSWORD=[redacted]")
		return list
	} else {
		return make([]string, 0)
	}
}

func GetRedactedOrEmptyPasswordString(pw string) string {
	if pw != "" {
		return "<redacted>"
	}
	return ""
}
