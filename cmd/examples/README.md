# Examples

This directory contains example code demonstrating how to use the confluent-go package.

## Running the Examples

### Prerequisites
- Confluent CLI installed and in your PATH
- Valid Confluent Cloud or Platform credentials
- An existing environment and cluster

### Build and Run

```bash
# From the root directory
make build

# Run the example
./dist/example
```

## Example Breakdown

The `main.go` file demonstrates:

1. **List Resources**: Enumerating environments and clusters
2. **Service Account Management**: Creating, managing, and deleting service accounts with API keys
3. **Topic Management**: Creating, configuring, and listing topics
4. **ACL Management**: Managing access control lists

Each example is self-contained and includes error handling.

## Customizing Examples

Edit the function calls in `main.go` to:
- Use specific cluster IDs instead of the first cluster
- Create different types of resources
- Integrate with your own Kubernetes operator code

## Integration with Kubernetes Operators

For use in your Kubernetes operator, consider:

1. Creating a singleton client in your operator's main function
2. Passing the client to your reconcilers
3. Using context with timeouts for long-running operations
4. Implementing proper error handling and retry logic

Example reconciler pattern:

```go
type YourReconciler struct {
    ConfluentClient *client.Client
}

func (r *YourReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
    mgr := resources.NewClusterManager(r.ConfluentClient)
    // Your reconciliation logic here
}
```
