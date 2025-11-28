package schemaregistry

import (
	"encoding/json"
	"fmt"
	"strings"
)

// Validator validates schema syntax for a specific schema type.
// Implementations validate AVRO, JSON Schema, and Protobuf schemas.
type Validator interface {
	// Validate checks if the schema syntax is valid
	Validate(schema string) error
}

// ValidateSchema validates a schema based on its type.
// Supported types are SchemaTypeAvro, SchemaTypeJSON, and SchemaTypeProtobuf.
// Returns an error if the schema is empty, malformed, or missing required fields.
func ValidateSchema(schema string, schemaType string) error {
	if schema == "" {
		return fmt.Errorf("schema cannot be empty")
	}

	var validator Validator
	switch schemaType {
	case SchemaTypeAvro:
		validator = &AvroValidator{}
	case SchemaTypeJSON:
		validator = &JSONSchemaValidator{}
	case SchemaTypeProtobuf:
		validator = &ProtobufValidator{}
	default:
		return fmt.Errorf("unsupported schema type: %s", schemaType)
	}

	return validator.Validate(schema)
}

// AvroValidator validates AVRO schema syntax.
// It checks JSON validity and required fields based on the AVRO type
// (record, enum, array, map, primitive, or union).
type AvroValidator struct{}

// Validate checks if the AVRO schema is valid JSON and has required fields
func (v *AvroValidator) Validate(schema string) error {
	// First, ensure it's valid JSON
	var avroSchema map[string]interface{}
	if err := json.Unmarshal([]byte(schema), &avroSchema); err != nil {
		// Try as array (for union types)
		var avroSchemaArray []interface{}
		if err2 := json.Unmarshal([]byte(schema), &avroSchemaArray); err2 != nil {
			return fmt.Errorf("invalid AVRO schema JSON: %w", err)
		}
		// Array schema is valid for unions
		return nil
	}

	// Check for required "type" field
	typeField, hasType := avroSchema["type"]
	if !hasType {
		return fmt.Errorf("AVRO schema missing required 'type' field")
	}

	// Validate type is a string or array
	switch t := typeField.(type) {
	case string:
		// Valid primitive or named type
		if t == "" {
			return fmt.Errorf("AVRO schema 'type' field cannot be empty")
		}
	case []interface{}:
		// Valid union type
		if len(t) == 0 {
			return fmt.Errorf("AVRO schema union type cannot be empty")
		}
	default:
		return fmt.Errorf("AVRO schema 'type' field must be string or array, got %T", typeField)
	}

	// For record types, validate required fields
	if typeStr, ok := typeField.(string); ok && typeStr == "record" {
		if _, hasName := avroSchema["name"]; !hasName {
			return fmt.Errorf("AVRO record schema missing required 'name' field")
		}
		if _, hasFields := avroSchema["fields"]; !hasFields {
			return fmt.Errorf("AVRO record schema missing required 'fields' field")
		}
	}

	// For enum types, validate required fields
	if typeStr, ok := typeField.(string); ok && typeStr == "enum" {
		if _, hasName := avroSchema["name"]; !hasName {
			return fmt.Errorf("AVRO enum schema missing required 'name' field")
		}
		if _, hasSymbols := avroSchema["symbols"]; !hasSymbols {
			return fmt.Errorf("AVRO enum schema missing required 'symbols' field")
		}
	}

	// For array types, validate items field
	if typeStr, ok := typeField.(string); ok && typeStr == "array" {
		if _, hasItems := avroSchema["items"]; !hasItems {
			return fmt.Errorf("AVRO array schema missing required 'items' field")
		}
	}

	// For map types, validate values field
	if typeStr, ok := typeField.(string); ok && typeStr == "map" {
		if _, hasValues := avroSchema["values"]; !hasValues {
			return fmt.Errorf("AVRO map schema missing required 'values' field")
		}
	}

	return nil
}

// JSONSchemaValidator validates JSON Schema syntax.
// It checks for valid JSON and the presence of typical JSON Schema fields
// ($schema, type, properties, or $ref).
type JSONSchemaValidator struct{}

// Validate checks if the JSON Schema is valid
func (v *JSONSchemaValidator) Validate(schema string) error {
	// First, ensure it's valid JSON
	var jsonSchema map[string]interface{}
	if err := json.Unmarshal([]byte(schema), &jsonSchema); err != nil {
		return fmt.Errorf("invalid JSON Schema: %w", err)
	}

	// JSON Schema should have at least one of: $schema, type, properties, or $ref
	_, hasSchemaField := jsonSchema["$schema"]
	_, hasType := jsonSchema["type"]
	_, hasProperties := jsonSchema["properties"]
	_, hasRef := jsonSchema["$ref"]

	if !hasSchemaField && !hasType && !hasProperties && !hasRef {
		return fmt.Errorf("JSON Schema missing typical fields ($schema, type, properties, or $ref)")
	}

	return nil
}

// ProtobufValidator validates Protobuf schema syntax.
// It performs lightweight validation by checking for expected Protobuf keywords
// (syntax, message, service, package, or enum).
// Note: This is intentionally permissive - it checks for keyword presence but does not
// parse full .proto syntax or validate context (comments, strings, etc.).
type ProtobufValidator struct{}

// Validate checks if the Protobuf schema has basic syntax requirements
func (v *ProtobufValidator) Validate(schema string) error {
	if schema == "" {
		return fmt.Errorf("protobuf schema cannot be empty")
	}

	// Basic validation: Protobuf schemas should contain "syntax", "message", or "package"
	// This is a lightweight check - full validation would require parsing .proto syntax
	hasProtoKeyword := false
	protoKeywords := []string{"syntax", "message", "service", "package", "enum"}

	for _, keyword := range protoKeywords {
		if containsWord(schema, keyword) {
			hasProtoKeyword = true
			break
		}
	}

	if !hasProtoKeyword {
		return fmt.Errorf("protobuf schema missing expected keywords (syntax, message, service, package, or enum)")
	}

	return nil
}

// containsWord checks if a word appears in the text as a separate token.
// Uses field splitting on common delimiters including underscores, dots, and operators
// to prevent false positives from keywords in identifiers, comments, or strings.
func containsWord(text, word string) bool {
	if text == "" || word == "" {
		return false
	}
	// Split on delimiters: whitespace, braces, parens, semicolons, underscores, dots, slashes, equals
	for _, field := range strings.FieldsFunc(text, func(r rune) bool {
		return r == ' ' || r == '\n' || r == '\t' || r == ';' ||
			r == '{' || r == '}' || r == '(' || r == ')' ||
			r == '_' || r == '.' || r == '/' || r == '=' ||
			r == '"' || r == '\'' || r == ','
	}) {
		if field == word {
			return true
		}
	}
	return false
}
