package restore

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"os"
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
)

const NamePreRestoreLock = "pre-restore-lock"
const NameRestore = "restore"
const NameRestoreCleanup = "restore-cleanup"
const NamePostRestoreUnlock = "post-restore-unlock"

func RemoveJob(w http.ResponseWriter, r *http.Request) {
	log.Println("Restore job deletion request received.")
	if !security.BasicAuth(w, r) {
		return
	}

	body, err := utils.UnmarshallIntoRestoreBody(w, r)
	if err != nil || utils.IsUUIDEmptyInRestoreBodyWithResponse(w, r, body) {
		return
	}

	if jobs.RemoveRestoreJob(body.UUID) {
		w.WriteHeader(200)
	} else {
		w.WriteHeader(410)
	}

	log.Println("Restore job deletion request completed.")
}

func HandleAsyncRequest(w http.ResponseWriter, r *http.Request) {
	log.Println("Async Restore request received.")

	if !security.BasicAuth(w, r) {
		return
	}

	body, err := utils.UnmarshallIntoRestoreBody(w, r)
	if err != nil || utils.IsUUIDEmptyInRestoreBodyWithResponse(w, r, body) {
		return
	}

	job, exists := jobs.GetRestoreJob(body.UUID)
	if exists {
		log.Println("Job does exist -> showing current result.")
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(200)
		json.NewEncoder(w).Encode(job)
	} else {
		// No job exists yet -> create new one
		log.Println("Job does not exist yet -> creating a new one.")

		missingFields := !httpBodies.CheckForMissingFieldsInRestoreBody(body)
		if missingFields {
			err = errors.New("body is missing essential fields")
			errorlog.LogError("Restore failed during body deserialization due to '", err.Error(), "'")
			var response = httpBodies.RestoreResponse{Status: httpBodies.Status_failed, Message: "Restore failed.", State: "Body Deserialization", ErrorMessage: err.Error(),
				StartTime: "", EndTime: "", ExecutionTime: 0,
			}
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(400)
			json.NewEncoder(w).Encode(response)
			return
		}

		job, err := jobs.AddNewRestoreJob(body.UUID)
		if err != nil {
			errorlog.LogError("Creating a new job failed due to '", err.Error(), "'")
			var response = httpBodies.RestoreResponse{Status: httpBodies.Status_failed, Message: "Restore failed.", State: "Job creation", ErrorMessage: err.Error(),
				StartTime: "", EndTime: "", ExecutionTime: 0,
			}
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(409)
			json.NewEncoder(w).Encode(response)
			return
		}

		// Starting new go routine to handle the backup request
		log.Println("Starting new go routine to handle restore request for", body.UUID)
		go RestoreRequest(body, job)

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(201)
	}
	log.Println("Restore request completed.")
}

