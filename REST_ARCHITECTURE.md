# REST-Based Architecture Guide

This document explains the REST-based architecture of confluent-go and how it works.

## Overview

This package makes direct HTTP REST API calls to:
- **Confluent Cloud APIs** (https://api.confluent.cloud)
- **Confluent Platform APIs** (self-hosted installations)

The Confluent CLI served as a reference for understanding:
- What resources need to be managed (Clusters, Topics, Service Accounts, etc.)
- What operations are needed on those resources (CRUD operations)
- How to structure API requests and responses

## API Endpoints Used

### Cluster Management (CMK API v2)
- `GET /cmk/v2/clusters?environment={envId}` - List clusters in environment
- `GET /cmk/v2/clusters/{clusterId}` - Get cluster details
- `POST /cmk/v2/clusters` - Create cluster
- `PATCH /cmk/v2/clusters/{clusterId}` - Update cluster
- `DELETE /cmk/v2/clusters/{clusterId}` - Delete cluster

### Topic Management (Kafka API v3)
- `GET /kafka/v3/clusters/{clusterId}/topics` - List topics
- `GET /kafka/v3/clusters/{clusterId}/topics/{topicName}` - Get topic
- `POST /kafka/v3/clusters/{clusterId}/topics` - Create topic
- `PATCH /kafka/v3/clusters/{clusterId}/topics/{topicName}` - Update topic
- `DELETE /kafka/v3/clusters/{clusterId}/topics/{topicName}` - Delete topic
- `GET /kafka/v3/clusters/{clusterId}/topics/{topicName}/configs` - Get configs
- `PUT /kafka/v3/clusters/{clusterId}/topics/{topicName}/configs` - Update configs

### Service Account Management (IAM API v2)
- `GET /iam/v2/service-accounts` - List service accounts
- `GET /iam/v2/service-accounts/{saId}` - Get service account
- `POST /iam/v2/service-accounts` - Create service account
- `PATCH /iam/v2/service-accounts/{saId}` - Update service account
- `DELETE /iam/v2/service-accounts/{saId}` - Delete service account

### API Key Management (IAM API v2)
- `GET /iam/v2/api-keys` - List API keys
- `GET /iam/v2/api-keys?owner={saId}` - List keys for service account
- `POST /iam/v2/api-keys` - Create API key
- `DELETE /iam/v2/api-keys/{keyId}` - Delete API key

### ACL Management (Kafka API v3)
- `GET /kafka/v3/clusters/{clusterId}/acls` - List ACLs
- `POST /kafka/v3/clusters/{clusterId}/acls` - Create ACL
- `DELETE /kafka/v3/clusters/{clusterId}/acls` - Delete ACL

### Environment Management (Org API v2)
- `GET /org/v2/environments` - List environments
- `GET /org/v2/environments/{envId}` - Get environment
- `POST /org/v2/environments` - Create environment
- `PATCH /org/v2/environments/{envId}` - Update environment
- `DELETE /org/v2/environments/{envId}` - Delete environment

### Schema Registry (SR API v1)
- `GET /subjects` - List all subjects
- `GET /subjects/{subject}/versions` - List versions for a subject
- `GET /subjects/{subject}/versions/latest` - Get latest schema version
- `GET /subjects/{subject}/versions/{version}` - Get specific schema version
- `GET /schemas/ids/{id}` - Get schema by global ID
- `POST /subjects/{subject}/versions` - Register a new schema
- `DELETE /subjects/{subject}?permanent=true` - Delete subject (soft/hard)
- `POST /compatibility/subjects/{subject}/versions/latest` - Test compatibility
- `GET /config` - Get global compatibility level
- `PUT /config` - Set global compatibility level
- `GET /config/{subject}` - Get subject compatibility level
- `PUT /config/{subject}` - Set subject compatibility level
- `GET /mode` - Get global mode
- `PUT /mode` - Set global mode
- `GET /mode/{subject}` - Get subject mode
- `PUT /mode/{subject}` - Set subject mode

**Configuration:**
- Base path: default `"/schema-registry/v1"`
- Cloud BaseURL: `https://api.confluent.cloud`
- On-prem BaseURL: SR URL (e.g., `https://sr.example.com`)
- Types: prefer constants `SchemaTypeAvro|JSON|Protobuf`
- Compatibility: prefer constants `CompatNone|Backward|BackwardTransitive|Forward|ForwardTransitive|Full|FullTransitive`
- Mode: prefer constants `ModeReadWrite|ReadOnly|Import`

## Authentication

All API calls use HTTP Basic Authentication with:
- **Username**: API Key ID
- **Password**: API Key Secret

```go
cfg := client.Config{
    BaseURL: "https://api.confluent.cloud",
    APIKey: "your-api-key-id",
    APISecret: "your-api-key-secret",
}
```

## Request/Response Pattern

### Request Structure
```go
req := client.Request{
    Method: "GET",
    Path: "/kafka/v3/clusters/lkc-abc123/topics",
    Body: nil,
    Headers: map[string]string{}, // optional custom headers
}
resp, err := client.Do(ctx, req)
```

### Response Structure
```go
type Response struct {
    StatusCode int                 // HTTP status code
    Body       []byte              // Raw response body
    Headers    http.Header         // Response headers
}

// Decode JSON response
var result struct {
    Data []Topic `json:"data"`
}
resp.DecodeJSON(&result)
```

## Error Handling

The client returns errors with context:

```go
if err != nil {
    // err contains HTTP status and response body details
    // e.g. "API error (status 400): Topic already exists"
}
```

## Resource Managers

Each resource manager follows this pattern:

```go
manager := resources.NewXyzManager(client)

// List operations
items, err := manager.ListXyz(ctx, ...)

// Get operations
item, err := manager.GetXyz(ctx, id)

// Create operations
item, err := manager.CreateXyz(ctx, ...)

// Update operations
item, err := manager.UpdateXyz(ctx, id, ...)

// Delete operations
err := manager.DeleteXyz(ctx, id)
```

## Advantages of REST-Based Approach

✅ **No CLI Dependency** - Runs anywhere without Confluent CLI installed  
✅ **Direct API Access** - Lower latency with native HTTP calls  
✅ **Programmatic** - Native Go code, better integration with Go projects  
✅ **Type-Safe** - Strong typing with JSON marshaling  
✅ **Testable** - Easy to mock HTTP calls for testing  
✅ **Scalable** - Can handle concurrent requests efficiently  

## Platform Compatibility

### Confluent Cloud
- BaseURL: `https://api.confluent.cloud`
- Requires Confluent Cloud API key
- Supports all resources and operations

### Confluent Platform (Self-Hosted)
- BaseURL: `https://your-platform-host:port`
- Uses Platform REST APIs (may differ from Cloud APIs)
- Check your Platform documentation for available endpoints

## Architecture Overview

| Aspect | Legacy Approach | REST-Based |
|--------|-----------|-----------|
| External Dependencies | External tooling required | None (pure Go) |
| Performance | Slower (process startup overhead) | Faster (direct HTTP) |
| Concurrency | Limited | High (concurrent requests) |
| Network | Indirect | Direct HTTP with control |
| Error Messages | String parsing | Structured API responses |

## Configuration Best Practices

### From Environment Variables
```go
cfg := client.Config{
    BaseURL: os.Getenv("CONFLUENT_BASE_URL"),
    APIKey: os.Getenv("CONFLUENT_API_KEY"),
    APISecret: os.Getenv("CONFLUENT_API_SECRET"),
}
```

### From Kubernetes Secret
```go
cfg := client.Config{
    BaseURL: "https://api.confluent.cloud",
    APIKey: string(secret.Data["api-key"]),
    APISecret: string(secret.Data["api-secret"]),
}
```

### With Custom HTTP Client (for advanced use cases)
```go
httpClient := &http.Client{
    Timeout: 30 * time.Second,
    Transport: &http.Transport{
        MaxConnsPerHost: 10,
    },
}

cfg := client.Config{
    BaseURL: "https://api.confluent.cloud",
    APIKey: "...",
    APISecret: "...",
    HTTPClient: httpClient,
}
```

## Advanced Features

### Custom Headers
```go
req := client.Request{
    Method: "GET",
    Path: "/cmk/v2/clusters",
    Headers: map[string]string{
        "X-Custom-Header": "value",
    },
}
```

### Request/Response Context
All operations support Go context:
```go
ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
defer cancel()

clusters, err := clusterMgr.ListClusters(ctx, envId)
```

## Troubleshooting

### Authentication Errors (401)
- Verify API key and secret are correct
- Check API key hasn't expired
- Ensure proper URL format

### Resource Not Found (404)
- Verify resource ID is correct
- Check environment/cluster context
- May indicate insufficient permissions

### Rate Limiting (429)
- Implement backoff and retry logic
- Consider batch operations
- Check Confluent Cloud API quotas

### Connection Errors
- Verify network connectivity
- Check firewall rules (if self-hosted)
- Validate BaseURL format

## Next Steps

1. See **README.md** for quick start guide
2. Check **QUICK_REFERENCE.md** for common operations
3. Review **cmd/examples/main.go** for practical examples
4. Explore individual manager implementations in **pkg/resources/**
