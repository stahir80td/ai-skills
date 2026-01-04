package validation

import (
	"context"
	"fmt"
	"time"

	"github.com/your-github-org/ai-scaffolder/core/go/errors"
	"github.com/your-github-org/ai-scaffolder/core/go/logger"
	"go.uber.org/zap"
)

// Validator provides data quality validation capabilities
type Validator struct {
	logger *logger.Logger
}

// Config holds validator configuration
type Config struct {
	Logger *logger.Logger
}

// ValidationResult represents the outcome of a validation check
type ValidationResult struct {
	IsValid      bool
	ErrorCode    string
	ErrorMessage string
	FailedChecks []string
	Metadata     map[string]interface{}
}

// SchemaField defines expected structure for a field
type SchemaField struct {
	Name       string
	Type       string // "string", "int", "float", "bool", "timestamp"
	Required   bool
	AllowNull  bool
	MinValue   *float64
	MaxValue   *float64
	MinLength  *int
	MaxLength  *int
	Pattern    string // Regex pattern for string validation
	EnumValues []string
}

// Schema defines the expected structure of data
type Schema struct {
	Fields []SchemaField
}

// NewValidator creates a new data validator with dependency injection
func NewValidator(cfg Config) *Validator {
	return &Validator{
		logger: cfg.Logger,
	}
}

// ValidateSchema validates data against a schema definition
func (v *Validator) ValidateSchema(ctx context.Context, data map[string]interface{}, schema *Schema) *ValidationResult {
	result := &ValidationResult{
		IsValid:      true,
		FailedChecks: []string{},
		Metadata:     make(map[string]interface{}),
	}

	for _, field := range schema.Fields {
		value, exists := data[field.Name]

		// Check required fields
		if field.Required && !exists {
			result.IsValid = false
			result.FailedChecks = append(result.FailedChecks,
				fmt.Sprintf("Required field '%s' is missing", field.Name))
			continue
		}

		if !exists {
			continue // Optional field not present
		}

		// Check null values
		if value == nil {
			if !field.AllowNull {
				result.IsValid = false
				result.FailedChecks = append(result.FailedChecks,
					fmt.Sprintf("Field '%s' cannot be null", field.Name))
			}
			continue
		}

		// Type validation
		if !v.validateType(value, field.Type) {
			result.IsValid = false
			result.FailedChecks = append(result.FailedChecks,
				fmt.Sprintf("Field '%s' has invalid type, expected %s", field.Name, field.Type))
			continue
		}

		// Range validation for numeric types
		if field.Type == "int" || field.Type == "float" {
			numValue := v.toFloat64(value)
			if field.MinValue != nil && numValue < *field.MinValue {
				result.IsValid = false
				result.FailedChecks = append(result.FailedChecks,
					fmt.Sprintf("Field '%s' value %f is below minimum %f", field.Name, numValue, *field.MinValue))
			}
			if field.MaxValue != nil && numValue > *field.MaxValue {
				result.IsValid = false
				result.FailedChecks = append(result.FailedChecks,
					fmt.Sprintf("Field '%s' value %f exceeds maximum %f", field.Name, numValue, *field.MaxValue))
			}
		}

		// Length validation for strings
		if field.Type == "string" {
			strValue, ok := value.(string)
			if ok {
				length := len(strValue)
				if field.MinLength != nil && length < *field.MinLength {
					result.IsValid = false
					result.FailedChecks = append(result.FailedChecks,
						fmt.Sprintf("Field '%s' length %d is below minimum %d", field.Name, length, *field.MinLength))
				}
				if field.MaxLength != nil && length > *field.MaxLength {
					result.IsValid = false
					result.FailedChecks = append(result.FailedChecks,
						fmt.Sprintf("Field '%s' length %d exceeds maximum %d", field.Name, length, *field.MaxLength))
				}
			}
		}

		// Enum validation
		if len(field.EnumValues) > 0 {
			if !v.isInEnum(value, field.EnumValues) {
				result.IsValid = false
				result.FailedChecks = append(result.FailedChecks,
					fmt.Sprintf("Field '%s' value is not in allowed values", field.Name))
			}
		}
	}

	if !result.IsValid {
		result.ErrorCode = "VALIDATION-001"
		result.ErrorMessage = fmt.Sprintf("Schema validation failed with %d errors", len(result.FailedChecks))

		v.logger.Warn("Schema validation failed",
			zap.Int("failed_checks", len(result.FailedChecks)),
			zap.Strings("failures", result.FailedChecks),
		)
	} else {
		v.logger.Debug("Schema validation passed",
			zap.Int("fields_validated", len(schema.Fields)),
		)
	}

	return result
}

// ValidateRange checks if a numeric value is within the specified range
func (v *Validator) ValidateRange(ctx context.Context, value float64, min, max float64, fieldName string) *ValidationResult {
	result := &ValidationResult{
		IsValid:      true,
		FailedChecks: []string{},
		Metadata:     map[string]interface{}{"field": fieldName, "value": value},
	}

	if value < min {
		result.IsValid = false
		result.ErrorCode = "VALIDATION-002"
		result.ErrorMessage = fmt.Sprintf("%s value %f is below minimum %f", fieldName, value, min)
		result.FailedChecks = append(result.FailedChecks, result.ErrorMessage)

		v.logger.Warn("Range validation failed - below minimum",
			zap.String("field", fieldName),
			zap.Float64("value", value),
			zap.Float64("min", min),
		)
	}

	if value > max {
		result.IsValid = false
		result.ErrorCode = "VALIDATION-002"
		result.ErrorMessage = fmt.Sprintf("%s value %f exceeds maximum %f", fieldName, value, max)
		result.FailedChecks = append(result.FailedChecks, result.ErrorMessage)

		v.logger.Warn("Range validation failed - exceeds maximum",
			zap.String("field", fieldName),
			zap.Float64("value", value),
			zap.Float64("max", max),
		)
	}

	return result
}

