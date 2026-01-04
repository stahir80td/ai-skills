package errors

import (
	goerrors "errors"

	"github.com/your-github-org/ai-scaffolder/core/go/errors"
)

// Common errors for order validation
var (
	ErrInvalidCustomerID = goerrors.New("customer ID is required")
	ErrEmptyOrderItems   = goerrors.New("order must have at least one item")
	ErrOrderNotFound     = goerrors.New("order not found")
	ErrUserNotFound      = goerrors.New("user not found")
	ErrInvalidEmail      = goerrors.New("valid email address is required")
	ErrInvalidDeviceID   = goerrors.New("device ID is required")
	ErrSessionNotFound   = goerrors.New("session not found")
)

// ProductErrors is the error registry for product/patterns domain
var ProductErrors = errors.NewErrorRegistry()

func init() {
	// Product entity errors (PRD = Product)
	ProductErrors.Register(&errors.ErrorDefinition{
		Code:        "PAT-PRD-001",
		Severity:    errors.SeverityMedium,
		Description: "Product not found: %v",
		SODScore:    120, // 4 × 5 × 6
		Severity_S:  4,
		Occurrence:  5,
		Detect_D:    6,
		Mitigation:  "Verify product ID exists in database before lookup",
		Example:     "User requests product that was deleted or never existed",
	})

	ProductErrors.Register(&errors.ErrorDefinition{
		Code:        "PAT-PRD-002",
		Severity:    errors.SeverityLow,
		Description: "Product name is required",
		SODScore:    30, // 2 × 3 × 5
		Severity_S:  2,
		Occurrence:  3,
		Detect_D:    5,
		Mitigation:  "Validate product name in request payload",
		Example:     "API request missing required name field",
	})

	ProductErrors.Register(&errors.ErrorDefinition{
		Code:        "PAT-PRD-003",
		Severity:    errors.SeverityLow,
		Description: "Product price must be positive: %v",
		SODScore:    24, // 2 × 3 × 4
		Severity_S:  2,
		Occurrence:  3,
		Detect_D:    4,
		Mitigation:  "Validate price > 0 before persistence",
		Example:     "API request with negative or zero price",
	})

	ProductErrors.Register(&errors.ErrorDefinition{
		Code:        "PAT-PRD-004",
		Severity:    errors.SeverityLow,
		Description: "Product category is required",
		SODScore:    30, // 2 × 3 × 5
		Severity_S:  2,
		Occurrence:  3,
		Detect_D:    5,
		Mitigation:  "Validate category in request payload",
		Example:     "API request missing required category field",
	})

	ProductErrors.Register(&errors.ErrorDefinition{
		Code:        "PAT-PRD-005",
		Severity:    errors.SeverityLow,
		Description: "Stock quantity cannot be negative: %v",
		SODScore:    36, // 3 × 3 × 4
		Severity_S:  3,
		Occurrence:  3,
		Detect_D:    4,
		Mitigation:  "Validate quantity >= 0 before update",
		Example:     "Inventory update with negative quantity",
	})

	ProductErrors.Register(&errors.ErrorDefinition{
		Code:        "PAT-PRD-006",
		Severity:    errors.SeverityMedium,
		Description: "Invalid status transition from %v to %v",
		SODScore:    80, // 4 × 4 × 5
		Severity_S:  4,
		Occurrence:  4,
		Detect_D:    5,
		Mitigation:  "Implement state machine validation for status transitions",
		Example:     "Trying to transition from 'delivered' to 'pending'",
	})

	ProductErrors.Register(&errors.ErrorDefinition{
		Code:        "PAT-PRD-007",
		Severity:    errors.SeverityMedium,
		Description: "Cannot delete active product: %v",
		SODScore:    96, // 4 × 4 × 6
		Severity_S:  4,
		Occurrence:  4,
		Detect_D:    6,
		Mitigation:  "Check product status before deletion, require deactivation first",
		Example:     "Delete request for product currently listed in orders",
	})

	ProductErrors.Register(&errors.ErrorDefinition{
		Code:        "PAT-PRD-008",
		Severity:    errors.SeverityMedium,
		Description: "Insufficient stock: available=%v, requested=%v",
		SODScore:    60, // 3 × 5 × 4
		Severity_S:  3,
		Occurrence:  5,
		Detect_D:    4,
		Mitigation:  "Check stock availability before order creation",
		Example:     "Order for 10 units when only 5 available",
	})

	// Order entity errors (ORD = Order)
	ProductErrors.Register(&errors.ErrorDefinition{
		Code:        "PAT-ORD-001",
		Severity:    errors.SeverityMedium,
		Description: "Order not found: %v",
		SODScore:    100, // 4 × 5 × 5
		Severity_S:  4,
		Occurrence:  5,
		Detect_D:    5,
		Mitigation:  "Verify order ID exists before lookup",
		Example:     "Customer queries non-existent order",
	})

	ProductErrors.Register(&errors.ErrorDefinition{
		Code:        "PAT-ORD-002",
		Severity:    errors.SeverityLow,
		Description: "Order must have at least one item",
		SODScore:    24, // 2 × 3 × 4
		Severity_S:  2,
		Occurrence:  3,
		Detect_D:    4,
		Mitigation:  "Validate items array is not empty",
		Example:     "Create order request with empty items array",
	})

	ProductErrors.Register(&errors.ErrorDefinition{
		Code:        "PAT-ORD-003",
		Severity:    errors.SeverityMedium,
		Description: "Order already cancelled: %v",
		SODScore:    48, // 3 × 4 × 4
		Severity_S:  3,
		Occurrence:  4,
		Detect_D:    4,
		Mitigation:  "Check order status before modification",
		Example:     "Attempt to ship already cancelled order",
	})

	// User entity errors (USR = User)
	ProductErrors.Register(&errors.ErrorDefinition{
		Code:        "PAT-USR-001",
		Severity:    errors.SeverityMedium,
		Description: "User not found: %v",
		SODScore:    100, // 4 × 5 × 5
		Severity_S:  4,
		Occurrence:  5,
		Detect_D:    5,
		Mitigation:  "Verify user ID exists before operations",
		Example:     "Profile update for non-existent user",
	})

	ProductErrors.Register(&errors.ErrorDefinition{
		Code:        "PAT-USR-002",
		Severity:    errors.SeverityLow,
		Description: "Email is required",
		SODScore:    24, // 2 × 3 × 4
		Severity_S:  2,
		Occurrence:  3,
		Detect_D:    4,
		Mitigation:  "Validate email field in request",
		Example:     "User registration without email",
	})

	ProductErrors.Register(&errors.ErrorDefinition{
		Code:        "PAT-USR-003",
		Severity:    errors.SeverityMedium,
		Description: "Email already registered: %v",
		SODScore:    72, // 3 × 6 × 4
		Severity_S:  3,
		Occurrence:  6,
		Detect_D:    4,
		Mitigation:  "Check email uniqueness before registration",
		Example:     "Registration with existing email address",
	})

	// Infrastructure errors (INFRA)
	ProductErrors.Register(&errors.ErrorDefinition{
		Code:        "PAT-INFRA-001",
		Severity:    errors.SeverityHigh,
		Description: "Database connection failed: %v",
		SODScore:    280, // 7 × 5 × 8
		Severity_S:  7,
		Occurrence:  5,
		Detect_D:    8,
		Mitigation:  "Check database connectivity, connection string, credentials",
		Example:     "SQL Server unavailable during connection pool initialization",
	})

	ProductErrors.Register(&errors.ErrorDefinition{
		Code:        "PAT-INFRA-002",
		Severity:    errors.SeverityMedium,
		Description: "Cache operation failed: %v",
		SODScore:    120, // 4 × 5 × 6
		Severity_S:  4,
		Occurrence:  5,
		Detect_D:    6,
		Mitigation:  "Implement graceful degradation for cache failures",
		Example:     "Redis connection timeout during cache write",
	})

	ProductErrors.Register(&errors.ErrorDefinition{
		Code:        "PAT-INFRA-003",
		Severity:    errors.SeverityHigh,
		Description: "External service unavailable: %v",
		SODScore:    252, // 6 × 6 × 7
		Severity_S:  6,
		Occurrence:  6,
		Detect_D:    7,
		Mitigation:  "Implement circuit breaker and fallback mechanisms",
		Example:     "Kafka broker unreachable during event publish",
	})

	ProductErrors.Register(&errors.ErrorDefinition{
		Code:        "PAT-INFRA-004",
		Severity:    errors.SeverityMedium,
		Description: "Database query failed: %v",
		SODScore:    100, // 4 × 5 × 5
		Severity_S:  4,
		Occurrence:  5,
		Detect_D:    5,
		Mitigation:  "Review query syntax, check table existence, validate parameters",
		Example:     "SQL syntax error or missing table",
	})

	ProductErrors.Register(&errors.ErrorDefinition{
		Code:        "PAT-INFRA-005",
		Severity:    errors.SeverityHigh,
		Description: "Time-series database error: %v",
		SODScore:    224, // 7 × 4 × 8
		Severity_S:  7,
		Occurrence:  4,
		Detect_D:    8,
		Mitigation:  "Check ScyllaDB cluster health, verify keyspace configuration",
		Example:     "ScyllaDB query timeout during telemetry insert",
	})

	// Validation errors (VAL)
	ProductErrors.Register(&errors.ErrorDefinition{
		Code:        "PAT-VAL-001",
		Severity:    errors.SeverityLow,
		Description: "Invalid request format: %v",
		SODScore:    18, // 2 × 3 × 3
		Severity_S:  2,
		Occurrence:  3,
		Detect_D:    3,
		Mitigation:  "Validate JSON schema before processing",
		Example:     "Malformed JSON in request body",
	})

	ProductErrors.Register(&errors.ErrorDefinition{
		Code:        "PAT-VAL-002",
		Severity:    errors.SeverityLow,
		Description: "Missing required parameter: %v",
		SODScore:    24, // 2 × 4 × 3
		Severity_S:  2,
		Occurrence:  4,
		Detect_D:    3,
		Mitigation:  "Check required fields in request handler",
		Example:     "API call missing required query parameter",
	})

	ProductErrors.Register(&errors.ErrorDefinition{
		Code:        "PAT-VAL-003",
		Severity:    errors.SeverityLow,
		Description: "Invalid UUID format: %v",
		SODScore:    18, // 2 × 3 × 3
		Severity_S:  2,
		Occurrence:  3,
		Detect_D:    3,
		Mitigation:  "Validate UUID format before parsing",
		Example:     "Order ID with invalid UUID format",
	})

	// Telemetry errors (TEL)
	ProductErrors.Register(&errors.ErrorDefinition{
		Code:        "PAT-TEL-001",
		Severity:    errors.SeverityMedium,
		Description: "Device not found: %v",
		SODScore:    80, // 4 × 4 × 5
		Severity_S:  4,
		Occurrence:  4,
		Detect_D:    5,
		Mitigation:  "Verify device ID exists in device registry",
		Example:     "Telemetry query for unknown device",
	})

	ProductErrors.Register(&errors.ErrorDefinition{
		Code:        "PAT-TEL-002",
		Severity:    errors.SeverityMedium,
		Description: "Anomaly detected for device %v: %v",
		SODScore:    150, // 5 × 5 × 6
		Severity_S:  5,
		Occurrence:  5,
		Detect_D:    6,
		Mitigation:  "Alert operations team, check device status",
		Example:     "Temperature reading outside normal range",
	})
}

