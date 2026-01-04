"""
Tests for HTTP client
"""

import pytest
from unittest.mock import Mock, patch
import httpx

from core.logger import Logger
from core.errors import ServiceError

# Handle cassandra driver import issues that affect core.infrastructure (Python 3.13 compatibility)
try:
    from core.infrastructure import HTTPClient, HTTPConfig

    INFRASTRUCTURE_AVAILABLE = True
except ImportError as e:
    INFRASTRUCTURE_AVAILABLE = False
    pytestmark = pytest.mark.skip(
        reason=f"Infrastructure import failed (cassandra driver issue): {e}"
    )


@pytest.fixture
def logger():
    """Create a test logger"""
    return Logger("test-http", "INFO")


@pytest.fixture
def config(logger):
    """Create a valid HTTPConfig"""
    return HTTPConfig(
        base_url="http://localhost:8080",
        logger=logger,
        timeout_seconds=30.0,
        max_retries=3,
    )


def test_config_validation_empty_base_url(logger):
    """Test that empty base_url raises ServiceError"""
    with pytest.raises(ServiceError) as exc_info:
        HTTPConfig(
            base_url="",
            logger=logger,
        )
    assert exc_info.value.code == "INFRA-HTTP-CONFIG-ERROR"


def test_client_initialization(config):
    """Test client initializes with valid config"""
    client = HTTPClient(config)
    assert client.config == config
    assert client.client is not None


@patch("core.infrastructure.http_client.httpx.Client")
def test_get_success(mock_client_class, config):
    """Test successful GET request"""
    mock_response = Mock()
    mock_response.status_code = 200
    mock_response.elapsed.total_seconds.return_value = 0.1
    mock_response.raise_for_status = Mock()

    mock_http_client = Mock()
    mock_http_client.get.return_value = mock_response
    mock_client_class.return_value = mock_http_client

    client = HTTPClient(config)
    response = client.get("/api/test")

    assert response.status_code == 200
    mock_http_client.get.assert_called()


@patch("core.infrastructure.http_client.httpx.Client")
def test_get_with_params_and_headers(mock_client_class, config):
    """Test GET with query params and headers"""
    mock_response = Mock()
    mock_response.status_code = 200
    mock_response.elapsed.total_seconds.return_value = 0.1
    mock_response.raise_for_status = Mock()

    mock_http_client = Mock()
    mock_http_client.get.return_value = mock_response
    mock_client_class.return_value = mock_http_client

    client = HTTPClient(config)
    params = {"key": "value"}
    headers = {"Authorization": "Bearer token"}

    response = client.get("/api/test", params=params, headers=headers)

    assert response.status_code == 200
    call_args = mock_http_client.get.call_args
    assert call_args[1]["params"] == params
    assert call_args[1]["headers"] == headers


@patch("core.infrastructure.http_client.httpx.Client")
def test_get_timeout(mock_client_class, config):
    """Test GET request timeout"""
    mock_http_client = Mock()
    mock_http_client.get.side_effect = httpx.TimeoutException("Request timed out")
    mock_client_class.return_value = mock_http_client

    client = HTTPClient(config)

    with pytest.raises(ServiceError) as exc_info:
        client.get("/api/test")

    assert exc_info.value.code == "INFRA-HTTP-TIMEOUT"


@patch("core.infrastructure.http_client.httpx.Client")
def test_get_http_error(mock_client_class, config):
    """Test GET with HTTP error status"""
    mock_response = Mock()
    mock_response.status_code = 500
    mock_response.elapsed.total_seconds.return_value = 0.1
    mock_response.raise_for_status.side_effect = httpx.HTTPStatusError(
        "Server error", request=Mock(), response=mock_response
    )

    mock_http_client = Mock()
    mock_http_client.get.return_value = mock_response
    mock_client_class.return_value = mock_http_client

    client = HTTPClient(config)

    with pytest.raises(ServiceError) as exc_info:
        client.get("/api/test")

    assert exc_info.value.code == "INFRA-HTTP-ERROR"


@patch("core.infrastructure.http_client.httpx.Client")
def test_post_success(mock_client_class, config):
    """Test successful POST request"""
    mock_response = Mock()
    mock_response.status_code = 201
    mock_response.elapsed.total_seconds.return_value = 0.15
    mock_response.raise_for_status = Mock()

    mock_http_client = Mock()
    mock_http_client.post.return_value = mock_response
    mock_client_class.return_value = mock_http_client

    client = HTTPClient(config)
    json_data = {"key": "value"}
    response = client.post("/api/test", json=json_data)

    assert response.status_code == 201
    mock_http_client.post.assert_called()


