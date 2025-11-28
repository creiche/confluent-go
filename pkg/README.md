# confluent-go

This directory contains the Go packages for the confluent-go project.

## Package Organization

### `api/`
Contains all data type definitions and interfaces for Confluent resources. These types are used throughout the client.

### `client/`
Contains the core REST client implementation for Confluent APIs. This is the entry point for all operations.

### `resources/`
Contains resource-specific managers for different Confluent resource types:
- `cluster.go` - Cluster management (CRUD operations)
- `topic.go` - Topic management (create, delete, configure)
- `service_account.go` - Service account and API key management
- `acl.go` - Access control list management
- `environment.go` - Environment management

## Usage

All resource managers follow the same pattern:

```go
manager := resources.NewXyzManager(client)
result, err := manager.MethodName(ctx, args...)
```

## Adding New Resource Managers

To add support for a new resource type:

1. Create a new file in `pkg/resources/` named after the resource type
2. Create a manager struct that holds a pointer to the `client.Client`
3. Implement methods following the existing patterns
4. Document the manager in this file

## Testing

Each resource manager should have corresponding unit tests in the `test/` directory.
