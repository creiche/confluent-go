package schemaregistry

import (
	"testing"
)

// AVRO Schema Validation Tests

func TestAvroValidator_ValidRecordSchema(t *testing.T) {
	schema := `{
		"type": "record",
		"name": "User",
		"fields": [
			{"name": "id", "type": "int"},
			{"name": "name", "type": "string"}
		]
	}`
	
	err := ValidateSchema(schema, SchemaTypeAvro)
	if err != nil {
		t.Errorf("valid AVRO record schema should pass validation: %v", err)
	}
}

func TestAvroValidator_ValidPrimitiveSchema(t *testing.T) {
	schemas := []string{
		`{"type": "string"}`,
		`{"type": "int"}`,
		`{"type": "long"}`,
		`{"type": "boolean"}`,
		`{"type": "null"}`,
	}
	
	for _, schema := range schemas {
		err := ValidateSchema(schema, SchemaTypeAvro)
		if err != nil {
			t.Errorf("valid AVRO primitive schema %s should pass: %v", schema, err)
		}
	}
}

func TestAvroValidator_ValidUnionSchema(t *testing.T) {
	schema := `{"type": ["null", "string"]}`
	
	err := ValidateSchema(schema, SchemaTypeAvro)
	if err != nil {
		t.Errorf("valid AVRO union schema should pass validation: %v", err)
	}
}

func TestAvroValidator_ValidEnumSchema(t *testing.T) {
	schema := `{
		"type": "enum",
		"name": "Status",
		"symbols": ["ACTIVE", "INACTIVE", "PENDING"]
	}`
	
	err := ValidateSchema(schema, SchemaTypeAvro)
	if err != nil {
		t.Errorf("valid AVRO enum schema should pass validation: %v", err)
	}
}

func TestAvroValidator_ValidArraySchema(t *testing.T) {
	schema := `{
		"type": "array",
		"items": "string"
	}`
	
	err := ValidateSchema(schema, SchemaTypeAvro)
	if err != nil {
		t.Errorf("valid AVRO array schema should pass validation: %v", err)
	}
}

func TestAvroValidator_ValidMapSchema(t *testing.T) {
	schema := `{
		"type": "map",
		"values": "int"
	}`
	
	err := ValidateSchema(schema, SchemaTypeAvro)
	if err != nil {
		t.Errorf("valid AVRO map schema should pass validation: %v", err)
	}
}

func TestAvroValidator_InvalidJSON(t *testing.T) {
	schema := `{type: "record", invalid json}`
	
	err := ValidateSchema(schema, SchemaTypeAvro)
	if err == nil {
		t.Error("invalid JSON should fail validation")
	}
}

func TestAvroValidator_MissingType(t *testing.T) {
	schema := `{"name": "User", "fields": []}`
	
	err := ValidateSchema(schema, SchemaTypeAvro)
	if err == nil {
		t.Error("schema missing 'type' field should fail validation")
	}
}

func TestAvroValidator_RecordMissingName(t *testing.T) {
	schema := `{"type": "record", "fields": []}`
	
	err := ValidateSchema(schema, SchemaTypeAvro)
	if err == nil {
		t.Error("record schema missing 'name' should fail validation")
	}
}

func TestAvroValidator_RecordMissingFields(t *testing.T) {
	schema := `{"type": "record", "name": "User"}`
	
	err := ValidateSchema(schema, SchemaTypeAvro)
	if err == nil {
		t.Error("record schema missing 'fields' should fail validation")
	}
}

func TestAvroValidator_EnumMissingSymbols(t *testing.T) {
	schema := `{"type": "enum", "name": "Status"}`
	
	err := ValidateSchema(schema, SchemaTypeAvro)
	if err == nil {
		t.Error("enum schema missing 'symbols' should fail validation")
	}
}

func TestAvroValidator_ArrayMissingItems(t *testing.T) {
	schema := `{"type": "array"}`
	
	err := ValidateSchema(schema, SchemaTypeAvro)
	if err == nil {
		t.Error("array schema missing 'items' should fail validation")
	}
}

func TestAvroValidator_MapMissingValues(t *testing.T) {
	schema := `{"type": "map"}`
	
	err := ValidateSchema(schema, SchemaTypeAvro)
	if err == nil {
		t.Error("map schema missing 'values' should fail validation")
	}
}

// JSON Schema Validation Tests

