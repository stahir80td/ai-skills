package errors

import (
	"errors"
	"testing"
)

func TestNew(t *testing.T) {
	err := New("TEST-001", SeverityCritical, "test error message")

	if err.Code != "TEST-001" {
		t.Errorf("Code = %v, want %v", err.Code, "TEST-001")
	}
	if err.Severity != SeverityCritical {
		t.Errorf("Severity = %v, want %v", err.Severity, SeverityCritical)
	}
	if err.Message != "test error message" {
		t.Errorf("Message = %v, want %v", err.Message, "test error message")
	}
	if err.Underlying != nil {
		t.Error("Underlying should be nil")
	}
	if err.Context == nil {
		t.Error("Context should be initialized")
	}
}

func TestWrap(t *testing.T) {
	originalErr := errors.New("original error")
	err := Wrap(originalErr, "TEST-002", SeverityHigh, "wrapped error message")

	if err.Code != "TEST-002" {
		t.Errorf("Code = %v, want %v", err.Code, "TEST-002")
	}
	if err.Severity != SeverityHigh {
		t.Errorf("Severity = %v, want %v", err.Severity, SeverityHigh)
	}
	if err.Message != "wrapped error message" {
		t.Errorf("Message = %v, want %v", err.Message, "wrapped error message")
	}
	if err.Underlying != originalErr {
		t.Error("Underlying error not set correctly")
	}
}

func TestServiceError_Error(t *testing.T) {
	tests := []struct {
		name       string
		err        *ServiceError
		wantString string
	}{
		{
			name:       "without underlying error",
			err:        New("TEST-003", SeverityMedium, "test message"),
			wantString: "[TEST-003] MEDIUM: test message",
		},
		{
			name:       "with underlying error",
			err:        Wrap(errors.New("underlying"), "TEST-004", SeverityLow, "wrapped message"),
			wantString: "[TEST-004] LOW: wrapped message (caused by: underlying)",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.err.Error(); got != tt.wantString {
				t.Errorf("Error() = %v, want %v", got, tt.wantString)
			}
		})
	}
}

func TestServiceError_Unwrap(t *testing.T) {
	originalErr := errors.New("original")
	err := Wrap(originalErr, "TEST-005", SeverityCritical, "wrapped")

	unwrapped := err.Unwrap()
	if unwrapped != originalErr {
		t.Errorf("Unwrap() = %v, want %v", unwrapped, originalErr)
	}

	// Test with no underlying error
	err2 := New("TEST-006", SeverityHigh, "no underlying")
	if err2.Unwrap() != nil {
		t.Error("Unwrap() should return nil when no underlying error")
	}
}

func TestServiceError_WithContext(t *testing.T) {
	err := New("TEST-007", SeverityMedium, "test message")

	err.WithContext("user_id", "123")
	err.WithContext("request_id", "abc-def")
	err.WithContext("count", 42)

	if val, ok := err.GetContext("user_id"); !ok || val != "123" {
		t.Errorf("user_id context = %v, want %v", val, "123")
	}
	if val, ok := err.GetContext("request_id"); !ok || val != "abc-def" {
		t.Errorf("request_id context = %v, want %v", val, "abc-def")
	}
	if val, ok := err.GetContext("count"); !ok || val != 42 {
		t.Errorf("count context = %v, want %v", val, 42)
	}
}

func TestServiceError_GetContext(t *testing.T) {
	err := New("TEST-008", SeverityLow, "test message")
	err.WithContext("key1", "value1")

	tests := []struct {
		name      string
		key       string
		wantValue interface{}
		wantOk    bool
	}{
		{
			name:      "existing key",
			key:       "key1",
			wantValue: "value1",
			wantOk:    true,
		},
		{
			name:      "non-existing key",
			key:       "key2",
			wantValue: nil,
			wantOk:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotValue, gotOk := err.GetContext(tt.key)
			if gotOk != tt.wantOk {
				t.Errorf("GetContext() ok = %v, want %v", gotOk, tt.wantOk)
			}
			if gotValue != tt.wantValue {
				t.Errorf("GetContext() value = %v, want %v", gotValue, tt.wantValue)
			}
		})
	}
}

func TestNewErrorRegistry(t *testing.T) {
	registry := NewErrorRegistry()
	if registry == nil {
		t.Fatal("NewErrorRegistry returned nil")
	}
	if registry.definitions == nil {
		t.Error("definitions map should be initialized")
	}
}

func TestErrorRegistry_Register(t *testing.T) {
	registry := NewErrorRegistry()
	def := &ErrorDefinition{
		Code:        "INGEST-001",
		Severity:    SeverityCritical,
		Description: "Database connection failed",
		SODScore:    800,
		Severity_S:  10,
		Occurrence:  8,
		Detect_D:    10,
		Mitigation:  "Check database connectivity",
		Example:     "Database server is unreachable",
	}

	registry.Register(def)

	retrieved, ok := registry.Get("INGEST-001")
	if !ok {
		t.Fatal("error definition not found after registration")
	}
	if retrieved.Code != def.Code {
		t.Errorf("Code = %v, want %v", retrieved.Code, def.Code)
	}
	if retrieved.SODScore != def.SODScore {
		t.Errorf("SODScore = %v, want %v", retrieved.SODScore, def.SODScore)
	}
}

