package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

// Analytics-specific Prometheus metrics following core package patterns

var (
	// Model prediction metrics
	ModelPredictionDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "analytics_model_prediction_duration_seconds",
			Help:    "Duration of model predictions in seconds",
			Buckets: []float64{0.001, 0.005, 0.01, 0.05, 0.1, 0.5, 1.0, 2.0, 5.0},
		},
		[]string{"model_name", "model_version"},
	)

	ModelPredictionTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "analytics_model_prediction_total",
			Help: "Total number of model predictions",
		},
		[]string{"model_name", "model_version", "status"},
	)

	ModelPredictionErrors = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "analytics_model_prediction_errors_total",
			Help: "Total number of model prediction errors",
		},
		[]string{"model_name", "model_version", "error_type"},
	)

	// Model drift metrics
	ModelDriftScore = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "analytics_model_drift_score",
			Help: "Model drift score (0-1, higher means more drift)",
		},
		[]string{"model_name", "model_version", "drift_type"},
	)

	ModelAccuracy = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "analytics_model_accuracy",
			Help: "Current model accuracy",
		},
		[]string{"model_name", "model_version"},
	)

	// Data quality metrics
	DataQualityScore = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "analytics_data_quality_score",
			Help: "Data quality score (0-1, 1 is perfect)",
		},
		[]string{"dataset", "dimension"},
	)

	DataValidationErrors = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "analytics_data_validation_errors_total",
			Help: "Total number of data validation errors",
		},
		[]string{"dataset", "error_type"},
	)

	DataNullPercentage = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "analytics_data_null_percentage",
			Help: "Percentage of null values in dataset",
		},
		[]string{"dataset", "field"},
	)

	DataOutlierCount = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "analytics_data_outlier_count",
			Help: "Number of outliers detected in dataset",
		},
		[]string{"dataset", "field"},
	)

	// Feature store metrics
	FeatureComputeDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "analytics_feature_compute_duration_seconds",
			Help:    "Duration of feature computation in seconds",
			Buckets: []float64{0.001, 0.005, 0.01, 0.05, 0.1, 0.5, 1.0},
		},
		[]string{"feature_name"},
	)

	FeatureFreshnessAge = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "analytics_feature_freshness_age_seconds",
			Help: "Age of feature data in seconds",
		},
		[]string{"feature_name"},
	)

	FeatureRetrievalTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "analytics_feature_retrieval_total",
			Help: "Total number of feature retrievals",
		},
		[]string{"feature_name", "status"},
	)

	FeatureCacheHits = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "analytics_feature_cache_hits_total",
			Help: "Total number of feature cache hits",
		},
		[]string{"feature_name"},
	)

	FeatureCacheMisses = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "analytics_feature_cache_misses_total",
			Help: "Total number of feature cache misses",
		},
		[]string{"feature_name"},
	)

	// Aggregation metrics
	AggregationDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "analytics_aggregation_duration_seconds",
			Help:    "Duration of data aggregation in seconds",
			Buckets: []float64{0.01, 0.05, 0.1, 0.5, 1.0, 5.0, 10.0, 30.0},
		},
		[]string{"aggregation_type", "granularity"},
	)

	AggregationRecordsProcessed = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "analytics_aggregation_records_processed_total",
			Help: "Total number of records processed in aggregation",
		},
		[]string{"aggregation_type", "granularity"},
	)

	AggregationErrors = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "analytics_aggregation_errors_total",
			Help: "Total number of aggregation errors",
		},
		[]string{"aggregation_type", "error_type"},
	)

	// MLflow integration metrics
	MLflowRequestDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "analytics_mlflow_request_duration_seconds",
			Help:    "Duration of MLflow API requests in seconds",
			Buckets: []float64{0.01, 0.05, 0.1, 0.5, 1.0, 2.0, 5.0},
		},
		[]string{"operation"},
	)

	MLflowRequestTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "analytics_mlflow_request_total",
			Help: "Total number of MLflow API requests",
		},
		[]string{"operation", "status"},
	)

	MLflowCircuitBreakerState = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "analytics_mlflow_circuit_breaker_state",
			Help: "MLflow circuit breaker state (0=closed, 1=half_open, 2=open)",
		},
		[]string{},
	)

	// Time-series processing metrics
	TimeSeriesWindowCount = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "analytics_timeseries_window_count",
			Help: "Number of time-series windows created",
		},
		[]string{"window_type"},
	)

	TimeSeriesInterpolatedPoints = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "analytics_timeseries_interpolated_points_total",
			Help: "Total number of interpolated data points",
		},
		[]string{"interpolation_method"},
	)

	TimeSeriesProcessingDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "analytics_timeseries_processing_duration_seconds",
			Help:    "Duration of time-series processing in seconds",
			Buckets: []float64{0.001, 0.01, 0.1, 0.5, 1.0, 5.0},
		},
		[]string{"operation"},
	)

	// Anomaly detection metrics
	AnomaliesDetected = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "analytics_anomalies_detected_total",
			Help: "Total number of anomalies detected",
		},
		[]string{"device_id", "anomaly_type", "severity"},
	)

	AnomalyScore = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "analytics_anomaly_score",
			Help: "Anomaly score for device (higher means more anomalous)",
		},
		[]string{"device_id", "model_version"},
	)

	AnomalyDetectionDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "analytics_anomaly_detection_duration_seconds",
			Help:    "Duration of anomaly detection in seconds",
			Buckets: []float64{0.001, 0.01, 0.1, 0.5, 1.0, 2.0},
		},
		[]string{"model_version"},
	)

	// Query performance metrics
	QueryDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "analytics_query_duration_seconds",
			Help:    "Duration of analytics queries in seconds",
			Buckets: []float64{0.01, 0.05, 0.1, 0.5, 1.0, 5.0, 10.0, 30.0},
		},
		[]string{"query_type", "data_source"},
	)

	QueryResultSize = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "analytics_query_result_size_bytes",
			Help:    "Size of analytics query results in bytes",
			Buckets: []float64{1024, 10240, 102400, 1048576, 10485760, 104857600},
		},
		[]string{"query_type"},
	)

	QueryCacheHitRatio = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "analytics_query_cache_hit_ratio",
			Help: "Query cache hit ratio (0-1)",
		},
		[]string{"query_type"},
	)

	// Operational intelligence metrics
	SystemHealthScore = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "analytics_system_health_score",
			Help: "Overall system health score (0-1, 1 is healthy)",
		},
		[]string{"system_component"},
	)

	DeviceActivityScore = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "analytics_device_activity_score",
			Help: "Device activity score (0-1, 1 is highly active)",
		},
		[]string{"device_id"},
	)

	PredictiveMaintenanceScore = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "analytics_predictive_maintenance_score",
			Help: "Predictive maintenance score (0-1, higher means maintenance needed soon)",
		},
		[]string{"device_id", "component"},
	)
)

