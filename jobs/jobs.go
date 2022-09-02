package jobs

import (
	"log"

	"github.com/evoila/osb-backup-agent/configuration"
	"github.com/evoila/osb-backup-agent/errorlog"
	"github.com/evoila/osb-backup-agent/httpBodies"
	"github.com/evoila/osb-backup-agent/mutex"
)

var currentJobCount int
var jobCountMutex mutex.Mutex

var backupJobs map[string]*httpBodies.BackupResponse
var backupMutex mutex.Mutex

var restoreJobs map[string]*httpBodies.RestoreResponse
var restoreMutex mutex.Mutex

func SetUpJobStructure() {
	currentJobCount = 0
	backupJobs = make(map[string]*httpBodies.BackupResponse)
	restoreJobs = make(map[string]*httpBodies.RestoreResponse)
	jobCountMutex = make(mutex.Mutex, 1)
	backupMutex = make(mutex.Mutex, 1)
	restoreMutex = make(mutex.Mutex, 1)
	jobCountMutex.Release()
	backupMutex.Release()
	restoreMutex.Release()
}

func IncreaseCurrentJobCountWithCheck() bool {
	log.Println("Accessing job number mutex to increase job count.")
	jobCountMutex.Acquire()
	isAllowedToStart := currentJobCount < configuration.GetMaxJobNumber()
	if isAllowedToStart {
		log.Println("Current job count is", currentJobCount, "-> reserving a spot")
		currentJobCount++
	} else {
		log.Println("Current job count is", currentJobCount, "-> not allowed to reserve a spot")
	}
	log.Println("Unlocking job number mutex after increasing job count.")
	jobCountMutex.Release()
	return isAllowedToStart
}

func DecreaseCurrentJobCount() {
	log.Println("Accessing job number mutex to decrease job count.")
	jobCountMutex.Acquire()
	currentJobCount--
	log.Println("Unlocking job number mutex after decreasing job count to", currentJobCount)
	jobCountMutex.Release()
}

func GetBackupJob(UUID string) (*httpBodies.BackupResponse, bool) {
	log.Println("Accessing backup mutex for getting a job.")
	backupMutex.Acquire()

	job, existing := backupJobs[UUID]

	log.Println("Unlocking backup mutex after getting a job.")
	backupMutex.Release()

	return job, existing
}

func AddNewBackupJob(UUID string) (*httpBodies.BackupResponse, error) {
	_, exists := GetBackupJob(UUID)
	if exists {
		return nil, errorlog.LogError("backup job with UUID ", UUID, " already exists")
	}

	log.Println("Accessing backup mutex for adding a new job.")
	backupMutex.Acquire()

	newJob := &httpBodies.BackupResponse{Status: "RUNNING"}
	backupJobs[UUID] = newJob

	log.Println("Unlocking backup mutex after adding a new job.")
	backupMutex.Release()

	return newJob, nil
}

func UpdateBackupJob(UUID string, job *httpBodies.BackupResponse) error {
	_, exists := GetBackupJob(UUID)
	if !exists {
		return errorlog.LogError("backup job with UUID ", UUID, " does not exists")
	}

	log.Println("Accessing backup mutex for updating a job.")
	backupMutex.Acquire()

	backupJobs[UUID] = job

	log.Println("Unlocking backup mutex after updating a job.")
	backupMutex.Release()

	return nil
}

func RemoveBackupJob(UUID string) bool {
	if _, exists := GetBackupJob(UUID); !exists {
		return false
	}

	log.Println("Accessing backup mutex for deleting a job.")
	backupMutex.Acquire()

	delete(backupJobs, UUID)

	log.Println("Unlocking backup mutex after deleting a job.")
	backupMutex.Release()

	return true
}

func GetRestoreJob(UUID string) (*httpBodies.RestoreResponse, bool) {
	log.Println("Accessing restore mutex for getting a job.")
	restoreMutex.Acquire()

	job, existing := restoreJobs[UUID]

	log.Println("Unlocking restore mutex after getting a job.")
	restoreMutex.Release()

	return job, existing
}

func AddNewRestoreJob(UUID string) (*httpBodies.RestoreResponse, error) {
	_, exists := GetRestoreJob(UUID)
	if exists {
		return nil, errorlog.LogError("restore job with UUID ", UUID, " already exists")
	}

	log.Println("Accessing restore mutex for adding a new job.")
	restoreMutex.Acquire()

	newJob := &httpBodies.RestoreResponse{Status: "RUNNING"}
	restoreJobs[UUID] = newJob

	log.Println("Unlocking restore mutex after adding a new job.")
	restoreMutex.Release()

	return newJob, nil
}

func UpdateRestoreJob(UUID string, job *httpBodies.RestoreResponse) error {

	_, exists := GetRestoreJob(UUID)
	if !exists {
		return errorlog.LogError("restore job with UUID ", UUID, " does not exists")
	}

	log.Println("Accessing restore mutex for updating a job.")
	restoreMutex.Acquire()

	restoreJobs[UUID] = job

	log.Println("Unlocking restore mutex after updating a job.")
	restoreMutex.Release()
	return nil
}

func RemoveRestoreJob(UUID string) bool {
	if _, exists := GetRestoreJob(UUID); !exists {
		return false
	}

	log.Println("Accessing restore mutex for deleting a job.")
	restoreMutex.Acquire()

	delete(restoreJobs, UUID)

	log.Println("Unlocking restore mutex after deleting a job.")
	restoreMutex.Release()
	return true
}
