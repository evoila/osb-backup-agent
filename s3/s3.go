package s3

import (
	"fmt"
	"log"
	"os"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	"github.com/evoila/osb-backup-agent/errorlog"
	"github.com/evoila/osb-backup-agent/httpBodies"
	"github.com/evoila/osb-backup-agent/mutex"
)

var sessionMutex mutex.Mutex

func SetUpS3() {
	sessionMutex = make(mutex.Mutex, 1)
	sessionMutex.Release()
}

// getSession creates a AWS S3 session with the provided credentials.
// The ENV VARs are set and cleared in this method.
// A mutex regulates access to the credentials and ensures the creation of the session with the correct credentials.
func getSession(region, authkey, authSecret string) (*session.Session, error) {
	sessionMutex.Acquire()

	// Setting credentials of this request
	os.Setenv("AWS_ACCESS_KEY_ID", authkey)
	os.Setenv("AWS_SECRET_ACCESS_KEY", authSecret)

	log.Println("Creating S3 session ...")
	sess, err := session.NewSession(&aws.Config{
		Region: aws.String(region)},
	)

	//Clear credentials after use
	os.Setenv("AWS_ACCESS_KEY_ID", "")
	os.Setenv("AWS_SECRET_ACCESS_KEY", "")

	sessionMutex.Release()

	return sess, err
}

func UploadFile(filename, path string, body httpBodies.BackupBody) error {

	log.Println("Opening file at", path)
	file, err := os.Open(path)
	if err != nil {
		return errorlog.LogError("Failed to open file ", path, " due to '", err.Error(), "'")
	}
	defer file.Close()
	log.Println("Successfully opened file at", path)

	// -- Creating S3 session and uploader --
	sess, err := getSession(body.Destination.Region, body.Destination.AuthKey, body.Destination.AuthSecret)

	if err != nil {
		return errorlog.LogError("Unable to create a S3 session due to '", err.Error(), "'")
	}
	log.Println("Successfully created S3 session")

	log.Println("Setting up S3 uploader")
	var uploader = s3manager.NewUploader(sess)
	//var client = s3.New(sess)

	// -- Uploading the backup file to the given bucket --
	log.Println("Uploading", filename, "to", body.Destination.Bucket)

	_, err = uploader.Upload(&s3manager.UploadInput{
		Bucket: aws.String(body.Destination.Bucket),
		Key:    aws.String(filename),
		Body:   file,
	})

	if err != nil {
		return errorlog.LogError("Failed to upload to S3 due to '", err.Error(), "'")
	}
	log.Printf("Successfully uploaded %q to %q\n", filename, body.Destination.Bucket)

	return nil
}

func DownloadFile(filename, path string, body httpBodies.RestoreBody) error {

	log.Println("Creating file at", path)
	file, err := os.Create(path)
	if err != nil {
		return errorlog.LogError("Failed to create file ", path, " due to '", err.Error(), "'")
	}
	defer file.Close()

	sess, err := getSession(body.Destination.Region, body.Destination.AuthKey, body.Destination.AuthSecret)

	if err != nil {
		return errorlog.LogError("Unable to create a S3 session due to '", err.Error(), "'")
	}

	log.Println("Sucessfully created S3 session")

	log.Println("Setting up S3 downloader")
	var downloader = s3manager.NewDownloader(sess)
	//var client = s3.New(sess)

	numBytes, err := downloader.Download(file,
		&s3.GetObjectInput{
			Bucket: aws.String(body.Destination.Bucket),
			Key:    aws.String(filename),
		})
	if err != nil {
		return errorlog.LogError("Failed to download the file ", filename, "  due to '", err.Error(), "'")
	}

	log.Println("Successfully downloaded", file.Name(), "(", numBytes, "bytes )")

	return nil
}

func listObjectsOfBucket(bucket string, client *s3.S3) error {

	resp, err := client.ListObjects(&s3.ListObjectsInput{Bucket: aws.String(bucket)})
	if err != nil {
		return errorlog.LogError("Failed to list all buckets due to '", err.Error(), "'")
	}

	log.Println("Objects of bucket", bucket, ":")

	for _, item := range resp.Contents {
		fmt.Println("Name:         ", *item.Key)
		fmt.Println("Last modified:", *item.LastModified)
		fmt.Println("Size:         ", *item.Size)
		fmt.Println("Storage class:", *item.StorageClass)
		fmt.Println("")

	}

	return nil
}

func listAllBuckets(client *s3.S3) error {
	log.Println("Sending request for the bucket list.")
	result, err := client.ListBuckets(nil)
	if err != nil {
		return errorlog.LogError("Unable to list buckets due to '", err.Error(), "'")
	}
	log.Println("Listing all buckets.")
	fmt.Println("Buckets:")

	for _, b := range result.Buckets {
		fmt.Printf("* %s created on %s\n",
			aws.StringValue(b.Name), aws.TimeValue(b.CreationDate))
	}
	return nil
}
