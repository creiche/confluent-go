# Project Structure Overview

This document provides a high-level overview of the confluent-go project structure and how to use it.

## Directory Structure

```
confluent-go/
├── pkg/                          # Main package directory
│   ├── api/                       # Data types and interfaces
│   │   └── types.go               # All resource type definitions
│   ├── client/                    # Core client implementation
│   │   └── client.go              # Main Confluent client wrapper
│   └── resources/                 # Resource-specific managers
│       ├── cluster.go             # Cluster management
│       ├── topic.go               # Topic management
│       ├── service_account.go     # Service account & API key management
│       ├── acl.go                 # Access control list management
│       └── environment.go         # Environment management
├── cmd/
│   └── examples/                  # Example code and patterns
│       ├── main.go                # Basic usage examples
│       ├── operator_pattern.go    # Kubernetes operator pattern
│       └── README.md              # Example documentation
├── test/                          # Unit and integration tests
├── Makefile                       # Build and test targets
├── go.mod                         # Go module definition
├── go.sum                         # Go dependencies
├── README.md                      # Main documentation
├── CONTRIBUTING.md               # Contribution guidelines
├── LICENSE                        # MIT License
└── .gitignore                     # Git ignore rules
```

## Key Design Patterns

### 1. Manager Pattern
Each resource type has a dedicated manager that handles operations for that resource:

```go
// Create manager
topicMgr := resources.NewTopicManager(client)

// Use manager methods
topics, err := topicMgr.ListTopics(ctx, clusterID)
topic, err := topicMgr.GetTopic(ctx, clusterID, topicName)
err := topicMgr.CreateTopic(ctx, clusterID, topic)
```

### 2. Context-Aware Operations
All operations accept a context for timeout and cancellation:

```go
ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
defer cancel()

topic, err := topicMgr.GetTopic(ctx, clusterID, topicName)
```

### 3. Error Wrapping
Errors include context and are wrapped with `fmt.Errorf`:

```go
if err := topicMgr.CreateTopic(ctx, clusterID, topic); err != nil {
    // err contains detailed information about what failed
    log.Printf("Failed to create topic: %v", err)
}
```

## Resource Types Supported

### Core Resources
- **Cluster**: Create, list, describe, update, delete Kafka clusters
- **Topic**: Create, list, describe, configure, and delete topics
- **ServiceAccount**: Manage service accounts and their API keys
- **ACL**: Create and manage access control lists
- **Environment**: Create and manage environments

### Type Definitions
All resource types are defined in `pkg/api/types.go` and include JSON tags for unmarshaling CLI output.

## Usage Patterns

### Pattern 1: Standalone Usage
```go
cfg := client.Config{CliPath: "confluent"}
c, _ := client.NewClient(cfg)

mgr := resources.NewClusterManager(c)
clusters, _ := mgr.ListClusters(context.Background())
```

### Pattern 2: Kubernetes Operator
```go
type MyReconciler struct {
    ConfluentClient *client.Client
}

func (r *MyReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
    mgr := resources.NewTopicManager(r.ConfluentClient)
    // Reconciliation logic
}
```

### Pattern 3: Configuration Management
```go
// Store in environment or config file
config := client.Config{
    APIKey:    os.Getenv("CONFLUENT_API_KEY"),
    APISecret: os.Getenv("CONFLUENT_API_SECRET"),
    CliPath:   "/usr/local/bin/confluent",
}
```

## Extending the Package

### Adding a New Resource Manager
1. Define types in `pkg/api/types.go`
2. Create `pkg/resources/new_resource.go`
3. Implement manager with Create, Read, Update, Delete operations
4. Add tests in `test/`
5. Document in README and examples

Example:
```go
// pkg/resources/my_resource.go
package resources

type MyResourceManager struct {
    client *client.Client
}

func NewMyResourceManager(c *client.Client) *MyResourceManager {
    return &MyResourceManager{client: c}
}

func (m *MyResourceManager) List(ctx context.Context) ([]api.MyResource, error) {
    // Implementation
}
```

## Integration with Kubernetes

### Using with controller-runtime
```go
import "sigs.k8s.io/controller-runtime/pkg/client"

// In your reconciler
confluent := resources.NewTopicManager(c.ConfluentClient)
// Use confluent client for management
```

### Storing Credentials in Secrets
```go
// After creating service account and API key
secret := &corev1.Secret{
    ObjectMeta: metav1.ObjectMeta{Name: "confluent-creds"},
    Data: map[string][]byte{
        "api-key":    []byte(apiKey.ID),
        "api-secret": []byte(apiKey.Secret),
    },
}
```

## CLI Command Mapping

The package wraps these Confluent CLI commands:

| Resource | Operation | CLI Command |
|----------|-----------|-------------|
| Cluster | List | `confluent kafka cluster list` |
| Cluster | Describe | `confluent kafka cluster describe` |
| Topic | List | `confluent kafka topic list` |
| Topic | Create | `confluent kafka topic create` |
| ServiceAccount | List | `confluent service-account list` |
| ServiceAccount | Create | `confluent service-account create` |
| APIKey | Create | `confluent api-key create` |
| ACL | List | `confluent kafka acl list` |
| ACL | Create | `confluent kafka acl create` |

## Error Handling Strategy

All errors include:
- Operation context (what was attempted)
- Underlying error details
- Wrapped with `fmt.Errorf` for easy debugging

Example:
```
Error: failed to create topic: failed to execute confluent command: exit status 1
output: Topic already exists
```

## Performance Considerations

1. **CLI Invocation**: Each operation invokes the CLI, which may have startup overhead
2. **JSON Parsing**: Output is parsed from JSON for type safety
3. **Context Usage**: Always set reasonable timeouts for CLI operations
4. **Batch Operations**: Implement custom batch logic in your operator if needed

## Security Considerations

1. **Credentials**: Store API keys in Kubernetes Secrets, not in code
2. **RBAC**: Use Kubernetes RBAC to control operator permissions
3. **Service Accounts**: Create least-privilege service accounts for operators
4. **ACL Management**: Always validate and restrict ACL rules

## Testing

Run tests with:
```bash
make test        # Run all tests
make test-cover  # Run with coverage
```

Tests should follow Go conventions and include error cases.

## Next Steps

1. Review `cmd/examples/main.go` for basic usage
2. Review `cmd/examples/operator_pattern.go` for operator patterns
3. Read the main README.md for detailed API documentation
4. Check out the `pkg/api/types.go` to understand available resource types
5. Explore individual manager files to understand available operations
