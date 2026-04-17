# isp-kit

[![Go Reference](https://pkg.go.dev/badge/github.com/txix-open/isp-kit.svg)](https://pkg.go.dev/github.com/txix-open/isp-kit)
[![Go Report Card](https://goreportcard.com/badge/github.com/txix-open/isp-kit)](https://goreportcard.com/report/github.com/txix-open/isp-kit)

A powerful and lightweight application lifecycle management framework and utility suite for building scalable Go microservices.

## 🌟 Overview

**isp-kit** provides a unified ecosystem for developing production-ready services. From configuration management to distributed tracing, it offers a structured approach to common microservice patterns, allowing you to focus on business logic rather than boilerplate.

## 📦 Features

- **Application Lifecycle** - Lightweight application management with graceful startup/shutdown
- **Configuration Management** - Flexible config system with YAML, environment variables, and hot-reload support
- **Database Integration** - PostgreSQL client with transaction support, migrations, and automatic schema management
- **Message Queues** - High-level abstractions for Kafka, RabbitMQ, and STOMP with middleware chains
- **HTTP & gRPC** - Production-ready servers and clients with middleware, metrics, and tracing
- **Distributed Tracing** - OpenTelemetry integration across all transport layers
- **Metrics Collection** - Prometheus-based metrics for HTTP, gRPC, Kafka, RabbitMQ, SQL, and background jobs
- **Error Handling** - Structured error types for HTTP and gRPC with business error codes
- **Cluster Coordination** - Client-side functionality for distributed configuration and service discovery
- **Health Checks** - Standardized health check endpoints with component-level status

## 📚 Core Packages

### Application & Bootstrap

| Package | Description |
|---------|-------------|
| [`app`](https://pkg.go.dev/github.com/txix-open/isp-kit/app) | Lightweight application lifecycle management with Runner components |
| [`bootstrap`](https://pkg.go.dev/github.com/txix-open/isp-kit/bootstrap) | Unified initialization framework for configuration, logging, and infrastructure |
| [`config`](https://pkg.go.dev/github.com/txix-open/isp-kit/config) | Flexible configuration management with multiple sources and type-safe retrieval |

### Database & Storage

| Package | Description |
|---------|-------------|
| [`db`](https://pkg.go.dev/github.com/txix-open/isp-kit/db) | PostgreSQL client with sqlx/pgx integration and transaction support |
| [`dbx`](https://pkg.go.dev/github.com/txix-open/isp-kit/dbx) | Extended database client with migrations and schema management |
| [`dbrx`](https://pkg.go.dev/github.com/txix-open/isp-kit/dbrx) | Dynamic database client with hot-reload capability |

### Messaging

| Package | Description |
|---------|-------------|
| [`kafkax`](https://pkg.go.dev/github.com/txix-open/isp-kit/kafkax) | High-level Kafka abstraction with franz-go client |
| [`grmqx`](https://pkg.go.dev/github.com/txix-open/isp-kit/grmqx) | RabbitMQ wrapper with automatic topology declaration |
| [`stompx`](https://pkg.go.dev/github.com/txix-open/isp-kit/stompx) | STOMP protocol wrapper for message brokers |

### Communication

| Package | Description |
|---------|-------------|
| [`http`](https://pkg.go.dev/github.com/txix-open/isp-kit/http) | Core HTTP server with middleware support |
| [`http/httpcli`](https://pkg.go.dev/github.com/txix-open/isp-kit/http/httpcli) | High-level HTTP client with retries and middleware |
| [`grpc`](https://pkg.go.dev/github.com/txix-open/isp-kit/grpc) | gRPC server and client with hot-swappable handlers |
| [`grpc/client`](https://pkg.go.dev/github.com/txix-open/isp-kit/grpc/client) | gRPC client with load balancing and observability |

### Observability

| Package | Description |
|---------|-------------|
| [`log`](https://pkg.go.dev/github.com/txix-open/isp-kit/log) | Structured logging adapter based on Uber Zap |
| [`metrics`](https://pkg.go.dev/github.com/txix-open/isp-kit/metrics) | Prometheus metrics registry and storage types |
| [`observability/tracing`](https://pkg.go.dev/github.com/txix-open/isp-kit/observability/tracing) | OpenTelemetry distributed tracing integration |
| [`observability/sentry`](https://pkg.go.dev/github.com/txix-open/isp-kit/observability/sentry) | Sentry error tracking and event monitoring |

### Utilities

| Package | Description |
|---------|-------------|
| [`healthcheck`](https://pkg.go.dev/github.com/txix-open/isp-kit/healthcheck) | Health check registry and JSON endpoint |
| [`requestid`](https://pkg.go.dev/github.com/txix-open/isp-kit/requestid) | Request ID management across contexts |
| [`retry`](https://pkg.go.dev/github.com/txix-open/isp-kit/retry) | Exponential backoff retry utilities |
| [`shutdown`](https://pkg.go.dev/github.com/txix-open/isp-kit/shutdown) | Process termination signal handling |
