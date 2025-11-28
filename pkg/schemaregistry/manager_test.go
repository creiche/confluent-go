package schemaregistry

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/creiche/confluent-go/pkg/client"
)

func newTestClient(t *testing.T, handler http.HandlerFunc) *client.Client {
	t.Helper()
	srv := httptest.NewServer(handler)
	t.Cleanup(srv.Close)

	c, err := client.NewClient(client.Config{
		BaseURL:   srv.URL,
		APIKey:    "key",
		APISecret: "secret",
	})
	if err != nil {
		t.Fatalf("failed to create client: %v", err)
	}
	return c
}

func TestListSubjects(t *testing.T) {
	handler := func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet || !strings.HasSuffix(r.URL.Path, "/schema-registry/v1/subjects") {
			http.Error(w, "not found", http.StatusNotFound)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode([]string{"foo", "bar"})
	}
	c := newTestClient(t, handler)
	m := NewManager(c, "/schema-registry/v1")

	got, err := m.ListSubjects(context.Background())
	if err != nil {
		t.Fatalf("ListSubjects error: %v", err)
	}
	if len(got) != 2 || got[0] != "foo" || got[1] != "bar" {
		t.Fatalf("unexpected subjects: %#v", got)
	}
}

func TestGetLatestSchema(t *testing.T) {
	handler := func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet || !strings.HasSuffix(r.URL.Path, "/schema-registry/v1/subjects/test/versions/latest") {
			http.Error(w, "not found", http.StatusNotFound)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(Schema{ID: 1, Subject: "test", Version: 3, Schema: "{\"type\":\"string\"}", Type: SchemaTypeAvro})
	}
	c := newTestClient(t, handler)
	m := NewManager(c, "/schema-registry/v1")

	s, err := m.GetLatestSchema(context.Background(), "test")
	if err != nil {
		t.Fatalf("GetLatestSchema error: %v", err)
	}
	if s.ID != 1 || s.Subject != "test" || s.Version != 3 || s.Type != "AVRO" {
		t.Fatalf("unexpected schema: %#v", s)
	}
}

func TestGetSchemaByID(t *testing.T) {
	handler := func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet || !strings.HasSuffix(r.URL.Path, "/schema-registry/v1/schemas/ids/42") {
			http.Error(w, "not found", http.StatusNotFound)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(struct {
			Schema string `json:"schema"`
		}{Schema: "{\"type\":\"string\"}"})
	}
	c := newTestClient(t, handler)
	m := NewManager(c, "/schema-registry/v1")

	s, err := m.GetSchemaByID(context.Background(), 42)
	if err != nil {
		t.Fatalf("GetSchemaByID error: %v", err)
	}
	if s.ID != 42 || s.Schema == "" {
		t.Fatalf("unexpected schema by id: %#v", s)
	}
}

func TestRegisterSchema(t *testing.T) {
	handler := func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost || !strings.HasSuffix(r.URL.Path, "/schema-registry/v1/subjects/my-subject/versions") {
			http.Error(w, "not found", http.StatusNotFound)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(RegisterResponse{ID: 99})
	}
	c := newTestClient(t, handler)
	m := NewManager(c, "/schema-registry/v1")

	id, err := m.RegisterSchema(context.Background(), "my-subject", RegisterRequest{Schema: "{\"type\":\"string\"}", SchemaType: SchemaTypeAvro})
	if err != nil {
		t.Fatalf("RegisterSchema error: %v", err)
	}
	if id != 99 {
		t.Fatalf("unexpected id: %d", id)
	}
}

func TestTestCompatibility(t *testing.T) {
	handler := func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost || !strings.HasSuffix(r.URL.Path, "/schema-registry/v1/compatibility/subjects/my-subject/versions/latest") {
			http.Error(w, "not found", http.StatusNotFound)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(CompatibilityResponse{IsCompatible: true})
	}
	c := newTestClient(t, handler)
	m := NewManager(c, "/schema-registry/v1")

	ok, err := m.TestCompatibility(context.Background(), "my-subject", RegisterRequest{Schema: "{\"type\":\"string\"}", SchemaType: SchemaTypeAvro})
	if err != nil {
		t.Fatalf("TestCompatibility error: %v", err)
	}
	if !ok {
		t.Fatalf("expected compatible")
	}
}

