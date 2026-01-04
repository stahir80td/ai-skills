package health

import (
	"context"
	"testing"
	"time"

	"github.com/your-github-org/ai-scaffolder/core/go/logger"
)

func TestChecker_Register(t *testing.T) {
	appLogger, _ := logger.NewProduction("health-test", "1.0.0")
	defer appLogger.Sync()

	checker := NewChecker(appLogger, 5*time.Second)

	checker.Register("test_service", func(ctx context.Context) error {
		return nil
	})

	results := checker.Check(context.Background())
	if len(results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(results))
	}

	if results["test_service"].Status != StatusHealthy {
		t.Fatalf("expected healthy status")
	}
}

func TestChecker_Timeout(t *testing.T) {
	appLogger, _ := logger.NewProduction("health-test", "1.0.0")
	defer appLogger.Sync()

	checker := NewChecker(appLogger, 5*time.Second)
	checker.SetTimeout(50 * time.Millisecond)

	checker.Register("slow_check", func(ctx context.Context) error {
		<-ctx.Done()
		return ctx.Err()
	})

	results := checker.Check(context.Background())
	if results["slow_check"].Status != StatusUnhealthy {
		t.Fatalf("expected unhealthy status due to timeout, got %s: %s",
			results["slow_check"].Status, results["slow_check"].Message)
	}
}

func TestChecker_IsHealthy(t *testing.T) {
	appLogger, _ := logger.NewProduction("health-test", "1.0.0")
	defer appLogger.Sync()

	checker := NewChecker(appLogger, 5*time.Second)

	checker.Register("service1", func(ctx context.Context) error {
		return nil
	})

	if !checker.IsHealthy(context.Background()) {
		t.Fatal("expected healthy")
	}
}
