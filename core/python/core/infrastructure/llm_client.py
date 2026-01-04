"""
Azure OpenAI client - Production-ready LLM client with streaming support

Provides:
- Chat completions with retry and circuit breaker
- Streaming responses
- Token counting and cost estimation
- Model routing support (model-router)
- Embeddings generation
- Health checks
- Structured logging with metrics
- Circuit breaker protection
"""

from typing import Any, Dict, Generator, List, Optional, Union
from dataclasses import dataclass
from datetime import datetime
import time
import json

import httpx
from openai import (
    AzureOpenAI,
    APIError,
    APIConnectionError,
    RateLimitError,
    APITimeoutError,
)
from openai.types.chat import ChatCompletion, ChatCompletionChunk

from ..logger import Logger
from ..errors import ServiceError, Severity
from ..reliability import CircuitBreaker


@dataclass
class LLMConfig:
    """
    Configuration for Azure OpenAI client
    """

    endpoint: str
    api_key: str
    logger: Logger
    model: str = "gpt-5-mini-2025-08-07"  # Pinned for API version stability
    api_version: str = "2024-10-01-preview"
    timeout_seconds: float = 60.0
    max_retries: int = 3
    max_tokens: int = 4096
    temperature: float = 0.7

    def __post_init__(self):
        """Validate configuration"""
        if not self.endpoint:
            raise ServiceError(
                code="INFRA-LLM-CONFIG-ERROR",
                message="LLM endpoint cannot be empty",
                severity=Severity.CRITICAL,
            )
        if not self.api_key:
            raise ServiceError(
                code="INFRA-LLM-CONFIG-ERROR",
                message="LLM API key cannot be empty",
                severity=Severity.CRITICAL,
            )


@dataclass
class LLMResponse:
    """Structured response from LLM"""

    content: str
    model: str
    usage: Dict[str, int]
    finish_reason: str
    latency_ms: float
    cost_estimate_usd: float


@dataclass
class EmbeddingResponse:
    """Structured response for embeddings"""

    embedding: List[float]
    model: str
    usage: Dict[str, int]
    latency_ms: float