func TestErrorsArePropagated(t *testing.T) {
	handler := func(w http.ResponseWriter, r *http.Request) {
		// Return a 404 for any request to assert api.Error propagation
		w.Header().Set("Content-Type", "application/json")
		http.Error(w, `{"error_code":40401,"message":"not found"}`, http.StatusNotFound)
	}
	c := newTestClient(t, handler)
	m := NewManager(c, "/schema-registry/v1")

	if _, err := m.ListSubjects(context.Background()); err == nil {
		t.Fatalf("expected error for ListSubjects")
	}
	if _, err := m.GetLatestSchema(context.Background(), "x"); err == nil {
		t.Fatalf("expected error for GetLatestSchema")
	}
	if _, err := m.GetSchemaByID(context.Background(), 1); err == nil {
		t.Fatalf("expected error for GetSchemaByID")
	}
	if _, err := m.RegisterSchema(context.Background(), "x", RegisterRequest{Schema: "{}"}); err == nil {
		t.Fatalf("expected error for RegisterSchema")
	}
	if _, err := m.TestCompatibility(context.Background(), "x", RegisterRequest{Schema: "{}"}); err == nil {
		t.Fatalf("expected error for TestCompatibility")
	}
}

func TestListVersionsAndGetVersion(t *testing.T) {
	handler := func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch {
		case r.Method == http.MethodGet && strings.HasSuffix(r.URL.Path, "/schema-registry/v1/subjects/test/versions"):
			_ = json.NewEncoder(w).Encode([]int{1, 2, 3})
		case r.Method == http.MethodGet && strings.HasSuffix(r.URL.Path, "/schema-registry/v1/subjects/test/versions/2"):
			_ = json.NewEncoder(w).Encode(Schema{ID: 100, Subject: "test", Version: 2, Schema: "{\"type\":\"string\"}", Type: SchemaTypeAvro})
		default:
			http.Error(w, "not found", http.StatusNotFound)
		}
	}
	c := newTestClient(t, handler)
	m := NewManager(c, "/schema-registry/v1")

	vers, err := m.ListVersions(context.Background(), "test")
	if err != nil || len(vers) != 3 || vers[1] != 2 {
		t.Fatalf("unexpected versions: %#v err=%v", vers, err)
	}
	s, err := m.GetSchemaVersion(context.Background(), "test", 2)
	if err != nil || s.Version != 2 || s.Subject != "test" {
		t.Fatalf("unexpected schema version: %#v err=%v", s, err)
	}
}

func TestDeleteSubject(t *testing.T) {
	handler := func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodDelete && strings.HasPrefix(r.URL.Path, "/schema-registry/v1/subjects/my-subject") {
			w.WriteHeader(http.StatusOK)
			return
		}
		http.Error(w, "not found", http.StatusNotFound)
	}
	c := newTestClient(t, handler)
	m := NewManager(c, "/schema-registry/v1")

	if err := m.DeleteSubject(context.Background(), "my-subject", false); err != nil {
		t.Fatalf("DeleteSubject error: %v", err)
	}
	if err := m.DeleteSubject(context.Background(), "my-subject", true); err != nil {
		t.Fatalf("DeleteSubject permanent error: %v", err)
	}
}