func TestJSONSchemaValidator_ValidSchema(t *testing.T) {
	schema := `{
		"$schema": "http://json-schema.org/draft-07/schema#",
		"type": "object",
		"properties": {
			"id": {"type": "integer"},
			"name": {"type": "string"}
		},
		"required": ["id"]
	}`
	
	err := ValidateSchema(schema, SchemaTypeJSON)
	if err != nil {
		t.Errorf("valid JSON Schema should pass validation: %v", err)
	}
}

func TestJSONSchemaValidator_ValidWithType(t *testing.T) {
	schema := `{"type": "string"}`
	
	err := ValidateSchema(schema, SchemaTypeJSON)
	if err != nil {
		t.Errorf("JSON Schema with type should pass validation: %v", err)
	}
}

func TestJSONSchemaValidator_ValidWithProperties(t *testing.T) {
	schema := `{"properties": {"name": {"type": "string"}}}`
	
	err := ValidateSchema(schema, SchemaTypeJSON)
	if err != nil {
		t.Errorf("JSON Schema with properties should pass validation: %v", err)
	}
}

func TestJSONSchemaValidator_ValidWithRef(t *testing.T) {
	schema := `{"$ref": "#/definitions/User"}`
	
	err := ValidateSchema(schema, SchemaTypeJSON)
	if err != nil {
		t.Errorf("JSON Schema with $ref should pass validation: %v", err)
	}
}

func TestJSONSchemaValidator_InvalidJSON(t *testing.T) {
	schema := `{type: "object", invalid}`
	
	err := ValidateSchema(schema, SchemaTypeJSON)
	if err == nil {
		t.Error("invalid JSON should fail validation")
	}
}

func TestJSONSchemaValidator_MissingTypicalFields(t *testing.T) {
	schema := `{"title": "Something", "description": "A schema"}`
	
	err := ValidateSchema(schema, SchemaTypeJSON)
	if err == nil {
		t.Error("JSON Schema without typical fields should fail validation")
	}
}

// Protobuf Schema Validation Tests

func TestProtobufValidator_ValidMessageSchema(t *testing.T) {
	schema := `
syntax = "proto3";

message User {
	int32 id = 1;
	string name = 2;
}
`
	
	err := ValidateSchema(schema, SchemaTypeProtobuf)
	if err != nil {
		t.Errorf("valid Protobuf schema should pass validation: %v", err)
	}
}

func TestProtobufValidator_ValidWithPackage(t *testing.T) {
	schema := `package example;`
	
	err := ValidateSchema(schema, SchemaTypeProtobuf)
	if err != nil {
		t.Errorf("Protobuf schema with package should pass validation: %v", err)
	}
}

func TestProtobufValidator_ValidWithEnum(t *testing.T) {
	schema := `
enum Status {
	ACTIVE = 0;
	INACTIVE = 1;
}
`
	
	err := ValidateSchema(schema, SchemaTypeProtobuf)
	if err != nil {
		t.Errorf("Protobuf schema with enum should pass validation: %v", err)
	}
}

func TestProtobufValidator_ValidWithService(t *testing.T) {
	schema := `
service UserService {
	rpc GetUser (UserRequest) returns (UserResponse);
}
`
	
	err := ValidateSchema(schema, SchemaTypeProtobuf)
	if err != nil {
		t.Errorf("Protobuf schema with service should pass validation: %v", err)
	}
}

func TestProtobufValidator_MissingKeywords(t *testing.T) {
	schema := `just some random text without proto keywords`
	
	err := ValidateSchema(schema, SchemaTypeProtobuf)
	if err == nil {
		t.Error("Protobuf schema without keywords should fail validation")
	}
}

func TestProtobufValidator_EmptySchema(t *testing.T) {
	schema := ``
	
	err := ValidateSchema(schema, SchemaTypeProtobuf)
	if err == nil {
		t.Error("empty Protobuf schema should fail validation")
	}
}

// General Validation Tests

func TestValidateSchema_EmptySchema(t *testing.T) {
	err := ValidateSchema("", SchemaTypeAvro)
	if err == nil {
		t.Error("empty schema should fail validation")
	}
}

func TestValidateSchema_UnsupportedType(t *testing.T) {
	schema := `{"type": "record"}`
	err := ValidateSchema(schema, "UNKNOWN")
	if err == nil {
		t.Error("unsupported schema type should fail validation")
	}
}
