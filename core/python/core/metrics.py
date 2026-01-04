"""
Service Metrics - Python equivalent of Go core/metrics

Provides:
- Prometheus metrics collection
- Request/error tracking
- Latency histograms
- Incident metrics
"""

from typing import Optional
from prometheus_client import Counter, Histogram, Gauge, CollectorRegistry
import time


class ServiceMetrics:
    """
    Service metrics for Prometheus
    Matches Go metrics.ServiceMetrics pattern
    """

    def __init__(
        self,
        namespace: str,
        subsystem: str,
        registry: Optional[CollectorRegistry] = None,
    ):
        self.namespace = namespace
        self.subsystem = subsystem
        self.registry = registry

        # Four Golden Signals
        self.requests_total = Counter(
            "requests_total",
            "Total number of requests",
            ["operation", "service", "method"],
            namespace=namespace,
            subsystem=subsystem,
            registry=registry,
        )

        self.request_duration_seconds = Histogram(
            "request_duration_seconds",
            "Request duration in seconds",
            ["operation", "service", "method"],
            buckets=(0.001, 0.005, 0.01, 0.025, 0.05, 0.1, 0.25, 0.5, 1, 2.5, 5, 10),
            namespace=namespace,
            subsystem=subsystem,
            registry=registry,
        )

        self.errors_total = Counter(
            "errors_total",
            "Total number of errors",
            ["error_type", "service", "component"],
            namespace=namespace,
            subsystem=subsystem,
            registry=registry,
        )

        self.active_requests = Gauge(
            "active_requests",
            "Number of active requests",
            ["operation", "service"],
            namespace=namespace,
            subsystem=subsystem,
            registry=registry,
        )

        # Cache metrics
        self.cache_hits_total = Counter(
            "cache_hits_total",
            "Total cache hits",
            ["cache_name", "service"],
            namespace=namespace,
            subsystem=subsystem,
            registry=registry,
        )

        self.cache_misses_total = Counter(
            "cache_misses_total",
            "Total cache misses",
            ["cache_name", "service"],
            namespace=namespace,
            subsystem=subsystem,
            registry=registry,
        )

    def record_request(
        self, operation: str, service: str, method: str, duration_seconds: float
    ):
        """Record a request with duration"""
        self.requests_total.labels(
            operation=operation, service=service, method=method
        ).inc()
        self.request_duration_seconds.labels(
            operation=operation, service=service, method=method
        ).observe(duration_seconds)

    def record_error(self, error_type: str, service: str, component: str):
        """Record an error"""
        self.errors_total.labels(
            error_type=error_type, service=service, component=component
        ).inc()

    def record_cache_hit(self, cache_name: str, service: str):
        """Record a cache hit"""
        self.cache_hits_total.labels(cache_name=cache_name, service=service).inc()

    def record_cache_miss(self, cache_name: str, service: str):
        """Record a cache miss"""
        self.cache_misses_total.labels(cache_name=cache_name, service=service).inc()

    def inc_active_requests(self, operation: str, service: str):
        """Increment active requests"""
        self.active_requests.labels(operation=operation, service=service).inc()

    def dec_active_requests(self, operation: str, service: str):
        """Decrement active requests"""
        self.active_requests.labels(operation=operation, service=service).dec()

    def timing(self, metric_name: str, value_ms: float, tags: "MetricTags" = None):
        """
        Record a timing metric in milliseconds.

        Args:
            metric_name: Name of the metric (e.g., "seek.router.latency_ms")
            value_ms: Duration in milliseconds
            tags: Optional metric tags
        """
        # Convert ms to seconds for Prometheus histogram
        duration_seconds = value_ms / 1000.0
        tag_dict = tags.to_dict() if tags else {}

        operation = tag_dict.get("operation", metric_name)
        service = tag_dict.get("service", self.namespace)
        method = tag_dict.get("method", "timing")

        self.request_duration_seconds.labels(
            operation=operation, service=service, method=method
        ).observe(duration_seconds)

    def gauge(self, metric_name: str, value: float, tags: "MetricTags" = None):
        """
        Set a gauge metric value.

        Args:
            metric_name: Name of the metric
            value: Gauge value to set
            tags: Optional metric tags
        """
        tag_dict = tags.to_dict() if tags else {}
        operation = tag_dict.get("operation", metric_name)
        service = tag_dict.get("service", self.namespace)

        self.active_requests.labels(operation=operation, service=service).set(value)

    def increment(self, metric_name: str, value: int = 1, tags: "MetricTags" = None):
        """
        Increment a counter metric.

        Args:
            metric_name: Name of the metric
            value: Amount to increment by (default 1)
            tags: Optional metric tags
        """
        tag_dict = tags.to_dict() if tags else {}
        operation = tag_dict.get("operation", metric_name)
        service = tag_dict.get("service", self.namespace)
        method = tag_dict.get("method", "count")

        self.requests_total.labels(
            operation=operation, service=service, method=method
        ).inc(value)


class IncidentMetrics:
    """
    Incident lifecycle metrics
    Matches Go metrics.IncidentMetrics pattern
    """

    def __init__(
        self,
        service_name: str,
        registry: Optional[CollectorRegistry] = None,
    ):
        self.service_name = service_name
        self.registry = registry

        self.incident_active = Gauge(
            "incident_active",
            "Number of active incidents",
            ["service", "severity", "incident_type"],
            registry=registry,
        )

        self.incident_mttr_minutes = Histogram(
            "incident_mttr_minutes",
            "Mean Time To Resolution in minutes",
            ["service", "severity"],
            buckets=(1, 5, 10, 15, 30, 60, 120, 240, 480, 960),
            registry=registry,
        )

        self.incident_total = Counter(
            "incident_total",
            "Total number of incidents",
            ["service", "severity", "incident_type"],
            registry=registry,
        )

    def start_incident(self, severity: str, incident_type: str):
        """Start tracking an incident"""
        self.incident_active.labels(
            service=self.service_name, severity=severity, incident_type=incident_type
        ).inc()
        self.incident_total.labels(
            service=self.service_name, severity=severity, incident_type=incident_type
        ).inc()

    def resolve_incident(
        self, severity: str, incident_type: str, duration_minutes: float
    ):
        """Resolve an incident and record MTTR"""
        self.incident_active.labels(
            service=self.service_name, severity=severity, incident_type=incident_type
        ).dec()
        self.incident_mttr_minutes.labels(
            service=self.service_name, severity=severity
        ).observe(duration_minutes)


def create_nop_metrics() -> ServiceMetrics:
    """Create no-op metrics for testing - matches Go NewNopMetrics"""
    # Return a metrics instance with null registry (won't actually record)
    return ServiceMetrics("test", "test", registry=CollectorRegistry())


# Alias for backwards compatibility and consistent naming
MetricsClient = ServiceMetrics


class MetricTags:
    """Helper class for metric tags/labels"""

    def __init__(
        self,
        service: str = "",
        operation: str = "",
        component: str = "",
        method: str = "",
        **extra_tags,
    ):
        self.service = service
        self.operation = operation
        self.component = component
        self.method = method
        self.extra = extra_tags

    def to_dict(self) -> dict:
        """Convert to dictionary for use with Prometheus labels"""
        base = {
            "service": self.service,
            "operation": self.operation,
            "component": self.component,
            "method": self.method,
        }
        # Filter out empty values
        base = {k: v for k, v in base.items() if v}
        base.update(self.extra)
        return base
