package httpBodies

import (
	"fmt"
	"log"

	"github.com/evoila/go-backup-agent/errorlog"
)

type BackupResponse struct {
	Message  string `json:"message"`
	FileName string `json:"filename"`
	Region   string `json:"region"`
	Bucket   string `json:"bucket"`
}

type ErrorResponse struct {
	Message      string `json:"message"`
	State        string `json:"state"`
	ErrorMessage string `json:"error message"`
}

type BackupBody struct {
	Destination DestinationInformation
	Backup      DbInformation
}

type RestoreBody struct {
	Destination DestinationInformation
	Restore     DbInformation
}

type DestinationInformation struct {
	Type       string
	Bucket     string
	Region     string
	AuthKey    string
	AuthSecret string
	File       string
}

type DbInformation struct {
	Host       string
	User       string
	Password   string
	Database   string
	Parameters []map[string]interface{}
}

func PrintOutBackupBody(body BackupBody) {

	log.Println("Backup Request Body: {\n",
		"    \"destination\" : {\n",
		errorlog.Concat([]string{"        \"type\" : \"", body.Destination.Type, "\",\n"}, ""),
		errorlog.Concat([]string{"        \"bucket\" : \"", body.Destination.Bucket, "\",\n"}, ""),
		errorlog.Concat([]string{"        \"region\" : \"", body.Destination.Region, "\",\n"}, ""),
		errorlog.Concat([]string{"        \"authKey\" : \"", body.Destination.AuthKey, "\",\n"}, ""),
		"        \"authSecret\" : <redacted>\n",
		"    },\n",
		"    \"backup\" : {\n",
		errorlog.Concat([]string{"        \"host\" : \"", body.Backup.Host, "\",\n"}, ""),
		errorlog.Concat([]string{"        \"user\" : \"", body.Backup.User, "\",\n"}, ""),
		"        \"password\" : <redacted>\n",
		errorlog.Concat([]string{"        \"database\" : \"", body.Backup.Database, "\",\n"}, ""),
		"        \"parameters\" : ",
		getParametersAsLogStringSlice(body.Backup.Parameters),
		"    }\n",
		"}")
}

func PrintOutRestoreBody(body RestoreBody) {

	log.Println("Restore Request Body: {\n",
		"    \"destination\" : {\n",
		errorlog.Concat([]string{"        \"type\" : \"", body.Destination.Type, "\",\n"}, ""),
		errorlog.Concat([]string{"        \"bucket\" : \"", body.Destination.Bucket, "\",\n"}, ""),
		errorlog.Concat([]string{"        \"region\" : \"", body.Destination.Region, "\",\n"}, ""),
		errorlog.Concat([]string{"        \"authKey\" : \"", body.Destination.AuthKey, "\",\n"}, ""),
		"        \"authSecret\" : <redacted>\n",
		errorlog.Concat([]string{"        \"file\" : \"", body.Destination.File, "\",\n"}, ""),
		"    },\n",
		"    \"backup\" : {\n",
		errorlog.Concat([]string{"        \"host\" : \"", body.Restore.Host, "\",\n"}, ""),
		errorlog.Concat([]string{"        \"user\" : \"", body.Restore.User, "\",\n"}, ""),
		"        \"password\" : <redacted>\n",
		errorlog.Concat([]string{"        \"database\" : \"", body.Restore.Database, "\",\n"}, ""),
		"        \"parameters\" : ",
		getParametersAsLogStringSlice(body.Restore.Parameters),
		"    }\n",
		"}")

}

func getParametersAsLogStringSlice(parameters []map[string]interface{}) []string {
	// Non string simple types will still be returned surrounded by "" !!!
	var strs []string
	for entry, entryValue := range parameters {
		for key, value := range entryValue {
			parsedValue := fmt.Sprintf("%v", value)
			if entry != len(parameters)-1 {
				strs = append(strs, errorlog.Concat([]string{"\n            {\"", key, "\": \"", parsedValue, "\" },"}, ""))
			} else {
				strs = append(strs, errorlog.Concat([]string{"\n            {\"", key, "\": \"", parsedValue, "\" }\n"}, ""))
			}

		}
	}
	return strs
}

func GetParametersAsEnvVarStringSlice(parameters []map[string]interface{}) (envParameters []string) {
	for _, entryValue := range parameters {
		for key, value := range entryValue {
			parsedValue := fmt.Sprintf("%s=%v", key, value)
			log.Println("Parsed additional parameter:", parsedValue)
			envParameters = append(envParameters, parsedValue)
		}
	}
	return
}
