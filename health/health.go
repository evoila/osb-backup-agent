package health

import (
	"log"
	"net/http"

	"github.com/evoila/go-backup-agent/security"
)

func HealthCheck(w http.ResponseWriter, r *http.Request) {
	log.Println("Health check was requested.")

	if !security.BasicAuth(w, r) {
		return
	}
	log.Println("Sending signs of life.")
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(200)
	w.Write([]byte("{ \"message\" : \"Client is running\" }"))
	log.Println("Finished status request.")

}
