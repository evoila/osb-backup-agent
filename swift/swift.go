package swift

import (
	"log"
	"os"

	"github.com/evoila/osb-backup-agent/errorlog"
	"github.com/evoila/osb-backup-agent/httpBodies"
	"github.com/ncw/swift"
)

func UploadFile(filename, path string, body httpBodies.BackupBody) error {

	log.Println("Opening file at", path)
	file, err := os.Open(path)
	if err != nil {
		return errorlog.LogError("Failed to open file ", path, " due to '", err.Error(), "'")
	}
	defer file.Close()
	log.Println("Successfully opened file at", path)

	c, err := createSwiftConnection(body.Destination)
	if err != nil {
		return errorlog.LogError("Failed to create a authenticated connection to swift due to '", err.Error(), "'")
	}

	log.Println("Putting file to swift...")
	_, err = c.ObjectPut(body.Destination.Container_name, filename, file, true, "", "", nil)
	if err != nil {
		return errorlog.LogError("Failed to put file to swift due to '", err.Error(), "'")
	}
	log.Printf("Successfully uploaded %q to %q at %q\n", filename, body.Destination.Container_name, body.Destination.Project_name)

	return nil
}

func DownloadFile(filename, path string, body httpBodies.RestoreBody) error {
	log.Println("Creating file at", path)
	file, err := os.Create(path)
	if err != nil {
		return errorlog.LogError("Failed to create file ", path, " due to '", err.Error(), "'")
	}
	defer file.Close()

	log.Println("Not implemented but found swift restore order")
	c, err := createSwiftConnection(body.Destination)

	if err != nil {
		return errorlog.LogError("Failed to create a authenticated connection to swift due to '", err.Error(), "'")
	}

	log.Println("Getting file from swift...")
	_, err = c.ObjectGet(body.Destination.Container_name, filename, file, true, nil)

	if err != nil {
		return errorlog.LogError("Failed to download the file ", filename, "  due to '", err.Error(), "'")
	}

	log.Println("Successfully downloaded", file.Name(), "from swift.")

	return nil
}

func createSwiftConnection(destination httpBodies.DestinationInformation) (swift.Connection, error) {
	// Create a connection
	c := swift.Connection{
		UserName: destination.Username,
		ApiKey:   destination.Password,
		AuthUrl:  destination.AuthUrl,
		Domain:   destination.Domain,
		Tenant:   destination.Project_name, // Tenant is equal to the project name in this connection
	}

	// Authenticate
	err := c.Authenticate()
	if err != nil {
		return c, err
	}
	log.Println("Successfully authenticated swift connection.")
	return c, nil
}
