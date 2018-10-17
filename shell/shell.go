package shell

import (
	"bytes"
	"errors"
	"log"
	"os"
	"os/exec"
	"strconv"
	"strings"

	"github.com/evoila/go-backup-agent/errorlog"
)

//var directory = "/var/vcap/jobs/backup"
var Directory = "testscripts"

func ExecuteScriptForStage(stageName string, jsonParams []string, params ...string) (bool, error) {
	var found, fileName = CheckForBothExistingFiles(Directory, stageName)
	if !found {
		return false, errors.New(errorlog.Concat([]string{"No script found for the ", stageName, " stage."}, ""))
	}

	out, errOut, err := ExecShellScript(GetPathToFile(Directory, fileName), jsonParams, params)

	if err != nil {
		errorlog.LogError("Calling the shell script ", fileName,
			" resulted in an error with the debug message \n'", err.Error(),
			"'\nwith the script's Stdout\n'", out.String(),
			"'\nand the script's Stderr\n'", errOut.String(), "'")
	}

	log.Println("Script's Stdout:", out.String())
	return true, nil
}

func ExecShellScript(path string, jsonParams []string, params []string) (bytes.Buffer, bytes.Buffer, error) {
	log.Println("Executing the", path, "script.")

	var cmd *exec.Cmd

	if len(params) > 0 {
		if len(params) == 4 {
			log.Println("Using follwing parameters: [", params[0], params[1], "<redacted>", params[3], "]")
			cmd = exec.Command("bash", path, params[0], params[1], params[2], params[3])

		} else if len(params) == 5 {
			log.Println("Using follwing parameters: [", params[0], params[1], "<redacted>", params[3], params[4], "]")
			cmd = exec.Command("bash", path, params[0], params[1], params[2], params[3], params[4])
		} else {
			var o, e bytes.Buffer
			return o, e, errors.New(errorlog.Concat([]string{"Wrong amount of parameters were given: ", strconv.Itoa(len(params))}, ""))
		}
	} else {
		log.Println("No further parameters given.")
		cmd = exec.Command("bash", path)

	}

	addEnvVars(jsonParams, cmd)
	log.Println("Adding following environment variables to the execution environment:", jsonParams)
	cmd.Stdin = strings.NewReader("")
	var out bytes.Buffer
	var errOut bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &errOut
	err := cmd.Run()
	return out, errOut, err
}

func CheckForExistingFile(directory, fileName string) bool {
	var path = GetPathToFile(directory, fileName)
	log.Println("Looking for file at", path)
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return false
	}
	return true
}

func CheckForBothExistingFiles(directory, fileName string) (bool, string) {
	if CheckForExistingFile(directory, fileName) {
		log.Println("File", fileName, "found.")
		return true, fileName
	}
	log.Println("File", fileName, "not found.")

	var extendedFileName = errorlog.Concat([]string{fileName, ".sh"}, "")
	if CheckForExistingFile(directory, extendedFileName) {
		log.Println("File", extendedFileName, "found.")
		return true, extendedFileName
	}
	log.Println("File", extendedFileName, "not found.")
	return false, extendedFileName

}

func GetPathToFile(directory, fileName string) string {
	return errorlog.Concat([]string{directory, fileName}, "/")
}

func addEnvVars(params []string, cmd *exec.Cmd) {
	// Currently not setting os ENV VAR for the shells !
	//cmd.Env = os.Environ()

	for _, param := range params {
		cmd.Env = append(cmd.Env, param)
	}

}
