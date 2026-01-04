package health

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/your-github-org/ai-scaffolder/core/go/logger"
	"go.uber.org/zap"
)

// Status represents health check status
type Status string

const (
	StatusHealthy   Status = "healthy"
	StatusUnhealthy Status = "unhealthy"
)

// CheckFunc is a health check function
type CheckFunc func(ctx context.Context) error

// CheckResult represents the result of a health check
type CheckResult struct {
	Name     string        `json:"name"`
	Status   Status        `json:"status"`
	Message  string        `json:"message"`
	Duration time.Duration `json:"duration_ms"`
	Error    string        `json:"error,omitempty"`
}

// Checker manages health checks
type Checker struct {
	logger  *logger.Logger
	checks  map[string]CheckFunc
	mu      sync.RWMutex
	timeout time.Duration
}

// NewChecker creates a new health checker
func NewChecker(log *logger.Logger, timeout time.Duration) *Checker {
	if log == nil {
		log, _ = logger.NewProduction("health-checker", "1.0")
	}

	return &Checker{
		logger:  log,
		checks:  make(map[string]CheckFunc),
		timeout: timeout, // NO HARDCODING - from config (MINIMUM 60s)
	}
}

// Register registers a health check
func (c *Checker) Register(name string, checkFn CheckFunc) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.logger.WithComponent("HealthChecker").Debug("Registering health check",
		zap.String("check_name", name),
		zap.Int("total_checks", len(c.checks)+1))

	c.checks[name] = checkFn
}

// Check runs all health checks
func (c *Checker) Check(ctx context.Context) map[string]*CheckResult {
	c.mu.RLock()
	checksCount := len(c.checks)
	c.mu.RUnlock()

	c.logger.WithComponent("HealthChecker").Debug("Starting health checks",
		zap.Int("total_checks", checksCount),
		zap.Duration("timeout", c.timeout))

	results := make(map[string]*CheckResult)
	var wg sync.WaitGroup
	var mu sync.Mutex

	c.mu.RLock()
	for name, checkFn := range c.checks {
		wg.Add(1)
		go func(checkName string, fn CheckFunc) {
			defer wg.Done()

			result := c.executeCheck(ctx, checkName, fn)
			mu.Lock()
			results[checkName] = result
			mu.Unlock()
		}(name, checkFn)
	}
	c.mu.RUnlock()

	wg.Wait()

	healthy := true
	failedChecks := 0
	for _, result := range results {
		if result.Status != StatusHealthy {
			healthy = false
			failedChecks++
		}
	}

	if healthy {
		c.logger.WithComponent("HealthChecker").Debug("All health checks passed",
			zap.Int("check_count", len(results)),
			zap.String("overall_status", "healthy"))
	} else {
		c.logger.WithComponent("HealthChecker").Warn("Some health checks failed",
			zap.Int("total_checks", len(results)),
			zap.Int("failed_checks", failedChecks),
			zap.String("overall_status", "unhealthy"))
	}

	return results
}

// executeCheck runs a single health check
func (c *Checker) executeCheck(ctx context.Context, name string, checkFn CheckFunc) *CheckResult {
	result := &CheckResult{
		Name: name,
	}

	checkCtx, cancel := context.WithTimeout(ctx, c.timeout)
	defer cancel()

	start := time.Now()
	err := checkFn(checkCtx)
	result.Duration = time.Since(start)

	if err != nil {
		result.Status = StatusUnhealthy
		result.Message = fmt.Sprintf("Check failed: %v", err)
		result.Error = err.Error()

		c.logger.WithComponent("HealthChecker").Warn("Health check failed",
			zap.String("check_name", name),
			zap.String("error_code", "INFRA-HEALTH-CHECK-FAILED"),
			zap.Error(err),
			zap.Duration("duration_ms", result.Duration),
			zap.String("status", string(result.Status)))
	} else {
		result.Status = StatusHealthy
		result.Message = "Check passed"

		c.logger.WithComponent("HealthChecker").Debug("Health check passed",
			zap.String("check_name", name),
			zap.Duration("duration_ms", result.Duration),
			zap.String("status", string(result.Status)))
	}

	return result
}

// IsHealthy returns true if all checks are healthy
func (c *Checker) IsHealthy(ctx context.Context) bool {
	results := c.Check(ctx)
	for _, result := range results {
		if result.Status != StatusHealthy {
			return false
		}
	}
	return true
}

// SetTimeout sets the timeout for individual checks
func (c *Checker) SetTimeout(timeout time.Duration) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.timeout = timeout
}
