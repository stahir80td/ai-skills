"""Tests for analytics features module"""

import pytest
import numpy as np
from core.analytics.features import (
    compute_rolling_average,
    compute_percentile,
    compute_delta,
    compute_statistics,
    compute_rate_of_change,
    detect_outliers,
)


def test_compute_rolling_average(sample_values):
    """Test rolling average computation"""
    result = compute_rolling_average(sample_values, window_size=3)

    assert len(result) > 0
    assert isinstance(result, list)
    assert all(isinstance(x, float) for x in result)


def test_compute_rolling_average_empty():
    """Test rolling average with empty input"""
    result = compute_rolling_average([], window_size=3)
    assert result == []


def test_compute_rolling_average_small_window(sample_values):
    """Test rolling average with window larger than data"""
    result = compute_rolling_average(sample_values[:5], window_size=10)
    assert len(result) == 1


def test_compute_percentile(sample_values):
    """Test percentile computation"""
    p50 = compute_percentile(sample_values, 50)
    p95 = compute_percentile(sample_values, 95)
    p99 = compute_percentile(sample_values, 99)

    assert p50 > 0
    assert p95 >= p50
    assert p99 >= p95


def test_compute_percentile_empty():
    """Test percentile with empty input"""
    result = compute_percentile([], 50)
    assert result == 0.0


def test_compute_delta():
    """Test delta computation"""
    result = compute_delta(150.0, 100.0)

    assert result["absolute_delta"] == 50.0
    assert result["percent_change"] == 50.0


def test_compute_delta_negative():
    """Test delta with decrease"""
    result = compute_delta(75.0, 100.0)

    assert result["absolute_delta"] == -25.0
    assert result["percent_change"] == -25.0


def test_compute_delta_zero_previous():
    """Test delta when previous value is zero"""
    result = compute_delta(100.0, 0.0)

    assert result["absolute_delta"] == 100.0
    assert result["percent_change"] == 100.0


def test_compute_statistics(sample_values):
    """Test comprehensive statistics"""
    stats = compute_statistics(sample_values)

    assert "mean" in stats
    assert "median" in stats
    assert "std" in stats
    assert "min" in stats
    assert "max" in stats
    assert "p95" in stats
    assert "p99" in stats
    assert "count" in stats

    assert stats["count"] == len(sample_values)
    assert stats["min"] == min(sample_values)
    assert stats["max"] == max(sample_values)


def test_compute_statistics_empty():
    """Test statistics with empty data"""
    stats = compute_statistics([])

    assert stats["count"] == 0
    assert stats["mean"] == 0.0


def test_compute_rate_of_change():
    """Test rate of change computation"""
    values = [10.0, 20.0, 35.0, 45.0]
    rates = compute_rate_of_change(values, time_delta_seconds=1.0)

    assert len(rates) == len(values) - 1
    assert rates[0] == 10.0  # (20-10)/1
    assert rates[1] == 15.0  # (35-20)/1


def test_compute_rate_of_change_empty():
    """Test rate of change with insufficient data"""
    rates = compute_rate_of_change([10.0], time_delta_seconds=1.0)
    assert rates == []


def test_detect_outliers():
    """Test outlier detection"""
    # Normal data with one outlier
    values = [10.0, 12.0, 11.0, 13.0, 100.0, 12.0, 11.0]
    outliers = detect_outliers(values, threshold_std=2.0)

    assert len(outliers) > 0
    assert 4 in outliers  # The 100.0 value


def test_detect_outliers_no_outliers(sample_values):
    """Test outlier detection with normal data"""
    outliers = detect_outliers(sample_values, threshold_std=3.0)

    # Should have few or no outliers with reasonable threshold
    assert isinstance(outliers, list)


def test_detect_outliers_insufficient_data():
    """Test outlier detection with too little data"""
    outliers = detect_outliers([10.0, 20.0], threshold_std=2.0)
    assert outliers == []


def test_detect_outliers_constant_data():
    """Test outlier detection with constant values"""
    values = [5.0, 5.0, 5.0, 5.0, 5.0]
    outliers = detect_outliers(values, threshold_std=2.0)

    # No outliers in constant data
    assert outliers == []
