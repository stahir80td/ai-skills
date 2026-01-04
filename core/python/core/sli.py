"""
SLI Tracker - Python equivalent of Go core/sli

Provides:
- Service Level Indicator tracking
- Availability, latency, error rate metrics
- Request outcome recording
- Prometheus integration
"""

from dataclasses import dataclass, field
from datetime import datetime, timezone
from typing import Optional
from prometheus_client import Counter, Histogram, Gauge, CollectorRegistry
import time


@dataclass
class RequestOutcome:
    """
    Outcome of a request for SLI tracking
    Matches Go sli.RequestOutcome struct
    """

    success: bool
    error_code: str = ""
    error_severity: str = ""
    latency_seconds: float = 0.0
    operation: str = ""
    timestamp: datetime = field(default_factory=lambda: datetime.now(timezone.utc))
    user_id: str = ""
    device_id: str = ""


@dataclass
class SLIMetrics:
    """
    Current SLI metrics snapshot
    Matches Go sli.Metrics struct
    """

    # Availability metrics
    total_requests: int = 0
    success_requests: int = 0
    failed_requests: int = 0
    availability: float = 0.0  # Percentage (0-100)

    # Latency metrics (milliseconds)
    latency_p50: float = 0.0
    latency_p95: float = 0.0
    latency_p99: float = 0.0
    latency_avg: float = 0.0

    # Error rate
    error_rate: float = 0.0
    error_rate_percent: float = 0.0

    # Throughput
    requests_per_second: float = 0.0

    # Time window
    window_start: Optional[datetime] = None
    window_end: Optional[datetime] = None


class SLITracker:
    """
    SLI tracker for measuring service health
    Matches Go sli.Tracker interface
    """

    def __init__(
        self,
        service_name: str,
        registry: Optional[CollectorRegistry] = None,
    ):
        self.service_name = service_name
        self.registry = registry

        # Prometheus metrics
        self.requests_total = Counter(
            "sli_requests_total",
            "Total requests for SLI tracking",
            ["service", "operation"],
            registry=registry,
        )

        self.requests_success = Counter(
            "sli_requests_success_total",
            "Total successful requests",
            ["service", "operation"],
            registry=registry,
        )

        self.requests_failed = Counter(
            "sli_requests_failed_total",
            "Total failed requests",
            ["service", "operation", "error_code", "severity"],
            registry=registry,
        )

        self.request_duration = Histogram(
            "sli_request_duration_seconds",
            "Request duration for SLI tracking",
            ["service", "operation"],
            buckets=(0.001, 0.005, 0.01, 0.025, 0.05, 0.1, 0.25, 0.5, 1, 2.5, 5, 10),
            registry=registry,
        )

        self.availability_percent = Gauge(
            "sli_availability_percent",
            "Current availability SLI in percent",
            ["service"],
            registry=registry,
        )

        self.latency_p95_ms = Gauge(
            "sli_latency_p95_milliseconds",
            "P95 latency SLI in milliseconds",
            ["service", "operation"],
            registry=registry,
        )

        self.latency_p99_ms = Gauge(
            "sli_latency_p99_milliseconds",
            "P99 latency SLI in milliseconds",
            ["service", "operation"],
            registry=registry,
        )

        self.error_rate_percent = Gauge(
            "sli_error_rate_percent",
            "Current error rate SLI in percent",
            ["service"],
            registry=registry,
        )

    def record_request(self, outcome: RequestOutcome):
        """
        Record a request outcome for SLI tracking
        Matches Go Tracker.RecordRequest
        """
        operation = outcome.operation or "default"

        # Total requests
        self.requests_total.labels(service=self.service_name, operation=operation).inc()

        # Success/failure
        if outcome.success:
            self.requests_success.labels(
                service=self.service_name, operation=operation
            ).inc()
        else:
            self.requests_failed.labels(
                service=self.service_name,
                operation=operation,
                error_code=outcome.error_code,
                severity=outcome.error_severity,
            ).inc()

        # Latency
        if outcome.latency_seconds > 0:
            self.request_duration.labels(
                service=self.service_name, operation=operation
            ).observe(outcome.latency_seconds)

    def record_latency(self, duration_seconds: float, operation: str = "default"):
        """
        Record request latency
        Matches Go Tracker.RecordLatency
        """
        self.request_duration.labels(
            service=self.service_name, operation=operation
        ).observe(duration_seconds)

    def get_metrics(self) -> SLIMetrics:
        """
        Get current SLI metrics snapshot
        Matches Go Tracker.GetMetrics

        Note: In production, this would query Prometheus for actual metrics.
        For now, returns a basic snapshot.
        """
        return SLIMetrics()


def create_nop_tracker(service_name: str) -> SLITracker:
    """Create no-op SLI tracker for testing"""
    # Use a separate registry that won't be exposed
    return SLITracker(service_name, registry=CollectorRegistry())
