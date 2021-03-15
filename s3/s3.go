package s3

import (
	"context"
	"log"
	"net/url"
	"strings"

	"github.com/evoila/osb-backup-agent/errorlog"
	"github.com/evoila/osb-backup-agent/httpBodies"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

// getClient creates a S3 client with the provided credentials.
// A mutex regulates access to the credentials and ensures the creation of the session with the correct credentials.
func getClient(endpoint, authkey, authSecret, region string, useSSL bool) (*minio.Client, error) {
	var minioClient *minio.Client
	var err error

	// --- The minio client can not handle an endpoint with a scheme (for example https://my.s3.server -> 'https://'), so we need to remove the scheme before using the URL ---
	endpointURL, err := url.Parse(endpoint)
	if err != nil {
		return nil, errorlog.LogError("Unable to parse endpoint URL due to: '", err.Error(), "'")
	}
	if endpointURL.Scheme != "" {
		log.Println("Endpoint '" + endpoint + "' contains http(s) scheme, therefore trying to split URL and scheme.")
		endpointWithoutScheme := strings.Split(endpoint, endpointURL.Scheme+"://")
		endpoint = endpointWithoutScheme[1] // First element is "", because there are (or rather should be) no characters before the scheme.
		log.Println("Successfully split. Now using", endpoint, "as new endpoint.")

		// --- If scheme of the endpoint conflicts with the given useSSL boolean, we can not guess, which one the user wants to use
		if endpointURL.Scheme == "http" && useSSL {
			return nil, errorlog.LogError("The given endpoint contains the scheme 'http', but the requests dictates 'https'.")
		}
		if endpointURL.Scheme == "https" && !useSSL {
			return nil, errorlog.LogError("The given endpoint contains the scheme 'https', but the requests dictates 'http'.")
		}
	}

	minioClient, err = minio.New(endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(authkey, authSecret, ""),
		Secure: useSSL,
		Region: region, // setting a region here overwrites the clients selfidentification process of the region -> to use "" is valid here
	})

	return minioClient, err
}

func UploadFile(filename, path string, body httpBodies.BackupBody) error {

	ctx := context.Background()

	// -- Initialize client object --
	minioClient, err := getClient(body.Destination.Endpoint, body.Destination.AuthKey, body.Destination.AuthSecret, body.Destination.Region, !body.Destination.SkipSSL)
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
	minioClient, err := getClient(body.Destination.Endpoint, body.Destination.AuthKey, body.Destination.AuthSecret, body.Destination.Region, !body.Destination.SkipSSL)
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

	// -- Downloading the restore file from the given bucket --
	err = minioClient.FGetObject(ctx, body.Destination.Bucket, filename, path, minio.GetObjectOptions{})
	if err != nil {
		return errorlog.LogError("Failed to download from S3 due to '", err.Error(), "'")
	}
	log.Printf("Successfully downloaded %s from bucket %s\n", filename, body.Destination.Bucket)

	return nil
}
