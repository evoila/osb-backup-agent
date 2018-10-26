package webclient

import (
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"

	"github.com/evoila/osb-backup-agent/backup"
	"github.com/evoila/osb-backup-agent/configuration"
	"github.com/evoila/osb-backup-agent/health"
	"github.com/evoila/osb-backup-agent/restore"
	"github.com/gorilla/mux"
)

var port int
var router mux.Router

func StartWebAgent() {
	router := mux.NewRouter()
	setUpEndpoints(router)

	log.Println("Preparing web client ...")
	port = configuration.GetPort()
	if port < 0 {
		log.Println("[ERROR]", "Port is invalid. Stopping the agent.")
		os.Exit(1)
	}
	var portAsString = strings.Join([]string{":", strconv.Itoa(port)}, "")
	log.Println("Successfully prepared the web client")

	log.Println("Starting and running web client on port", GetUsedPort())
	log.Fatal(http.ListenAndServe(portAsString, router))
}

func setUpEndpoints(router *mux.Router) {
	log.Println("Setting up endpoints:")
	log.Println("GET /status")
	router.HandleFunc("/status", health.HealthCheck).Methods("GET")
	log.Println("POST /backup")
	router.HandleFunc("/backup", backup.BackupRequest).Methods("POST")
	log.Println("PUT /restore")
	router.HandleFunc("/restore", restore.RestoreRequest).Methods("PUT")
	log.Println("End points are set up.")
}

func GetUsedPort() int {
	return port
}
