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
	"github.com/evoila/osb-backup-agent/jobs"
	"github.com/evoila/osb-backup-agent/s3"
	"github.com/evoila/osb-backup-agent/security"
	"github.com/evoila/osb-backup-agent/shell"
	"github.com/evoila/osb-backup-agent/timeutil"
	"github.com/evoila/osb-backup-agent/utils"
	"github.com/gorilla/mux"
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

func RemoveJob(w http.ResponseWriter, r *http.Request) {
	log.Println("-- Backup job deletion request received. --")

	if !security.BasicAuth(w, r) {
		return
	}

	body, err := utils.UnmarshallIntoBackupBody(w, r)
	if err != nil || utils.IsIdEmptyInBackupBodyWithResponse(w, r, body) {
		return
	}

	if jobs.RemoveBackupJob(body.Id) {
		w.WriteHeader(200)
	} else {
		w.WriteHeader(410)
	}

	log.Println("-- Backup job deletion request completed. --")
}

func HandlePolling(w http.ResponseWriter, r *http.Request) {
	log.Println("-- Backup status request received. --")

	if !security.BasicAuth(w, r) {
		return
	}

	vars := mux.Vars(r)

	Id, exists := vars["id"]
	if !exists {
		w.WriteHeader(400)
		return
	}

	job, existingJob := jobs.GetBackupJob(Id)
	if !existingJob {
		w.WriteHeader(404)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(200)
	json.NewEncoder(w).Encode(job)
	log.Println("-- Backup status request completed. --")
}

func HandleAsyncRequest(w http.ResponseWriter, r *http.Request) {
	log.Println("-- Async Backup request received. --")

	if !security.BasicAuth(w, r) {
		return
	}

	body, err := utils.UnmarshallIntoBackupBody(w, r)
	if err != nil {
		return
	}

	if utils.IsIdEmptyInBackupBodyWithResponse(w, r, body) {

		return
	}

	job, exists := jobs.GetBackupJob(body.Id)
	if exists {
		log.Println("Job does exist -> showing current result.")
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(409)
		json.NewEncoder(w).Encode(job)
	} else {
		// No job exists yet -> create new one
		log.Println("Job does not exist yet -> creating a new one.")

		missingFields := !httpBodies.CheckForMissingFieldsInBackupBody(body)
		if missingFields {
			err = errors.New("body is missing essential fields")
			errorlog.LogError("Backup failed during body deserialization due to '", err.Error(), "'")
			var response = httpBodies.BackupResponse{Status: httpBodies.Status_failed, Message: "Backup failed.", State: "Body Deserialization", ErrorMessage: err.Error()}
			jobs.AddNewBackupJob(body.Id)
			jobs.UpdateBackupJob(body.Id, &response)

			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(400)
			json.NewEncoder(w).Encode(response)
			return
		}

		job, err := jobs.AddNewBackupJob(body.Id)
		if err != nil {
			errorlog.LogError("Creating a new job failed due to '", err.Error(), "'")
			var response = httpBodies.BackupResponse{Status: httpBodies.Status_failed, Message: "Backup failed.", State: "Job creation", ErrorMessage: err.Error(),
				StartTime: "", EndTime: "", ExecutionTime: 0,
			}
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(409)
			json.NewEncoder(w).Encode(response)
			return
		}

		// Starting new go routine to handle the backup request
		log.Println("Starting new go routine to handle backup request for", body.Id)
		go Backup(body, job)

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(201)
	}
	log.Println("-- Backup request completed. --")
}

func Backup(body httpBodies.BackupBody, job *httpBodies.BackupResponse) *httpBodies.BackupResponse {

	log.Println("Database", body.Backup.Database, "is supposed to get a new backup.")
	httpBodies.PrintOutBackupBody(body)

	response, _ := jobs.GetBackupJob(body.Id)
	response.Message = "backup is running"
	response.Status = httpBodies.Status_running
	response.Bucket = body.Destination.Bucket
	response.Region = body.Destination.Region
	jobs.UpdateBackupJob(body.Id, response)

	// Set up variables for filling response bodies later on
	var fileSize int64
	var err error

	// Get environment parameters from request body
	var envParameters = httpBodies.GetParametersAsEnvVarStringSlice(body.Backup.Parameters)

	// Set start time
	currentTime := time.Now()
	executionTime := currentTime.UnixNano()
	startTime := timeutil.GetTimestamp(&currentTime)

	response.StartTime = startTime
	jobs.UpdateBackupJob(body.Id, response)

	// Start execution of scripts
	var status = true
	if status {
		response.State = NamePreBackupLock
		jobs.UpdateBackupJob(body.Id, response)

		log.Println("> Starting", response.State, "stage.")
		status, response.PreBackupLockLog, response.PreBackupLockErrorLog, err = shell.ExecuteScriptForStage(NamePreBackupLock, envParameters)
		jobs.UpdateBackupJob(body.Id, response)
		log.Println("> Finishing", response.State, "stage.")
	}
	if status {
		response.State = NamePreBackupCheck
		jobs.UpdateBackupJob(body.Id, response)

		log.Println("> Starting", response.State, "stage.")
		status, response.PreBackupCheckLog, response.PreBackupCheckErrorLog, err = shell.ExecuteScriptForStage(NamePreBackupCheck, envParameters)
		jobs.UpdateBackupJob(body.Id, response)
		log.Println("> Finishing", response.State, "stage.")
	}
	if status {
		response.State = NameBackup
		jobs.UpdateBackupJob(body.Id, response)

		log.Println("> Starting", response.State, "stage.")
		var path = GetBackupPath(body.Backup.Host, body.Backup.Database)
		status, response.BackupLog, response.BackupErrorLog, err = shell.ExecuteScriptForStage(NameBackup, envParameters,
			body.Backup.Host, body.Backup.Username, body.Backup.Password, body.Backup.Database, path)
		jobs.UpdateBackupJob(body.Id, response)
		if err == nil {

			if body.Destination.Type == "S3" {
				response.FileName, fileSize, err = uploadToS3(body)
				if err != nil {
					status = false
					log.Println("[ERROR] Uploading to S3 failed due to '", err.Error(), "'")
				}
				response.FileSize = httpBodies.FileSize{Size: fileSize, Unit: "byte"}
				jobs.UpdateBackupJob(body.Id, response)
			} else {
				status = false
				err = errors.New("type is not supported")
			}
		}
		log.Println("> Finishing", response.State, "stage.")
	}
	if status {
		response.State = NameBackupCleanup
		jobs.UpdateBackupJob(body.Id, response)

		log.Println("> Starting", response.State, "stage.")
		status, response.BackupCleanupLog, response.BackupCleanupErrorLog, err = shell.ExecuteScriptForStage(NameBackupCleanup, envParameters)
		jobs.UpdateBackupJob(body.Id, response)
		log.Println("> Finishing", response.State, "stage.")
	}
	if status {
		response.State = NamePostBackupUnlock
		jobs.UpdateBackupJob(body.Id, response)

		log.Println("> Starting", response.State, "stage.")
		status, response.PostBackupUnlockLog, response.PostBackupUnlockErrorLog, err = shell.ExecuteScriptForStage(NamePostBackupUnlock, envParameters)
		jobs.UpdateBackupJob(body.Id, response)
		log.Println("> Finishing", response.State, "stage.")
	}

	// Set end time and calculate execution time
	currentTime = time.Now()
	executionTime = timeutil.GetTimeDifferenceInMilliseconds(executionTime, currentTime.UnixNano())
	endTime := timeutil.GetTimestamp(&currentTime)

	response.ExecutionTime = executionTime
	response.EndTime = endTime
	response.State = "finished"
	jobs.UpdateBackupJob(body.Id, response)

	// Write standard or error response according to status
	if status {
		response.Status = httpBodies.Status_success
		response.Message = "backup successfully carried out"
		log.Println("Backup successfully created")

		log.Println("Updating backup job", body.Id, "with an response.")
		jobs.UpdateBackupJob(body.Id, response)
	} else {
		var errorMessage = "Unknown error"
		if err != nil {
			errorMessage = err.Error()
		}
		errorlog.LogError("Backup failed due to '", errorMessage, "'")

		response.Status = httpBodies.Status_failed
		response.Message = "restore failed"
		response.ErrorMessage = errorMessage

		log.Println("Restore incompletely carried out")

		log.Println("Updating backup job", body.Id, "with an error response.")
		jobs.UpdateBackupJob(body.Id, response)

	}
	log.Println("Finished backup for", body.Id)
	return response
}

func uploadToS3(body httpBodies.BackupBody) (string, int64, error) {
	var fileName = GetBackupFilename(body.Backup.Host, body.Backup.Database)
	var backupDirectory = configuration.GetBackupDirectory()
	var path = GetBackupPath(body.Backup.Host, body.Backup.Database)
	if !shell.CheckForExistingFile(backupDirectory, fileName) {
		return "", 0, errorlog.LogError("File not found: ", path)
	}
	log.Println("Using file at", path)
	size, err := shell.GetFileSize(path)
	if err != nil {
		return fileName, 0, errorlog.LogError("Reading file size failed due to '", err.Error(), "'")
	}
	err = s3.UploadFile(fileName, path, body)

	return fileName, size, err
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
