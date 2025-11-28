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

## Schema Registry Error Handling

Schema Registry APIs use specific error codes to indicate different failure conditions. The `schemaregistry` package provides helper functions that work directly with `*api.Error` to check for specific SR error codes.

### Schema Registry Error Codes

| Code | Constant | Description |
|------|----------|-------------|
| 40401 | `ErrorCodeSubjectNotFound` | Subject does not exist |
| 40402 | `ErrorCodeVersionNotFound` | Schema version not found |
| 40403 | `ErrorCodeSchemaNotFound` | Schema ID not found |
| 40404 | `ErrorCodeSubjectSoftDeleted` | Subject was soft-deleted |
| 42201 | `ErrorCodeInvalidSchema` | Schema syntax is invalid |
| 42202 | `ErrorCodeInvalidSubject` | Subject name is invalid |
| 42203 | `ErrorCodeInvalidCompatibility` | Compatibility level is invalid |
| 42204 | `ErrorCodeInvalidMode` | Mode value is invalid |
| 409 | `ErrorCodeIncompatibleSchema` | Schema is incompatible with existing version |

### Helper Functions

The `schemaregistry` package provides helper functions to check specific error types:

```go
import "github.com/creiche/confluent-go/pkg/schemaregistry"

// Check for subject not found (40401)
if schemaregistry.IsSubjectNotFound(err) {
    log.Printf("Subject does not exist: %v", err)
}

// Check for version not found (40402)
if schemaregistry.IsVersionNotFound(err) {
    log.Printf("Schema version not found: %v", err)
}

// Check for schema ID not found (40403)
if schemaregistry.IsSchemaNotFound(err) {
    log.Printf("Schema ID not found: %v", err)
}

// Check for soft-deleted subject (40404)
if schemaregistry.IsSubjectSoftDeleted(err) {
    log.Printf("Subject was soft-deleted, use permanent=true: %v", err)
}

// Check for invalid schema syntax (42201)
if schemaregistry.IsInvalidSchema(err) {
    log.Printf("Schema syntax is invalid: %v", err)
}

// Check for invalid subject name (42202)
if schemaregistry.IsInvalidSubject(err) {
    log.Printf("Subject name is invalid: %v", err)
}

// Check for incompatible schema (409)
if schemaregistry.IsIncompatibleSchema(err) {
    log.Printf("Schema is incompatible with existing version: %v", err)
}

// Check for invalid compatibility level (42203)
if schemaregistry.IsInvalidCompatibility(err) {
    log.Printf("Invalid compatibility level: %v", err)
}

// Check for invalid mode (42204)
if schemaregistry.IsInvalidMode(err) {
    log.Printf("Invalid mode: %v", err)
}
```

### Schema Registry Error Handling Examples

#### Handling Missing Subjects

```go
schema, err := srMgr.GetLatestSchema(ctx, "my-subject")
if err != nil {
    if schemaregistry.IsSubjectNotFound(err) {
        log.Printf("Subject 'my-subject' does not exist, creating...")
        // Register initial schema
        id, err := srMgr.RegisterSchema(ctx, "my-subject", schemaregistry.RegisterRequest{
            Schema:     mySchemaJSON,
            SchemaType: schemaregistry.SchemaTypeAvro,
        })
        // Handle registration...
    } else {
        return fmt.Errorf("failed to get schema: %w", err)
    }
}
```

#### Handling Schema Validation Errors

```go
id, err := srMgr.RegisterSchema(ctx, subject, schemaregistry.RegisterRequest{
    Schema:     schemaJSON,
    SchemaType: schemaregistry.SchemaTypeAvro,
})
if err != nil {
    if schemaregistry.IsInvalidSchema(err) {
        log.Printf("Schema validation failed: %v", err)
        // Log the schema for debugging
        log.Printf("Invalid schema: %s", schemaJSON)
        return fmt.Errorf("schema syntax error: %w", err)
    } else if schemaregistry.IsIncompatibleSchema(err) {
        log.Printf("Schema is incompatible with existing version: %v", err)
        // Could test compatibility first in the future
        return fmt.Errorf("schema compatibility error: %w", err)
    }
    return err
}
```

