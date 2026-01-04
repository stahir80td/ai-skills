package metrics

import (
	"testing"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/testutil"
)

func TestNewServiceMetrics(t *testing.T) {
	config := Config{
		ServiceName: "test-service",
		Namespace:   "test",
		Subsystem:   "subsystem",
	}

	metrics := NewServiceMetrics(config)

	if metrics == nil {
		t.Fatal("NewServiceMetrics returned nil")
	}
	if metrics.serviceName != "test-service" {
		t.Errorf("serviceName = %v, want %v", metrics.serviceName, "test-service")
	}
	if metrics.requestDuration == nil {
		t.Error("requestDuration should be initialized")
	}
	if metrics.requestTotal == nil {
		t.Error("requestTotal should be initialized")
	}
	if metrics.errorTotal == nil {
		t.Error("errorTotal should be initialized")
	}
	if metrics.resourceUtilization == nil {
		t.Error("resourceUtilization should be initialized")
	}
}

func TestNewServiceMetrics_DefaultNamespace(t *testing.T) {
	// Create with unregistered metrics by using a unique registry
	registry := prometheus.NewRegistry()

	config := Config{
		ServiceName: "test-service-2",
	}

	metrics := NewServiceMetrics(config)

	if metrics == nil {
		t.Fatal("NewServiceMetrics returned nil")
	}

	// Verify default namespace is used
	// We can't easily check the namespace directly, but we can verify metrics were created
	if metrics.requestDuration == nil {
		t.Error("requestDuration should use default namespace")
	}

	_ = registry // Suppress unused warning
}

func TestNewServiceMetrics_CustomBuckets(t *testing.T) {
	customBuckets := []float64{0.001, 0.01, 0.1, 1.0}
	config := Config{
		ServiceName:    "test-service-3",
		Namespace:      "test_buckets",
		Subsystem:      "custom",
		LatencyBuckets: customBuckets,
	}

	metrics := NewServiceMetrics(config)

	if metrics == nil {
		t.Fatal("NewServiceMetrics returned nil")
	}
	// Buckets are set during initialization, can't easily verify them
	// but we can verify the histogram was created
	if metrics.requestDuration == nil {
		t.Error("requestDuration should be initialized with custom buckets")
	}
}

func TestRecordRequest(t *testing.T) {
	config := Config{
		ServiceName: "test-record",
		Namespace:   "test_record",
	}

	metrics := NewServiceMetrics(config)

	// Record a request
	metrics.RecordRequest("GET", "/api/devices", "200", 150*time.Millisecond)

	// Verify counter increased
	counter := testutil.ToFloat64(metrics.requestTotal.With(prometheus.Labels{
		"service":  "test-record",
		"method":   "GET",
		"endpoint": "/api/devices",
		"status":   "200",
	}))

	if counter != 1 {
		t.Errorf("request counter = %v, want 1", counter)
	}

	// Record another request with same labels
	metrics.RecordRequest("GET", "/api/devices", "200", 200*time.Millisecond)

	counter = testutil.ToFloat64(metrics.requestTotal.With(prometheus.Labels{
		"service":  "test-record",
		"method":   "GET",
		"endpoint": "/api/devices",
		"status":   "200",
	}))

	if counter != 2 {
		t.Errorf("request counter = %v, want 2", counter)
	}
}

func TestRecordRequest_DifferentLabels(t *testing.T) {
	config := Config{
		ServiceName: "test-labels",
		Namespace:   "test_labels",
	}

	metrics := NewServiceMetrics(config)

	// Record requests with different labels
	metrics.RecordRequest("GET", "/api/devices", "200", 100*time.Millisecond)
	metrics.RecordRequest("POST", "/api/devices", "201", 200*time.Millisecond)
	metrics.RecordRequest("GET", "/api/users", "200", 150*time.Millisecond)

	// Each should have separate counters
	counter1 := testutil.ToFloat64(metrics.requestTotal.With(prometheus.Labels{
		"service":  "test-labels",
		"method":   "GET",
		"endpoint": "/api/devices",
		"status":   "200",
	}))

	counter2 := testutil.ToFloat64(metrics.requestTotal.With(prometheus.Labels{
		"service":  "test-labels",
		"method":   "POST",
		"endpoint": "/api/devices",
		"status":   "201",
	}))

	if counter1 != 1 {
		t.Errorf("GET /api/devices counter = %v, want 1", counter1)
	}
	if counter2 != 1 {
		t.Errorf("POST /api/devices counter = %v, want 1", counter2)
	}
}

