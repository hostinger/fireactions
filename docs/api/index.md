# Interacting with Fireactions via API

Fireactions provides a REST API for interacting with the server.

## Authentication

The API supports Basic Authentication when enabled. Include the username and password in the request using Basic Auth headers.

```bash
curl -u username:password http://localhost:8080/api/v1/pools
```

## Base URL

Versioned resource-management endpoints (e.g., routes handling pools, microvms, and reload) are prefixed with `/api/v1`, while utility endpoints like `/healthz` and `/version` are top-level routes without the `/api/v1` prefix.

## Endpoints

### Health Check

Check if the server is healthy and responding.

```http
GET /healthz
```

**Response:**
```json
{
  "status": "OK"
}
```

**Curl example:**
```bash
curl http://localhost:8080/healthz
```

### Version

Get server version information.

```http
GET /version
```

**Response:**
```json
{
  "version": "fireactions v1.0.0 (commit: abc123, date: 2026-01-23)"
}
```

**Curl example:**
```bash
curl http://localhost:8080/version
```

---

## Pool Management

### List all pools

Returns a list of all configured pools with their current status.

```http
GET /api/v1/pools
```

**Response:**
```json
{
  "pools": [
    {
      "name": "default",
      "replicas": 5,
      "current_replicas": 5,
      "desired_replicas": 5,
      "organization": "my-org",
      "group_id": 123,
      "labels": ["self-hosted", "linux"],
      "image": "ghcr.io/my-org/runner:latest",
      "status": {
        "state": "Active",
        "message": "Pool is active"
      }
    }
  ]
}
```

**Curl example:**
```bash
curl -u username:password http://localhost:8080/api/v1/pools
```

### Get pool details

Returns detailed information about a specific pool.

```http
GET /api/v1/pools/:pool
```

**Parameters:**
- `pool` (path) - Pool name

**Response:**
```json
{
  "pool": {
    "name": "default",
    "replicas": 5,
    "current_replicas": 5,
    "desired_replicas": 5,
    "organization": "my-org",
    "group_id": 123,
    "labels": ["self-hosted", "linux"],
    "image": "ghcr.io/my-org/runner:latest",
    "status": {
      "state": "Active",
      "message": "Pool is active"
    }
  }
}
```

**Curl example:**
```bash
curl -u username:password http://localhost:8080/api/v1/pools/default
```

### Scale a pool

Set the desired number of replicas for a pool. The pool will scale up or down to match the specified number.

```http
POST /api/v1/pools/:pool/scale
```

**Parameters:**
- `pool` (path) - Pool name

**Request Body:**
```json
{
  "replicas": 10
}
```

**Response:**
```json
{
  "message": "Pool replicas updated successfully"
}
```

**Curl examples:**

Scale to 10 replicas:
```bash
curl -X POST -u username:password \
  -H "Content-Type: application/json" \
  -d '{"replicas": 10}' \
  http://localhost:8080/api/v1/pools/default/scale
```

Scale down to 0:
```bash
curl -X POST -u username:password \
  -H "Content-Type: application/json" \
  -d '{"replicas": 0}' \
  http://localhost:8080/api/v1/pools/default/scale
```

**Note:** You can scale down to 0 to stop all VMs in a pool. The `replicas` field must be a non-negative integer.

### Pause a pool

Pause a pool to prevent it from scaling. Running VMs will continue to operate, but no new VMs will be started.

```http
POST /api/v1/pools/:pool/pause
```

**Parameters:**
- `pool` (path) - Pool name

**Response:**
```json
{
  "message": "Pool paused successfully"
}
```

**Curl example:**
```bash
curl -X POST -u username:password \
  http://localhost:8080/api/v1/pools/default/pause
```

### Resume a pool

Resume a paused pool to allow it to scale again.

```http
POST /api/v1/pools/:pool/resume
```

**Parameters:**
- `pool` (path) - Pool name

**Response:**
```json
{
  "message": "Pool resumed successfully"
}
```

**Curl example:**
```bash
curl -X POST -u username:password \
  http://localhost:8080/api/v1/pools/default/resume
```

---

## MicroVM Management

### List all MicroVMs

List all MicroVMs across all pools or within a specific pool.

```http
GET /api/v1/microvms
```

or

```http
GET /api/v1/pools/:pool/microvms
```

**Parameters:**
- `pool` (path, optional) - Pool name to filter by

**Response:**
```json
{
  "micro_vms": [
    {
      "vmid": "default-abc123",
      "pool": "default",
      "ip_addr": "192.168.1.100",
      "runner_id": 456789,
      "created_at": "2026-01-23T10:00:00Z"
    }
  ]
}
```

**Curl examples:**

List all VMs:
```bash
curl -u username:password http://localhost:8080/api/v1/microvms
```

List VMs in a specific pool:
```bash
curl -u username:password http://localhost:8080/api/v1/pools/default/microvms
```

### Get MicroVM details

Get detailed information about a specific MicroVM by its ID.

```http
GET /api/v1/microvms/:id
```

**Parameters:**
- `id` (path) - VM ID

**Response:**
```json
{
  "micro_vm": {
    "vmid": "default-abc123",
    "pool": "default",
    "ip_addr": "192.168.1.100",
    "runner_id": 456789,
    "created_at": "2026-01-23T10:00:00Z"
  }
}
```

**Curl example:**
```bash
curl -u username:password http://localhost:8080/api/v1/microvms/default-abc123
```

---

## Configuration Management

### Reload configuration

Reload the server configuration from disk without downtime. This allows updating pool configurations without restarting the server or interrupting running VMs.

```http
POST /api/v1/reload
```

**Response:**
```json
{
  "message": "Pools reloaded successfully"
}
```

**Curl example:**
```bash
curl -X POST -u username:password \
  http://localhost:8080/api/v1/reload
```

---

## Error Responses

All endpoints may return error responses in the following format:

```json
{
  "error": "Error message describing what went wrong"
}
```

Common HTTP status codes:
- `200 OK` - Request succeeded
- `400 Bad Request` - Invalid request parameters
- `401 Unauthorized` - Authentication required or failed
- `404 Not Found` - Resource not found (pool or VM doesn't exist)
- `500 Internal Server Error` - Server error

## Examples

### Complete Workflow

```bash
# Set credentials
USERNAME="admin"
PASSWORD="secret"
BASE_URL="http://localhost:8080"

# Check server health
curl $BASE_URL/healthz

# List all pools
curl -u $USERNAME:$PASSWORD $BASE_URL/api/v1/pools

# Get specific pool
curl -u $USERNAME:$PASSWORD $BASE_URL/api/v1/pools/production

# Scale pool to 10 replicas
curl -X POST -u $USERNAME:$PASSWORD \
  -H "Content-Type: application/json" \
  -d '{"replicas": 10}' \
  $BASE_URL/api/v1/pools/production/scale

# List all VMs
curl -u $USERNAME:$PASSWORD $BASE_URL/api/v1/microvms

# Get VM details
curl -u $USERNAME:$PASSWORD $BASE_URL/api/v1/microvms/production-abc123

# Pause pool
curl -X POST -u $USERNAME:$PASSWORD \
  $BASE_URL/api/v1/pools/production/pause

# Resume pool
curl -X POST -u $USERNAME:$PASSWORD \
  $BASE_URL/api/v1/pools/production/resume

# Reload configuration
curl -X POST -u $USERNAME:$PASSWORD \
  $BASE_URL/api/v1/reload
```
