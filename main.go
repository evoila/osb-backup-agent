package main

import (
	"log"
	"os"

	"github.com/evoila/osb-backup-agent/configuration"
	"github.com/evoila/osb-backup-agent/errorlog"
	"github.com/evoila/osb-backup-agent/webclient"
)

func main() {

	// Default logs in Golang are written onto stderr -> changing it here at the very start
	log.SetOutput(os.Stdout)
	// errorlog package uses its own log.Logger
	errorlog.InitErrorLog()

	log.Println("Go-backup-agent is starting...")
	PrintOutConfig()
	webclient.StartWebAgent()
}

func PrintOutConfig() {
	var username = configuration.GetUsername()
	var pw = ""
	if configuration.GetPassword() != "" {
		pw = "<redacted>"
	}
	var port = configuration.GetPort()
	var backupDirectory = configuration.GetBackupDirectory()
	var restoreDirectory = configuration.GetRestoreDirectory()
	var scriptsPath = configuration.GetScriptsPath()
	var allowedToDeleteFiles = configuration.IsAllowedToDeleteFiles()
	var skipStorage = configuration.IsInstructedToSkipStorage()
	var maxJobNumber = configuration.GetMaxJobNumber()
	var defaultS3Endpoint = configuration.GetDefaultS3Endpoint()
	log.Println("Using following configuration: ",
		"\nclient_username :", username,
		"\nclient_password :", pw,
		"\nclient_port :", port,
		"\ndirectory_backup :", backupDirectory,
		"\ndirectory_restore :", restoreDirectory,
		"\nscripts_path :", scriptsPath,
		"\nallowed_to_delete_files :", allowedToDeleteFiles,
		"\nskip_storage :", skipStorage,
		"\nmax_job_number :", maxJobNumber,
		"\ndefault_s3_endpoint :", defaultS3Endpoint)
}
