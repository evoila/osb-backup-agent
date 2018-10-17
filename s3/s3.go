package s3

import (
	"fmt"
	"log"
	"os"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	"github.com/evoila/go-backup-agent/errorlog"
	"github.com/evoila/go-backup-agent/httpBodies"
)

func UploadFile(filename, path string, body httpBodies.BackupBody) error {

	log.Println("Opening file at", path)
	file, err := os.Open(path)
	if err != nil {
		return errorlog.LogError("Failed to open file ", path, " due to '", err.Error(), "'")
	}
	defer file.Close()
	log.Println("Successfully opened file at", path)

	// -- Creating session, service client and uploader --
	os.Setenv("AWS_ACCESS_KEY_ID", body.Destination.AuthKey)
	os.Setenv("AWS_SECRET_ACCESS_KEY", body.Destination.AuthSecret)

	//Clear credentials after use
	defer os.Setenv("AWS_ACCESS_KEY_ID", "")
	defer os.Setenv("AWS_SECRET_ACCESS_KEY", "")

	log.Println("Creating S3 session ...")
	sess, err := session.NewSession(&aws.Config{
		Region: aws.String(body.Destination.Region)},
	)

	if err != nil {
		return errorlog.LogError("Unable to create a S3 session due to due to '", err.Error(), "'")
	}
	log.Println("Successfully created S3 session")

	log.Println("Setting up S3 uploader")
	var uploader = s3manager.NewUploader(sess)
	var client = s3.New(sess)

	// -- Listing all objects of the given bucket --
	// Surely not needed later
	listObjectsOfBucket(body.Destination.Bucket, client)

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

	listObjectsOfBucket(body.Destination.Bucket, client)

	return nil
}

func DownloadFile(filename, path string, body httpBodies.RestoreBody) error {

	log.Println("Creating file at", path)
	file, err := os.Create(path)
	if err != nil {
		return errorlog.LogError("Failed to create file ", path, "due to '", err.Error(), "'")
	}
	defer file.Close()

	// -- Creating session, service client and uploader --
	os.Setenv("AWS_ACCESS_KEY_ID", body.Destination.AuthKey)
	os.Setenv("AWS_SECRET_ACCESS_KEY", body.Destination.AuthSecret)

	//Clear credentials after use
	defer os.Setenv("AWS_ACCESS_KEY_ID", "")
	defer os.Setenv("AWS_SECRET_ACCESS_KEY", "")

	log.Println("Creating S3 session ...")
	sess, err := session.NewSession(&aws.Config{
		Region: aws.String(body.Destination.Region)},
	)

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
	// -- Listing all buckets --
	// Could be removed later on

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
