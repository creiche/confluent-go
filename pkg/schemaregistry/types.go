package schemaregistry

// Subject represents a Schema Registry subject.
// A subject is a named scope in which schemas evolve.
type Subject struct {
	Name string `json:"subject"`
}

// Schema represents a schema stored in Schema Registry.
// It contains the schema definition along with its metadata (ID, version, subject).
type Schema struct {
	ID      int    `json:"id,omitempty"`
	Subject string `json:"subject,omitempty"`
	Version int    `json:"version,omitempty"`
	Schema  string `json:"schema"`
	Type    string `json:"schemaType,omitempty"`
}

// RegisterRequest is the request payload for registering a schema.
// The schema will be validated client-side before being sent to the Schema Registry.
type RegisterRequest struct {
	Schema     string            `json:"schema"`
	SchemaType string            `json:"schemaType,omitempty"`
	References []SchemaReference `json:"references,omitempty"`
}

// RegisterResponse is the response containing the assigned schema ID.
type RegisterResponse struct {
	ID int `json:"id"`
}

// SchemaReference models a schema reference used by Avro/JSON/Protobuf schemas.
// References allow schemas to refer to other registered schemas.
type SchemaReference struct {
	Name    string `json:"name"`
	Subject string `json:"subject"`
	Version int    `json:"version"`
}

// CompatibilityResponse indicates whether two schemas are compatible
// according to the configured compatibility level.
type CompatibilityResponse struct {
	IsCompatible bool `json:"is_compatible"`
}

// Compatibility levels for Schema Registry configuration.
const (
	CompatNone               = "NONE"
	CompatBackward           = "BACKWARD"
	CompatBackwardTransitive = "BACKWARD_TRANSITIVE"
	CompatForward            = "FORWARD"
	CompatForwardTransitive  = "FORWARD_TRANSITIVE"
	CompatFull               = "FULL"
	CompatFullTransitive     = "FULL_TRANSITIVE"
)

// Supported schema types.
const (
	SchemaTypeAvro     = "AVRO"
	SchemaTypeJSON     = "JSON"
	SchemaTypeProtobuf = "PROTOBUF"
)
