# Lock Management

[![Build Status](https://drone-ci-kramergroup.serveo.net/api/badges/kramergroup-workflows/locks/status.svg)](https://drone-ci-kramergroup.serveo.net/kramergroup-workflows/locks)

This repository contains client-side code of a simple lock service. The locks can be used in many situation, but
it is primarily intended to

- provide capabilties to resume suspended Argo workflows.

A hosted API implementation maintains locks. These can be managed (created, released, deleted) using the `lockcli` client or by directly calling the API endpoint defined below.

An agent (`lockagt`) handles interaction with Argo and has to run in the same Kubernetes cluster as the Argo workflow manager (or be given access to the clusters API through proper configuration of the `kubeconfig` file). The `lockagt` uses polling and can be used behind firewalls. This allows to resume Argo workflows in clusters behind firewalls as a consequence of external events.

## Life cycle of a lock

A lock can go through these life cycle stages:

1. Creation: a new lock with status `locked` is created
2. Release: upon release, a lock status changes to `released`. A released lock triggers action from `lockagt`
3. Deletion: once the agent has acted successfully on a `released` lock, the lock is deleted

## API Definition

The lock management defines the following REST API:

### Create a lock

```http
POST /lock HTTP/1.1
Content-Type: application/json

{
  "workflow": "workflow-name",
  "namespace": "namespace",
}
```

A success response will look like this:

```http
HTTP/1.1 200 Success

{
  "id": "e9205d4a-2545-4c58-86ae-d0e7b3ecec23",
  "status": "locked",
  "created": "2019-05-28T08:30:24.595Z",
  "lastChange": "2019-05-28T08:30:24.595Z"
}
```

### Get lock information

```http
GET /lock/?id=<lock-id> HTTP/1.1
```

A successful response will look like this:

```http
HTTP/1.1 200 Success

{
  "id": "e9205d4a-2545-4c58-86ae-d0e7b3ecec23",
  "status": "locked",
  "created": "2019-05-28T08:30:24.595Z",
  "lastChange": "2019-05-28T08:30:24.595Z"
}
```

### Release a lock

```http
PATCH /lock/?id=<lock-id> HTTP/1.1
```

A successful response will look like this:

```http
HTTP/1.1 200 Success

{
  "id": "e9205d4a-2545-4c58-86ae-d0e7b3ecec23",
  "status": "released",
  "created": "2019-05-28T08:30:24.595Z",
  "lastChange": "2019-05-28T08:30:24.595Z"
}
```

### Delete a lock

```http
DELETE /lock/?id=<lock-id> HTTP/1.1
```

A successful response will look like this:

```http
HTTP/1.1 200 Success

{}
```

## Lock Client (lockcli)

The `lockcli` provides a simple command-line interface for the lock manager. A typical call looks like

```bash
lockcli --api-endpoint URL [get|create|delete|release] [command parameters]
```

| Command   | Parameter            | Description                                                            |
| --------- | -------------------- | ---------------------------------------------------------------------- |
| `get`     | `id`                 | Retrieve the lock with `id`; returns a JSON representation of the lock |
| `release` | `id`                 | Release the lock with `id`                                             |
| `delete`  | `id`                 | Delete the lock with `id`                                              |
| `create`  | `workflow namespace` | Create a new lock for `workflow` in `namespace`                        |

> The API endpoint URL can also be defined via the `API_ENDPOINT` environment variable.

## Lock Agent (lockagt)

The lock agent syncs suspended workflows and locks. It periodically polls the lock API for released locks and resumes associated workflows. Once a workflow is resuming, the lock is deleted.

## Configuration

| Environment variable | Flag             | Description                                  |
| -------------------- | ---------------- | -------------------------------------------- |
| `API_ENDPOINT`       | `--api-endpoint` | The lock API endpoint                        |
|                      | `--kubeconfig`   | Location of the `kubeconfig` file (optional) |
