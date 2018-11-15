package shell

import (
	"bytes"
	"errors"
	"log"
	"os"
	"os/exec"
	"strconv"
	"strings"

	"github.com/evoila/osb-backup-agent/configuration"
	"github.com/evoila/osb-backup-agent/errorlog"
)

var Directory = configuration.GetScriptsPath()

func ExecuteScriptForStage(stageName string, jsonParams []string, params ...string) (found bool, logs string, errlogs string, err error) {
	var fileName string
	found, fileName = CheckForBothExistingFiles(Directory, stageName)
	if !found {
		return found, "", "", errors.New(errorlog.Concat([]string{"No script found for the ", stageName, " stage."}, ""))
	}

	out, errOut, err := ExecShellScript(GetPathToFile(Directory, fileName), jsonParams, params)

	if err != nil {
		errorlog.LogError("Calling the shell script ", fileName,
			" resulted in an error with the debug message \n'", err.Error(),
			"'\nwith the script's Stdout\n'", out.String(),
			"'\nand the script's Stderr\n'", errOut.String(), "'")
	}

	log.Println("Script's Stdout:", out.String())
	log.Println("Script's Sterr:", errOut.String())
	return true, out.String(), errOut.String(), err
}

func ExecShellScript(path string, jsonParams []string, params []string) (bytes.Buffer, bytes.Buffer, error) {
	log.Println("Executing the", path, "script.")

	var cmd *exec.Cmd

	if len(params) > 0 {
		if len(params) == 9 { // restore
			log.Println("Using following parameters: [", params[0], params[1], "<redacted>", params[3], params[4], params[5], params[6], params[7], "<redacted>", "]")
			cmd = exec.Command("bash", path, params[0], params[1], params[2], params[3], params[4], params[5], params[6], params[7], params[8])
		} else if len(params) == 8 { // Backup
			log.Println("Using following parameters: [", params[0], params[1], "<redacted>", params[3], params[4], params[5], params[6], "<redacted>", "]")
			cmd = exec.Command("bash", path, params[0], params[1], params[2], params[3], params[4], params[5], params[6], params[7])
		} else if len(params) == 2 { // backup-cleanup,
			log.Println("Using following parameters: [", params[0], params[1], "]")
			cmd = exec.Command("bash", path, params[0], params[1])
		} else if len(params) == 1 { // pre-backup-check, pre-backup-lock, post-backup-unlock, restore-cleanup
			log.Println("Using following parameter: ", params[0])
			cmd = exec.Command("bash", path, params[0])
		} else { // pre-restore-lock, post-restore-unlock
			var o, e bytes.Buffer
			return o, e, errors.New(errorlog.Concat([]string{"Wrong amount of parameters were given: ", strconv.Itoa(len(params))}, ""))
		}
	} else {
		log.Println("No further parameters given.")
		cmd = exec.Command("bash", path)

	}

	log.Println("Adding following environment variables to the execution environment:", jsonParams)
	addEnvVars(jsonParams, cmd)
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

func GetFileSize(path string) (int64, error) {
	file, err := os.Stat(path)
	if err != nil {
		return 0, errorlog.LogError("Accessing file stats of ", path, " failed due to '", err.Error(), "'")
	}
	return file.Size(), nil
}

func addEnvVars(params []string, cmd *exec.Cmd) {
	// Currently not setting os ENV VAR for the shells !
	//cmd.Env = os.Environ()
	for _, param := range params {
		cmd.Env = append(cmd.Env, param)
	}

}
