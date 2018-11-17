package restore

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/evoila/osb-backup-agent/configuration"
	"github.com/evoila/osb-backup-agent/errorlog"
	"github.com/evoila/osb-backup-agent/httpBodies"
	"github.com/evoila/osb-backup-agent/jobs"
	"github.com/evoila/osb-backup-agent/s3"
	"github.com/evoila/osb-backup-agent/security"
	"github.com/evoila/osb-backup-agent/shell"
	"github.com/evoila/osb-backup-agent/swift"
	"github.com/evoila/osb-backup-agent/timeutil"
	"github.com/evoila/osb-backup-agent/utils"
	"github.com/gorilla/mux"
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
	if err != nil || utils.IsIdEmptyInRestoreBodyWithResponse(w, r, body) {
		return
	}

	if jobs.RemoveRestoreJob(body.Id) {
		w.WriteHeader(200)
	} else {
		w.WriteHeader(410)
	}

	log.Println("Restore job deletion request completed.")
}

func HandlePolling(w http.ResponseWriter, r *http.Request) {
	log.Println("-- Restore status request received. --")

	if !security.BasicAuth(w, r) {
		return
	}

	vars := mux.Vars(r)

	Id, exists := vars["id"]
	if !exists {
		w.WriteHeader(400)
		return
	}

	job, existingJob := jobs.GetRestoreJob(Id)
	if !existingJob {
		w.WriteHeader(404)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(200)
	json.NewEncoder(w).Encode(job)
	log.Println("-- Restore status request completed. --")
}

func HandleAsyncRequest(w http.ResponseWriter, r *http.Request) {
	log.Println("-- Async Restore request received. --")

	if !security.BasicAuth(w, r) {
		return
	}

	body, err := utils.UnmarshallIntoRestoreBody(w, r)
	if err != nil || utils.IsIdEmptyInRestoreBodyWithResponse(w, r, body) {
		return
	}

	job, exists := jobs.GetRestoreJob(body.Id)
	if exists {
		log.Println("Job does exist -> showing current result.")
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(409)
		json.NewEncoder(w).Encode(job)
	} else {
		// No job exists yet -> create new one
		log.Println("Job does not exist yet -> creating a new one.")

		if !utils.IsSupportedType(w, r, body.Destination, "Restore") {
			return
		}

		missingFields := !httpBodies.CheckForMissingFieldsInRestoreBody(body)
		if missingFields {
			err = errors.New("body is missing essential fields")
			errorlog.LogError("Restore failed during body deserialization due to '", err.Error(), "'")
			var response = httpBodies.RestoreResponse{Status: httpBodies.Status_failed, Message: "Restore failed.", State: "Body Deserialization", ErrorMessage: err.Error()}

			jobs.AddNewRestoreJob(body.Id)
			jobs.UpdateRestoreJob(body.Id, &response)

			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(400)
			json.NewEncoder(w).Encode(response)
			return
		}

		if !utils.IsAllowedToSpawnNewJob(w, r) {
			return
		}

		job, err := jobs.AddNewRestoreJob(body.Id)
		if err != nil {
			errorlog.LogError("Creating a new job failed due to '", err.Error(), "'")
			var response = httpBodies.RestoreResponse{Status: httpBodies.Status_failed, Message: "Restore failed.", State: "Job creation", ErrorMessage: err.Error(),
				StartTime: "", EndTime: "", ExecutionTime: 0,
			}
			jobs.DecreaseCurrentJobCount()
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(409)
			json.NewEncoder(w).Encode(response)
			return
		}

		// Starting new go routine to handle the restore request
		log.Println("Starting new go routine to handle restore request for", body.Id)
		go Restore(body, job)

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(201)
	}
	log.Println("-- Restore request completed. --")
}

func Restore(body httpBodies.RestoreBody, job *httpBodies.RestoreResponse) *httpBodies.RestoreResponse {

	log.Println("Database", body.Restore.Database, "is supposed to get a restore.")
	httpBodies.PrintOutRestoreBody(body)

	response, _ := jobs.GetRestoreJob(body.Id)
	response.Message = "restore is running"
	response.Status = httpBodies.Status_running
	response.Type = body.Destination.Type
	response.Compression = body.Compression
	jobs.UpdateRestoreJob(body.Id, response)

	// Set up variables for filling response bodies later on
	var err error

	// Get environment parameters from request body
	var envParameters = httpBodies.GetParametersAsEnvVarStringSlice(body.Restore.Parameters)

	// Set start time
	currentTime := time.Now()
	executionTime := currentTime.UnixNano()
	startTime := timeutil.GetTimestamp(&currentTime)

	response.StartTime = startTime
	jobs.UpdateRestoreJob(body.Id, response)

	var status = true
	if status {
		response.State = NamePreRestoreLock
		jobs.UpdateRestoreJob(body.Id, response)

		log.Println("> Starting", response.State, "stage.")
		status, response.PreRestoreLockLog, response.PreRestoreLockErrorLog, err = shell.ExecuteScriptForStage(NamePreRestoreLock, envParameters, body.Id)
		jobs.UpdateRestoreJob(body.Id, response)
		log.Println("> Finishing", response.State, "stage.")
	}
	if status {
		response.State = NameRestore
		jobs.UpdateRestoreJob(body.Id, response)

		log.Println("> Starting", response.State, "stage.")

		if body.Destination.Type == "S3" {
			err = download(body, body.Destination.Type)
		} else if body.Destination.Type == "SWIFT" {
			err = download(body, body.Destination.Type)
		} else {
			status = false
			err = errors.New("type is not supported")
		}

		if err != nil {
			status = false
			err = errorlog.LogError("Downloading from "+body.Destination.Type+" failed due to '", err.Error(), "'")
		} else {
			status, response.RestoreLog, response.RestoreErrorLog, err = shell.ExecuteScriptForStage(NameRestore, envParameters,
				body.Restore.Host, body.Restore.Username, body.Restore.Password, body.Restore.Database,
				body.Destination.Filename, body.Id, strconv.FormatBool(body.Compression), body.Encryption_key)
			jobs.UpdateRestoreJob(body.Id, response)
		}

		log.Println("> Finishing", response.State, "stage.")
	}
	if status {
		response.State = NameRestoreCleanup
		jobs.UpdateRestoreJob(body.Id, response)

		log.Println("> Starting", response.State, "stage.")
		status, response.RestoreCleanupLog, response.RestoreCleanupErrorLog, err = shell.ExecuteScriptForStage(NameRestoreCleanup, envParameters, body.Id)
		jobs.UpdateRestoreJob(body.Id, response)
		log.Println("> Finishing", response.State, "stage.")
	}
	if status {
		response.State = NamePostRestoreUnlock
		jobs.UpdateRestoreJob(body.Id, response)

		log.Println("> Starting", response.State, "stage.")
		status, response.PostRestoreUnlockLog, response.PostRestoreUnlockErrorLog, err = shell.ExecuteScriptForStage(NamePostRestoreUnlock, envParameters)
		jobs.UpdateRestoreJob(body.Id, response)
		log.Println("> Finishing", response.State, "stage.")
	}

	// Set end time and calculate execution time
	currentTime = time.Now()
	executionTime = (currentTime.UnixNano() - executionTime) / 1000 / 1000 //convert from ns to ms
	endTime := timeutil.GetTimestamp(&currentTime)

	response.ExecutionTime = executionTime
	response.EndTime = endTime
	response.State = "finished"
	jobs.UpdateRestoreJob(body.Id, response)

	// Write standard or error response according to status
	if status {
		response.Status = httpBodies.Status_success
		response.Message = "restore successfully carried out"

		log.Println("Restore successfully carried out")

		log.Println("Updating restore job", body.Id, "with an response.")
		jobs.UpdateRestoreJob(body.Id, response)
	} else {
		var errorMessage = "Unknown error"
		if err != nil {
			errorMessage = err.Error()
		}
		err = errorlog.LogError("Restore failed due to '", errorMessage, "'")

		response.Status = httpBodies.Status_failed
		response.Message = "restore failed"
		response.ErrorMessage = err.Error()

		log.Println("Restore incompletely carried out")

		log.Println("Updating restore job", body.Id, "with an error response.")
		jobs.UpdateRestoreJob(body.Id, response)
	}
	jobs.DecreaseCurrentJobCount()
	log.Println("Finished restore for", body.Id)
	return response

}

func download(body httpBodies.RestoreBody, downloadType string) error {
	var restoreDirectory = configuration.GetRestoreDirectory() + "/" + body.Id
	var path = errorlog.Concat([]string{restoreDirectory, "/", body.Destination.Filename}, "")
	var err error
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

	if downloadType == "S3" {
		log.Println("Using S3 as destination.")
		err = s3.DownloadFile(body.Destination.Filename, path, body)
	} else {
		log.Println("Using swift as destination.")
		err = swift.DownloadFile(body.Destination.Filename, path, body)
	}

	return err
}
