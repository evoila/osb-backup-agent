package s3

import (
	"context"
	"log"

	"github.com/evoila/osb-backup-agent/errorlog"
	"github.com/evoila/osb-backup-agent/httpBodies"
	"github.com/evoila/osb-backup-agent/mutex"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

var sessionMutex mutex.Mutex

func SetUpS3() {
	sessionMutex = make(mutex.Mutex, 1)
	sessionMutex.Release()
}

// getClient creates a S3 client with the provided credentials.
// A mutex regulates access to the credentials and ensures the creation of the session with the correct credentials.
func getClient(endpoint, authkey, authSecret, region string, useSSL bool) (*minio.Client, error) {
	sessionMutex.Acquire()

	var minioClient *minio.Client
	var err error

	minioClient, err = minio.New(endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(authkey, authSecret, ""),
		Secure: useSSL,
		Region: region, // setting a region here overwrites the clients selfidentification process of the region -> to use "" is valid here
	})

	sessionMutex.Release()

	return minioClient, err
}

func UploadFile(filename, path string, body httpBodies.BackupBody) error {

	ctx := context.Background()

	// -- Initialize client object --
	minioClient, err := getClient(body.Destination.Endpoint, body.Destination.AuthKey, body.Destination.AuthSecret, body.Destination.Region, body.Destination.UseSSL)
	if err != nil {
		return errorlog.LogError("Unable to create a S3 session due to '", err.Error(), "'")
	}
	log.Println("Successfully created S3 client")

	// -- Check to see if bucket exists --
	exists, err := minioClient.BucketExists(ctx, body.Destination.Bucket)
	if err == nil && exists {
		log.Println("Bucket", body.Destination.Bucket, "exists and used user owns it.")
	} else if err == nil && !exists {
		return errorlog.LogError("S3 bucket seems not to exist due to '", err.Error(), "'")
	} else if err != nil {
		return errorlog.LogError("Failed to check S3 bucket existance due to '", err.Error(), "'")
	}

	// -- Uploading the backup file to the given bucket --
	uploadInfo, err := minioClient.FPutObject(ctx, body.Destination.Bucket, filename, path, minio.PutObjectOptions{})
	if err != nil {
		return errorlog.LogError("Failed to upload to S3 due to '", err.Error(), "'")
	}
	log.Printf("Successfully uploaded %s of size %d to bucket %s\n", filename, uploadInfo.Size, uploadInfo.Bucket)

	return nil
}

func DownloadFile(filename, path string, body httpBodies.RestoreBody) error {

	ctx := context.Background()

	// -- Initialize client object --
	minioClient, err := getClient(body.Destination.Endpoint, body.Destination.AuthKey, body.Destination.AuthSecret, body.Destination.Region, body.Destination.UseSSL)
	if err != nil {
		return errorlog.LogError("Unable to create a S3 session due to '", err.Error(), "'")
	}
	log.Println("Successfully created S3 client")

	// -- Check to see if bucket exists --
	exists, err := minioClient.BucketExists(ctx, body.Destination.Bucket)
	if err == nil && exists {
		log.Println("Bucket", body.Destination.Bucket, "exists and used user owns it.")
	} else if err == nil && !exists {
		return errorlog.LogError("S3 bucket seems not to exist.")
	} else if err != nil {
		return errorlog.LogError("Failed to check S3 bucket existance due to '", err.Error(), "'")
	}

	err = minioClient.FGetObject(ctx, body.Destination.Bucket, filename, path, minio.GetObjectOptions{})
	if err != nil {
		return errorlog.LogError("Failed to download from S3 due to '", err.Error(), "'")
	}
	log.Printf("Successfully downloaded %s from bucket %s\n", filename, body.Destination.Bucket)

	return nil
}
