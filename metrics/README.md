# Package `metrics`

Пакет `metrics` предоставляет удобный способ регистрации и экспонирования метрик Prometheus, включая стандартные метрики Go-приложения. Поддерживает кастомную регистрацию метрик, их повторное использование и предоставляет HTTP-обработчики для вывода значений и описаний метрик.

## Types

### Registry

Структура `Registry` представляет собой обёртку над `prometheus.Registry`, обеспечивающую потокобезопасную регистрацию метрик, экспонирование и описание зарегистрированных метрик.

**Methods:**

#### `func NewRegistry() *Registry`

Создаёт новый `Registry` с предустановленными метриками:

* `go_*`
* `process_*`
* `build_info`

#### `(r *Registry) GetOrRegister(metric Metric) Metric`

Регистрирует метрику, если она ещё не была зарегистрирована. В противном случае возвращает уже существующую.

#### `(r *Registry) MetricsHandler() http.Handler`

Возвращает HTTP-обработчик (`/metrics`), совместимый с Prometheus.

#### `(r *Registry) MetricsDescriptionHandler() http.Handler`

Возвращает HTTP-обработчик, печатающий описания всех зарегистрированных метрик в текстовом виде. Используется, например, для дебага.

### Metric

Интерфейс `Metric` — это алиас для `prometheus.Collector`.

### `func GetOrRegister[M Metric](registry *Registry, metric M) M`

Обобщённая функция, упрощающая регистрацию метрик с сохранением типа.

## Global variables

### `var DefaultRegistry = NewRegistry()`

Глобальный экземпляр `Registry` со всеми предустановленными метриками.

### `var DefaultObjectives map[float64]float64`

Типовые квантили для создания `Summary`-метрик

## Functions

### `func Milliseconds(duration time.Duration) float64`

Преобразует `time.Duration` в миллисекунды как `float64`. Гарантирует, что `0` миллисекунд будет преобразовано в `1` (чтобы избежать нулевых значений в метриках).

## Usage

### Default usage flow

```go
import (
    "net/http"

    "github.com/txix-open/isp-kit/metrics"
    "github.com/prometheus/client_golang/prometheus"
)

func main() {
    registry := metrics.NewRegistry()

    http.Handle("/metrics", registry.MetricsHandler())
    http.ListenAndServe(":8080", nil)
}
```

### Registering custom metric

```go
var myCounter = prometheus.NewCounter(prometheus.CounterOpts{
    Name: "my_custom_counter_total",
    Help: "A custom counter example",
})

func main() {
    metrics.DefaultRegistry.GetOrRegister(myCounter).(prometheus.Counter).Inc()
}
```

### Using description handler

```go
http.Handle("/metrics-description", metrics.DefaultRegistry.MetricsDescriptionHandler())
```