func TestCompatibilityGetSet(t *testing.T) {
	handler := func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch {
		case r.Method == http.MethodGet && strings.HasSuffix(r.URL.Path, "/schema-registry/v1/config"):
			_ = json.NewEncoder(w).Encode(map[string]string{"compatibility": "FULL"})
		case r.Method == http.MethodPut && strings.HasSuffix(r.URL.Path, "/schema-registry/v1/config"):
			w.WriteHeader(http.StatusOK)
		case r.Method == http.MethodGet && strings.HasSuffix(r.URL.Path, "/schema-registry/v1/config/my-subject"):
			_ = json.NewEncoder(w).Encode(map[string]string{"compatibility": "BACKWARD"})
		case r.Method == http.MethodPut && strings.HasSuffix(r.URL.Path, "/schema-registry/v1/config/my-subject"):
			w.WriteHeader(http.StatusOK)
		default:
			http.Error(w, "not found", http.StatusNotFound)
		}
	}
	c := newTestClient(t, handler)
	m := NewManager(c, "/schema-registry/v1")

	glob, err := m.GetGlobalCompatibility(context.Background())
	if err != nil || glob != "FULL" {
		t.Fatalf("unexpected global compat: %s err=%v", glob, err)
	}
	if err := m.SetGlobalCompatibility(context.Background(), "BACKWARD"); err != nil {
		t.Fatalf("SetGlobalCompatibility error: %v", err)
	}
	subj, err := m.GetSubjectCompatibility(context.Background(), "my-subject")
	if err != nil || subj != "BACKWARD" {
		t.Fatalf("unexpected subject compat: %s err=%v", subj, err)
	}
	if err := m.SetSubjectCompatibility(context.Background(), "my-subject", "FULL"); err != nil {
		t.Fatalf("SetSubjectCompatibility error: %v", err)
	}
}

// SR-specific error handling tests

func TestSRError_SubjectNotFound(t *testing.T) {
	handler := func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"error_code": 40401,
			"message":    "Subject 'unknown' not found",
		})
	}
	c := newTestClient(t, handler)
	m := NewManager(c, "/schema-registry/v1")

	_, err := m.GetLatestSchema(context.Background(), "unknown")
	if err == nil {
		t.Fatal("expected error for unknown subject")
	}
	if !IsSubjectNotFound(err) {
		t.Errorf("IsSubjectNotFound should return true, got error: %v", err)
	}
}

func TestSRError_VersionNotFound(t *testing.T) {
	handler := func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"error_code": 40402,
			"message":    "Version 99 not found",
		})
	}
	c := newTestClient(t, handler)
	m := NewManager(c, "/schema-registry/v1")

	_, err := m.GetSchemaVersion(context.Background(), "test", 99)
	if err == nil {
		t.Fatal("expected error for unknown version")
	}
	if !IsVersionNotFound(err) {
		t.Errorf("IsVersionNotFound should return true, got error: %v", err)
	}
}

func TestSRError_SchemaNotFound(t *testing.T) {
	handler := func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"error_code": 40403,
			"message":    "Schema 999 not found",
		})
	}
	c := newTestClient(t, handler)
	m := NewManager(c, "/schema-registry/v1")

	_, err := m.GetSchemaByID(context.Background(), 999)
	if err == nil {
		t.Fatal("expected error for unknown schema ID")
	}
	if !IsSchemaNotFound(err) {
		t.Errorf("IsSchemaNotFound should return true, got error: %v", err)
	}
}

func TestSRError_InvalidSchema(t *testing.T) {
	handler := func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnprocessableEntity)
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"error_code": 42201,
			"message":    "Invalid schema: field type not supported",
		})
	}
	c := newTestClient(t, handler)
	m := NewManager(c, "/schema-registry/v1")

	// Use a schema that passes local validation but might fail server-side validation
	_, err := m.RegisterSchema(context.Background(), "test", RegisterRequest{
		Schema:     `{"type":"record","name":"Test","fields":[{"name":"invalid","type":"unsupported_type"}]}`,
		SchemaType: SchemaTypeAvro,
	})
	if err == nil {
		t.Fatal("expected error for invalid schema")
	}
	if !IsInvalidSchema(err) {
		t.Errorf("IsInvalidSchema should return true, got error: %v", err)
	}
}

