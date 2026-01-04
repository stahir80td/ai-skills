"""
Data Validation - Python equivalent of Go core/analytics/validation

Provides data quality checks and validation
"""

from datetime import datetime, timedelta
from typing import List, Optional, Tuple
import numpy as np


class ValidationError(Exception):
    """Validation error with details"""

    def __init__(self, message: str, field: str = ""):
        self.message = message
        self.field = field
        super().__init__(self.message)


def validate_time_range(
    start_time: datetime,
    end_time: datetime,
    max_range_days: int = 90,
) -> Tuple[bool, Optional[str]]:
    """
    Validate time range for queries
    Matches Go validation.ValidateTimeRange

    Args:
        start_time: Query start time
        end_time: Query end time
        max_range_days: Maximum allowed range in days

    Returns:
        Tuple of (is_valid, error_message)
    """
    # Check if end is after start
    if end_time <= start_time:
        return False, "end_time must be after start_time"

    # Check range duration
    duration = end_time - start_time
    max_duration = timedelta(days=max_range_days)

    if duration > max_duration:
        return False, f"time range exceeds maximum of {max_range_days} days"

    # Check if times are in the future
    now = datetime.now(start_time.tzinfo)
    if start_time > now:
        return False, "start_time cannot be in the future"

    return True, None


def validate_data_quality(
    values: List[float],
    min_data_points: int = 10,
    max_null_percent: float = 10.0,
) -> Tuple[bool, Optional[str]]:
    """
    Validate data quality for analytics
    Matches Go validation.ValidateDataQuality

    Args:
        values: List of numeric values
        min_data_points: Minimum required data points
        max_null_percent: Maximum allowed null/NaN percentage

    Returns:
        Tuple of (is_valid, error_message)
    """
    if not values:
        return False, "no data provided"

    # Check minimum data points
    if len(values) < min_data_points:
        return False, f"insufficient data points (minimum: {min_data_points})"

    # Check for null/NaN values
    arr = np.array(values, dtype=float)
    null_count = np.isnan(arr).sum()
    null_percent = (null_count / len(values)) * 100

    if null_percent > max_null_percent:
        return (
            False,
            f"too many null values ({null_percent:.1f}% > {max_null_percent}%)",
        )

    return True, None


def validate_value_range(
    value: float,
    min_value: Optional[float] = None,
    max_value: Optional[float] = None,
    field_name: str = "value",
) -> None:
    """
    Validate that a value is within expected range
    Raises ValidationError if invalid

    Args:
        value: Value to validate
        min_value: Minimum allowed value
        max_value: Maximum allowed value
        field_name: Name of field for error messages
    """
    if min_value is not None and value < min_value:
        raise ValidationError(
            f"{field_name} must be >= {min_value} (got {value})",
            field=field_name,
        )

    if max_value is not None and value > max_value:
        raise ValidationError(
            f"{field_name} must be <= {max_value} (got {value})",
            field=field_name,
        )


def validate_required_fields(data: dict, required_fields: List[str]) -> None:
    """
    Validate that all required fields are present in data
    Raises ValidationError if any field is missing

    Args:
        data: Data dictionary
        required_fields: List of required field names
    """
    missing_fields = [f for f in required_fields if f not in data or data[f] is None]

    if missing_fields:
        raise ValidationError(
            f"missing required fields: {', '.join(missing_fields)}",
        )


def check_data_completeness(
    values: List[float],
    expected_count: int,
) -> float:
    """
    Calculate data completeness percentage

    Args:
        values: List of values (may contain NaN)
        expected_count: Expected number of data points

    Returns:
        Completeness percentage (0-100)
    """
    if expected_count == 0:
        return 0.0

    arr = np.array(values, dtype=float)
    valid_count = len(arr) - np.isnan(arr).sum()

    return (valid_count / expected_count) * 100


def detect_data_gaps(
    timestamps: List[datetime],
    expected_interval_seconds: float,
    max_gap_multiplier: float = 2.0,
) -> List[Tuple[int, int]]:
    """
    Detect gaps in time-series data

    Args:
        timestamps: List of timestamps (must be sorted)
        expected_interval_seconds: Expected time between data points
        max_gap_multiplier: Gap is detected if interval > expected * multiplier

    Returns:
        List of (start_index, end_index) tuples where gaps occur
    """
    if len(timestamps) < 2:
        return []

    gaps = []
    max_gap = timedelta(seconds=expected_interval_seconds * max_gap_multiplier)

    for i in range(1, len(timestamps)):
        interval = timestamps[i] - timestamps[i - 1]
        if interval > max_gap:
            gaps.append((i - 1, i))

    return gaps