// Helper functions for recording metrics

// RecordModelPrediction records a model prediction with timing
func RecordModelPrediction(modelName, modelVersion, status string, durationSeconds float64) {
	ModelPredictionDuration.WithLabelValues(modelName, modelVersion).Observe(durationSeconds)
	ModelPredictionTotal.WithLabelValues(modelName, modelVersion, status).Inc()
}

// RecordDataQuality records data quality metrics
func RecordDataQuality(dataset, dimension string, score float64) {
	DataQualityScore.WithLabelValues(dataset, dimension).Set(score)
}

// RecordFeatureCompute records feature computation metrics
func RecordFeatureCompute(featureName string, durationSeconds float64) {
	FeatureComputeDuration.WithLabelValues(featureName).Observe(durationSeconds)
}

// RecordAggregation records aggregation metrics
func RecordAggregation(aggregationType, granularity string, recordCount int64, durationSeconds float64) {
	AggregationDuration.WithLabelValues(aggregationType, granularity).Observe(durationSeconds)
	AggregationRecordsProcessed.WithLabelValues(aggregationType, granularity).Add(float64(recordCount))
}

// RecordAnomalyDetection records anomaly detection results
func RecordAnomalyDetection(deviceID, anomalyType, severity string, score float64, durationSeconds float64, modelVersion string) {
	AnomaliesDetected.WithLabelValues(deviceID, anomalyType, severity).Inc()
	AnomalyScore.WithLabelValues(deviceID, modelVersion).Set(score)
	AnomalyDetectionDuration.WithLabelValues(modelVersion).Observe(durationSeconds)
}

// RecordQuery records query performance metrics
func RecordQuery(queryType, dataSource string, durationSeconds float64, resultSizeBytes int64) {
	QueryDuration.WithLabelValues(queryType, dataSource).Observe(durationSeconds)
	QueryResultSize.WithLabelValues(queryType).Observe(float64(resultSizeBytes))
}