func TestSRError_IncompatibleSchema(t *testing.T) {
	handler := func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusConflict)
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"error_code": 409,
			"message":    "Schema incompatible with previous version",
		})
	}
	c := newTestClient(t, handler)
	m := NewManager(c, "/schema-registry/v1")

	_, err := m.RegisterSchema(context.Background(), "test", RegisterRequest{Schema: "{\"type\":\"int\"}", SchemaType: SchemaTypeAvro})
	if err == nil {
		t.Fatal("expected error for incompatible schema")
	}
	if !IsIncompatibleSchema(err) {
		t.Errorf("IsIncompatibleSchema should return true, got error: %v", err)
	}
}

func TestSRError_InvalidCompatibility(t *testing.T) {
	handler := func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnprocessableEntity)
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"error_code": 42203,
			"message":    "Invalid compatibility level",
		})
	}
	c := newTestClient(t, handler)
	m := NewManager(c, "/schema-registry/v1")

	err := m.SetGlobalCompatibility(context.Background(), "INVALID_LEVEL")
	if err == nil {
		t.Fatal("expected error for invalid compatibility")
	}
	if !IsInvalidCompatibility(err) {
		t.Errorf("IsInvalidCompatibility should return true, got error: %v", err)
	}
}

func TestSRError_InvalidSubject(t *testing.T) {
	handler := func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnprocessableEntity)
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"error_code": 42202,
			"message":    "Invalid subject name",
		})
	}
	c := newTestClient(t, handler)
	m := NewManager(c, "/schema-registry/v1")

	_, err := m.RegisterSchema(context.Background(), "invalid-@#$-subject", RegisterRequest{Schema: "{\"type\":\"string\"}", SchemaType: SchemaTypeAvro})
	if err == nil {
		t.Fatal("expected error for invalid subject")
	}
	if !IsInvalidSubject(err) {
		t.Errorf("IsInvalidSubject should return true, got error: %v", err)
	}
}

func TestSRError_SubjectSoftDeleted(t *testing.T) {
	handler := func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"error_code": 40404,
			"message":    "Subject 'deleted-subject' was soft deleted",
		})
	}
	c := newTestClient(t, handler)
	m := NewManager(c, "/schema-registry/v1")

	_, err := m.GetLatestSchema(context.Background(), "deleted-subject")
	if err == nil {
		t.Fatal("expected error for soft deleted subject")
	}
	if !IsSubjectSoftDeleted(err) {
		t.Errorf("IsSubjectSoftDeleted should return true, got error: %v", err)
	}
}

// Client-side validation tests

func TestRegisterSchema_ClientSideValidation(t *testing.T) {
	// Handler should never be called due to client-side validation
	handler := func(w http.ResponseWriter, r *http.Request) {
		t.Fatal("handler should not be called - validation should fail client-side")
	}
	c := newTestClient(t, handler)
	m := NewManager(c, "/schema-registry/v1")

	// Test with invalid JSON
	_, err := m.RegisterSchema(context.Background(), "test", RegisterRequest{
		Schema:     "bad-json",
		SchemaType: SchemaTypeAvro,
	})
	if err == nil {
		t.Fatal("expected validation error for invalid JSON")
	}
	// Use strings.Contains from standard library
	if !strings.Contains(err.Error(), "validation failed") {
		t.Errorf("expected 'validation failed' in error, got: %v", err)
	}
}

func TestTestCompatibility_ClientSideValidation(t *testing.T) {
	// Handler should never be called due to client-side validation
	handler := func(w http.ResponseWriter, r *http.Request) {
		t.Fatal("handler should not be called - validation should fail client-side")
	}
	c := newTestClient(t, handler)
	m := NewManager(c, "/schema-registry/v1")

	// Test with empty schema
	_, err := m.TestCompatibility(context.Background(), "test", RegisterRequest{
		Schema:     "",
		SchemaType: SchemaTypeAvro,
	})
	if err == nil {
		t.Fatal("expected validation error for empty schema")
	}
	// Use strings.Contains from standard library
	if !strings.Contains(err.Error(), "validation failed") {
		t.Errorf("expected 'validation failed' in error, got: %v", err)
	}
}