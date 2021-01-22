# Table of Contents
1. [osb-backup-agent](#osb-backup-agent)
2. [Installation](#installation)
3. [Configuration](#configuration)
4. [Endpoints](#endpoints)
7. [Functionality](#functionality)

# osb-backup-agent #

This project holds a small go web agent for backup and restore actions for bosh, but does not contain any logic for specific services or applications. The agent simply allows to trigger scripts in a predefined directory and uploads to or downloads from a cloud storage.

Supported cloud storages: **S3**, **SWIFT**

For further information about the usage of API for the backup and restore see [Backup](docs/backup.md) and [Restore](docs/restore.md).

## Installation ##
As Go Module:
* Download Go Module via `GO111MODULE=on go get github.com/evoila/osb-backup-agent` or download the repository and place them accordingly on your GOPATH
* Get its dependencies via the Go modules functions
* Run `go build`
* Set environment variables as necessary
* Execute the created binary

## Configuration ##
The agent uses environment variables to configurate its parameters.

| Variable | Example | Description |
|----|----|----|
| client_username | admin | The username for authorization of http requests |
| client_password | admin | The password for authorization of http requests |
| client_port | 8000 | The port the client will use for the http interface. Defaults to 8000 |
| directory_backup | /tmp/backups | The directory in which the agent looks for files to upload to the cloud storage. For every job, a directory with the id of the job as its name will be created in here. |
| directory_restore | /tmp/restores | The directory in which the agent will put the downloaded restore files from the cloud storage. For every job, a directory with the id of the job as its name will be created in here. |
| scrips_path | /tmp/scrips | The directory in which the agent will look for the backup scrips. Defaults to `/var/vcap/jobs/backup-agent/backup`  |
| allowed_to_delete_files | true | Flag for permission to delete already existing files. Defaults to `false`. | 
| skip_storage | true | Flag for instruction to skip upload and download. Defaults to `false`. | 
| max_job_number | 10 | Maximum number of running jobs at a time. Defaults to 10. |


## Endpoints ##
The agent supports http endpoints for status, backup and restore. These endpoints are secured by BasicAuth.

|Endpoint|Method|Body|Description|
|----|----|----|----|
|/status|GET| - |Simple check whether the agent is running. |
|/backup|POST| See [Backup](docs/backup.md)|Trigger the backup procedure for the service.|
|/backup/{id}|GET| - |Returns the status of the requested backup job.|
|/backup|DELETE| See [Backup](docs/backup.md) |Removes a result of a backup job.|
|/restore|PUT| See [Restore](docs/restore.md) |Trigger the restore procedure for the service.|
|/restore/{id}|GET| - |Returns the status of the requested restore job.|
|/restore|DELETE| See [Restore](docs/restore.md) |Removes a result of a restore job.|


## Functionality ##
The agent calls a predefined set of shell scripts in order to trigger the backup or restore procedure. Generally speaking there are three stages: Pre, Action, Post. 
These files have to be located or will be placed in the respective directories set by the environment variables.

**Note**:
The osb-backup-agent does **not** store any information in a persistent way! All information are stored in memory and are thereby ephemeral.

The upload or download functionality can be skipped by using the `skipStorage` field in the respective request bodies or via the configuration property `skip_storage`. If this is the case, the destination information for the selected storage are set as environment variables for each script.



|Environment variable key for S3| Field in destination body|
|----|----|
|S3_BUCKET|bucket|
|S3_ENDPOINT|endpoint|
|S3_REGION|region|
|S3_AUTHKEY|authKey|
|S3_AUTHSECRET|authSecret|
|S3_USESSL|useSSL|

|Environment variable key for SWIFT| Field in destination body|
|----|----|
|SWIFT_AUTHURL|authUrl|
|SWIFT_DOMAIN|domain|
|SWIFT_CONTAINERNAME|container_name|
|SWIFT_PROJECTNAME|project_name|
|SWIFT_USERNAME|username|
|SWIFT_PASSWORD|password|


#### Backup ####
The agent runs following shell scripts from top to bottom:
- `pre-backup-lock`
- `pre-backup-check`
- `backup`
- `backup-cleanup`
- `post-backup-unlock`

In the backup stage, the agent generates a name (consists of `YYYY_MM_DD_HH_MM_<host>_<dbname>`) for the backup file and forwards its path (`backup_directory/<job_id>/<generated_file_name>`) to the back script. After the script generated the file to upload, the agent uploads the first encountered file in the dedicated directory (`backup_directory/<job_id>`) to the cloud storage using the given information and credentials.

##### Script Parameters #####
- `pre-backup-lock databasename`
- `pre-backup-check databasename`
- `backup host username password databasename file_name_without_type job_id compression_flag encryption_key`
- `backup-cleanup databasename job_id`
- `post-backup-unlock databasename`

The job_id parameter is equal to to the id given by the request, which triggered the backup or restore.
Be aware that encryption key can be empty and uppon adding more parameters after the encryption_key, the order could not match anymore. In future there might be need for named parameters.


#### Restore ####
The agent runs following shell scripts from top to bottom:
- `pre-restore-lock`
- `restore`
- `restore-cleanup`
- `post-restore-unlock`

In the restore stage, the agent downloads a file with the given file name (out of the request body) from the used cloud storage to the dedicated directory (`restore_directory/<job_id>/`).

##### Script Parameters #####
- `pre-restore-lock job_id`
- `restore host username password databasename filename job_id compression_flag encryption_key`
- `restore-cleanup job_id`
- `post-restore-unlock`

In the restore stage, before the dedicated script starts the actual restore, the agent downloads the backed up restore file from the cloud storage, using the given information and credentials, and puts it in the dedicated directory.

Be aware that encryption key can be empty and uppon adding more parameters after the encryption_key, the order could not match anymore. In future there might be need for named parameters.

<p align="center">
	    <span>&nbsp; | &nbsp;</span> 
    <span><a href="docs/backup.md">Backup -></a></span>
</p>