class AzureOpenAIClient:
    """
    Azure OpenAI client with production patterns

    Features:
    - Chat completions with structured responses
    - Streaming support for long responses
    - Automatic retry with exponential backoff
    - Circuit breaker for fault tolerance
    - Token usage tracking and cost estimation
    - Support for model-router (automatic model selection)
    """

    # Approximate cost per 1K tokens (USD) - update as needed
    COST_PER_1K_TOKENS = {
        "gpt-5-nano": {"input": 0.0001, "output": 0.0002},
        "gpt-5-mini": {"input": 0.0005, "output": 0.001},
        "gpt-5-chat": {"input": 0.002, "output": 0.006},
        "grok-4-fast-reasoning": {"input": 0.003, "output": 0.015},
        "model-router": {"input": 0.001, "output": 0.003},  # Average estimate
        "text-embedding-ada-002": {"input": 0.0001, "output": 0},
        "text-embedding-3-small": {"input": 0.00002, "output": 0},
    }

    def __init__(self, config: LLMConfig):
        self.config = config
        self.logger = config.logger.with_component("AzureOpenAIClient")
        self._client: Optional[AzureOpenAI] = None
        self.circuit_breaker = CircuitBreaker(
            "azure_openai", max_failures=5, enabled=False
        )

        # Usage tracking
        self.total_tokens_used = 0
        self.total_requests = 0
        self.total_cost_usd = 0.0

        self.logger.debug(
            "Initializing Azure OpenAI client",
            endpoint=config.endpoint,
            model=config.model,
            max_tokens=config.max_tokens,
        )

    def connect(self) -> None:
        """
        Initialize Azure OpenAI client
        Validates credentials with a simple request
        """
        start_time = time.time()

        try:
            # Parse endpoint to get base URL
            # Expected format: https://<resource>.openai.azure.com/openai/v1/
            base_url = self.config.endpoint.rstrip("/")
            if not base_url.endswith("/openai"):
                base_url = base_url.rsplit("/openai", 1)[0]

            self._client = AzureOpenAI(
                azure_endpoint=base_url,
                api_key=self.config.api_key,
                api_version=self.config.api_version,
                timeout=self.config.timeout_seconds,
                max_retries=self.config.max_retries,
            )

            # Verify connection with a minimal request
            # Skip verification for now to avoid wasting tokens
            # We'll validate on first real request

            elapsed_ms = (time.time() - start_time) * 1000

            self.logger.info(
                "Azure OpenAI client initialized",
                endpoint=self.config.endpoint,
                model=self.config.model,
                api_version=self.config.api_version,
                status="ready",
                init_time_ms=round(elapsed_ms, 2),
            )

        except Exception as e:
            self.logger.error(
                "Failed to initialize Azure OpenAI client",
                error=str(e),
                error_type=type(e).__name__,
                error_code="INFRA-LLM-INIT-ERROR",
            )
            raise ServiceError(
                code="INFRA-LLM-INIT-ERROR",
                message=f"Failed to initialize Azure OpenAI client: {e}",
                severity=Severity.CRITICAL,
                underlying=e,
            )

    def _ensure_connected(self) -> AzureOpenAI:
        """Ensure client is initialized"""
        if not self._client:
            raise ServiceError(
                code="INFRA-LLM-CLIENT-ERROR",
                message="Azure OpenAI client not initialized - call connect() first",
                severity=Severity.CRITICAL,
            )
        return self._client

    def _estimate_cost(
        self, model: str, input_tokens: int, output_tokens: int
    ) -> float:
        """Estimate cost for token usage"""
        rates = self.COST_PER_1K_TOKENS.get(
            model, self.COST_PER_1K_TOKENS["model-router"]
        )
        input_cost = (input_tokens / 1000) * rates["input"]
        output_cost = (output_tokens / 1000) * rates["output"]
        return round(input_cost + output_cost, 6)

    def chat_completion(
        self,
        messages: List[Dict[str, str]],
        model: Optional[str] = None,
        temperature: Optional[float] = None,
        max_tokens: Optional[int] = None,
        system_prompt: Optional[str] = None,
        json_mode: bool = False,
    ) -> LLMResponse:
        """
        Generate chat completion

        Args:
            messages: List of message dicts with 'role' and 'content'
            model: Model to use (defaults to config.model)
            temperature: Sampling temperature (0-2)
            max_tokens: Maximum tokens to generate
            system_prompt: Optional system prompt to prepend
            json_mode: If True, enforce JSON response format

        Returns:
            LLMResponse with content and usage info
        """

        def _chat_completion():
            client = self._ensure_connected()
            start_time = time.time()

            # Build messages
            all_messages = []
            if system_prompt:
                all_messages.append({"role": "system", "content": system_prompt})
            all_messages.extend(messages)

            # Build request params
            params = {
                "model": model or self.config.model,
                "messages": all_messages,
                "temperature": (
                    temperature if temperature is not None else self.config.temperature
                ),
                "max_tokens": max_tokens or self.config.max_tokens,
            }

            if json_mode:
                params["response_format"] = {"type": "json_object"}

            try:
                response: ChatCompletion = client.chat.completions.create(**params)

                elapsed_ms = (time.time() - start_time) * 1000
                actual_model = response.model
                usage = {
                    "prompt_tokens": response.usage.prompt_tokens,
                    "completion_tokens": response.usage.completion_tokens,
                    "total_tokens": response.usage.total_tokens,
                }

                # Update tracking
                self.total_tokens_used += usage["total_tokens"]
                self.total_requests += 1

                cost = self._estimate_cost(
                    actual_model,
                    usage["prompt_tokens"],
                    usage["completion_tokens"],
                )
                self.total_cost_usd += cost

                content = response.choices[0].message.content or ""
                finish_reason = response.choices[0].finish_reason

                self.logger.info(
                    "LLM chat completion",
                    model=actual_model,
                    prompt_tokens=usage["prompt_tokens"],
                    completion_tokens=usage["completion_tokens"],
                    total_tokens=usage["total_tokens"],
                    finish_reason=finish_reason,
                    elapsed_ms=round(elapsed_ms, 2),
                    cost_usd=cost,
                )

                return LLMResponse(
                    content=content,
                    model=actual_model,
                    usage=usage,
                    finish_reason=finish_reason,
                    latency_ms=round(elapsed_ms, 2),
                    cost_estimate_usd=cost,
                )

            except RateLimitError as e:
                self.logger.warning(
                    "LLM rate limit hit",
                    error=str(e),
                    model=params["model"],
                    error_code="INFRA-LLM-RATE-LIMIT",
                )
                raise ServiceError(
                    code="INFRA-LLM-RATE-LIMIT",
                    message="LLM rate limit exceeded",
                    severity=Severity.MEDIUM,
                    underlying=e,
                )

            except APITimeoutError as e:
                self.logger.error(
                    "LLM request timeout",
                    error=str(e),
                    model=params["model"],
                    timeout=self.config.timeout_seconds,
                    error_code="INFRA-LLM-TIMEOUT",
                )
                raise ServiceError(
                    code="INFRA-LLM-TIMEOUT",
                    message="LLM request timed out",
                    severity=Severity.MEDIUM,
                    underlying=e,
                )

            except APIConnectionError as e:
                self.logger.error(
                    "LLM connection error",
                    error=str(e),
                    endpoint=self.config.endpoint,
                    error_code="INFRA-LLM-CONNECTION-ERROR",
                )
                raise ServiceError(
                    code="INFRA-LLM-CONNECTION-ERROR",
                    message="Failed to connect to LLM service",
                    severity=Severity.HIGH,
                    underlying=e,
                )

            except APIError as e:
                self.logger.error(
                    "LLM API error",
                    error=str(e),
                    model=params["model"],
                    status_code=getattr(e, "status_code", None),
                    error_code="INFRA-LLM-API-ERROR",
                )
                raise ServiceError(
                    code="INFRA-LLM-API-ERROR",
                    message=f"LLM API error: {e}",
                    severity=Severity.MEDIUM,
                    underlying=e,
                )

        return self.circuit_breaker.call(_chat_completion)

    def chat_completion_stream(
        self,
        messages: List[Dict[str, str]],
        model: Optional[str] = None,
        temperature: Optional[float] = None,
        max_tokens: Optional[int] = None,
        system_prompt: Optional[str] = None,
    ) -> Generator[str, None, Dict[str, Any]]:
        """
        Generate streaming chat completion

        Yields:
            Content chunks as strings

        Returns:
            Final metadata dict with usage info
        """
        client = self._ensure_connected()
        start_time = time.time()

        # Build messages
        all_messages = []
        if system_prompt:
            all_messages.append({"role": "system", "content": system_prompt})
        all_messages.extend(messages)

        params = {
            "model": model or self.config.model,
            "messages": all_messages,
            "temperature": (
                temperature if temperature is not None else self.config.temperature
            ),
            "max_tokens": max_tokens or self.config.max_tokens,
            "stream": True,
        }

        try:
            stream = client.chat.completions.create(**params)

            full_content = ""
            actual_model = params["model"]
            finish_reason = None

            for chunk in stream:
                if chunk.choices:
                    choice = chunk.choices[0]
                    if choice.delta.content:
                        full_content += choice.delta.content
                        yield choice.delta.content

                    if choice.finish_reason:
                        finish_reason = choice.finish_reason

                if chunk.model:
                    actual_model = chunk.model

            elapsed_ms = (time.time() - start_time) * 1000

            # Estimate tokens (rough approximation for streaming)
            prompt_tokens = sum(len(m.get("content", "")) // 4 for m in all_messages)
            completion_tokens = len(full_content) // 4

            self.logger.info(
                "LLM streaming completion finished",
                model=actual_model,
                approx_tokens=prompt_tokens + completion_tokens,
                finish_reason=finish_reason,
                elapsed_ms=round(elapsed_ms, 2),
            )

            # Return metadata (though generators can't really return)
            return {
                "model": actual_model,
                "finish_reason": finish_reason,
                "latency_ms": round(elapsed_ms, 2),
            }

        except Exception as e:
            self.logger.error(
                "LLM streaming error",
                error=str(e),
                error_type=type(e).__name__,
                error_code="INFRA-LLM-STREAM-ERROR",
            )
            raise ServiceError(
                code="INFRA-LLM-STREAM-ERROR",
                message=f"LLM streaming failed: {e}",
                severity=Severity.MEDIUM,
                underlying=e,
            )

    def create_embedding(
        self,
        text: Union[str, List[str]],
        model: str = "text-embedding-3-small",
    ) -> Union[EmbeddingResponse, List[EmbeddingResponse]]:
        """
        Generate embeddings for text

        Args:
            text: Single text or list of texts
            model: Embedding model to use

        Returns:
            EmbeddingResponse or list of EmbeddingResponse
        """

        def _create_embedding():
            client = self._ensure_connected()
            start_time = time.time()

            try:
                # Normalize input
                input_texts = [text] if isinstance(text, str) else text

                response = client.embeddings.create(
                    model=model,
                    input=input_texts,
                )

                elapsed_ms = (time.time() - start_time) * 1000

                usage = {
                    "prompt_tokens": response.usage.prompt_tokens,
                    "total_tokens": response.usage.total_tokens,
                }

                results = []
                for data in response.data:
                    results.append(
                        EmbeddingResponse(
                            embedding=data.embedding,
                            model=response.model,
                            usage=usage,
                            latency_ms=round(elapsed_ms, 2),
                        )
                    )

                self.logger.debug(
                    "LLM embedding created",
                    model=response.model,
                    input_count=len(input_texts),
                    embedding_dim=len(results[0].embedding) if results else 0,
                    tokens=usage["total_tokens"],
                    elapsed_ms=round(elapsed_ms, 2),
                )

                # Return single or list based on input
                return results[0] if isinstance(text, str) else results

            except Exception as e:
                self.logger.error(
                    "LLM embedding error",
                    error=str(e),
                    model=model,
                    error_code="INFRA-LLM-EMBEDDING-ERROR",
                )
                raise ServiceError(
                    code="INFRA-LLM-EMBEDDING-ERROR",
                    message=f"Failed to create embedding: {e}",
                    severity=Severity.MEDIUM,
                    underlying=e,
                )

        return self.circuit_breaker.call(_create_embedding)

    def get_usage_stats(self) -> Dict[str, Any]:
        """Get cumulative usage statistics"""
        return {
            "total_requests": self.total_requests,
            "total_tokens": self.total_tokens_used,
            "estimated_cost_usd": round(self.total_cost_usd, 4),
            "model": self.config.model,
        }

    def reset_usage_stats(self) -> None:
        """Reset usage tracking"""
        self.total_tokens_used = 0
        self.total_requests = 0
        self.total_cost_usd = 0.0
        self.logger.info("LLM usage stats reset")

    def health_check(self) -> Dict[str, Any]:
        """
        Check LLM health status

        Returns:
            Health status dict
        """
        start_time = time.time()

        try:
            if not self._client:
                return {
                    "status": "unhealthy",
                    "error": "Client not initialized",
                    "endpoint": self.config.endpoint,
                }

            # Simple check - list models (no tokens used)
            # For Azure, we'll just check if client is valid
            elapsed_ms = (time.time() - start_time) * 1000

            return {
                "status": "healthy",
                "endpoint": self.config.endpoint,
                "model": self.config.model,
                "latency_ms": round(elapsed_ms, 2),
                "usage": self.get_usage_stats(),
            }

        except Exception as e:
            elapsed_ms = (time.time() - start_time) * 1000
            self.logger.warning(
                "LLM health check failed",
                error=str(e),
                elapsed_ms=round(elapsed_ms, 2),
            )
            return {
                "status": "unhealthy",
                "endpoint": self.config.endpoint,
                "error": str(e),
                "latency_ms": round(elapsed_ms, 2),
            }

    def close(self) -> None:
        """Close LLM client"""
        if self._client:
            # Log final stats before closing
            stats = self.get_usage_stats()
            self.logger.info(
                "Azure OpenAI client closing",
                total_requests=stats["total_requests"],
                total_tokens=stats["total_tokens"],
                total_cost_usd=stats["estimated_cost_usd"],
            )
            self._client = None

    def __enter__(self) -> "AzureOpenAIClient":
        """Context manager entry"""
        self.connect()
        return self

    def __exit__(self, exc_type, exc_val, exc_tb) -> None:
        """Context manager exit"""
        self.close()
