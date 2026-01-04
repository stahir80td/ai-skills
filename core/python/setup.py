"""
ai Core Package - Python Edition
Production-ready shared utilities for all Python services
"""

from setuptools import setup, find_packages

setup(
    name="ai-core",
    version="1.0.3",
    description="ai Core package for Python services with enterprise patterns",
    author="ai Team",
    packages=find_packages(),
    python_requires=">=3.10",
    install_requires=[
        "prometheus-client>=0.19.0",
        "structlog>=24.1.0",
        "python-json-logger>=2.0.7",
        "httpx>=0.26.0",
        "tenacity>=8.2.3",
        "pydantic>=2.5.3",
        "pydantic-settings>=2.1.0",
        # Infrastructure clients
        "scylla-driver>=3.29.0",
        "redis>=5.0.0",
        "kafka-python>=2.0.2",
        "pymongo>=4.6.0",
        "pyodbc>=5.0.0",
        "azure-identity>=1.15.0",
        "azure-keyvault-secrets>=4.7.0",
        "openai>=1.10.0",
    ],
    extras_require={
        "dev": [
            "pytest>=7.4.3",
            "pytest-asyncio>=0.23.2",
            "pytest-cov>=4.1.0",
            "mypy>=1.8.0",
            "black>=23.12.1",
            "ruff>=0.1.9",
        ],
        "analytics": [
            "numpy>=1.26.2",
            "pandas>=2.1.4",
            "scikit-learn>=1.3.2",
        ],
    },
    classifiers=[
        "Development Status :: 5 - Production/Stable",
        "Intended Audience :: Developers",
        "Programming Language :: Python :: 3.10",
        "Programming Language :: Python :: 3.11",
        "Programming Language :: Python :: 3.12",
    ],
)