#### Handling Soft-Deleted Subjects

```go
err := srMgr.DeleteSubject(ctx, "old-subject", false)
if err != nil {
    log.Printf("Soft delete failed: %v", err)
    return err
}

// Later, trying to access...
schema, err := srMgr.GetLatestSchema(ctx, "old-subject")
if err != nil {
    if schemaregistry.IsSubjectSoftDeleted(err) {
        log.Printf("Subject was soft-deleted, performing hard delete...")
        // Hard delete to fully remove
        err = srMgr.DeleteSubject(ctx, "old-subject", true)
        if err != nil {
            return fmt.Errorf("hard delete failed: %w", err)
        }
    }
}
```

#### Testing Schema Compatibility Before Registration

```go
compatible, err := srMgr.TestCompatibility(ctx, subject, schemaregistry.RegisterRequest{
    Schema:     newSchemaJSON,
    SchemaType: schemaregistry.SchemaTypeAvro,
})
if err != nil {
    if schemaregistry.IsSubjectNotFound(err) {
        // No existing schema, safe to register
        log.Printf("No existing schema, registering initial version...")
    } else {
        return fmt.Errorf("compatibility check failed: %w", err)
    }
}

if !compatible {
    return fmt.Errorf("schema would be incompatible with current version")
}

// Safe to register
id, err := srMgr.RegisterSchema(ctx, subject, schemaregistry.RegisterRequest{
    Schema:     newSchemaJSON,
    SchemaType: schemaregistry.SchemaTypeAvro,
})
```

#### Validating Compatibility Levels

```go
err := srMgr.SetGlobalCompatibility(ctx, "CUSTOM_LEVEL")
if err != nil {
    if schemaregistry.IsInvalidCompatibility(err) {
        log.Printf("Invalid compatibility level, using BACKWARD instead")
        err = srMgr.SetGlobalCompatibility(ctx, schemaregistry.CompatBackward)
        if err != nil {
            return fmt.Errorf("failed to set compatibility: %w", err)
        }
    } else {
        return err
    }
}
```

### Working with Schema Registry Errors

All Schema Registry manager methods return `*api.Error` directly. The error helpers work by extracting the `error_code` field from the error's `Details` map:

```go
schema, err := srMgr.GetLatestSchema(ctx, "my-subject")
if err != nil {
    // Use helper functions directly on api.Error
    if schemaregistry.IsSubjectNotFound(err) {
        log.Printf("Subject not found")
    }
    
    // Or access the underlying api.Error
    var apiErr *api.Error
    if errors.As(err, &apiErr) {
        log.Printf("HTTP Status: %d", apiErr.Code)
        log.Printf("Message: %s", apiErr.Message)
        
        // Can also use api.Error helpers
        if apiErr.IsNotFound() {
            // Handle 404 generically
        }
    }
    
    // Or get the SR error code directly
    if code, ok := schemaregistry.GetSRCode(err); ok {
        log.Printf("SR Error Code: %d", code)
    }
}
```

### Best Practices for Schema Registry

1. **Always check for subject existence before operations**:
   ```go
   if schemaregistry.IsSubjectNotFound(err) {
       // Handle missing subject
   }
   ```

2. **Test compatibility before registration**:
   ```go
   compatible, err := srMgr.TestCompatibility(ctx, subject, req)
   if !compatible {
       // Handle incompatibility
   }
   }
   ```

3. **Validate schemas before registration**:
   ```go
   if schemaregistry.IsInvalidSchema(err) {
       // Log schema for debugging
   }
   ```

4. **Handle soft vs hard deletes explicitly**:
   ```go
   if schemaregistry.IsSubjectSoftDeleted(err) {
       // Use permanent=true for hard delete
       srMgr.DeleteSubject(ctx, subject, true)
   }
   ```

5. **Use typed constants for compatibility levels and schema types**:
   ```go
   // Good
   schemaregistry.CompatBackward
   schemaregistry.SchemaTypeAvro
   
   // Avoid
   "BACKWARD"
   "AVRO"
   ```