// ValidateNotNull checks if required fields are not null
func (v *Validator) ValidateNotNull(ctx context.Context, data map[string]interface{}, requiredFields []string) *ValidationResult {
	result := &ValidationResult{
		IsValid:      true,
		FailedChecks: []string{},
		Metadata:     make(map[string]interface{}),
	}

	for _, field := range requiredFields {
		value, exists := data[field]

		if !exists || value == nil {
			result.IsValid = false
			result.FailedChecks = append(result.FailedChecks,
				fmt.Sprintf("Required field '%s' is missing or null", field))
		}
	}

	if !result.IsValid {
		result.ErrorCode = "VALIDATION-003"
		result.ErrorMessage = fmt.Sprintf("Null validation failed for %d fields", len(result.FailedChecks))

		v.logger.Warn("Null validation failed",
			zap.Int("failed_fields", len(result.FailedChecks)),
			zap.Strings("failures", result.FailedChecks),
		)
	}

	return result
}

// DetectOutliers identifies statistical outliers using IQR method
func (v *Validator) DetectOutliers(ctx context.Context, values []float64, threshold float64) []int {
	if len(values) < 4 {
		return []int{}
	}

	// Calculate quartiles
	sorted := make([]float64, len(values))
	copy(sorted, values)

	// Simple sort for quartile calculation
	for i := 0; i < len(sorted); i++ {
		for j := i + 1; j < len(sorted); j++ {
			if sorted[i] > sorted[j] {
				sorted[i], sorted[j] = sorted[j], sorted[i]
			}
		}
	}

	q1Index := len(sorted) / 4
	q3Index := 3 * len(sorted) / 4

	q1 := sorted[q1Index]
	q3 := sorted[q3Index]
	iqr := q3 - q1

	lowerBound := q1 - threshold*iqr
	upperBound := q3 + threshold*iqr

	outlierIndices := []int{}
	for i, val := range values {
		if val < lowerBound || val > upperBound {
			outlierIndices = append(outlierIndices, i)
		}
	}

	v.logger.Debug("Detected outliers",
		zap.Int("total_values", len(values)),
		zap.Int("outliers", len(outlierIndices)),
		zap.Float64("lower_bound", lowerBound),
		zap.Float64("upper_bound", upperBound),
	)

	return outlierIndices
}

// ValidateTimestamp checks if timestamp is within acceptable range
func (v *Validator) ValidateTimestamp(ctx context.Context, timestamp time.Time, maxAge time.Duration) *ValidationResult {
	result := &ValidationResult{
		IsValid:      true,
		FailedChecks: []string{},
		Metadata:     map[string]interface{}{"timestamp": timestamp},
	}

	now := time.Now()
	age := now.Sub(timestamp)

	// Check if timestamp is in the future
	if timestamp.After(now) {
		result.IsValid = false
		result.ErrorCode = "VALIDATION-004"
		result.ErrorMessage = "Timestamp is in the future"
		result.FailedChecks = append(result.FailedChecks, result.ErrorMessage)

		v.logger.Warn("Timestamp validation failed - future timestamp",
			zap.Time("timestamp", timestamp),
			zap.Duration("age", age),
		)
	}

	// Check if timestamp is too old
	if age > maxAge {
		result.IsValid = false
		result.ErrorCode = "VALIDATION-005"
		result.ErrorMessage = fmt.Sprintf("Timestamp age %v exceeds maximum allowed %v", age, maxAge)
		result.FailedChecks = append(result.FailedChecks, result.ErrorMessage)

		v.logger.Warn("Timestamp validation failed - too old",
			zap.Time("timestamp", timestamp),
			zap.Duration("age", age),
			zap.Duration("max_age", maxAge),
		)
	}

	return result
}

// ToServiceError converts validation result to service error if invalid
func (v *Validator) ToServiceError(result *ValidationResult) error {
	if result.IsValid {
		return nil
	}

	return &errors.ServiceError{
		Code:     result.ErrorCode,
		Message:  result.ErrorMessage,
		Severity: errors.SeverityMedium,
		Context: map[string]interface{}{
			"failed_checks": result.FailedChecks,
		},
	}
}

// Helper methods

func (v *Validator) validateType(value interface{}, expectedType string) bool {
	switch expectedType {
	case "string":
		_, ok := value.(string)
		return ok
	case "int":
		switch value.(type) {
		case int, int32, int64:
			return true
		}
		return false
	case "float":
		switch value.(type) {
		case float32, float64:
			return true
		}
		return false
	case "bool":
		_, ok := value.(bool)
		return ok
	case "timestamp":
		_, ok := value.(time.Time)
		if !ok {
			_, ok = value.(string) // Accept ISO8601 strings
		}
		return ok
	default:
		return true
	}
}

func (v *Validator) toFloat64(value interface{}) float64 {
	switch v := value.(type) {
	case float64:
		return v
	case float32:
		return float64(v)
	case int:
		return float64(v)
	case int32:
		return float64(v)
	case int64:
		return float64(v)
	default:
		return 0
	}
}

func (v *Validator) isInEnum(value interface{}, enumValues []string) bool {
	strValue := fmt.Sprintf("%v", value)
	for _, allowed := range enumValues {
		if strValue == allowed {
			return true
		}
	}
	return false
}
