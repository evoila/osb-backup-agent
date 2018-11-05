package backup

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/evoila/osb-backup-agent/configuration"
	"github.com/evoila/osb-backup-agent/errorlog"
	"github.com/evoila/osb-backup-agent/httpBodies"
	"github.com/evoila/osb-backup-agent/s3"
	"github.com/evoila/osb-backup-agent/security"
	"github.com/evoila/osb-backup-agent/shell"
)

// NamePreBackupLock : Name of the script to call for the pre-backup-lock stage
const NamePreBackupLock = "pre-backup-lock"

// NamePreBackupCheck : Name of the script to call for the pre-backup-check stage
const NamePreBackupCheck = "pre-backup-check"

// NameBackup : Name of the script to call for the backup stage
const NameBackup = "backup"

// NameBackupCleanup :  Name of the script to call for the backup-cleanup stage
const NameBackupCleanup = "backup-cleanup"

// NamePostBackupUnlock : Name of the script to call for the post-backup-unlock stage
const NamePostBackupUnlock = "post-backup-unlock"

// BackupRequest : Request handler for backup requests.
func BackupRequest(w http.ResponseWriter, r *http.Request) {
	log.Println("Backup request received.")

	if !security.BasicAuth(w, r) {
		return
	}

	decoder := json.NewDecoder(r.Body)
	var body httpBodies.BackupBody
	err := decoder.Decode(&body)

	missingFields := !httpBodies.CheckForMissingFieldsInBackupBody(body)
	if err != nil || missingFields {
		if err == nil {
			err = errors.New("Body is missing essential fields.")
		}
		errorlog.LogError("Backup failed during body deserialization due to '", err.Error(), "'")
		var response = httpBodies.ErrorResponse{Message: "Backup failed.", State: "Body Deserialization", ErrorMessage: err.Error()}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(400)
		json.NewEncoder(w).Encode(response)
		return
	}
	log.Println("Database", body.Backup.Database, "is supposed to get a new backup.")
	httpBodies.PrintOutBackupBody(body)

	var status = true
	var state, filename string

	var envParameters = httpBodies.GetParametersAsEnvVarStringSlice(body.Backup.Parameters)

	if status {
		state = NamePreBackupLock
		log.Println("> Starting", state, "stage.")
		status, err = shell.ExecuteScriptForStage(NamePreBackupLock, envParameters)
		log.Println("> Finishing", state, "stage.")
	}
	if status {
		state = NamePreBackupCheck
		log.Println("> Starting", state, "stage.")
		status, err = shell.ExecuteScriptForStage(NamePreBackupCheck, envParameters)
		log.Println("> Finishing", state, "stage.")
	}
	if status {
		state = NameBackup
		log.Println("> Starting", state, "stage.")
		var path = GetBackupPath(body.Backup.Host, body.Backup.Database)
		status, err = shell.ExecuteScriptForStage(NameBackup, envParameters,
			body.Backup.Host, body.Backup.Username, body.Backup.Password, body.Backup.Database, path)
		if err == nil {

			if body.Destination.Type == "S3" {
				filename, err = uploadToS3(body)
				if err != nil {
					status = false
					log.Println("[ERROR] Uploading to S3 failed due to '", err.Error(), "'")
				}
			} else {
				status = false
				err = errors.New("type is not supported")
			}
		}
		log.Println("> Finishing", state, "stage.")
	}
	if status {
		state = NameBackupCleanup
		log.Println("> Starting", state, "stage.")
		status, err = shell.ExecuteScriptForStage(NameBackupCleanup, envParameters)
		log.Println("> Finishing", state, "stage.")
	}
	if status {
		state = NamePostBackupUnlock
		log.Println("> Starting", state, "stage.")
		status, err = shell.ExecuteScriptForStage(NamePostBackupUnlock, envParameters)
		log.Println("> Finishing", state, "stage.")
	}

	w.Header().Set("Content-Type", "application/json")
	if status {
		log.Println("Backup successfully created")
		var response = &httpBodies.BackupResponse{Message: "backup successfully created",
			FileName: filename, Region: body.Destination.Region, Bucket: body.Destination.Bucket}
		json.NewEncoder(w).Encode(response)
	} else {
		var errorMessage = "Unknown error"
		if err != nil {
			errorMessage = err.Error()
		}
		errorlog.LogError("Backup failed due to '", errorMessage, "'")
		var response = httpBodies.ErrorResponse{Message: "backup failed.", State: state, ErrorMessage: errorMessage}
		w.WriteHeader(500)
		json.NewEncoder(w).Encode(response)

	}
	log.Println("Finished backup request.")
}

func uploadToS3(body httpBodies.BackupBody) (string, error) {
	var fileName = GetBackupFilename(body.Backup.Host, body.Backup.Database)
	var backupDirectory = configuration.GetBackupDirectory()
	var path = GetBackupPath(body.Backup.Host, body.Backup.Database)
	if !shell.CheckForExistingFile(backupDirectory, fileName) {
		return "", errorlog.LogError("File not found: ", path)
	}
	log.Println("Using file at", path)

	err := s3.UploadFile(fileName, path, body)

	return fileName, err
}

func GetBackupPath(host, database string) string {
	var backupDirectory = configuration.GetBackupDirectory()
	var fileName = GetBackupFilename(host, database)
	var path = errorlog.Concat([]string{backupDirectory, "/", fileName}, "")
	return path
}

func GetBackupFilename(host, database string) string {
	// using the UTC of the local machine!!!
	currentTime := time.Now().UTC()
	log.Printf("Getting filename by current UTC as is %v-%v-%02vT%v:%v:%v+00:00\n", currentTime.Year(), int(currentTime.Month()), currentTime.Day(), currentTime.Hour(), currentTime.Minute(), currentTime.Second())
	return fmt.Sprintf("%s_%v_%v_%02v_%s%s", host, currentTime.Year(), int(currentTime.Month()), currentTime.Day(), database, ".tar.gz")
}
