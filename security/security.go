package security

import (
	"crypto/subtle"
	"log"
	"net/http"

	"github.com/evoila/osb-backup-agent/configuration"
)

func BasicAuth(w http.ResponseWriter, r *http.Request) bool {

	user, pass, ok := r.BasicAuth()
	var username = configuration.GetUsername()
	var password = configuration.GetPassword()

	if !ok || subtle.ConstantTimeCompare([]byte(user), []byte(username)) != 1 || subtle.ConstantTimeCompare([]byte(pass), []byte(password)) != 1 {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(401)
		w.Write([]byte("{ \"message\" : \"Unauthorised.\" }"))
		log.Println("Request is not correctly authorised. Should be " + username + ":" + password + " is " + user + ":" password )
		return false
	}
	log.Println("Request is correctly authorised.")
	return true
}
