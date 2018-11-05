package restore

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"os"

	"github.com/evoila/osb-backup-agent/configuration"
	"github.com/evoila/osb-backup-agent/errorlog"
	"github.com/evoila/osb-backup-agent/httpBodies"
	"github.com/evoila/osb-backup-agent/s3"
	"github.com/evoila/osb-backup-agent/security"
	"github.com/evoila/osb-backup-agent/shell"
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

	decoder := json.NewDecoder(r.Body)
	var body httpBodies.RestoreBody
	err := decoder.Decode(&body)

	missingFields := !httpBodies.CheckForMissingFieldsInRestoreBody(body)
	if err != nil || missingFields {
		if err == nil {
			err = errors.New("body is missing essential fields")
		}
		errorlog.LogError("Restore failed during body deserialization due to '", err.Error(), "'")
		var response = httpBodies.ErrorResponse{Message: "Restore failed.", State: "Body Deserialization", ErrorMessage: err.Error()}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(400)
		json.NewEncoder(w).Encode(response)
		return
	}
	log.Println("Database", body.Restore.Database, "is supposed to get a restore.")
	httpBodies.PrintOutRestoreBody(body)

	var status = true
	var state string

	var envParameters = httpBodies.GetParametersAsEnvVarStringSlice(body.Restore.Parameters)

	if status {
		state = NamePreRestoreLock
		log.Println("> Starting", state, "stage.")
		status, err = shell.ExecuteScriptForStage(NamePreRestoreLock, envParameters)
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
				status, err = shell.ExecuteScriptForStage(NameRestore, envParameters,
					body.Restore.Host, body.Restore.Username, body.Restore.Password, body.Restore.Database, configuration.GetRestoreDirectory(), body.Destination.File)
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
		status, err = shell.ExecuteScriptForStage(NameRestoreCleanup, envParameters)
		log.Println("> Finishing", state, "stage.")
	}
	if status {
		state = NamePostRestoreUnlock
		log.Println("> Starting", state, "stage.")
		status, err = shell.ExecuteScriptForStage(NamePostRestoreUnlock, envParameters)
		log.Println("> Finishing", state, "stage.")
	}

	w.Header().Set("Content-Type", "application/json")
	if status {
		log.Println("Restore successfully carried out")
		var response = &response{Message: "restore successfully carried out."}
		json.NewEncoder(w).Encode(response)
	} else {
		var errorMessage = "Unknown error"
		if err != nil {
			errorMessage = err.Error()
		}
		errorlog.LogError("Restore failed due to '", errorMessage, "'")
		var response = httpBodies.ErrorResponse{"restore failed.", state, errorMessage}
		w.WriteHeader(500)
		json.NewEncoder(w).Encode(response)

	}
	log.Println("Finished restore request.")

}

func downloadFromS3(body httpBodies.RestoreBody) error {
	var restoreDirectory = configuration.GetRestoreDirectory()
	var path = errorlog.Concat([]string{restoreDirectory, "/", body.Destination.File}, "")
	if shell.CheckForExistingFile(restoreDirectory, body.Destination.File) {
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

	err := s3.DownloadFile(body.Destination.File, path, body)

	return err
}
