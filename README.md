# confluent-go

![Go Version](https://img.shields.io/badge/go-1.22%2B-blue.svg)
![License](https://img.shields.io/github/license/creiche/confluent-go)
![Build](https://github.com/creiche/confluent-go/actions/workflows/build.yml/badge.svg)
![Tests](https://github.com/creiche/confluent-go/actions/workflows/test.yml/badge.svg)
![PR Checks](https://github.com/creiche/confluent-go/actions/workflows/pr-checks.yml/badge.svg)
![Security](https://github.com/creiche/confluent-go/actions/workflows/security.yml/badge.svg)
![Release](https://github.com/creiche/confluent-go/actions/workflows/release.yml/badge.svg)

REST-based Go client for Confluent Cloud and Platform APIs.

## Overview

This library provides a pure-HTTP, type-safe client for Confluent Cloud and Platform APIs. It supports robust error handling and includes retry/backoff for transient failures.

Key packages:
- `pkg/client`: low-level HTTP client (`Do`, `Response.DecodeJSON`) with Basic Auth.
- `pkg/resources`: resource managers for Clusters (CMK v2), Topics/ACLs (Kafka REST v3), Service Accounts & API Keys (IAM v2), Environments (Org v2).
- `pkg/retry`: configurable retry strategy with exponential backoff and jitter.
- `pkg/schemaregistry`: Schema Registry with complete operations, validation, and mode configuration.

Related docs: see `ERROR_HANDLING.md`, `REST_ARCHITECTURE.md`, and `PROJECT_STRUCTURE.md` for deeper reference.

## Quick Start

```go
// 1. Create REST client
cfg := client.Config{
    BaseURL:   "https://api.confluent.cloud",
    APIKey:    "your-key",
    APISecret: "your-secret",
}
c, err := client.NewClient(cfg)

// 2. Use a resource manager
clusters, err := resources.NewClusterManager(c).ListClusters(ctx, envID)
```

## Schema Registry

Schema Registry support lives in `pkg/schemaregistry` and reuses the shared REST client. Core operations include subjects, schemas, versions, deletion, compatibility, and mode configuration. **Schemas are automatically validated client-side before registration.**

```go
sr := schemaregistry.NewManager(c, "/schema-registry/v1")

// Subjects & schemas
subs, err := sr.ListSubjects(ctx)
latest, err := sr.GetLatestSchema(ctx, "my-subject")
byID, err := sr.GetSchemaByID(ctx, 42)

// Register a schema (automatically validated)
id, err := sr.RegisterSchema(ctx, "my-subject", schemaregistry.RegisterRequest{
  Schema:     `{"type":"record","name":"Foo","fields":[{"name":"bar","type":"string"}]}`,
  SchemaType: schemaregistry.SchemaTypeAvro,
})

// Versions
versions, err := sr.ListVersions(ctx, "my-subject")
v2, err := sr.GetSchemaVersion(ctx, "my-subject", 2)

// Delete (soft/hard)
_ = sr.DeleteSubject(ctx, "my-subject", false) // soft delete
_ = sr.DeleteSubject(ctx, "my-subject", true)  // permanent delete

// Compatibility (global and per-subject)
glob, err := sr.GetGlobalCompatibility(ctx)
err = sr.SetGlobalCompatibility(ctx, schemaregistry.CompatFull)
subj, err := sr.GetSubjectCompatibility(ctx, "my-subject")
err = sr.SetSubjectCompatibility(ctx, "my-subject", schemaregistry.CompatBackward)

// Mode (global and per-subject): READWRITE, READONLY, IMPORT
mode, err := sr.GetGlobalMode(ctx)
err = sr.SetGlobalMode(ctx, schemaregistry.ModeReadOnly) // prevent schema changes
subjMode, err := sr.GetSubjectMode(ctx, "my-subject")
err = sr.SetSubjectMode(ctx, "my-subject", schemaregistry.ModeReadWrite)
```

### Schema Validation

Schemas are automatically validated before registration or compatibility testing. Validation catches common syntax errors early:

```go
// AVRO validation: checks JSON syntax, required fields (type, name, fields)
id, err := sr.RegisterSchema(ctx, "user-value", schemaregistry.RegisterRequest{
  Schema:     `{"type":"record","name":"User","fields":[{"name":"id","type":"int"}]}`,
  SchemaType: schemaregistry.SchemaTypeAvro,
})
// Returns validation error immediately if schema is malformed

// JSON Schema validation: checks JSON syntax, typical fields ($schema, type, properties)
id, err := sr.RegisterSchema(ctx, "config-value", schemaregistry.RegisterRequest{
  Schema:     `{"type":"object","properties":{"version":{"type":"string"}}}`,
  SchemaType: schemaregistry.SchemaTypeJSON,
})

// Protobuf validation: checks for proto keywords (syntax, message, service, package)
id, err := sr.RegisterSchema(ctx, "event-value", schemaregistry.RegisterRequest{
  Schema:     `syntax = "proto3"; message Event { int32 id = 1; }`,
  SchemaType: schemaregistry.SchemaTypeProtobuf,
})
```

Validation errors are returned immediately without making an API call:

```go
_, err := sr.RegisterSchema(ctx, "bad-schema", schemaregistry.RegisterRequest{
  Schema:     `{invalid json}`,
  SchemaType: schemaregistry.SchemaTypeAvro,
})
// err: schema validation failed: invalid AVRO schema JSON: ...
```

### Configuration Notes

- Base path: If omitted, `NewManager` defaults to `"/schema-registry/v1"`.
- Cloud URL: Use your Confluent Cloud base URL (e.g., `https://api.confluent.cloud`).
- On-prem URL: Point `client.Config.BaseURL` to your SR endpoint (e.g., `https://sr.example.com`).
- Constants: Prefer `schemaregistry.SchemaTypeAvro|JSON|Protobuf`, `schemaregistry.Compat*`, and `schemaregistry.Mode*` constants over raw strings.

### Error Handling

Schema Registry operations return typed errors with helper functions:

```go
schema, err := sr.GetLatestSchema(ctx, "my-subject")
if err != nil {
  if schemaregistry.IsSubjectNotFound(err) {
    // Subject doesn't exist - register initial schema
  } else if schemaregistry.IsSubjectSoftDeleted(err) {
    // Subject was soft-deleted - use permanent delete
  }
  return err
}

id, err := sr.RegisterSchema(ctx, subject, req)
if err != nil {
  if schemaregistry.IsInvalidSchema(err) {
    // Schema syntax error - check schema JSON
  } else if schemaregistry.IsIncompatibleSchema(err) {
    // Schema incompatible - test compatibility first
  }
  return err
}
```

Available error helpers: `IsSubjectNotFound`, `IsVersionNotFound`, `IsSchemaNotFound`, `IsSubjectSoftDeleted`, `IsInvalidSchema`, `IsIncompatibleSchema`, `IsInvalidCompatibility`, `IsInvalidSubject`, `IsInvalidMode`. See `ERROR_HANDLING.md` for complete examples.

## Examples

- `cmd/examples/main.go` — REST client usage across managers
- `cmd/examples/operator_pattern.go` — Kubernetes operator-style reconciliation

## Documentation

- Error Handling: see `ERROR_HANDLING.md`
- Architecture & Endpoints: see `REST_ARCHITECTURE.md`
- Project Layout & Conventions: see `PROJECT_STRUCTURE.md`
- Tests & Coverage: see `TESTS_SUMMARY.md`

## Error Handling

All HTTP 4xx/5xx responses return a structured `*api.Error` with helpers like `IsNotFound()`, `IsRateLimited()`, and `RetryAfter()`. See `ERROR_HANDLING.md` for examples and best practices.

## Retry & Backoff

Use `retry.Strategy` to wrap operations with exponential backoff and jitter. It honors `Retry-After` headers from Confluent APIs. See `pkg/retry/retry.go` and tests for usage patterns.

## Installation

```zsh
go get github.com/creiche/confluent-go
```

Go version: 1.22+

## Supported APIs

- Confluent Cloud Management (CMK v2) — clusters
- Kafka REST (v3) — topics, ACLs
- IAM (v2) — service accounts, API keys
- Org (v2) — environments
- Schema Registry (v1) — subjects/schemas/compatibility/modes

## Contributing

Contributions are welcome. Please read `CONTRIBUTING.md` and ensure tests pass (`go test ./...`).
