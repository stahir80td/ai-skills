"""
Feature Computation - Python equivalent of Go core/analytics/features

Provides statistical feature computations for ML pipelines
"""

from typing import List, Optional
import numpy as np
from datetime import timedelta


def compute_rolling_average(values: List[float], window_size: int) -> List[float]:
    """
    Compute rolling average over a window
    Matches Go features.ComputeRollingAverage

    Args:
        values: List of numeric values
        window_size: Size of rolling window

    Returns:
        List of rolling averages
    """
    if not values or window_size <= 0:
        return []

    arr = np.array(values)
    if len(arr) < window_size:
        return [float(np.mean(arr))]

    # Use numpy's convolve for efficient rolling window
    weights = np.ones(window_size) / window_size
    return np.convolve(arr, weights, mode="valid").tolist()


def compute_percentile(values: List[float], percentile: float) -> float:
    """
    Compute percentile of values
    Matches Go features.ComputePercentile

    Args:
        values: List of numeric values
        percentile: Percentile to compute (0-100)

    Returns:
        Percentile value
    """
    if not values:
        return 0.0

    return float(np.percentile(values, percentile))


def compute_delta(current_value: float, previous_value: float) -> dict:
    """
    Compute delta and percent change
    Matches Go features.ComputeDelta

    Args:
        current_value: Current value
        previous_value: Previous value

    Returns:
        Dictionary with absolute_delta and percent_change
    """
    absolute_delta = current_value - previous_value

    if previous_value == 0:
        percent_change = 0.0 if current_value == 0 else 100.0
    else:
        percent_change = (absolute_delta / previous_value) * 100

    return {
        "absolute_delta": absolute_delta,
        "percent_change": percent_change,
    }


def compute_statistics(values: List[float]) -> dict:
    """
    Compute comprehensive statistics for a dataset

    Returns:
        Dictionary with mean, median, std, min, max, p95, p99
    """
    if not values:
        return {
            "mean": 0.0,
            "median": 0.0,
            "std": 0.0,
            "min": 0.0,
            "max": 0.0,
            "p95": 0.0,
            "p99": 0.0,
            "count": 0,
        }

    arr = np.array(values)
    return {
        "mean": float(np.mean(arr)),
        "median": float(np.median(arr)),
        "std": float(np.std(arr)),
        "min": float(np.min(arr)),
        "max": float(np.max(arr)),
        "p95": float(np.percentile(arr, 95)),
        "p99": float(np.percentile(arr, 99)),
        "count": len(values),
    }


def compute_rate_of_change(
    values: List[float], time_delta_seconds: float
) -> List[float]:
    """
    Compute rate of change over time

    Args:
        values: List of values
        time_delta_seconds: Time between measurements

    Returns:
        List of rates (value per second)
    """
    if len(values) < 2:
        return []

    rates = []
    for i in range(1, len(values)):
        delta = values[i] - values[i - 1]
        rate = delta / time_delta_seconds if time_delta_seconds > 0 else 0
        rates.append(rate)

    return rates


def detect_outliers(values: List[float], threshold_std: float = 3.0) -> List[int]:
    """
    Detect outliers using standard deviation method

    Args:
        values: List of numeric values
        threshold_std: Number of standard deviations for outlier threshold

    Returns:
        List of indices where outliers are detected
    """
    if len(values) < 3:
        return []

    arr = np.array(values)
    mean = np.mean(arr)
    std = np.std(arr)

    if std == 0:
        return []

    z_scores = np.abs((arr - mean) / std)
    outlier_indices = np.where(z_scores > threshold_std)[0]

    return outlier_indices.tolist()
