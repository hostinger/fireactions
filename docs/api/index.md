# Interacting with Fireactions via API

Fireactions provides a REST API for interacting with the server.

## Authentication

The API requires a valid API key to be included in the `X-API-Key` header of the request. The API key can be obtained from the Fireactions server by setting the `api_key` configuration option.

## Endpoints

### Get a list of all configured pools

This endpoint returns a list of all pools that have been configured on the server.

```http
GET /api/v1/pools
```

Curl example:

```bash
curl -H "X-API-Key: <API_KEY>" http://localhost:8080/api/v1/pools
```

### Get the status of a pool

This endpoint returns details of a specific pool.

```http
GET /api/v1/pools/:pool
```

Curl example:

```bash
curl -H "X-API-Key: <API_KEY>" http://localhost:8080/api/v1/pools/my-pool
```

### Scale a pool

This endpoint scales a pool up by 1 instance.

```http
POST /api/v1/pools/:pool/scale
```

Curl example:

```bash
curl -X POST -H "X-API-Key: <API_KEY>" http://localhost:8080/api/v1/pools/my-pool/scale
```

### Pause a pool

This endpoint pauses a pool, preventing it from scaling.

```http
POST /api/v1/pools/:pool/pause
```

Curl example:

```bash
curl -X POST -H "X-API-Key: <API_KEY>" http://localhost:8080/api/v1/pools/my-pool/pause
```

### Resume a pool

This endpoint resumes a pool, allowing it to scale.

```http
POST /api/v1/pools/:pool/resume
```

Curl example:

```bash
curl -X POST -H "X-API-Key: <API_KEY>" http://localhost:8080/api/v1/pools/my-pool/resume
```
