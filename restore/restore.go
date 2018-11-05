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
	"github.com/evoila/osb-backup-agent/s3"
	"github.com/evoila/osb-backup-agent/security"
	"github.com/evoila/osb-backup-agent/shell"
	"github.com/evoila/osb-backup-agent/timeutil"
)

type response struct {
	Message string
}

const NamePreRestoreLock = "pre-restore-lock"
const NameRestore = "restore"
const NameRestoreCleanup = "restore-cleanup"
const NamePostRestoreUnlock = "post-restore-unlock"

func RestoreRequest(w http.ResponseWriter, r *http.Request) {
	log.Println("Restore request received.")

	if !security.BasicAuth(w, r) {
		return
	}

	// Decode request body and scan for empty fields
	decoder := json.NewDecoder(r.Body)
	var body httpBodies.RestoreBody
	err := decoder.Decode(&body)
	missingFields := !httpBodies.CheckForMissingFieldsInRestoreBody(body)

	// Handle error or missing fields of request body
	if err != nil || missingFields {
		if err == nil {
			err = errors.New("body is missing essential fields")
		}
		errorlog.LogError("Restore failed during body deserialization due to '", err.Error(), "'")
		var response = httpBodies.RestoreErrorResponse{Status: httpBodies.Status_failed, Message: "Restore failed.", State: "Body Deserialization", ErrorMessage: err.Error(),
			StartTime: "", EndTime: "", ExecutionTime: 0,
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(400)
		json.NewEncoder(w).Encode(response)
		return
	}

	log.Println("Database", body.Restore.Database, "is supposed to get a restore.")
	httpBodies.PrintOutRestoreBody(body)

	// Set up variables for filling response bodies later on
	var state string
	outputStatus := httpBodies.Status_failed
	preRestoreLockLog, restoreLog, restoreCleanupLog, postRestoreUnlockLog := "", "", "", ""
	preRestoreLockErrorLog, restoreErrorLog, restoreCleanupErrorLog, postRestoreUnlockErrorLog := "", "", "", ""

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
	w.Header().Set("Content-Type", "application/json")
	if status {
		state = "finished"
		outputStatus = httpBodies.Status_success
		log.Println("Restore successfully carried out")
		var response = &httpBodies.RestoreResponse{Status: outputStatus, Message: "restore successfully carried out",
			StartTime: startTime, EndTime: endTime, ExecutionTime: executionTime,
			// Logs
			PreRestoreLockLog: preRestoreLockLog, RestoreLog: restoreLog, RestoreCleanupLog: restoreCleanupLog, PostRestoreUnlockLog: postRestoreUnlockLog,
			// Error logs
			PreRestoreLockErrorLog: preRestoreLockErrorLog, RestoreErrorLog: restoreErrorLog, RestoreCleanupErrorLog: restoreCleanupErrorLog, PostRestoreUnlockErrorLog: postRestoreUnlockErrorLog,
		}
		json.NewEncoder(w).Encode(response)
	} else {
		var errorMessage = "Unknown error"
		if err != nil {
			errorMessage = err.Error()
		}
		errorlog.LogError("Restore failed due to '", errorMessage, "'")
		var response = httpBodies.RestoreErrorResponse{
			Status: outputStatus, Message: "restore failed", State: state, ErrorMessage: errorMessage,
			// Logs
			PreRestoreLockLog: preRestoreLockLog, RestoreLog: restoreLog, RestoreCleanupLog: restoreCleanupLog, PostRestoreUnlockLog: postRestoreUnlockLog,
			// Error logs
			PreRestoreLockErrorLog: preRestoreLockErrorLog, RestoreErrorLog: restoreErrorLog, RestoreCleanupErrorLog: restoreCleanupErrorLog, PostRestoreUnlockErrorLog: postRestoreUnlockErrorLog,
		}
		w.WriteHeader(500)
		json.NewEncoder(w).Encode(response)

	}
	log.Println("Finished restore request.")

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
