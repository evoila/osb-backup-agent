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
| client_port | 8000 | The port the client will use for the http interface. Defaults to 8000 |
| directory_backup | /tmp/backups | The directory in which the agent looks for files to upload to the cloud storage. |
| directory_restore | /tmp/restores | The directory in which the agent will put the downloaded restore file from the cloud storage. |
| scrips_path | /tmp/scrips | The directory in which the agent will look for the backup scrips. Defaults to `/var/vcap/jobs/backup-agent/backup`  |
| allowed_to_delete_files | true | Flag for permission to delete already existing files. Defaults to `false` | 


## Endpoints ##
The agent supports three http endpoints for status, backup and restore. The endpoints are secured by BasicAuth.

|Endpoint|Method|Body|Description|
|----|----|----|----|
|/status|GET| - |Simple check whether the agent is running. |
|/backup|POST| See Backup below |Trigger the backup procedure for the service.|
|/backup/{id}|GET| - |Returns the status of the requested backup job.|
|/backup|DELETE| See Job deletion body below |Removes a result of a backup job.|
|/restore|PUT| See Restore below |Trigger the restore procedure for the service.|
|/restore/{id}|GET| - |Returns the status of the requested restore job.|
|/restore|DELETE| See Job deletion body below |Removes a result of a restore job.|

Following status codes can be expected for multiple endpoints:

### Backup ###

The backup functionality consists of three calls: Triggering a backup, requesting its status and removing the job from the agent.

#### Trigger Backup ####
This call triggers an asynchronous backup procedure. 

Endpoint: POST /backup

##### Status Codes and their meaning ####
The backup agent intentionally returns the following status codes. Codes that differ are likely to be unexpected and not intended to be returned.

| Code | Body | Description |
| --- | --- | --- |
| 201 | - | A backup was triggered and is getting run asynchronously. |
| 400| See Polling Body| The information in the body are not sufficient. |
| 401| See Simple response body rigth below| The provided credentials are not correct. |
| 409 | See Polling Body| There already exists a job with the given id.|

#### Polling Backup Status ####
This call request the status of the dedicated job identified by the given id.

Endpoint: GET /backup/{id}

##### Status Codes and their meaning #####
The backup agent intentionally returns the following status codes. Codes that differ are likely to be unexpected and not intended to be returned.
| Code | Body | Description |
| --- | --- | --- |
| 200 | See Polling Body | A matching job was found and its status is returned. |
| 400| - | No valid id was provided. |
| 401| See Simple response body| The provided credentials are not correct. |
| 404 | - | There exists no job for the given id.|


#### Backup Job Deletion ####
This call requests the deletion of a result of a backup job. This should be done to either use the id again or free the space for the agent.

Endpoint: DELETE /backup

##### Status Codes and their meaning #####
The backup agent intentionally returns the following status codes. Codes that differ are likely to be unexpected and not intended to be returned.
| Code | Body | Description |
| --- | --- | --- |
| 200 | - | A matching job was found and deleted. |
| 400| - | The information in the body are not sufficient. |
| 401| See Simple response body| The provided credentials are not correct. |
| 410 | - | No matching job was found.|






### Restore ###
The restore functionality consists of three calls: Triggering a backup, requesting its status and removing the job from the agent.

#### Trigger Restore ####
This call triggers an asynchronous restore procedure. 

Endpoint: PUT /restore


##### Response Codes and their meaning #####
See Trigger Backup Status Codes and their meaning


#### Polling Restore Status ####
This call request the status of the dedicated job identified by the given id.

Endpoint: GET /restore/{id}

##### Status Codes and their meaning #####
See Backup Polling Status Codes and their meaning


#### Restore Job Deletion ####
This call requests the deletion of a result of a backup job. This should be done to either use the id again or free the space for the agent.

Endpoint: DELETE /restore

##### Status Codes and their meaning #####
See Backup Job Deletion Status Codes and their meaning


## Request Bodies ##

### Trigger Backup Body ###
```json
{
    "id" : "778f038c-e1c5-11e8-9f32-f2801f1b9fd1",
    "destination" : {
        "type": "S3",
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


### Trigger Restore Body ###
```json
{
    "id" : "778f038c-e1c5-11e8-9f32-f2801f1b9fd1",
    "destination" : {
        "type": "S3",
        "bucket": "bucketName",
        "region": "regionName",
        "authKey": "key",
        "authSecret": "secret",
        "filename": "filename"
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

### Job Deletion Body ###

```json
{
	"id" : "778f038c-e1c5-11e8-9f32-f2801f1b9fd1"
}
```

## Response Bodies ##


### Simple Response Body ###
```json
{
    "message" : "simple response message"
}
```

### Backup Polling Body ###
Please not that the ``error_message`` field will not show up in the json, if it is empty.

```json
{
    "status": "SUCCEEDED / FAILED / RUNNING",
    "message": "backup successfully carried out",
    "state": "finished / name of the current phase",
    "error_message": "contains message dedicated to the occuring error, will not show up if empty",
    "region": "S3 region",
    "bucket": "S3 bucket",
    "filename": "host_YYYY_MM_DD_database.tar.gz",
    "filesize": {
        "size": 42,
        "unit": "byte"
    },
    "start_time": "YYYY-MM-DDTHH:MM:SS+00:00",
    "end_time": "YYYY-MM-DDTHH:MM:SS+00:00",
    "execution_time_ms": 42000,
    "pre_backup_lock_log": "stdout of the dedicated script",
    "pre_backup_lock_errorlog": "stderr of the dedicated script",
    "pre_backup_check_log": "stdout of the dedicated script",
    "pre_backup_check_errorlog": "stderr of the dedicated script",
    "backup_log": "stdout of the dedicated script",
    "backup_errorlog": "stderr of the dedicated script",
    "backup_cleanup_log": "stdout of the dedicated script",
    "backup_cleanup_errorlog": "stderr of the dedicated script",
    "post_backup_unlock_log": "stdout of the dedicated script",
    "post_backup_unlock_errorlog": "stderr of the dedicated script"
}
```

### Restore Polling Body ###
Please not that the ``error_message`` field will not show up in the json, if it is empty.

```json
{
    "status": "SUCCEEDED / FAILED / RUNNING",
    "message": "restore successfully carried out",
    "state": "finished / name of the current phase",
    "error_message": "contains message dedicated to the occuring error, will not show up if empty",
    "start_time": "YYYY-MM-DDTHH:MM:SS+00:00",
    "end_time": "YYYY-MM-DDTHH:MM:SS+00:00",
    "execution_time_ms": 42000,
    "pre_restore_lock_log": "stdout of the dedicated script",
    "pre_restore_lock_errorlog": "stderr of the dedicated script",
    "restore_log": "stdout of the dedicated script",
    "restore_errorlog": "stderr of the dedicated script",
    "restore_cleanup_log": "stdout of the dedicated script",
    "restore_cleanup_errorlog": "stderr of the dedicated script",
    "post_restore_unlock_log": "stdout of the dedicated script",
    "post_restore_unlock_errorlog": "stderr of the dedicated script"
}
```

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
--