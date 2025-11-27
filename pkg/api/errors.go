// Package api defines the API types and error handling for Confluent Cloud and Platform APIs.
package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
)

// Error represents a Confluent API error with structured information.
type Error struct {
	// Code is the HTTP status code
	Code int
	// ErrorCode is the Confluent-specific error code (if provided by API)
	ErrorCode string
	// Message is the error message
	Message string
	// Details contains additional error details from the API
	Details map[string]interface{}
	// Err is the underlying error (if any)
	Err error
}

// Error implements the error interface.
func (e *Error) Error() string {
	if e.ErrorCode != "" {
		return fmt.Sprintf("confluent error %s (%d): %s", e.ErrorCode, e.Code, e.Message)
	}
	return fmt.Sprintf("confluent error (%d): %s", e.Code, e.Message)
}

// Is implements error comparison for use with errors.Is().
func (e *Error) Is(target error) bool {
	t, ok := target.(*Error)
	if !ok {
		return false
	}
	return e.Code == t.Code
}

// Unwrap returns the underlying error.
func (e *Error) Unwrap() error {
	return e.Err
}

// IsBadRequest returns true if this is a 400 Bad Request error.
func (e *Error) IsBadRequest() bool {
	return e.Code == http.StatusBadRequest
}

// IsUnauthorized returns true if this is a 401 Unauthorized error.
func (e *Error) IsUnauthorized() bool {
	return e.Code == http.StatusUnauthorized
}

// IsForbidden returns true if this is a 403 Forbidden error.
func (e *Error) IsForbidden() bool {
	return e.Code == http.StatusForbidden
}

// IsNotFound returns true if this is a 404 Not Found error.
func (e *Error) IsNotFound() bool {
	return e.Code == http.StatusNotFound
}

// IsConflict returns true if this is a 409 Conflict error.
func (e *Error) IsConflict() bool {
	return e.Code == http.StatusConflict
}

// IsRateLimited returns true if this is a 429 Too Many Requests error.
func (e *Error) IsRateLimited() bool {
	return e.Code == http.StatusTooManyRequests
}

// IsInternalServerError returns true if this is a 500+ error.
func (e *Error) IsInternalServerError() bool {
	return e.Code >= http.StatusInternalServerError
}

// RetryAfter returns the Retry-After duration in seconds if available.
// Returns 0 if not available.
func (e *Error) RetryAfter() int {
	if e.Code != http.StatusTooManyRequests {
		return 0
	}
	if details, ok := e.Details["retry_after"].(string); ok {
		if seconds, err := strconv.Atoi(details); err == nil {
			return seconds
		}
	}
	// Default retry after for rate limiting
	return 60
}

// apiErrorResponse represents the standard Confluent API error response format.
type apiErrorResponse struct {
	ErrorCode string `json:"error_code"`
	Message   string `json:"message"`
}

// NewError creates a new Confluent API error from an HTTP status code and response body.
func NewError(statusCode int, responseBody []byte, headers http.Header) *Error {
	err := &Error{
		Code:    statusCode,
		Details: make(map[string]interface{}),
	}

	// Try to parse Confluent-specific error response
	if len(responseBody) > 0 {
		var apiErr apiErrorResponse
		if json.Unmarshal(responseBody, &apiErr) == nil {
			err.ErrorCode = apiErr.ErrorCode
			err.Message = apiErr.Message
		}

		// Also try to parse as generic JSON
		var jsonBody map[string]interface{}
		if json.Unmarshal(responseBody, &jsonBody) == nil {
			err.Details = jsonBody
		} else {
			// If not JSON, use raw response as message
			err.Message = string(responseBody)
		}
	}

	// Extract retry-after header if present
	if retryAfter := headers.Get("Retry-After"); retryAfter != "" {
		err.Details["retry_after"] = retryAfter
	}

	// Set default message if not provided
	if err.Message == "" {
		err.Message = http.StatusText(statusCode)
	}

	// Set error code based on HTTP status if not provided
	if err.ErrorCode == "" {
		err.ErrorCode = StatusCodeToErrorCode(statusCode)
	}

	return err
}

// StatusCodeToErrorCode converts an HTTP status code to a Confluent error code string.
func StatusCodeToErrorCode(statusCode int) string {
	switch statusCode {
	case http.StatusBadRequest:
		return "INVALID_REQUEST"
	case http.StatusUnauthorized:
		return "UNAUTHORIZED"
	case http.StatusForbidden:
		return "FORBIDDEN"
	case http.StatusNotFound:
		return "NOT_FOUND"
	case http.StatusConflict:
		return "CONFLICT"
	case http.StatusTooManyRequests:
		return "RATE_LIMIT_EXCEEDED"
	case http.StatusInternalServerError:
		return "INTERNAL_SERVER_ERROR"
	case http.StatusBadGateway:
		return "BAD_GATEWAY"
	case http.StatusServiceUnavailable:
		return "SERVICE_UNAVAILABLE"
	case http.StatusGatewayTimeout:
		return "GATEWAY_TIMEOUT"
	default:
		if statusCode >= 500 {
			return "SERVER_ERROR"
		}
		if statusCode >= 400 {
			return "CLIENT_ERROR"
		}
		return "UNKNOWN_ERROR"
	}
}

// IsRetryable returns true if the error is retryable (rate limiting or server errors).
func (e *Error) IsRetryable() bool {
	return e.IsRateLimited() || e.IsInternalServerError()
}

// Common error codes
const (
	ErrorCodeInvalidRequest     = "INVALID_REQUEST"
	ErrorCodeUnauthorized       = "UNAUTHORIZED"
	ErrorCodeForbidden          = "FORBIDDEN"
	ErrorCodeNotFound           = "NOT_FOUND"
	ErrorCodeConflict           = "CONFLICT"
	ErrorCodeRateLimitExceeded  = "RATE_LIMIT_EXCEEDED"
	ErrorCodeInternalServer     = "INTERNAL_SERVER_ERROR"
	ErrorCodeBadGateway         = "BAD_GATEWAY"
	ErrorCodeServiceUnavailable = "SERVICE_UNAVAILABLE"
	ErrorCodeGatewayTimeout     = "GATEWAY_TIMEOUT"
)

// ParseErrorFromResponse parses an error response and returns an error message and error code.
func ParseErrorFromResponse(statusCode int, body []byte) (message string, errorCode string) {
	var apiErr apiErrorResponse
	if err := json.Unmarshal(body, &apiErr); err == nil && apiErr.Message != "" {
		return apiErr.Message, apiErr.ErrorCode
	}

	// Try to extract error from generic JSON response
	var jsonBody map[string]interface{}
	if err := json.Unmarshal(body, &jsonBody); err == nil {
		if msg, ok := jsonBody["message"].(string); ok {
			return msg, StatusCodeToErrorCode(statusCode)
		}
		if msg, ok := jsonBody["error"].(string); ok {
			return msg, StatusCodeToErrorCode(statusCode)
		}
	}

	// Fallback to raw body
	bodyStr := strings.TrimSpace(string(body))
	if bodyStr != "" {
		return bodyStr, StatusCodeToErrorCode(statusCode)
	}

	return http.StatusText(statusCode), StatusCodeToErrorCode(statusCode)
}