func TestErrorRegistry_Get(t *testing.T) {
	registry := NewErrorRegistry()
	def := &ErrorDefinition{
		Code:     "TEST-009",
		Severity: SeverityHigh,
	}
	registry.Register(def)

	tests := []struct {
		name   string
		code   string
		wantOk bool
	}{
		{
			name:   "existing code",
			code:   "TEST-009",
			wantOk: true,
		},
		{
			name:   "non-existing code",
			code:   "TEST-999",
			wantOk: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, gotOk := registry.Get(tt.code)
			if gotOk != tt.wantOk {
				t.Errorf("Get() ok = %v, want %v", gotOk, tt.wantOk)
			}
		})
	}
}

func TestErrorRegistry_GetAll(t *testing.T) {
	registry := NewErrorRegistry()
	def1 := &ErrorDefinition{Code: "TEST-010", Severity: SeverityCritical}
	def2 := &ErrorDefinition{Code: "TEST-011", Severity: SeverityHigh}

	registry.Register(def1)
	registry.Register(def2)

	all := registry.GetAll()
	if len(all) != 2 {
		t.Errorf("GetAll() returned %d definitions, want 2", len(all))
	}

	if _, ok := all["TEST-010"]; !ok {
		t.Error("TEST-010 not found in GetAll()")
	}
	if _, ok := all["TEST-011"]; !ok {
		t.Error("TEST-011 not found in GetAll()")
	}
}

func TestErrorRegistry_CreateError(t *testing.T) {
	registry := NewErrorRegistry()
	def := &ErrorDefinition{
		Code:        "INGEST-002",
		Severity:    SeverityMedium,
		Description: "Invalid payload: %s",
	}
	registry.Register(def)

	t.Run("from registered definition", func(t *testing.T) {
		err := registry.CreateError("INGEST-002", "missing field 'device_id'")

		if err.Code != "INGEST-002" {
			t.Errorf("Code = %v, want %v", err.Code, "INGEST-002")
		}
		if err.Severity != SeverityMedium {
			t.Errorf("Severity = %v, want %v", err.Severity, SeverityMedium)
		}
		if err.Message != "Invalid payload: missing field 'device_id'" {
			t.Errorf("Message = %v, want formatted message", err.Message)
		}
	})

	t.Run("from unknown code", func(t *testing.T) {
		err := registry.CreateError("UNKNOWN-999")

		if err.Code != "UNKNOWN-999" {
			t.Errorf("Code = %v, want %v", err.Code, "UNKNOWN-999")
		}
		if err.Severity != SeverityMedium {
			t.Error("Unknown error should default to MEDIUM severity")
		}
	})
}

func TestErrorRegistry_WrapError(t *testing.T) {
	registry := NewErrorRegistry()
	def := &ErrorDefinition{
		Code:        "INGEST-003",
		Severity:    SeverityHigh,
		Description: "Failed to process message: %s",
	}
	registry.Register(def)

	originalErr := errors.New("JSON parsing failed")

	t.Run("wrap with registered definition", func(t *testing.T) {
		err := registry.WrapError(originalErr, "INGEST-003", "malformed JSON")

		if err.Code != "INGEST-003" {
			t.Errorf("Code = %v, want %v", err.Code, "INGEST-003")
		}
		if err.Severity != SeverityHigh {
			t.Errorf("Severity = %v, want %v", err.Severity, SeverityHigh)
		}
		if err.Underlying != originalErr {
			t.Error("Underlying error not set")
		}
		if err.Message != "Failed to process message: malformed JSON" {
			t.Errorf("Message = %v, want formatted message", err.Message)
		}
	})

	t.Run("wrap with unknown code", func(t *testing.T) {
		err := registry.WrapError(originalErr, "UNKNOWN-888")

		if err.Code != "UNKNOWN-888" {
			t.Errorf("Code = %v, want %v", err.Code, "UNKNOWN-888")
		}
		if err.Underlying != originalErr {
			t.Error("Underlying error not set")
		}
	})
}

func TestCalculateSOD(t *testing.T) {
	tests := []struct {
		name          string
		severity      int
		occurrence    int
		detectability int
		wantScore     int
	}{
		{
			name:          "critical error",
			severity:      10,
			occurrence:    8,
			detectability: 10,
			wantScore:     800,
		},
		{
			name:          "low impact error",
			severity:      2,
			occurrence:    3,
			detectability: 5,
			wantScore:     30,
		},
		{
			name:          "maximum score",
			severity:      10,
			occurrence:    10,
			detectability: 10,
			wantScore:     1000,
		},
		{
			name:          "minimum score",
			severity:      1,
			occurrence:    1,
			detectability: 1,
			wantScore:     1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			score := CalculateSOD(tt.severity, tt.occurrence, tt.detectability)
			if score != tt.wantScore {
				t.Errorf("CalculateSOD() = %v, want %v", score, tt.wantScore)
			}
		})
	}
}
