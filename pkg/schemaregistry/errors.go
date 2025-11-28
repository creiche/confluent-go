package schemaregistry

import (
	"errors"

	"github.com/creiche/confluent-go/pkg/api"
)

// Schema Registry error codes as defined by the Confluent Schema Registry API.
// These codes are returned in the error_code field of API error responses.
//
// See: https://docs.confluent.io/platform/current/schema-registry/develop/api.html#errors
const (
	// Subject errors
	ErrorCodeSubjectNotFound    = 40401
	ErrorCodeSubjectSoftDeleted = 40404
	ErrorCodeInvalidSubject     = 42202

	// Version errors
	ErrorCodeVersionNotFound = 40402

	// Schema errors
	ErrorCodeSchemaNotFound     = 40403
	ErrorCodeInvalidSchema      = 42201
	ErrorCodeIncompatibleSchema = 409

	// Compatibility errors
	ErrorCodeInvalidCompatibility = 42203

	// Mode errors
	ErrorCodeInvalidMode = 42204
)

// GetSRCode extracts the Schema Registry error code from an error.
// Returns the error code and true if found, otherwise 0 and false.
func GetSRCode(err error) (int, bool) {
	if err == nil {
		return 0, false
	}

	// Extract from api.Error Details (SR returns error_code as integer in JSON)
	var apiErr *api.Error
	if errors.As(err, &apiErr) {
		if errorCodeFloat, ok := apiErr.Details["error_code"].(float64); ok {
			return int(errorCodeFloat), true
		}
	}

	return 0, false
}

// Helper methods for common error conditions

// IsSubjectNotFound returns true if the error is a subject not found error (40401)
func IsSubjectNotFound(err error) bool {
	code, ok := GetSRCode(err)
	return ok && code == ErrorCodeSubjectNotFound
}

// IsSubjectSoftDeleted returns true if the error is a soft-deleted subject error (40404)
func IsSubjectSoftDeleted(err error) bool {
	code, ok := GetSRCode(err)
	return ok && code == ErrorCodeSubjectSoftDeleted
}

// IsVersionNotFound returns true if the error is a version not found error (40402)
func IsVersionNotFound(err error) bool {
	code, ok := GetSRCode(err)
	return ok && code == ErrorCodeVersionNotFound
}

// IsSchemaNotFound returns true if the error is a schema not found error (40403)
func IsSchemaNotFound(err error) bool {
	code, ok := GetSRCode(err)
	return ok && code == ErrorCodeSchemaNotFound
}

// IsInvalidSchema returns true if the error is an invalid schema error (42201)
func IsInvalidSchema(err error) bool {
	code, ok := GetSRCode(err)
	return ok && code == ErrorCodeInvalidSchema
}

// IsIncompatibleSchema returns true if the error is an incompatible schema error (409)
func IsIncompatibleSchema(err error) bool {
	code, ok := GetSRCode(err)
	return ok && code == ErrorCodeIncompatibleSchema
}

// IsInvalidCompatibility returns true if the error is an invalid compatibility level error (42203)
func IsInvalidCompatibility(err error) bool {
	code, ok := GetSRCode(err)
	return ok && code == ErrorCodeInvalidCompatibility
}

// IsInvalidSubject returns true if the error is an invalid subject error (42202)
func IsInvalidSubject(err error) bool {
	code, ok := GetSRCode(err)
	return ok && code == ErrorCodeInvalidSubject
}

// IsInvalidMode returns true if the error is an invalid mode error (42204)
func IsInvalidMode(err error) bool {
	code, ok := GetSRCode(err)
	return ok && code == ErrorCodeInvalidMode
}