func TestRecordError(t *testing.T) {
	config := Config{
		ServiceName: "test-error",
		Namespace:   "test_error",
	}

	metrics := NewServiceMetrics(config)

	// Record an error
	metrics.RecordError("INGEST-001", "CRITICAL", "DatabaseHandler")

	// Verify error counter increased
	counter := testutil.ToFloat64(metrics.errorTotal.With(prometheus.Labels{
		"service":    "test-error",
		"error_code": "INGEST-001",
		"severity":   "CRITICAL",
		"component":  "DatabaseHandler",
	}))

	if counter != 1 {
		t.Errorf("error counter = %v, want 1", counter)
	}

	// Record same error again
	metrics.RecordError("INGEST-001", "CRITICAL", "DatabaseHandler")

	counter = testutil.ToFloat64(metrics.errorTotal.With(prometheus.Labels{
		"service":    "test-error",
		"error_code": "INGEST-001",
		"severity":   "CRITICAL",
		"component":  "DatabaseHandler",
	}))

	if counter != 2 {
		t.Errorf("error counter = %v, want 2", counter)
	}
}

func TestUpdateResourceUtilization(t *testing.T) {
	config := Config{
		ServiceName: "test-resource",
		Namespace:   "test_resource",
	}

	metrics := NewServiceMetrics(config)

	// Update CPU utilization
	metrics.UpdateResourceUtilization("cpu", 75.5)

	gauge := testutil.ToFloat64(metrics.resourceUtilization.With(prometheus.Labels{
		"service":       "test-resource",
		"resource_type": "cpu",
	}))

	if gauge != 75.5 {
		t.Errorf("cpu utilization = %v, want 75.5", gauge)
	}

	// Update memory utilization
	metrics.UpdateResourceUtilization("memory", 60.2)

	gauge = testutil.ToFloat64(metrics.resourceUtilization.With(prometheus.Labels{
		"service":       "test-resource",
		"resource_type": "memory",
	}))

	if gauge != 60.2 {
		t.Errorf("memory utilization = %v, want 60.2", gauge)
	}
}

func TestActiveRequests(t *testing.T) {
	config := Config{
		ServiceName: "test-active",
		Namespace:   "test_active",
	}

	metrics := NewServiceMetrics(config)

	// Initial value should be 0
	active := testutil.ToFloat64(metrics.activeRequests)
	if active != 0 {
		t.Errorf("initial active requests = %v, want 0", active)
	}

	// Increment
	metrics.IncActiveRequests()
	active = testutil.ToFloat64(metrics.activeRequests)
	if active != 1 {
		t.Errorf("active requests after inc = %v, want 1", active)
	}

	// Increment again
	metrics.IncActiveRequests()
	active = testutil.ToFloat64(metrics.activeRequests)
	if active != 2 {
		t.Errorf("active requests after 2nd inc = %v, want 2", active)
	}

	// Decrement
	metrics.DecActiveRequests()
	active = testutil.ToFloat64(metrics.activeRequests)
	if active != 1 {
		t.Errorf("active requests after dec = %v, want 1", active)
	}
}

func TestSetQueueDepth(t *testing.T) {
	config := Config{
		ServiceName: "test-queue",
		Namespace:   "test_queue",
	}

	metrics := NewServiceMetrics(config)

	// Set queue depth
	metrics.SetQueueDepth(42)

	depth := testutil.ToFloat64(metrics.queueDepth)
	if depth != 42 {
		t.Errorf("queue depth = %v, want 42", depth)
	}

	// Update queue depth
	metrics.SetQueueDepth(15)

	depth = testutil.ToFloat64(metrics.queueDepth)
	if depth != 15 {
		t.Errorf("queue depth = %v, want 15", depth)
	}
}

