# Error Handling Guide

This document describes the error handling patterns used in the `confluent-go` package.

## Overview

The package uses structured error types to provide detailed information about API failures. All API errors are returned as `*api.Error`, which implements the `error` interface and provides helper methods for determining error types.

## Error Types

### `*api.Error` Structure

```go
type Error struct {
    Code      int                    // HTTP status code (400, 401, 404, etc.)
    ErrorCode string                 // Confluent-specific error code
    Message   string                 // Error message
    Details   map[string]interface{} // Additional error details
    Err       error                  // Underlying error
}
```

## Error Detection Methods

The `*api.Error` type provides several helper methods for detecting specific error types:

### Common Error Types

```go
// Returns true if this is a 400 Bad Request error
err.IsBadRequest()

// Returns true if this is a 401 Unauthorized error
err.IsUnauthorized()

// Returns true if this is a 403 Forbidden error
err.IsForbidden()

// Returns true if this is a 404 Not Found error
err.IsNotFound()

// Returns true if this is a 409 Conflict error
err.IsConflict()

// Returns true if this is a 429 Too Many Requests error
err.IsRateLimited()

// Returns true if this is a 500+ server error
err.IsInternalServerError()
```

### Retry Logic

```go
// Returns true if the error is retryable (rate limiting or server errors)
if err.IsRetryable() {
    // Implement exponential backoff
}

// Returns the Retry-After duration in seconds for rate limiting
retryAfter := err.RetryAfter()
if retryAfter > 0 {
    time.Sleep(time.Duration(retryAfter) * time.Second)
}
```

## Usage Examples

### Handling Authentication Failures

```go
clusters, err := clusterMgr.ListClusters(ctx, envID)
if err != nil {
    apiErr, ok := err.(*api.Error)
    if ok && apiErr.IsUnauthorized() {
        log.Fatal("Invalid API credentials")
    }
    return err
}
```

### Handling Not Found Errors

```go
cluster, err := clusterMgr.GetCluster(ctx, "invalid-id")
if err != nil {
    apiErr, ok := err.(*api.Error)
    if ok && apiErr.IsNotFound() {
        log.Printf("Cluster not found: %s", apiErr.Message)
        return nil // Gracefully handle missing resource
    }
    return err
}
```

### Handling Rate Limiting with Retry

```go
topic, err := topicMgr.GetTopic(ctx, clusterID, "my-topic")
if err != nil {
    apiErr, ok := err.(*api.Error)
    if ok && apiErr.IsRateLimited() {
        retryAfter := apiErr.RetryAfter()
        log.Printf("Rate limited, retrying after %d seconds", retryAfter)
        time.Sleep(time.Duration(retryAfter) * time.Second)
        // Retry the operation
        return topicMgr.GetTopic(ctx, clusterID, "my-topic")
    }
    return err
}
```

### Handling Validation Errors

```go
err := topicMgr.CreateTopic(ctx, clusterID, topic)
if err != nil {
    apiErr, ok := err.(*api.Error)
    if ok && apiErr.IsBadRequest() {
        log.Printf("Invalid topic configuration: %s", apiErr.Message)
        // Log Details for debugging
        for k, v := range apiErr.Details {
            log.Printf("  %s: %v", k, v)
        }
        return err
    }
    return err
}
```

### Handling Server Errors with Exponential Backoff

```go
import "time"

func createClusterWithRetry(ctx context.Context, mgr *resources.ClusterManager, 
    envID, name, clusterType, cloud, region string, maxRetries int) (*api.Cluster, error) {
    
    for attempt := 0; attempt < maxRetries; attempt++ {
        cluster, err := mgr.CreateCluster(ctx, envID, name, clusterType, cloud, region)
        if err == nil {
            return cluster, nil
        }
        
        apiErr, ok := err.(*api.Error)
        if !ok || !apiErr.IsRetryable() {
            return nil, err // Not retryable, fail immediately
        }
        
        // Exponential backoff: 1s, 2s, 4s, 8s, etc.
        backoff := time.Duration(math.Pow(2, float64(attempt))) * time.Second
        log.Printf("Attempt %d failed, retrying in %v: %s", 
            attempt+1, backoff, err)
        
        select {
        case <-time.After(backoff):
            continue
        case <-ctx.Done():
            return nil, ctx.Err()
        }
    }
    
    return nil, fmt.Errorf("failed to create cluster after %d attempts", maxRetries)
}
```

## Error Codes

The package uses the following error codes (from HTTP status codes):

