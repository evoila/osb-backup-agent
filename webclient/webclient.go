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
	"github.com/evoila/osb-backup-agent/jobs"
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
	jobs.SetUpJobStructure()
	log.Println("Successfully prepared the web client")

	log.Println("Starting and running web client on port", GetUsedPort())
	log.Fatal(http.ListenAndServe(portAsString, router))
}

func setUpEndpoints(router *mux.Router) {
	log.Println("Setting up endpoints:")
	log.Println("GET /status")
	router.HandleFunc("/status", health.HealthCheck).Methods("GET")

	log.Println("GET /backup/{id}")
	router.HandleFunc("/backup/{id}", backup.HandlePolling).Methods("GET")
	log.Println("POST /backup")
	router.HandleFunc("/backup", backup.HandleAsyncRequest).Methods("POST")
	log.Println("DELETE /backup")
	router.HandleFunc("/backup", backup.RemoveJob).Methods("DELETE")

	log.Println("GET /restore/{id}")
	router.HandleFunc("/restore/{id}", restore.HandlePolling).Methods("GET")
	log.Println("PUT /restore")
	router.HandleFunc("/restore", restore.HandleAsyncRequest).Methods("PUT")
	log.Println("DELETE /restore")
	router.HandleFunc("/restore", restore.RemoveJob).Methods("DELETE")
	log.Println("End points are set up.")
}

func GetUsedPort() int {
	return port
}
