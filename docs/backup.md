# Table of Contents

1. [Trigger Backup](#trigger-backup)
2. [Polling Backup Status](#polling-backup-status)
3. [Backup Job Deletion](#backup-job-deletion)
4. [Other Response Bodies](#other-response-bodies)

---

# Backup #

The backup functionality consists of three calls: Triggering a backup, requesting its status and removing the job from the agent.

## Trigger Backup ##
This call triggers an asynchronous backup procedure. 

Endpoint: POST /backup

### Trigger Backup Request Body ###
Fields that are not dedicated to the chosen type will be ignored.
Please note that objects in the parameters object can not have nested objects, arrays, lists, maps and so on inside. Only use simple types here as these values will be set as environment variables for the scripts to work with. **Be aware that the parameters are logged on the console of the agent! Do not use sensitive data, if you do not want to have it logged!** 
Furthermore will the compression, useSSL and skipStorage fields default to false, if no explicit value is present.

#### S3
```json
{
    "id" : "778f038c-e1c5-11e8-9f32-f2801f1b9fd1",
    "compression" : true,
    "encryption_key" : "example-encryption-key",
    "destination" : {
        "type": "S3",
        "skipStorage" : false,

        "bucket": "bucketName",
        "endpoint" : "http://custom.s3.endpoint",
        "useSSL" : true,
        "authKey": "key",
        "authSecret": "secret",
    },
    "backup" : {
        "host": "host",
        "username": "username",
        "password": "password",
        "database": "database name",
        "parameters": [
            { "key": "arbitraryValue" },
            { "retries": 2}
        ]
    }
}
```
#### Openstack Swift
```json
{
    "id" : "778f038c-e1c5-11e8-9f32-f2801f1b9fd1",
    "compression" : true,
    "encryption_key" : "example-encryption-key",
    "destination" : {
        "type": "SWIFT",
        "skipStorage" : false,

        "authUrl" : "auth url",
        "domain" : "domain name",
        "container_name" : "name of the container",
        "project_name" : "name of the project == tenant",
        "username" : "swift username",
        "password" : "swift API key"
    },
    "backup" : {
        "host": "host",
        "username": "username",
        "password": "password",
        "database": "database name",
        "parameters": [
            { "key": "arbitraryValue" },
            { "retries": 2}
        ]
    }
}
```

### Status Codes and their meaning ###
The backup agent intentionally returns the following status codes. Codes that differ are likely to be unexpected and not intended to be returned.

| Code | Body | Description |
| --- | --- | --- |
| 201 | - | A backup was triggered and is getting run asynchronously. |
| 400| See Polling Body| The information in the body are not sufficient. |
| 401| See Simple Response Body | The provided credentials are not correct. |
| 409 | See Polling Body| There already exists a job with the given id.|
| 429 | See Error Message Response Body| Not allowed to spawn a new job, because it would break the maximum job limit.|

## Polling Backup Status ##
This call request the status of the dedicated job identified by the given id.

Endpoint: GET /backup/{id}

### Backup Polling Response Body ###
Please be aware of the fact that the ``error_message`` field will not show up in the json, if it is empty. Same goes for fields that are dedicated to a specific backup destination type, which will be ignored if empty.

```json
{
    "status": "SUCCEEDED / FAILED / RUNNING",
    "message": "backup successfully carried out",
    "state": "finished / name of the current phase",
    "error_message": "contains message dedicated to the occuring error, will not show up if empty",

    "type": "S3 / SWIFT",

    "compression": true,
    "skip_storage": false,
    "useSSL": false,

    "endpoint" : "S3 endpoint",
    "bucket": "S3 bucket",

    "authUrl": "auth url for swift",
    "domain": "domain name for swift",
    "container_name": "name of the container for swift",
    "project_name": "name of the project for swift",

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

### Status Codes and their meaning ###
The backup agent intentionally returns the following status codes. Codes that differ are likely to be unexpected and not intended to be returned.

| Code | Body | Description |
| --- | --- | --- |
| 200 | See Polling Body | A matching job was found and its status is returned. |
| 400| - | No valid id was provided. |
| 401| See Simple response body| The provided credentials are not correct. |
| 404 | - | There exists no job for the given id.|


## Backup Job Deletion ##
This call requests the deletion of a result of a backup job. This should be done to either use the id again or free the space for the agent. This does NEITHER delete the generated backup file locally NOR in the cloud storage.

Endpoint: DELETE /backup

### Job Deletion Request Body ###

```json
{
	"id" : "778f038c-e1c5-11e8-9f32-f2801f1b9fd1"
}
```

### Status Codes and their meaning ###
The backup agent intentionally returns the following status codes. Codes that differ are likely to be unexpected and not intended to be returned.

| Code | Body | Description |
| --- | --- | --- |
| 200 | - | A matching job was found and deleted. |
| 400| - | The information in the body are not sufficient. |
| 401| See Simple response body| The provided credentials are not correct. |
| 410 | - | No matching job was found.|


## Other Response Bodies ##


### Simple Response Body ###
```json
{
    "message" : "simple response message"
}
```

### Error Message Response Body ###
```json
{
    "message": "descriptive message",
    "state": "state during error occurrence",
    "error_message": "message describing the occurred error"
}
```

---

<p align="center">
    <span ><a href="../README.md"><- README</a></span>
	    <span>&nbsp; | &nbsp;</span> 
    <span><a href="restore.md">Restore -></a></span>
</p>
