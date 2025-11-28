// Package schemaregistry provides a client for Confluent Schema Registry API.
//
// The package supports core Schema Registry operations including:
//   - Subject management (list, get, delete)
//   - Schema registration and retrieval
//   - Schema versioning
//   - Compatibility testing and configuration (global and per-subject)
//   - Mode configuration (global and per-subject): READWRITE, READONLY, IMPORT
//   - Client-side schema validation for AVRO, JSON Schema, and Protobuf
//
// Schemas are automatically validated before registration to catch syntax errors early.
// All errors from Schema Registry include typed error codes for precise error handling.
//
// Example usage:
//
//	sr := schemaregistry.NewManager(client, "/schema-registry/v1")
//
//	// Register a schema (automatically validated)
//	id, err := sr.RegisterSchema(ctx, "user-value", schemaregistry.RegisterRequest{
//		Schema:     `{"type":"record","name":"User","fields":[{"name":"id","type":"int"}]}`,
//		SchemaType: schemaregistry.SchemaTypeAvro,
//	})
//
//	// Handle errors with typed helpers
//	if schemaregistry.IsSubjectNotFound(err) {
//		// Handle missing subject
//	}
package schemaregistry

import (
	"context"
	"fmt"
	"net/url"

	"github.com/creiche/confluent-go/pkg/client"
)

// Manager provides high-level operations against Schema Registry.
type Manager struct {
	c        *client.Client
	basePath string
}

// NewManager creates a new Schema Registry manager using the shared REST client.
// basePath is typically "/schema-registry/v1" for Confluent Cloud.
func NewManager(c *client.Client, basePath string) *Manager {
	if basePath == "" {
		basePath = "/schema-registry/v1"
	}
	return &Manager{c: c, basePath: basePath}
}

// ListSubjects returns all subjects registered.
func (m *Manager) ListSubjects(ctx context.Context) ([]string, error) {
	var subjects []string
	req := client.Request{Method: "GET", Path: fmt.Sprintf("%s/subjects", m.basePath)}
	resp, err := m.c.Do(ctx, req)
	if err != nil {
		return nil, err
	}
	if err := resp.DecodeJSON(&subjects); err != nil {
		return nil, err
	}
	return subjects, nil
}

// GetLatestSchema returns the latest schema for a subject.
func (m *Manager) GetLatestSchema(ctx context.Context, subject string) (*Schema, error) {
	var s Schema
	req := client.Request{Method: "GET", Path: fmt.Sprintf("%s/subjects/%s/versions/latest", m.basePath, url.PathEscape(subject))}
	resp, err := m.c.Do(ctx, req)
	if err != nil {
		return nil, err
	}
	if err := resp.DecodeJSON(&s); err != nil {
		return nil, err
	}
	return &s, nil
}

// GetSchemaByID fetches a schema by its global ID.
func (m *Manager) GetSchemaByID(ctx context.Context, id int) (*Schema, error) {
	var body struct {
		Schema string `json:"schema"`
	}
	req := client.Request{Method: "GET", Path: fmt.Sprintf("%s/schemas/ids/%d", m.basePath, id)}
	resp, err := m.c.Do(ctx, req)
	if err != nil {
		return nil, err
	}
	if err := resp.DecodeJSON(&body); err != nil {
		return nil, err
	}
	return &Schema{ID: id, Schema: body.Schema}, nil
}

// RegisterSchema registers a new schema under a subject and returns the assigned ID.
// The schema is validated before registration to catch syntax errors early.
// If SchemaType is empty, it defaults to AVRO (matching Schema Registry API behavior).
func (m *Manager) RegisterSchema(ctx context.Context, subject string, payload RegisterRequest) (int, error) {
	// Default schema type to AVRO if omitted, matching Schema Registry API
	schemaType := payload.SchemaType
	if schemaType == "" {
		schemaType = SchemaTypeAvro
	}
	// Validate schema syntax before sending to SR
	if err := ValidateSchema(payload.Schema, schemaType); err != nil {
		return 0, fmt.Errorf("schema validation failed: %w", err)
	}

	var out RegisterResponse
	req := client.Request{Method: "POST", Path: fmt.Sprintf("%s/subjects/%s/versions", m.basePath, url.PathEscape(subject)), Body: payload}
	resp, err := m.c.Do(ctx, req)
	if err != nil {
		return 0, err
	}
	if err := resp.DecodeJSON(&out); err != nil {
		return 0, err
	}
	return out.ID, nil
}

// TestCompatibility checks compatibility of the provided schema against the latest.
// The schema is validated before the compatibility check.
// If SchemaType is empty, it defaults to AVRO (matching Schema Registry API behavior).
func (m *Manager) TestCompatibility(ctx context.Context, subject string, payload RegisterRequest) (bool, error) {
	// Default schema type to AVRO if omitted, matching Schema Registry API
	schemaType := payload.SchemaType
	if schemaType == "" {
		schemaType = SchemaTypeAvro
	}
	// Validate schema syntax before testing compatibility
	if err := ValidateSchema(payload.Schema, schemaType); err != nil {
		return false, fmt.Errorf("schema validation failed: %w", err)
	}

	var out CompatibilityResponse
	req := client.Request{Method: "POST", Path: fmt.Sprintf("%s/compatibility/subjects/%s/versions/latest", m.basePath, url.PathEscape(subject)), Body: payload}
	resp, err := m.c.Do(ctx, req)
	if err != nil {
		return false, err
	}
	if err := resp.DecodeJSON(&out); err != nil {
		return false, err
	}
	return out.IsCompatible, nil
}

