"""Tests for analytics validation module"""

import pytest
from datetime import datetime, timedelta
from core.analytics.validation import (
    validate_time_range,
    validate_data_quality,
    validate_value_range,
    validate_required_fields,
    check_data_completeness,
    detect_data_gaps,
    ValidationError,
)


def test_validate_time_range_valid():
    """Test valid time range"""
    start = datetime.now() - timedelta(days=7)
    end = datetime.now()

    is_valid, error = validate_time_range(start, end, max_range_days=30)

    assert is_valid is True
    assert error is None


def test_validate_time_range_end_before_start():
    """Test invalid time range (end before start)"""
    start = datetime.now()
    end = datetime.now() - timedelta(days=1)

    is_valid, error = validate_time_range(start, end)

    assert is_valid is False
    assert "after start_time" in error


def test_validate_time_range_too_large():
    """Test time range exceeds maximum"""
    start = datetime.now() - timedelta(days=100)
    end = datetime.now()

    is_valid, error = validate_time_range(start, end, max_range_days=90)

    assert is_valid is False
    assert "exceeds maximum" in error


def test_validate_time_range_future():
    """Test start time in the future"""
    start = datetime.now() + timedelta(days=1)
    end = datetime.now() + timedelta(days=2)

    is_valid, error = validate_time_range(start, end)

    assert is_valid is False
    assert "future" in error


def test_validate_data_quality_valid(sample_values):
    """Test valid data quality"""
    is_valid, error = validate_data_quality(
        sample_values,
        min_data_points=5,
        max_null_percent=10.0,
    )

    assert is_valid is True
    assert error is None


def test_validate_data_quality_empty():
    """Test empty data"""
    is_valid, error = validate_data_quality([])

    assert is_valid is False
    assert "no data" in error


def test_validate_data_quality_insufficient():
    """Test insufficient data points"""
    is_valid, error = validate_data_quality(
        [1.0, 2.0, 3.0],
        min_data_points=10,
    )

    assert is_valid is False
    assert "insufficient" in error


def test_validate_data_quality_too_many_nulls():
    """Test too many null values"""
    import numpy as np

    values = [1.0, 2.0, np.nan, np.nan, np.nan, 3.0, 4.0]

    is_valid, error = validate_data_quality(
        values,
        min_data_points=5,
        max_null_percent=20.0,
    )

    assert is_valid is False
    assert "null values" in error


def test_validate_value_range_valid():
    """Test valid value range"""
    # Should not raise exception
    validate_value_range(50.0, min_value=0.0, max_value=100.0)


def test_validate_value_range_too_low():
    """Test value below minimum"""
    with pytest.raises(ValidationError) as exc_info:
        validate_value_range(-5.0, min_value=0.0, field_name="temperature")

    assert "temperature" in str(exc_info.value)
    assert ">=" in str(exc_info.value)


def test_validate_value_range_too_high():
    """Test value above maximum"""
    with pytest.raises(ValidationError) as exc_info:
        validate_value_range(150.0, max_value=100.0, field_name="percentage")

    assert "percentage" in str(exc_info.value)
    assert "<=" in str(exc_info.value)


def test_validate_required_fields_valid():
    """Test all required fields present"""
    data = {"name": "test", "value": 42, "status": "active"}
    required = ["name", "value"]

    # Should not raise exception
    validate_required_fields(data, required)


def test_validate_required_fields_missing():
    """Test missing required field"""
    data = {"name": "test"}
    required = ["name", "value", "status"]

    with pytest.raises(ValidationError) as exc_info:
        validate_required_fields(data, required)

    error_msg = str(exc_info.value)
    assert "missing required fields" in error_msg


def test_validate_required_fields_null_value():
    """Test required field with null value"""
    data = {"name": "test", "value": None}
    required = ["name", "value"]

    with pytest.raises(ValidationError):
        validate_required_fields(data, required)


def test_check_data_completeness_full():
    """Test 100% data completeness"""
    values = [1.0, 2.0, 3.0, 4.0, 5.0]
    completeness = check_data_completeness(values, expected_count=5)

    assert completeness == 100.0


def test_check_data_completeness_partial():
    """Test partial data completeness"""
    import numpy as np

    values = [1.0, 2.0, np.nan, 4.0, np.nan]
    completeness = check_data_completeness(values, expected_count=5)

    assert completeness == 60.0  # 3 out of 5


def test_check_data_completeness_zero_expected():
    """Test completeness with zero expected count"""
    completeness = check_data_completeness([1.0, 2.0], expected_count=0)
    assert completeness == 0.0


def test_detect_data_gaps(sample_timestamps):
    """Test gap detection in time series"""
    # Create a gap by removing some timestamps
    timestamps_with_gap = sample_timestamps[:3] + sample_timestamps[6:]

    gaps = detect_data_gaps(
        timestamps_with_gap,
        expected_interval_seconds=60.0,
        max_gap_multiplier=2.0,
    )

    assert len(gaps) > 0


def test_detect_data_gaps_no_gaps(sample_timestamps):
    """Test no gaps in uniform time series"""
    gaps = detect_data_gaps(
        sample_timestamps,
        expected_interval_seconds=60.0,
        max_gap_multiplier=2.0,
    )

    # Should have no gaps with uniform 60s intervals
    assert len(gaps) == 0


def test_detect_data_gaps_insufficient_data():
    """Test gap detection with too few timestamps"""
    timestamps = [datetime.now()]
    gaps = detect_data_gaps(timestamps, expected_interval_seconds=60.0)

    assert gaps == []
