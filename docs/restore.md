# Table of Contents

1. [Trigger Restore](#trigger-restore)
2. [Polling Restore Status](#polling-restore-status)
3. [Restore Job Deletion](#restore-job-deletion)
4. [Other Response Bodies](#other-response-bodies)

---

# Restore #
The restore functionality consists of three calls: Triggering a restore, requesting its status and removing the job from the agent.

**Note:**
The backup agent intentionally returns the status codes described below. Codes that differ are likely to be unexpected and not intended to be returned.

## Trigger Restore ##
This call triggers an asynchronous restore procedure. 

Endpoint: PUT /restore

### Trigger Restore Request Body ###

Please note that objects in the parameters object can not have nested objects, arrays, lists, maps and so on inside. Only use simple types here as these values will be set as environment variables for the shell scripts to work with. **Be aware that the parameters are logged on the console of the agent! Do not use sensitive data, if you do not want to have it logged!**
Furthermore will the compression, skipSSL and skipStorage fields default to false, if no explicit value is present.

#### S3
```json
{
    "id" : "778f038c-e1c5-11e8-9f32-f2801f1b9fd1",
    "compression" : true,
    "encryption_key" : "example-encryption-key",
    "destination" : {
        "type": "S3",
        "skipStorage" : false,
        "filename": "filename",

        "bucket": "bucketName",
        "endpoint" : "http://custom.s3.endpoint -> defaults to the agents 'default_s3_endpoint' config parameter, if field is empty",
        "skipSSL": false,
        "authKey": "key",
        "authSecret": "secret",
    },
    "restore" : {
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
        "filename": "filename",
        
        "authUrl" : "auth url",
        "domain" : "domain name ",
        "container_name" : "name of the container",
        "project_name" : "name of the project == tenant",
        "username" : "swift username",
        "password" : "swift API key"
    },
    "restore" : {
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

### Response Codes and their meaning ###

| Code | Body | Description |
| --- | --- | --- |
| 201 | - | A restore was triggered and is getting run asynchronously. |
| 400| See Polling Body| The information in the body are not sufficient. |
| 401| See Simple Response Body | The provided credentials are not correct. |
| 409 | See Polling Body| There already exists a job with the given id.|
| 429 | See Error Message Response Body| Not allowed to spawn a new job, because it would break the maximum job limit.|


## Polling Restore Status ##
This call request the status of the dedicated job identified by the given id.

Endpoint: GET /restore/{id}

### Restore Polling Response Body ###
Please be aware of the fact that the ``error_message`` field will not show up in the json, if it is empty. 

```json
{
    "status": "SUCCEEDED / FAILED / RUNNING",
    "message": "restore successfully carried out",
    "state": "finished / name of the current phase",
    "error_message": "contains message dedicated to the occuring error, will not show up if empty",
    "type": "S3",
    "compression": true,
    "skip_storage": false,
    "useSSL": false,
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


### Status Codes and their meaning ###

| Code | Body | Description |
| --- | --- | --- |
| 200 | See Polling Body | A matching job was found and its status is returned. |
| 400| - | No valid id was provided. |
| 401| See Simple response body| The provided credentials are not correct. |
| 404 | - | There exists no job for the given id.|


## Restore Job Deletion ##
This call requests the deletion of a result of a restore job. This should be done to either use the id again or free the space for the agent. This does NEITHER delete the downloaded restore file locally NOR in the cloud storage.

Endpoint: DELETE /restore

### Job Deletion Request Body ###

```json
{
	"id" : "778f038c-e1c5-11e8-9f32-f2801f1b9fd1"
}
```

### Status Codes and their meaning ###

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
    <span ><a href="./backup.md"><- Backup</a></span>
	    <span>&nbsp; | &nbsp;</span> 
</p>