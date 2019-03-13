package shell

import (
	"bytes"
	"errors"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"strconv"
	"strings"

	"github.com/evoila/osb-backup-agent/configuration"
	"github.com/evoila/osb-backup-agent/errorlog"
)

var Directory = configuration.GetScriptsPath()

func ExecuteScriptForStage(stageName string, destinationParams []string, jsonParams []string, params ...string) (status bool, logs string, errlogs string, err error) {
	var fileName string
	status, fileName = CheckForBothExistingFiles(Directory, stageName)
	if !status {
		return status, "", "", errors.New(errorlog.Concat([]string{"No script found for the ", stageName, " stage."}, ""))
	}

	out, errOut, err := ExecShellScript(GetPathToFile(Directory, fileName), destinationParams, jsonParams, params)

	if err != nil {
		errorlog.LogError("Calling the shell script ", fileName,
			" resulted in an error with the debug message \n'", err.Error(),
			"'\nwith the script's Stdout\n'", out.String(),
			"'\nand the script's Stderr\n'", errOut.String(), "'")
		return false, out.String(), errOut.String(), err
	}

	log.Println("Script's Stdout:", out.String())
	log.Println("Script's Sterr:", errOut.String())
	return true, out.String(), errOut.String(), err
}

func ExecShellScript(path string, destinationParams []string, jsonParams []string, params []string) (bytes.Buffer, bytes.Buffer, error) {
	log.Println("Executing the", path, "script.")

	var cmd *exec.Cmd

	if len(params) > 0 {
		if len(params) == 8 { // backup, restore
			log.Println("Using following parameters: [", params[0], params[1], "<redacted>", params[3], params[4], params[5], params[6], "<redacted>", "]")
			cmd = exec.Command("bash", path, params[0], params[1], params[2], params[3], params[4], params[5], params[6], params[7])
		} else if len(params) == 2 { // backup-cleanup,
			log.Println("Using following parameters: [", params[0], params[1], "]")
			cmd = exec.Command("bash", path, params[0], params[1])
		} else if len(params) == 1 { // pre-backup-check, pre-backup-lock, post-backup-unlock, pre-restore-lock, restore-cleanup
			log.Println("Using following parameter: ", params[0])
			cmd = exec.Command("bash", path, params[0])
		} else { // post-restore-unlock
			var o, e bytes.Buffer
			return o, e, errors.New(errorlog.Concat([]string{"Wrong amount of parameters were given: ", strconv.Itoa(len(params))}, ""))
		}
	} else {
		log.Println("No further parameters given.")
		cmd = exec.Command("bash", path)

	}

	if len(destinationParams) > 0 {
		log.Println("Adding destination information as environment variables to the execution environment. Please refer to previous logs during the job preparation to see which variables are added.")
		addEnvVars(destinationParams, cmd)
	}

	log.Println("Adding following parameters as environment variables to the execution environment:", jsonParams)
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

func GetAllExistingFiles(directory string) ([]os.FileInfo, error) {
	files, err := ioutil.ReadDir(directory)
	if err != nil {
		return nil, err
	}
	return files, err
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

// GetCompleteFileName returns the name of the first file in the given directory, that starts with the given fileNameWithoutType
// Use an empty string for fileNameWithoutType to get the first file in the directory.
func GetCompleteFileName(directory, fileNameWithoutType string) (string, error) {
	files, err := ioutil.ReadDir(directory)
	if err != nil {
		return "", err
	}

	for _, f := range files {
		if strings.HasPrefix(f.Name(), fileNameWithoutType) {
			return f.Name(), nil
		}
	}
	return "", errors.New("Could not find a file starting with " + fileNameWithoutType)
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
	// Currently not setting ENV VARs of the current os for the shells ! Only the given additional ones !
	// cmd.Env = os.Environ()
	if params == nil {
		return
	}
	for _, param := range params {
		cmd.Env = append(cmd.Env, param)
	}

}