| Code | Error Code | Description |
|------|-----------|-------------|
| 400 | `INVALID_REQUEST` | Invalid request parameters |
| 401 | `UNAUTHORIZED` | Authentication failed |
| 403 | `FORBIDDEN` | User lacks permissions |
| 404 | `NOT_FOUND` | Resource not found |
| 409 | `CONFLICT` | Resource already exists or state conflict |
| 429 | `RATE_LIMIT_EXCEEDED` | Too many requests |
| 500 | `INTERNAL_SERVER_ERROR` | Server error |
| 502 | `BAD_GATEWAY` | Gateway error |
| 503 | `SERVICE_UNAVAILABLE` | Service temporarily unavailable |
| 504 | `GATEWAY_TIMEOUT` | Gateway timeout |

## Logging and Monitoring

### Structured Logging Example

```go
import "log/slog"

err := clusterMgr.DeleteCluster(ctx, clusterID)
if err != nil {
    apiErr, ok := err.(*api.Error)
    if ok {
        slog.Error("API error",
            "code", apiErr.Code,
            "error_code", apiErr.ErrorCode,
            "message", apiErr.Message,
            "details", apiErr.Details,
        )
    } else {
        slog.Error("Unexpected error", "error", err)
    }
}
```

### Error Categorization

```go
func categorizeError(err error) string {
    if err == nil {
        return "success"
    }
    
    apiErr, ok := err.(*api.Error)
    if !ok {
        return "unexpected_error"
    }
    
    switch {
    case apiErr.IsUnauthorized():
        return "auth_failure"
    case apiErr.IsRateLimited():
        return "rate_limit"
    case apiErr.IsInternalServerError():
        return "server_error"
    case apiErr.IsNotFound():
        return "not_found"
    case apiErr.IsBadRequest():
        return "validation_error"
    default:
        return "api_error"
    }
}
```

## Best Practices

### 1. Always Check Error Type

Don't just check `if err != nil`. Use type assertion to access `*api.Error` for better error handling:

```go
// Good
if err != nil {
    apiErr, ok := err.(*api.Error)
    if ok && apiErr.IsNotFound() {
        // Handle not found
    } else if ok {
        // Handle other API errors
    } else {
        // Handle unexpected errors
    }
}

// Less effective
if err != nil {
    log.Fatal(err) // Lost error details
}
```

### 2. Handle Rate Limiting Gracefully

Rate limiting (429) is expected in production environments. Always implement retry logic:

```go
if apiErr.IsRateLimited() {
    retryAfter := apiErr.RetryAfter()
    backoff := time.Duration(retryAfter) * time.Second
    time.Sleep(backoff)
    // Retry the operation
}
```

### 3. Distinguish Between Retryable and Non-Retryable Errors

```go
if apiErr.IsRetryable() {
    // Retry with exponential backoff
} else if apiErr.IsUnauthorized() || apiErr.IsForbidden() {
    // Don't retry: auth issues need manual intervention
    return err
} else {
    // May or may not be retryable, log and decide
    return err
}
```

### 4. Provide Context in Error Messages

```go
cluster, err := clusterMgr.GetCluster(ctx, clusterID)
if err != nil {
    // Good: includes context
    return fmt.Errorf("failed to get cluster %s: %w", clusterID, err)
    
    // Less good: loses context
    // return err
}
```

### 5. Log Enough Detail for Debugging

```go
apiErr := err.(*api.Error)
log.Printf("API Error: Code=%d, ErrorCode=%s, Message=%s, Details=%+v",
    apiErr.Code, apiErr.ErrorCode, apiErr.Message, apiErr.Details)
```

## Testing Error Handling

The package includes comprehensive tests for error handling. See `pkg/client/client_test.go` for examples of testing error scenarios:

- `TestClientDo_Error_IsNotFound()` - 404 handling
- `TestClientDo_Error_IsUnauthorized()` - 401 handling  
- `TestClientDo_Error_IsRateLimited()` - 429 handling
- `TestClientDo_Error_IsBadRequest()` - 400 handling
- `TestClientDo_Error_IsServerError()` - 500+ handling

## Compatibility

The `*api.Error` type implements Go's error interface and works with:

- `errors.Is()` for error comparison
- `errors.As()` for type assertion
- `fmt.Errorf()` with `%w` verb for error wrapping
- Standard error handling patterns

## Additional Resources

- [pkg/api/errors.go](pkg/api/errors.go) - Error type definitions
- [pkg/client/client_test.go](pkg/client/client_test.go) - Error handling tests
- [REST_ARCHITECTURE.md](REST_ARCHITECTURE.md) - API architecture guide