func TestNewRequestTimer(t *testing.T) {
	config := Config{
		ServiceName: "test-timer",
		Namespace:   "test_timer",
	}

	metrics := NewServiceMetrics(config)

	// Create timer
	timer := metrics.NewRequestTimer("GET", "/api/test")

	if timer == nil {
		t.Fatal("NewRequestTimer returned nil")
	}
	if timer.method != "GET" {
		t.Errorf("timer method = %v, want GET", timer.method)
	}
	if timer.endpoint != "/api/test" {
		t.Errorf("timer endpoint = %v, want /api/test", timer.endpoint)
	}

	// Active requests should be incremented
	active := testutil.ToFloat64(metrics.activeRequests)
	if active != 1 {
		t.Errorf("active requests = %v, want 1", active)
	}
}

func TestRequestTimer_Done(t *testing.T) {
	config := Config{
		ServiceName: "test-timer-done",
		Namespace:   "test_timer_done",
	}

	metrics := NewServiceMetrics(config)

	// Create and complete timer
	timer := metrics.NewRequestTimer("POST", "/api/create")
	time.Sleep(50 * time.Millisecond)
	timer.Done("201")

	// Verify request was recorded
	counter := testutil.ToFloat64(metrics.requestTotal.With(prometheus.Labels{
		"service":  "test-timer-done",
		"method":   "POST",
		"endpoint": "/api/create",
		"status":   "201",
	}))

	if counter != 1 {
		t.Errorf("request counter = %v, want 1", counter)
	}

	// Verify active requests decremented
	active := testutil.ToFloat64(metrics.activeRequests)
	if active != 0 {
		t.Errorf("active requests = %v, want 0", active)
	}
}

func TestRequestTimer_DoneWithError(t *testing.T) {
	config := Config{
		ServiceName: "test-timer-error",
		Namespace:   "test_timer_error",
	}

	metrics := NewServiceMetrics(config)

	// Create timer and complete with error
	timer := metrics.NewRequestTimer("GET", "/api/fail")
	time.Sleep(30 * time.Millisecond)
	timer.DoneWithError("500", "APIGW-004", "HIGH", "HTTPHandler")

	// Verify request was recorded
	reqCounter := testutil.ToFloat64(metrics.requestTotal.With(prometheus.Labels{
		"service":  "test-timer-error",
		"method":   "GET",
		"endpoint": "/api/fail",
		"status":   "500",
	}))

	if reqCounter != 1 {
		t.Errorf("request counter = %v, want 1", reqCounter)
	}

	// Verify error was recorded
	errCounter := testutil.ToFloat64(metrics.errorTotal.With(prometheus.Labels{
		"service":    "test-timer-error",
		"error_code": "APIGW-004",
		"severity":   "HIGH",
		"component":  "HTTPHandler",
	}))

	if errCounter != 1 {
		t.Errorf("error counter = %v, want 1", errCounter)
	}

	// Verify active requests decremented
	active := testutil.ToFloat64(metrics.activeRequests)
	if active != 0 {
		t.Errorf("active requests = %v, want 0", active)
	}
}

func TestMultipleTimers(t *testing.T) {
	config := Config{
		ServiceName: "test-multi-timer",
		Namespace:   "test_multi",
	}

	metrics := NewServiceMetrics(config)

	// Start multiple timers
	timer1 := metrics.NewRequestTimer("GET", "/api/1")
	timer2 := metrics.NewRequestTimer("GET", "/api/2")
	timer3 := metrics.NewRequestTimer("POST", "/api/3")

	// Active requests should be 3
	active := testutil.ToFloat64(metrics.activeRequests)
	if active != 3 {
		t.Errorf("active requests = %v, want 3", active)
	}

	// Complete first timer
	timer1.Done("200")
	active = testutil.ToFloat64(metrics.activeRequests)
	if active != 2 {
		t.Errorf("active requests after 1st done = %v, want 2", active)
	}

	// Complete remaining timers
	timer2.Done("200")
	timer3.Done("201")

	active = testutil.ToFloat64(metrics.activeRequests)
	if active != 0 {
		t.Errorf("active requests after all done = %v, want 0", active)
	}
}
