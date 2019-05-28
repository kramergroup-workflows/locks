# Lock Agent

The lock agent syncs suspended workflows and locks. It periodically polls the lock API for released locks and resumes associated workflows. Once a workflow is resuming, the lock is deleted.

## Configuration 

| Environment variable | Description           |
| -------------------- | --------------------- |
| `API_ENDPOINT`       | The lock API endpoint |