func RestoreRequest(body httpBodies.RestoreBody, job *httpBodies.RestoreResponse) *httpBodies.RestoreResponse {

	log.Println("Database", body.Restore.Database, "is supposed to get a restore.")
	httpBodies.PrintOutRestoreBody(body)

	response, _ := jobs.GetRestoreJob(body.UUID)

	// Set up variables for filling response bodies later on
	var state string
	outputStatus := httpBodies.Status_failed
	preRestoreLockLog, restoreLog, restoreCleanupLog, postRestoreUnlockLog := "", "", "", ""
	preRestoreLockErrorLog, restoreErrorLog, restoreCleanupErrorLog, postRestoreUnlockErrorLog := "", "", "", ""
	var err error

	// Get environment parameters from request body
	var envParameters = httpBodies.GetParametersAsEnvVarStringSlice(body.Restore.Parameters)

	// Set start time
	currentTime := time.Now()
	executionTime := currentTime.UnixNano()
	startTime := timeutil.GetTimestamp(&currentTime)

	var status = true
	if status {
		state = NamePreRestoreLock
		log.Println("> Starting", state, "stage.")
		status, preRestoreLockLog, preRestoreLockErrorLog, err = shell.ExecuteScriptForStage(NamePreRestoreLock, envParameters)
		log.Println("> Finishing", state, "stage.")
	}
	if status {
		state = NameRestore
		log.Println("> Starting", state, "stage.")

		if body.Destination.Type == "S3" {
			err = downloadFromS3(body)
			if err != nil {
				status = false
				log.Println("[ERROR] Downloading from S3 failed due to '", err.Error(), "'")
			} else {
				status, restoreLog, restoreErrorLog, err = shell.ExecuteScriptForStage(NameRestore, envParameters,
					body.Restore.Host, body.Restore.Username, body.Restore.Password, body.Restore.Database, configuration.GetRestoreDirectory(), body.Destination.Filename)
			}
		} else {
			status = false
			err = errors.New("type is not supported")
		}

		log.Println("> Finishing", state, "stage.")
	}
	if status {
		state = NameRestoreCleanup
		log.Println("> Starting", state, "stage.")
		status, restoreCleanupLog, restoreCleanupErrorLog, err = shell.ExecuteScriptForStage(NameRestoreCleanup, envParameters)
		log.Println("> Finishing", state, "stage.")
	}
	if status {
		state = NamePostRestoreUnlock
		log.Println("> Starting", state, "stage.")
		status, postRestoreUnlockLog, postRestoreUnlockErrorLog, err = shell.ExecuteScriptForStage(NamePostRestoreUnlock, envParameters)
		log.Println("> Finishing", state, "stage.")
	}

	// Set end time and calculate execution time
	currentTime = time.Now()
	executionTime = (currentTime.UnixNano() - executionTime) / 1000 / 1000 //convert from ns to ms
	endTime := timeutil.GetTimestamp(&currentTime)

	// Write standard or error response according to status
	if status {
		state = "finished"
		outputStatus = httpBodies.Status_success
		log.Println("Restore successfully carried out")

		response = &httpBodies.RestoreResponse{Status: outputStatus, Message: "restore successfully carried out",
			StartTime: startTime, EndTime: endTime, ExecutionTime: executionTime,
			// Logs
			PreRestoreLockLog: preRestoreLockLog, RestoreLog: restoreLog, RestoreCleanupLog: restoreCleanupLog, PostRestoreUnlockLog: postRestoreUnlockLog,
			// Error logs
			PreRestoreLockErrorLog: preRestoreLockErrorLog, RestoreErrorLog: restoreErrorLog, RestoreCleanupErrorLog: restoreCleanupErrorLog, PostRestoreUnlockErrorLog: postRestoreUnlockErrorLog,
		}
		log.Println("Updating restore job", body.UUID, "with an response.")
		jobs.UpdateRestoreJob(body.UUID, response)
	} else {
		var errorMessage = "Unknown error"
		if err != nil {
			errorMessage = err.Error()
		}
		errorlog.LogError("Restore failed due to '", errorMessage, "'")
		response = &httpBodies.RestoreResponse{
			Status: outputStatus, Message: "restore failed", State: state, ErrorMessage: errorMessage,
			// Logs
			PreRestoreLockLog: preRestoreLockLog, RestoreLog: restoreLog, RestoreCleanupLog: restoreCleanupLog, PostRestoreUnlockLog: postRestoreUnlockLog,
			// Error logs
			PreRestoreLockErrorLog: preRestoreLockErrorLog, RestoreErrorLog: restoreErrorLog, RestoreCleanupErrorLog: restoreCleanupErrorLog, PostRestoreUnlockErrorLog: postRestoreUnlockErrorLog,
		}
		log.Println("Updating restore job", body.UUID, "with an error response.")
		jobs.UpdateRestoreJob(body.UUID, response)
	}
	log.Println("Finished restore for", body.UUID)
	return response

}

func downloadFromS3(body httpBodies.RestoreBody) error {
	var restoreDirectory = configuration.GetRestoreDirectory()
	var path = errorlog.Concat([]string{restoreDirectory, "/", body.Destination.Filename}, "")
	if shell.CheckForExistingFile(restoreDirectory, body.Destination.Filename) {
		if configuration.IsAllowedToDeleteFiles() {
			log.Println("Removing already existing file:", path)
			err := os.Remove(path)

			if err != nil {
				return errorlog.LogError(err.Error())
			}
		} else {
			return errorlog.LogError("File already exists: ", path)
		}
	}
	log.Println("Using file at", path)

	err := s3.DownloadFile(body.Destination.Filename, path, body)

	return err
}
