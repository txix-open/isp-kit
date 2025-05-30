# Package `grpc_metrics`

Пакет `grpc_metrics` предоставляет средства для мониторинга метрик GRPC-клиентов и серверов.

## Types

### ClientStorage

Хранилище метрик GRPC-клиента.

**Metrics:**

#### `grpc_client_request_duration_ms{endpoint="..."}`

Продолжительность вызова внешнего GRPC-сервиса.

**Methods:**

#### `func NewClientStorage(reg *metrics.Registry) *ClientStorage`

Создаёт и регистрирует метрику `client_request_duration_ms` с лейблом `endpoint` в переданном реестре.

#### `func (s *ClientStorage) ObserveDuration(endpoint string, duration time.Duration)`

Регистрирует длительность GRPC-вызова для заданного `endpoint`.

### ServerStorage

Хранилище метрик GRPC-сервера.

**Metrics:**

#### `grpc_request_duration_ms{endpoint="..."}`

Продолжительность обработки входящего GRPC-запроса.

#### `grpc_request_body_size{endpoint="..."}`

Размер тела входящего запроса.

#### `grpc_response_body_size{endpoint="..."}`

Размер тела ответа.

#### `grpc_status_code_count{endpoint="...",code="..."}`

Количество ответов по статус-кодам (`OK`, `InvalidArgument`, и т.д.).

**Methods:**

#### `func NewServerStorage(reg *metrics.Registry) *ServerStorage`

Создаёт и регистрирует все метрики сервера с нужными лейблами в переданном реестре.

#### `func (s *ServerStorage) ObserveDuration(endpoint string, duration time.Duration)`

Регистрирует длительность обработки GRPC-запроса для `endpoint`.

#### `func (s *ServerStorage) ObserveRequestBodySize(endpoint string, size int)`

Регистрирует размер тела запроса в байтах.

#### `func (s *ServerStorage) ObserveResponseBodySize(endpoint string, size int)`

Регистрирует размер тела ответа в байтах.

#### `func (s *ServerStorage) CountStatusCode(endpoint string, code codes.Code)`

Увеличивает счётчик ответов с соответствующим GRPC-кодом.

## Prometheus metrics example

```
# HELP grpc_client_request_duration_ms The latencies of calling external services via GRPC
# TYPE grpc_client_request_duration_ms summary
grpc_client_request_duration_ms{endpoint="UserService.GetUser"} 12.3

# HELP grpc_request_duration_ms The latency of the GRPC requests
# TYPE grpc_request_duration_ms summary
grpc_request_duration_ms{endpoint="UserService.GetUser"} 10.5

# HELP grpc_request_body_size The size of request body
# TYPE grpc_request_body_size summary
grpc_request_body_size{endpoint="UserService.GetUser"} 512

# HELP grpc_response_body_size The size of response body
# TYPE grpc_response_body_size summary
grpc_response_body_size{endpoint="UserService.GetUser"} 1024

# HELP grpc_status_code_count Counter of statuses codes
# TYPE grpc_status_code_count counter
grpc_status_code_count{endpoint="UserService.GetUser",code="OK"} 32
```

## Usage

### Default usage flow

```go
clientMetrics := grpc_metrics.NewClientStorage(metrics.DefaultRegistry)
defer func(start time.Time) {
    clientMetrics.ObserveDuration("UserService.GetUser", time.Since(start))
}(time.Now())

// ... GRPC call

serverMetrics := grpc_metrics.NewServerStorage(metrics.DefaultRegistry)
serverMetrics.ObserveDuration("UserService.GetUser", 12*time.Millisecond)
serverMetrics.ObserveRequestBodySize("UserService.GetUser", 512)
serverMetrics.ObserveResponseBodySize("UserService.GetUser", 1024)
serverMetrics.CountStatusCode("UserService.GetUser", codes.OK)
```
