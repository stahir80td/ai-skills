package errors

import (
	"fmt"
)

// Severity levels for errors
const (
	SeverityCritical = "CRITICAL" // System unavailable, data loss, security breach
	SeverityHigh     = "HIGH"     // Major functionality broken, significant impact
	SeverityMedium   = "MEDIUM"   // Moderate impact, workaround available
	SeverityLow      = "LOW"      // Minor issue, minimal impact
	SeverityInfo     = "INFO"     // Informational, not an error
)

// ServiceError represents a structured error with code, severity, and context
type ServiceError struct {
	Code       string                 // Error code (e.g., "INGEST-001")
	Message    string                 // Human-readable error message
	Severity   string                 // Severity level (CRITICAL, HIGH, MEDIUM, LOW, INFO)
	Underlying error                  // Original error (if any)
	Context    map[string]interface{} // Additional context (user_id, request_id, etc.)
}

// Error implements the error interface
func (e *ServiceError) Error() string {
	if e.Underlying != nil {
		return fmt.Sprintf("[%s] %s: %s (caused by: %v)", e.Code, e.Severity, e.Message, e.Underlying)
	}
	return fmt.Sprintf("[%s] %s: %s", e.Code, e.Severity, e.Message)
}

// Unwrap returns the underlying error for error wrapping
func (e *ServiceError) Unwrap() error {
	return e.Underlying
}

// WithContext adds additional context to the error
func (e *ServiceError) WithContext(key string, value interface{}) *ServiceError {
	if e.Context == nil {
		e.Context = make(map[string]interface{})
	}
	e.Context[key] = value
	return e
}

// GetContext retrieves a context value by key
func (e *ServiceError) GetContext(key string) (interface{}, bool) {
	if e.Context == nil {
		return nil, false
	}
	val, ok := e.Context[key]
	return val, ok
}

// New creates a new ServiceError
func New(code, severity, message string) *ServiceError {
	return &ServiceError{
		Code:     code,
		Message:  message,
		Severity: severity,
		Context:  make(map[string]interface{}),
	}
}

// Wrap wraps an existing error with ServiceError metadata
func Wrap(err error, code, severity, message string) *ServiceError {
	return &ServiceError{
		Code:       code,
		Message:    message,
		Severity:   severity,
		Underlying: err,
		Context:    make(map[string]interface{}),
	}
}

// ErrorDefinition represents a registered error with SOD scores
type ErrorDefinition struct {
	Code        string // Error code (e.g., "INGEST-001")
	Severity    string // Severity level
	Description string // Detailed description
	SODScore    int    // Severity × Occurrence × Detectability (1-1000)
	Severity_S  int    // Severity score (1-10)
	Occurrence  int    // Occurrence score (1-10)
	Detect_D    int    // Detectability score (1-10)
	Mitigation  string // How to resolve this error
	Example     string // Example scenario when this error occurs
}

// ErrorRegistry manages registered error definitions
type ErrorRegistry struct {
	definitions map[string]*ErrorDefinition
}

// NewErrorRegistry creates a new error registry
func NewErrorRegistry() *ErrorRegistry {
	return &ErrorRegistry{
		definitions: make(map[string]*ErrorDefinition),
	}
}

// Register adds an error definition to the registry
func (r *ErrorRegistry) Register(def *ErrorDefinition) {
	r.definitions[def.Code] = def
}

// Get retrieves an error definition by code
func (r *ErrorRegistry) Get(code string) (*ErrorDefinition, bool) {
	def, ok := r.definitions[code]
	return def, ok
}

// GetAll returns all registered error definitions
func (r *ErrorRegistry) GetAll() map[string]*ErrorDefinition {
	return r.definitions
}

// CreateError creates a ServiceError from a registered error definition
func (r *ErrorRegistry) CreateError(code string, messageArgs ...interface{}) *ServiceError {
	def, ok := r.Get(code)
	if !ok {
		return New(code, SeverityMedium, fmt.Sprintf("Unknown error: %s", code))
	}

	message := def.Description
	if len(messageArgs) > 0 {
		message = fmt.Sprintf(def.Description, messageArgs...)
	}

	return New(code, def.Severity, message)
}

// WrapError wraps an existing error using a registered error definition
func (r *ErrorRegistry) WrapError(err error, code string, messageArgs ...interface{}) *ServiceError {
	def, ok := r.Get(code)
	if !ok {
		return Wrap(err, code, SeverityMedium, fmt.Sprintf("Unknown error: %s", code))
	}

	message := def.Description
	if len(messageArgs) > 0 {
		message = fmt.Sprintf(def.Description, messageArgs...)
	}

	return Wrap(err, code, def.Severity, message)
}

// CalculateSOD calculates the SOD score (Severity × Occurrence × Detectability)
func CalculateSOD(severity, occurrence, detectability int) int {
	return severity * occurrence * detectability
}