@patch("core.infrastructure.http_client.httpx.Client")
def test_post_with_data_and_headers(mock_client_class, config):
    """Test POST with form data and headers"""
    mock_response = Mock()
    mock_response.status_code = 200
    mock_response.elapsed.total_seconds.return_value = 0.1
    mock_response.raise_for_status = Mock()

    mock_http_client = Mock()
    mock_http_client.post.return_value = mock_response
    mock_client_class.return_value = mock_http_client

    client = HTTPClient(config)
    data = {"field": "value"}
    headers = {"Content-Type": "application/x-www-form-urlencoded"}

    response = client.post("/api/test", data=data, headers=headers)

    assert response.status_code == 200


@patch("core.infrastructure.http_client.httpx.Client")
def test_put_success(mock_client_class, config):
    """Test successful PUT request"""
    mock_response = Mock()
    mock_response.status_code = 200
    mock_response.elapsed.total_seconds.return_value = 0.1
    mock_response.raise_for_status = Mock()

    mock_http_client = Mock()
    mock_http_client.put.return_value = mock_response
    mock_client_class.return_value = mock_http_client

    client = HTTPClient(config)
    json_data = {"key": "updated_value"}
    response = client.put("/api/test/123", json=json_data)

    assert response.status_code == 200


@patch("core.infrastructure.http_client.httpx.Client")
def test_delete_success(mock_client_class, config):
    """Test successful DELETE request"""
    mock_response = Mock()
    mock_response.status_code = 204
    mock_response.elapsed.total_seconds.return_value = 0.1
    mock_response.raise_for_status = Mock()

    mock_http_client = Mock()
    mock_http_client.delete.return_value = mock_response
    mock_client_class.return_value = mock_http_client

    client = HTTPClient(config)
    response = client.delete("/api/test/123")

    assert response.status_code == 204


@patch("core.infrastructure.http_client.httpx.Client")
def test_health_check_success(mock_client_class, config):
    """Test successful health check"""
    mock_response = Mock()
    mock_response.status_code = 200

    mock_http_client = Mock()
    mock_http_client.get.return_value = mock_response
    mock_client_class.return_value = mock_http_client

    client = HTTPClient(config)
    assert client.health() is True


@patch("core.infrastructure.http_client.httpx.Client")
def test_health_check_fallback_to_root(mock_client_class, config):
    """Test health check fallback to root endpoint"""
    mock_response = Mock()
    mock_response.status_code = 200

    mock_http_client = Mock()
    # First call to /health fails, second to / succeeds
    mock_http_client.get.side_effect = [Exception("Not found"), mock_response]
    mock_client_class.return_value = mock_http_client

    client = HTTPClient(config)
    assert client.health() is True


@patch("core.infrastructure.http_client.httpx.Client")
def test_health_check_failure(mock_client_class, config):
    """Test failed health check"""
    mock_http_client = Mock()
    mock_http_client.get.side_effect = Exception("Connection failed")
    mock_client_class.return_value = mock_http_client

    client = HTTPClient(config)
    assert client.health() is False


@patch("core.infrastructure.http_client.httpx.Client")
def test_close(mock_client_class, config):
    """Test closing HTTP client"""
    mock_http_client = Mock()
    mock_client_class.return_value = mock_http_client

    client = HTTPClient(config)
    client.close()

    mock_http_client.close.assert_called_once()


@patch("core.infrastructure.http_client.httpx.Client")
def test_circuit_breaker_integration(mock_client_class, config):
    """Test circuit breaker triggers after failures"""
    mock_http_client = Mock()
    mock_http_client.get.side_effect = Exception("Simulated error")
    mock_client_class.return_value = mock_http_client

    client = HTTPClient(config)

    # Trigger multiple failures
    for _ in range(6):
        try:
            client.get("/api/test")
        except:
            pass

    # Circuit should be open
    assert client.circuit_breaker.state == "open"


@patch("core.infrastructure.http_client.httpx.Client")
def test_retry_policy_integration(mock_client_class, config):
    """Test retry policy works on transient failures"""
    mock_response = Mock()
    mock_response.status_code = 200
    mock_response.elapsed.total_seconds.return_value = 0.1
    mock_response.raise_for_status = Mock()

    mock_http_client = Mock()
    # Fail twice, then succeed
    mock_http_client.get.side_effect = [
        httpx.TimeoutException("Timeout"),
        httpx.TimeoutException("Timeout"),
        mock_response,
    ]
    mock_client_class.return_value = mock_http_client

    client = HTTPClient(config)
    response = client.get("/api/test")

    # Should eventually succeed after retries
    assert response.status_code == 200
    # Should have called get 3 times (2 failures + 1 success)
    assert mock_http_client.get.call_count == 3
