package main

import (
	"log"

	"github.com/evoila/go-backup-agent/configuration"
	"github.com/evoila/go-backup-agent/webclient"
)

func main() {

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
	log.Println("Using following configuration: ",
		"\nclient_username :", username,
		"\nclient_password :", pw,
		"\nclient_port :", port,
		"\ndirectory_backup :", backupDirectory,
		"\ndirectory_restore :", restoreDirectory)

}
