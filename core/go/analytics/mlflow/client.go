package mlflow

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/your-github-org/ai-scaffolder/core/go/errors"
	"github.com/your-github-org/ai-scaffolder/core/go/logger"
	"github.com/your-github-org/ai-scaffolder/core/go/reliability"
	"go.uber.org/zap"
)

// Client provides MLflow model registry and tracking capabilities
type Client struct {
	baseURL        string
	httpClient     *http.Client
	circuitBreaker *reliability.CircuitBreaker
	logger         *logger.Logger
}

// Config holds MLflow client configuration
type Config struct {
	BaseURL string
	Timeout time.Duration
	Logger  *logger.Logger
}

// Model represents an MLflow model
type Model struct {
	Name            string            `json:"name"`
	Version         string            `json:"version"`
	Description     string            `json:"description,omitempty"`
	Tags            map[string]string `json:"tags,omitempty"`
	Status          string            `json:"status,omitempty"`
	CreationTime    int64             `json:"creation_timestamp,omitempty"`
	LastUpdatedTime int64             `json:"last_updated_timestamp,omitempty"`
	RunID           string            `json:"run_id,omitempty"`
	Source          string            `json:"source,omitempty"`
}

// Prediction represents a model prediction result
type Prediction struct {
	ModelName    string                 `json:"model_name"`
	ModelVersion string                 `json:"model_version"`
	Input        map[string]interface{} `json:"input"`
	Output       interface{}            `json:"output"`
	Timestamp    time.Time              `json:"timestamp"`
	LatencyMs    float64                `json:"latency_ms"`
}

// Experiment represents an MLflow experiment
type Experiment struct {
	ID             string `json:"experiment_id"`
	Name           string `json:"name"`
	ArtifactPath   string `json:"artifact_location"`
	LifecycleStage string `json:"lifecycle_stage"`
}

// Run represents an MLflow run
type Run struct {
	ID           string             `json:"run_id"`
	ExperimentID string             `json:"experiment_id"`
	Status       string             `json:"status"`
	StartTime    int64              `json:"start_time"`
	EndTime      int64              `json:"end_time"`
	Params       map[string]string  `json:"params,omitempty"`
	Metrics      map[string]float64 `json:"metrics,omitempty"`
	Tags         map[string]string  `json:"tags,omitempty"`
}

// NewClient creates a new MLflow client with circuit breaker
func NewClient(cfg Config) *Client {
	if cfg.Timeout == 0 {
		cfg.Timeout = 30 * time.Second
	}

	return &Client{
		baseURL: cfg.BaseURL,
		httpClient: &http.Client{
			Timeout: cfg.Timeout,
		},
		circuitBreaker: reliability.NewCircuitBreaker("mlflow", 5, 60*time.Second),
		logger:         cfg.Logger,
	}
}

// GetModel retrieves a model by name and version
func (c *Client) GetModel(ctx context.Context, name, version string) (*Model, error) {
	var model *Model

	err := c.circuitBreaker.Execute(func() error {
		url := fmt.Sprintf("%s/api/2.0/mlflow/model-versions/get-by-name?name=%s&version=%s",
			c.baseURL, name, version)

		req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
		if err != nil {
			return err
		}

		resp, err := c.httpClient.Do(req)
		if err != nil {
			c.logger.Error("Failed to get model from MLflow",
				zap.String("model_name", name),
				zap.String("version", version),
				zap.Error(err),
			)
			return err
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			body, _ := io.ReadAll(resp.Body)
			c.logger.Error("MLflow API returned error",
				zap.Int("status_code", resp.StatusCode),
				zap.String("response", string(body)),
			)
			return fmt.Errorf("mlflow api error: status %d", resp.StatusCode)
		}

		var response struct {
			ModelVersion *Model `json:"model_version"`
		}

		if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
			return err
		}

		model = response.ModelVersion

		c.logger.Info("Retrieved model from MLflow",
			zap.String("model_name", name),
			zap.String("version", version),
			zap.String("status", model.Status),
		)

		return nil
	})

	if err != nil {
		return nil, &errors.ServiceError{
			Code:       "MLFLOW-001",
			Message:    "Failed to retrieve model from MLflow",
			Severity:   errors.SeverityHigh,
			Underlying: err,
		}
	}

	return model, nil
}

// RegisterModel registers a new model in MLflow
func (c *Client) RegisterModel(ctx context.Context, model *Model) error {
	return c.circuitBreaker.Execute(func() error {
		url := fmt.Sprintf("%s/api/2.0/mlflow/registered-models/create", c.baseURL)

		payload := map[string]interface{}{
			"name":        model.Name,
			"description": model.Description,
			"tags":        model.Tags,
		}

		body, err := json.Marshal(payload)
		if err != nil {
			return err
		}

		req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewReader(body))
		if err != nil {
			return err
		}
		req.Header.Set("Content-Type", "application/json")

		resp, err := c.httpClient.Do(req)
		if err != nil {
			c.logger.Error("Failed to register model in MLflow",
				zap.String("model_name", model.Name),
				zap.Error(err),
			)
			return err
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
			bodyBytes, _ := io.ReadAll(resp.Body)
			c.logger.Error("MLflow API returned error on register",
				zap.Int("status_code", resp.StatusCode),
				zap.String("response", string(bodyBytes)),
			)
			return fmt.Errorf("mlflow api error: status %d", resp.StatusCode)
		}

		c.logger.Info("Registered model in MLflow",
			zap.String("model_name", model.Name),
		)

		return nil
	})
}