// ListVersions lists all versions for a subject.
func (m *Manager) ListVersions(ctx context.Context, subject string) ([]int, error) {
	var versions []int
	req := client.Request{Method: "GET", Path: fmt.Sprintf("%s/subjects/%s/versions", m.basePath, url.PathEscape(subject))}
	resp, err := m.c.Do(ctx, req)
	if err != nil {
		return nil, err
	}
	if err := resp.DecodeJSON(&versions); err != nil {
		return nil, err
	}
	return versions, nil
}

// GetSchemaVersion fetches a specific version for a subject.
func (m *Manager) GetSchemaVersion(ctx context.Context, subject string, version int) (*Schema, error) {
	var s Schema
	req := client.Request{Method: "GET", Path: fmt.Sprintf("%s/subjects/%s/versions/%d", m.basePath, url.PathEscape(subject), version)}
	resp, err := m.c.Do(ctx, req)
	if err != nil {
		return nil, err
	}
	if err := resp.DecodeJSON(&s); err != nil {
		return nil, err
	}
	return &s, nil
}

// DeleteSubject deletes a subject. When permanent=true a hard delete is performed.
func (m *Manager) DeleteSubject(ctx context.Context, subject string, permanent bool) error {
	path := fmt.Sprintf("%s/subjects/%s", m.basePath, url.PathEscape(subject))
	if permanent {
		path += "?permanent=true"
	}
	req := client.Request{Method: "DELETE", Path: path}
	_, err := m.c.Do(ctx, req)
	return err
}

// Compatibility levels commonly used by SR: NONE, BACKWARD, BACKWARD_TRANSITIVE, FORWARD, FORWARD_TRANSITIVE, FULL, FULL_TRANSITIVE.

// GetGlobalCompatibility returns the global compatibility level.
func (m *Manager) GetGlobalCompatibility(ctx context.Context) (string, error) {
	var out struct {
		Compatibility string `json:"compatibility"`
	}
	req := client.Request{Method: "GET", Path: fmt.Sprintf("%s/config", m.basePath)}
	resp, err := m.c.Do(ctx, req)
	if err != nil {
		return "", err
	}
	if err := resp.DecodeJSON(&out); err != nil {
		return "", err
	}
	return out.Compatibility, nil
}

// SetGlobalCompatibility sets the global compatibility level.
func (m *Manager) SetGlobalCompatibility(ctx context.Context, level string) error {
	body := map[string]string{"compatibility": level}
	req := client.Request{Method: "PUT", Path: fmt.Sprintf("%s/config", m.basePath), Body: body}
	_, err := m.c.Do(ctx, req)
	return err
}

// GetSubjectCompatibility returns compatibility level for a subject.
func (m *Manager) GetSubjectCompatibility(ctx context.Context, subject string) (string, error) {
	var out struct {
		Compatibility string `json:"compatibility"`
	}
	req := client.Request{Method: "GET", Path: fmt.Sprintf("%s/config/%s", m.basePath, url.PathEscape(subject))}
	resp, err := m.c.Do(ctx, req)
	if err != nil {
		return "", err
	}
	if err := resp.DecodeJSON(&out); err != nil {
		return "", err
	}
	return out.Compatibility, nil
}

// SetSubjectCompatibility sets compatibility level for a subject.
func (m *Manager) SetSubjectCompatibility(ctx context.Context, subject string, level string) error {
	body := map[string]string{"compatibility": level}
	req := client.Request{Method: "PUT", Path: fmt.Sprintf("%s/config/%s", m.basePath, url.PathEscape(subject)), Body: body}
	_, err := m.c.Do(ctx, req)
	return err
}

// Mode operations: READWRITE (default), READONLY (prevents registration), IMPORT (for replication)

// GetGlobalMode returns the global mode.
func (m *Manager) GetGlobalMode(ctx context.Context) (string, error) {
	var out struct {
		Mode string `json:"mode"`
	}
	req := client.Request{Method: "GET", Path: fmt.Sprintf("%s/mode", m.basePath)}
	resp, err := m.c.Do(ctx, req)
	if err != nil {
		return "", err
	}
	if err := resp.DecodeJSON(&out); err != nil {
		return "", err
	}
	return out.Mode, nil
}

// SetGlobalMode sets the global mode.
// Valid modes: ModeReadWrite, ModeReadOnly, ModeImport.
func (m *Manager) SetGlobalMode(ctx context.Context, mode string) error {
	body := map[string]string{"mode": mode}
	req := client.Request{Method: "PUT", Path: fmt.Sprintf("%s/mode", m.basePath), Body: body}
	_, err := m.c.Do(ctx, req)
	return err
}

// GetSubjectMode returns mode for a subject.
func (m *Manager) GetSubjectMode(ctx context.Context, subject string) (string, error) {
	var out struct {
		Mode string `json:"mode"`
	}
	req := client.Request{Method: "GET", Path: fmt.Sprintf("%s/mode/%s", m.basePath, url.PathEscape(subject))}
	resp, err := m.c.Do(ctx, req)
	if err != nil {
		return "", err
	}
	if err := resp.DecodeJSON(&out); err != nil {
		return "", err
	}
	return out.Mode, nil
}

// SetSubjectMode sets mode for a subject.
// Valid modes: ModeReadWrite, ModeReadOnly, ModeImport.
func (m *Manager) SetSubjectMode(ctx context.Context, subject string, mode string) error {
	body := map[string]string{"mode": mode}
	req := client.Request{Method: "PUT", Path: fmt.Sprintf("%s/mode/%s", m.basePath, url.PathEscape(subject)), Body: body}
	_, err := m.c.Do(ctx, req)
	return err
}
