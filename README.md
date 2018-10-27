# osb-backup-agent #

This project holds a small go web agent for backup and restore actions for bosh, but does not contain any logic for specific services or applications. The agent simply allows to trigger scripts in a predefined directory and uploads or downloads from a cloud storage.

Currently implemented cloud storages: **S3**

## Installation ##
Download this repository and then get its dependencies via ```glide update```.



## Configuration ##
The agent uses environment variables to configurate its parameters.

| Variable | Example | Description |
|----|----|----|
| client_username | admin | The username for authorization of http requests |
| client_password | admin | The password for authorization of http requests |
| client_port | 8000 | The port the client will use for the http interface |
| directory_backup | /tmp/backups | The directory in which the agent looks for files to upload to the cloud storage. |
| directory_restore | /tmp/restores | The directory in which the agent will put the downloaded restore file from the cloud storage. |
| scrips_path | /tmp/scrips | The directory in which the agent will look for the backup scrips. Defaults to `/var/vcap/jobs/backup-agent/backup`  |


## Endpoints ##
The agent supports three http endpoints for status, backup and restore. The endpoints are secured by BasicAuth.

|Endpoint|Method|Body|Description|
|----|----|----|----|
|/status|GET| - |Simple check whether the agent is running. |
|/backup|POST| See Backup body below |Trigger the backup procedure for the service.|
|/restore|PUT| See Restore body below |Trigger the restore procedure for the service.|

#### Backup body ####
```json
{
    "destination" : {
        "type": "s3",
        "bucket": "bucketName",
        "region": "regionName",
        "authKey": "key",
        "authSecret": "secret"
    },
    "backup" : {
        "host": "host",
        "user": "user",
        "password": "password",
        "database": "database name",
        "parameters": [
            { "key": "arbitraryValue" },
            { "retries": 2}
        ]
    }
}
```
Please note that objects in the parameters object can not have nested objects, arrays, lists, maps and so on inside. Only use simple types here as these values will be set as environment variables for the scripts to work with.

#### Restore body ####
```json
{
    "destination" : {
        "type": "s3",
        "bucket": "bucketName",
        "region": "regionName",
        "authKey": "key",
        "authSecret": "secret",
        "file": "filename"
    },
    "restore" : {
        "host": "host",
        "user": "user",
        "password": "password",
        "database": "database name",
        "parameters": [
            { "key": "arbitraryValue" },
            { "retries": 2}
        ]
    }
}
```
Please note that objects in the parameters object can not have nested objects, arrays, lists, maps and so on inside. Only use simple types here as these values will be set as environment variables for the shell scripts to work with.

## Functionality ##
The agent calls a predefined set of shell scripts in order to trigger the backup or restore procedure. Generally speaking there are three stages: Pre, Action, Post. 
These files have to be located or will be placed in the respective directories set by the environment variables.

#### Backup ####
The agent runs following shell scripts from top to bottom:
- `pre-backup-lock`
- `pre-backup-check`
- `backup`
- `backup-cleanup`
- `post-backup-unlock`

In the backup stage, after the script generated the file to upload (name consists of `<host>_YYYY_MM_DD_<dbname>.tar.gz`), the agent uploads the backup file from the set directory to the cloud storage using the given information and credentials.

#### Restore ####
The agent runs following shell scripts from top to bottom:
- `pre-restore-lock`
- `restore`
- `restore-cleanup`
- `post-restore-unlock`

In the restore stage, before the dedicated script starts the actual restore, the agent downloads the backed up restore file from the cloud storage, using the given information and credentials, and puts it in the dedicated directory.

## Version ##
v0.2