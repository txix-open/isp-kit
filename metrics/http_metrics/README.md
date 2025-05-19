# Package `http_metrics`

Пакет `http_metrics` предоставляет инструменты для сбора и регистрации метрик HTTP-клиента и HTTP-сервера с использованием Prometheus.

## Types

### `ClientStorage`

Хранилище метрик HTTP-клиента.

#### `func NewClientStorage(reg *metrics.Registry) *ClientStorage`

Создаёт новое хранилище метрик для HTTP-клиента.

**Metrics:**

#### `client_request_duration_ms`

Общая продолжительность запроса.

#### `client_connect_duration`

Продолжительность установления соединения.

#### `client_request_write_duration`

Продолжительность записи запроса.

#### `client_dns_duration`

Продолжительность DNS-запроса.

#### `client_response_read_duration`

Продолжительность чтения ответа.

#### `client_status_code_count`

Счётчик HTTP-кодов.

#### `client_error_count`

Счётчик ошибок клиента.

**Methods:**

#### `func (s *ClientStorage) ObserveDuration(endpoint string, duration time.Duration)`

Регистрирует общую задержку запроса.

#### `func (s *ClientStorage) ObserveConnEstablishment(endpoint string, duration time.Duration)`

Регистрирует задержку установления соединения.

#### `func (s *ClientStorage) ObserveRequestWriting(endpoint string, duration time.Duration)`

Регистрирует продолжительность записи запроса.

#### `func (s *ClientStorage) ObserveDnsLookup(endpoint string, duration time.Duration)`

Регистрирует задержку DNS-запроса.

#### `func (s *ClientStorage) ObserveResponseReading(endpoint string, duration time.Duration)`

Регистрирует продолжительность чтения ответа.

#### `func (s *ClientStorage) CountStatusCode(endpoint string, code int)`

Увеличивает счётчик HTTP-кодов ответа.

#### `func (s *ClientStorage) CountError(endpoint string, err error)`

Увеличивает счётчик ошибок клиента с маскированием URL/IP.

#### `func ClientEndpointToContext(ctx context.Context, endpoint string) context.Context`

Сохраняет endpoint клиента в контекст.

#### `func ClientEndpoint(ctx context.Context) string`

Извлекает endpoint клиента из контекста.

### `ServerStorage`

Хранилище метрик HTTP-сервера

#### `func NewServerStorage(reg *metrics.Registry) *ServerStorage`

Создаёт новое хранилище метрик для HTTP-сервера.

**Metrics:**

#### `request_duration_ms`

Продолжительность запроса.

#### `request_body_size`

Размер тела запроса.

#### `response_body_size`

Размер тела ответа.

#### `status_code_count`

Счётчик HTTP-кодов.

**Methods:**

#### `func (s *ServerStorage) ObserveDuration(method string, endpoint string, duration time.Duration)`

Регистрирует задержку запроса.

#### `func (s *ServerStorage) ObserveRequestBodySize(method string, endpoint string, size int)`

Регистрирует размер тела запроса.

#### `func (s *ServerStorage) ObserveResponseBodySize(method string, endpoint string, size int)`

Регистрирует размер тела ответа.

#### `func (s *ServerStorage) CountStatusCode(method string, endpoint string, code int)`

Увеличивает счётчик HTTP-кодов ответа.

#### `func ServerEndpointToContext(ctx context.Context, endpoint string) context.Context`

Сохраняет endpoint сервера в контекст.

#### `func ServerEndpoint(ctx context.Context) string`

Извлекает endpoint сервера из контекста.

## Prometheus metrics example

```
# HELP client_request_duration_ms The total duration of HTTP client requests
# TYPE client_request_duration_ms summary
client_request_duration_ms{endpoint="external-service"} 123.4

# HELP client_connect_duration The duration of connection establishment
# TYPE client_connect_duration summary
client_connect_duration{endpoint="external-service"} 12.3

# HELP client_request_write_duration The duration of writing the request
# TYPE client_request_write_duration summary
client_request_write_duration{endpoint="external-service"} 8.7

# HELP client_dns_duration The duration of DNS lookup
# TYPE client_dns_duration summary
client_dns_duration{endpoint="external-service"} 5.6

# HELP client_response_read_duration The duration of reading the response
# TYPE client_response_read_duration summary
client_response_read_duration{endpoint="external-service"} 17.2

# HELP client_status_code_count Count of HTTP response status codes
# TYPE client_status_code_count counter
client_status_code_count{endpoint="external-service",code="200"} 42

# HELP client_error_count Count of client errors
# TYPE client_error_count counter
client_error_count{endpoint="external-service",error="connection_timeout"} 3

# HELP request_duration_ms The total duration of HTTP server requests
# TYPE request_duration_ms summary
request_duration_ms{method="GET",endpoint="/api/resource"} 110.5

# HELP request_body_size Size of the HTTP request body
# TYPE request_body_size summary
request_body_size{method="GET",endpoint="/api/resource"} 512

# HELP response_body_size Size of the HTTP response body
# TYPE response_body_size summary
response_body_size{method="GET",endpoint="/api/resource"} 1024

# HELP status_code_count Count of HTTP response status codes (server)
# TYPE status_code_count counter
status_code_count{method="GET",endpoint="/api/resource",code="200"} 58
```

## Usage

### Client

```go
clientMetrics := http_metrics.NewClientStorage(metrics.DefaultRegistry)
...
clientMetrics.ObserveDuration("external-service", duration)
```

### Server

```go
serverMetrics := http_metrics.NewServerStorage(metrics.DefaultRegistry)
...
serverMetrics.ObserveDuration("GET", "/api/resource", duration)
```