// Convenience functions for creating specific errors

// NotFound creates a "not found" error
func NotFound(entityType string, id interface{}) *errors.ServiceError {
	switch entityType {
	case "product":
		return ProductErrors.CreateError("PAT-PRD-001", id)
	case "order":
		return ProductErrors.CreateError("PAT-ORD-001", id)
	case "user":
		return ProductErrors.CreateError("PAT-USR-001", id)
	case "device":
		return ProductErrors.CreateError("PAT-TEL-001", id)
	default:
		return ProductErrors.CreateError("PAT-PRD-001", id)
	}
}

// InvalidStatusTransition creates an invalid status transition error
func InvalidStatusTransition(from, to interface{}) *errors.ServiceError {
	return ProductErrors.CreateError("PAT-PRD-006", from, to)
}

// InsufficientStock creates an insufficient stock error
func InsufficientStock(available, requested int) *errors.ServiceError {
	return ProductErrors.CreateError("PAT-PRD-008", available, requested)
}

// DatabaseError wraps a database error
func DatabaseError(err error) *errors.ServiceError {
	return ProductErrors.WrapError(err, "PAT-INFRA-001")
}

// CacheError wraps a cache operation error
func CacheError(err error) *errors.ServiceError {
	return ProductErrors.WrapError(err, "PAT-INFRA-002")
}

// ExternalServiceError wraps an external service error
func ExternalServiceError(serviceName string, err error) *errors.ServiceError {
	return ProductErrors.WrapError(err, "PAT-INFRA-003", serviceName)
}

// ValidationError creates a validation error
func ValidationError(details string) *errors.ServiceError {
	return ProductErrors.CreateError("PAT-VAL-001", details)
}

// MissingParameter creates a missing parameter error
func MissingParameter(paramName string) *errors.ServiceError {
	return ProductErrors.CreateError("PAT-VAL-002", paramName)
}

// InvalidUUID creates an invalid UUID error
func InvalidUUID(value string) *errors.ServiceError {
	return ProductErrors.CreateError("PAT-VAL-003", value)
}

// AnomalyDetected creates an anomaly detected error
func AnomalyDetected(deviceID, anomalyType string) *errors.ServiceError {
	return ProductErrors.CreateError("PAT-TEL-002", deviceID, anomalyType)
}

// GetAllErrorCodes returns all registered error codes
func GetAllErrorCodes() []string {
	codes := []string{}
	for code := range ProductErrors.GetAll() {
		codes = append(codes, code)
	}
	return codes
}
