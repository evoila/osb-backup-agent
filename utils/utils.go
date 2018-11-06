package utils

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/evoila/osb-backup-agent/errorlog"
	"github.com/evoila/osb-backup-agent/httpBodies"
)

func UnmarshallIntoBackupBody(w http.ResponseWriter, r *http.Request) (httpBodies.BackupBody, error) {
	decoder := json.NewDecoder(r.Body)
	var body httpBodies.BackupBody
	err := decoder.Decode(&body)

	if err != nil || body.Id == "" {
		if err == nil {
			err = errors.New("id is empty")
		}
		errorlog.LogError("Backup failed during body deserialization due to '", err.Error(), "'")
		var response = httpBodies.BackupResponse{Status: httpBodies.Status_failed, Message: "Backup failed.", State: "Body Deserialization", ErrorMessage: err.Error(),
			StartTime: "", EndTime: "", ExecutionTime: 0,
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(400)
		json.NewEncoder(w).Encode(response)
		return body, err
	}
	return body, nil
}

func UnmarshallIntoRestoreBody(w http.ResponseWriter, r *http.Request) (httpBodies.RestoreBody, error) {
	decoder := json.NewDecoder(r.Body)
	var body httpBodies.RestoreBody
	err := decoder.Decode(&body)

	if err != nil || body.Id == "" {
		if err == nil {
			err = errors.New("id is empty")
		}
		errorlog.LogError("Restore failed during body deserialization due to '", err.Error(), "'")
		var response = httpBodies.RestoreResponse{Status: httpBodies.Status_failed, Message: "Restore failed.", State: "Body Deserialization", ErrorMessage: err.Error(),
			StartTime: "", EndTime: "", ExecutionTime: 0,
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(400)
		json.NewEncoder(w).Encode(response)
		return body, err
	}
	return body, nil
}

func IsIdEmptyInBackupBodyWithResponse(w http.ResponseWriter, r *http.Request, body httpBodies.BackupBody) bool {
	if body.Id == "" {
		err := errorlog.LogError("Backup failed during body deserialization due to '", "id is empty", "'")
		var response = httpBodies.BackupResponse{Status: httpBodies.Status_failed, Message: "Backup failed.", State: "Body Deserialization", ErrorMessage: err.Error(),
			StartTime: "", EndTime: "", ExecutionTime: 0,
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(400)
		json.NewEncoder(w).Encode(response)
		return true
	}
	return false
}

func IsIdEmptyInRestoreBodyWithResponse(w http.ResponseWriter, r *http.Request, body httpBodies.RestoreBody) bool {
	if body.Id == "" {
		err := errorlog.LogError("Restore failed during body deserialization due to '", "id is empty", "'")
		var response = httpBodies.RestoreResponse{Status: httpBodies.Status_failed, Message: "Restore failed.", State: "Body Deserialization", ErrorMessage: err.Error(),
			StartTime: "", EndTime: "", ExecutionTime: 0,
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(400)
		json.NewEncoder(w).Encode(response)
		return true
	}
	return false
}
