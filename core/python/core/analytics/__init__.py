"""
Analytics Utilities - Python equivalent of Go core/analytics

Provides:
- Feature computation utilities
- Data validation
- Time-series operations
"""

__version__ = "1.0.0"

from .features import (
    compute_rolling_average,
    compute_percentile,
    compute_delta,
    compute_statistics,
    compute_rate_of_change,
    detect_outliers,
)
from .validation import validate_time_range, validate_data_quality

__all__ = [
    "compute_rolling_average",
    "compute_percentile",
    "compute_delta",
    "compute_statistics",
    "compute_rate_of_change",
    "detect_outliers",
    "validate_time_range",
    "validate_data_quality",
]
