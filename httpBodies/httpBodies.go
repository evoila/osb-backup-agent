package httpBodies

import (
	"fmt"
	"log"

	"github.com/evoila/osb-backup-agent/errorlog"
)

const Status_success = "SUCCESS"
const Status_failed = "ERROR"

type BackupResponse struct {
	Status              string   `json:"status"`
	Message             string   `json:"message"`
	Region              string   `json:"region"`
	Bucket              string   `json:"bucket"`
	FileName            string   `json:"filename"`
	FileSize            FileSize `json:"filesize"`
	StartTime           string   `json:"start_time"`
	EndTime             string   `json:"end_time"`
	ExecutionTime       int64    `json:"execution_time_ms"`
	PreBackupLockLog    string   `json:"pre_backup_lock_log"`
	PreBackupCheckLog   string   `json:"pre_backup_check_log"`
	BackupLog           string   `json:"backup_log"`
	BackupCleanupLog    string   `json:"backup_cleanup_log"`
	PostBackupUnlockLog string   `json:"post_backup_unlock_log"`
}

type FileSize struct {
	Size int64  `json:"size"`
	Unit string `json:"unit"`
}

type BackupErrorResponse struct {
	Status              string `json:"status"`
	Message             string `json:"message"`
	State               string `json:"state"`
	ErrorMessage        string `json:"error_message"`
	StartTime           string `json:"start_time"`
	EndTime             string `json:"end_time"`
	ExecutionTime       int64  `json:"execution_time_ms"`
	PreBackupLockLog    string `json:"pre_backup_lock_log"`
	PreBackupCheckLog   string `json:"pre_backup_check_log"`
	BackupLog           string `json:"backup_log"`
	BackupCleanupLog    string `json:"backup_cleanup_log"`
	PostBackupUnlockLog string `json:"post_backup_unlock_log"`
}

type RestoreResponse struct {
	Status               string `json:"status"`
	Message              string `json:"message"`
	StartTime            string `json:"start_time"`
	EndTime              string `json:"end_time"`
	ExecutionTime        int64  `json:"execution_time_ms"`
	PreRestoreLockLog    string `json:"pre_restore_lock_log"`
	RestoreLog           string `json:"restore_log"`
	RestoreCleanupLog    string `json:"restore_cleanup_log"`
	PostRestoreUnlockLog string `json:"post_restore_unlock_log"`
}

type RestoreErrorResponse struct {
	Status               string `json:"status"`
	Message              string `json:"message"`
	State                string `json:"state"`
	ErrorMessage         string `json:"error_message"`
	StartTime            string `json:"start_time"`
	EndTime              string `json:"end_time"`
	ExecutionTime        int64  `json:"execution_time_ms"`
	PreRestoreLockLog    string `json:"pre_restore_lock_log"`
	RestoreLog           string `json:"restore_log"`
	RestoreCleanupLog    string `json:"restore_cleanup_log"`
	PostRestoreUnlockLog string `json:"post_restore_unlock_log"`
}

type ErrorResponse struct {
	Message      string `json:"message"`
	State        string `json:"state"`
	ErrorMessage string `json:"error_message"`
}

type BackupBody struct {
	Destination DestinationInformation
	Backup      DbInformation
}

type RestoreBody struct {
	Destination DestinationInformation
	Restore     DbInformation
}

type DestinationInformation struct {
	Type       string
	Bucket     string
	Region     string
	AuthKey    string
	AuthSecret string
	Filename   string
}

type DbInformation struct {
	Host       string
	Username   string
	Password   string
	Database   string
	Parameters []map[string]interface{}
}

func PrintOutBackupBody(body BackupBody) {

	log.Println("Backup Request Body: {\n",
		"    \"destination\" : {\n",
		errorlog.Concat([]string{"        \"type\" : \"", body.Destination.Type, "\",\n"}, ""),
		errorlog.Concat([]string{"        \"bucket\" : \"", body.Destination.Bucket, "\",\n"}, ""),
		errorlog.Concat([]string{"        \"region\" : \"", body.Destination.Region, "\",\n"}, ""),
		errorlog.Concat([]string{"        \"authKey\" : \"", body.Destination.AuthKey, "\",\n"}, ""),
		"        \"authSecret\" : <redacted>\n",
		"    },\n",
		"    \"backup\" : {\n",
		errorlog.Concat([]string{"        \"host\" : \"", body.Backup.Host, "\",\n"}, ""),
		errorlog.Concat([]string{"        \"user\" : \"", body.Backup.Username, "\",\n"}, ""),
		"        \"password\" : <redacted>\n",
		errorlog.Concat([]string{"        \"database\" : \"", body.Backup.Database, "\",\n"}, ""),
		"        \"parameters\" : ",
		getParametersAsLogStringSlice(body.Backup.Parameters),
		"    }\n",
		"}")
}

func CheckForMissingFieldsInRestoreBody(body RestoreBody) bool {
	return CheckForMissingFieldDestinationInformation(body.Destination, false) && CheckForMissingFieldsInDbInformation(body.Restore)
}

func CheckForMissingFieldsInBackupBody(body BackupBody) bool {
	return CheckForMissingFieldDestinationInformation(body.Destination, true) && CheckForMissingFieldsInDbInformation(body.Backup)
}

func CheckForMissingFieldDestinationInformation(body DestinationInformation, fileCanBeMissing bool) bool {
	return body.AuthKey != "" && body.AuthSecret != "" && body.Bucket != "" && (body.Filename != "" || fileCanBeMissing) && body.Region != "" && body.Type != ""
}

func CheckForMissingFieldsInDbInformation(body DbInformation) bool {
	return body.Database != "" && body.Host != "" && body.Password != "" && body.Username != ""
}

func PrintOutRestoreBody(body RestoreBody) {

	log.Println("Restore Request Body: {\n",
		"    \"destination\" : {\n",
		errorlog.Concat([]string{"        \"type\" : \"", body.Destination.Type, "\",\n"}, ""),
		errorlog.Concat([]string{"        \"bucket\" : \"", body.Destination.Bucket, "\",\n"}, ""),
		errorlog.Concat([]string{"        \"region\" : \"", body.Destination.Region, "\",\n"}, ""),
		errorlog.Concat([]string{"        \"authKey\" : \"", body.Destination.AuthKey, "\",\n"}, ""),
		"        \"authSecret\" : <redacted>\n",
		errorlog.Concat([]string{"        \"filename\" : \"", body.Destination.Filename, "\",\n"}, ""),
		"    },\n",
		"    \"backup\" : {\n",
		errorlog.Concat([]string{"        \"host\" : \"", body.Restore.Host, "\",\n"}, ""),
		errorlog.Concat([]string{"        \"user\" : \"", body.Restore.Username, "\",\n"}, ""),
		"        \"password\" : <redacted>\n",
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
			parsedValue := fmt.Sprintf("%s=%v", key, value)
			log.Println("Parsed additional parameter:", parsedValue)
			envParameters = append(envParameters, parsedValue)
		}
	}
	return
}
