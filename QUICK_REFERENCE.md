# Quick Reference

## Installation

```bash
go get github.com/creiche/confluent-go
```

## Basic Setup

```go
import "github.com/creiche/confluent-go/pkg/client"

cfg := client.Config{BaseURL: "https://api.confluent.cloud", APIKey: "<key>", APISecret: "<secret>"}
c, err := client.NewClient(cfg)
if err != nil {
    log.Fatal(err)
}
```

## Common Operations

### Clusters
```go
mgr := resources.NewClusterManager(c)

// List
clusters, err := mgr.ListClusters(ctx)

// Get
cluster, err := mgr.GetCluster(ctx, "lkc-xyz")

// Create
cluster, err := mgr.CreateCluster(ctx, "name", "standard", "aws", "us-east-1")

// Delete
err := mgr.DeleteCluster(ctx, "lkc-xyz")
```

### Topics
```go
mgr := resources.NewTopicManager(c)

// List
topics, err := mgr.ListTopics(ctx, clusterID)

// Create
topic := api.Topic{
    Name: "my-topic",
    PartitionCount: 3,
    ReplicationFactor: 1,
}
err := mgr.CreateTopic(ctx, clusterID, topic)

// Delete
err := mgr.DeleteTopic(ctx, clusterID, "my-topic")
```

### Service Accounts & API Keys
```go
mgr := resources.NewServiceAccountManager(c)

// Create SA
sa, err := mgr.CreateServiceAccount(ctx, "my-sa", "description")

// Create API Key
key, err := mgr.CreateAPIKey(ctx, sa.ID, "key description")

// List Keys
keys, err := mgr.ListAPIKeys(ctx, sa.ID)

// Delete Key
err := mgr.DeleteAPIKey(ctx, key.ID)
```

### ACLs
```go
mgr := resources.NewACLManager(c)

// List
acls, err := mgr.ListACLs(ctx, clusterID)

// Create
acl := api.ACLBinding{
    Principal: "User:12345",
    Operation: "Read",
    ResourceType: "Topic",
    ResourceName: "*",
    PatternType: "PREFIXED",
}
err := mgr.CreateACL(ctx, clusterID, acl)
```

### Environments
```go
mgr := resources.NewEnvironmentManager(c)

// List
envs, err := mgr.ListEnvironments(ctx)

// Get
env, err := mgr.GetEnvironment(ctx, "env-xyz")

// Create
env, err := mgr.CreateEnvironment(ctx, "env-name", "display name")
```

### Connectors
```go
mgr := resources.NewConnectorManager(c)

// List
connectors, err := mgr.ListConnectors(ctx, envID, connectClusterID)

// Create
config := map[string]string{
    "connector.class": "io.confluent.connect.jdbc.JdbcSourceConnector",
    "tasks.max": "1",
    "connection.url": "jdbc:postgresql://localhost:5432/mydb",
    // ... other config
}
connector, err := mgr.CreateConnector(ctx, envID, connectClusterID, "my-connector", config)

// Get status
status, err := mgr.GetConnectorStatus(ctx, envID, connectClusterID, "my-connector")

// Pause/Resume
err := mgr.PauseConnector(ctx, envID, connectClusterID, "my-connector")
err := mgr.ResumeConnector(ctx, envID, connectClusterID, "my-connector")

// Restart
err := mgr.RestartConnector(ctx, envID, connectClusterID, "my-connector")

// Delete
err := mgr.DeleteConnector(ctx, envID, connectClusterID, "my-connector")
```

## Error Handling

```go
if err != nil {
    // Errors include context about what failed
    log.Printf("Operation failed: %v", err)
    
    // Use %w for error wrapping compatibility
    return fmt.Errorf("failed to reconcile: %w", err)
}
```

## Context Usage

```go
// With timeout
ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
defer cancel()

// With cancel
ctx, cancel := context.WithCancel(context.Background())
defer cancel()

result, err := mgr.ListClusters(ctx)
```

## Testing

```bash
# Run tests
make test

# Run with coverage
make test-cover

# Format code
make fmt

# Lint code
make lint
```

## File Locations

| Purpose | Location |
|---------|----------|
| Core client | `pkg/client/client.go` |
| Data types | `pkg/api/types.go` |
| Cluster ops | `pkg/resources/cluster.go` |
| Topic ops | `pkg/resources/topic.go` |
| Service account ops | `pkg/resources/service_account.go` |
| ACL ops | `pkg/resources/acl.go` |
| Environment ops | `pkg/resources/environment.go` |
| Connector ops | `pkg/resources/connector.go` |
| Examples | `cmd/examples/` |
| Tests | `test/` |

## Kubernetes Operator Pattern

```go
type MyReconciler struct {
    ConfluentClient *client.Client
}

func (r *MyReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
    topicMgr := resources.NewTopicManager(r.ConfluentClient)
    
    // Your reconciliation logic here
    topics, err := topicMgr.ListTopics(ctx, clusterID)
    if err != nil {
        return ctrl.Result{}, err
    }
    
    return ctrl.Result{}, nil
}
```

## Credentials

Store credentials in Kubernetes Secrets:

```go
// Service account API key becomes:
secret := &corev1.Secret{
    Data: map[string][]byte{
        "api-key": []byte(apiKey.ID),
        "api-secret": []byte(apiKey.Secret),
    },
}
```

## Resource Type Reference

- `api.Cluster` - Kafka cluster configuration
- `api.Topic` - Kafka topic definition
- `api.ServiceAccount` - Service account
- `api.APIKey` - API key credentials
- `api.ACLBinding` - Access control list entry
- `api.Environment` - Confluent environment
- `api.ConnectorConfig` - Kafka Connect connector configuration
- `api.ConnectorStatus` - Connector and task status information

## Common Issues

1. **Auth invalid**: Verify API Key/Secret and BaseURL
2. **Network issues**: Check connectivity and proxy settings
3. **JSON parsing errors**: Validate API responses and types
4. **Timeout errors**: Increase context timeout or check network connectivity

## Further Reading

- See `README.md` for full documentation
- See `PROJECT_STRUCTURE.md` for architecture details
- See `cmd/examples/main.go` for more examples
- See `CONTRIBUTING.md` for development guidelines