// CreateModelVersion creates a new version of a model
func (c *Client) CreateModelVersion(ctx context.Context, name, runID, source string) (*Model, error) {
	var modelVersion *Model

	err := c.circuitBreaker.Execute(func() error {
		url := fmt.Sprintf("%s/api/2.0/mlflow/model-versions/create", c.baseURL)

		payload := map[string]interface{}{
			"name":   name,
			"source": source,
			"run_id": runID,
		}

		body, err := json.Marshal(payload)
		if err != nil {
			return err
		}

		req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewReader(body))
		if err != nil {
			return err
		}
		req.Header.Set("Content-Type", "application/json")

		resp, err := c.httpClient.Do(req)
		if err != nil {
			c.logger.Error("Failed to create model version in MLflow",
				zap.String("model_name", name),
				zap.String("run_id", runID),
				zap.Error(err),
			)
			return err
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
			bodyBytes, _ := io.ReadAll(resp.Body)
			return fmt.Errorf("mlflow api error: status %d, body: %s", resp.StatusCode, string(bodyBytes))
		}

		var response struct {
			ModelVersion *Model `json:"model_version"`
		}

		if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
			return err
		}

		modelVersion = response.ModelVersion

		c.logger.Info("Created model version in MLflow",
			zap.String("model_name", name),
			zap.String("version", modelVersion.Version),
			zap.String("run_id", runID),
		)

		return nil
	})

	if err != nil {
		return nil, &errors.ServiceError{
			Code:       "MLFLOW-002",
			Message:    "Failed to create model version in MLflow",
			Severity:   errors.SeverityHigh,
			Underlying: err,
		}
	}

	return modelVersion, nil
}

// LogMetric logs a metric to MLflow
func (c *Client) LogMetric(ctx context.Context, runID, key string, value float64, timestamp int64) error {
	return c.circuitBreaker.Execute(func() error {
		url := fmt.Sprintf("%s/api/2.0/mlflow/runs/log-metric", c.baseURL)

		payload := map[string]interface{}{
			"run_id":    runID,
			"key":       key,
			"value":     value,
			"timestamp": timestamp,
		}

		body, err := json.Marshal(payload)
		if err != nil {
			return err
		}

		req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewReader(body))
		if err != nil {
			return err
		}
		req.Header.Set("Content-Type", "application/json")

		resp, err := c.httpClient.Do(req)
		if err != nil {
			c.logger.Error("Failed to log metric to MLflow",
				zap.String("run_id", runID),
				zap.String("key", key),
				zap.Error(err),
			)
			return err
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			bodyBytes, _ := io.ReadAll(resp.Body)
			return fmt.Errorf("mlflow api error: status %d, body: %s", resp.StatusCode, string(bodyBytes))
		}

		c.logger.Debug("Logged metric to MLflow",
			zap.String("run_id", runID),
			zap.String("key", key),
			zap.Float64("value", value),
		)

		return nil
	})
}

// GetRun retrieves a run by ID
func (c *Client) GetRun(ctx context.Context, runID string) (*Run, error) {
	var run *Run

	err := c.circuitBreaker.Execute(func() error {
		url := fmt.Sprintf("%s/api/2.0/mlflow/runs/get?run_id=%s", c.baseURL, runID)

		req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
		if err != nil {
			return err
		}

		resp, err := c.httpClient.Do(req)
		if err != nil {
			c.logger.Error("Failed to get run from MLflow",
				zap.String("run_id", runID),
				zap.Error(err),
			)
			return err
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			bodyBytes, _ := io.ReadAll(resp.Body)
			return fmt.Errorf("mlflow api error: status %d, body: %s", resp.StatusCode, string(bodyBytes))
		}

		var response struct {
			Run *Run `json:"run"`
		}

		if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
			return err
		}

		run = response.Run

		c.logger.Debug("Retrieved run from MLflow",
			zap.String("run_id", runID),
			zap.String("status", run.Status),
		)

		return nil
	})

	if err != nil {
		return nil, &errors.ServiceError{
			Code:       "MLFLOW-003",
			Message:    "Failed to retrieve run from MLflow",
			Severity:   errors.SeverityMedium,
			Underlying: err,
		}
	}

	return run, nil
}

// HealthCheck checks if MLflow service is available
func (c *Client) HealthCheck(ctx context.Context) error {
	return c.circuitBreaker.Execute(func() error {
		url := fmt.Sprintf("%s/health", c.baseURL)

		req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
		if err != nil {
			return err
		}

		resp, err := c.httpClient.Do(req)
		if err != nil {
			c.logger.Warn("MLflow health check failed", zap.Error(err))
			return err
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			return fmt.Errorf("mlflow health check failed: status %d", resp.StatusCode)
		}

		return nil
	})
}
