# Package `stats`

Пакет `stats` предоставляет инструменты для сбора и экспорта метрик Kafka-паблишеров и консумеров в Prometheus.
Интегрируется с пакетами [`consumer`](../consumer) и [`publisher`](../publisher) для мониторинга производительности и
диагностики проблем.

Основные возможности

- Сбор детальных метрик работы паблишеров и консумеров
- Поддержка гистограмм и суммарных показателей
- Автоматическая регистрация метрик в Prometheus Registry

## Types

### ConsumerStorage

Реализует интерфейс `MetricStorage` для консумеров. Собирает:

- Количественные показатели (сообщения, ошибки, ребалансировки)
- Временные характеристики (задержки подключения, чтения, ожидания)
- Размеры данных (батчи, сообщения)

**Methods:**

#### `NewConsumerStorage(reg *metrics.Registry, consumerId string) *ConsumerStorage`

Конструктор хралища метрик для консумера.

**Metrics:**

```prometheus
# Основные счетчики
kafka_reader_dial_count
kafka_reader_fetch_count
kafka_reader_message_count
kafka_reader_error_count
kafka_reader_rebalance_count
kafka_reader_timeout_count

# Задержки (миллисекунды)
kafka_reader_avg_dial_time_duration_ms
kafka_reader_min_dial_time_duration_ms
kafka_reader_max_dial_time_duration_ms
kafka_reader_avg_read_time_duration_ms
kafka_reader_min_read_time_duration_ms
kafka_reader_max_read_time_duration_ms
kafka_reader_avg_wait_time_duration_ms
kafka_reader_min_wait_time_duration_ms
kafka_reader_max_wait_time_duration_ms


# Размеры данных
kafka_reader_avg_fetch_size_count
kafka_reader_min_fetch_size_count
kafka_reader_max_fetch_size_count
kafka_reader_avg_fetch_bytes_count
kafka_reader_min_fetch_bytes_count
kafka_reader_max_fetch_bytes_count
kafka_reader_message_bytes_count
kafka_reader_max_fetch_size_count

# Состояние
kafka_reader_lag_count
kafka_reader_offset_count
kafka_reader_queue_capacity_count
kafka_reader_queue_length_count
```

### PublisherStorage

Реализует интерфейс `MetricStorage` для продюсеров. Собирает:

- Статистику отправки сообщений
- Характеристики батчей
- Ошибки и повторы

**Methods:**

#### `NewPublisherStorage(reg *metrics.Registry, publisherId string) *PublisherStorage`

Конструктор хралища метрик для паблишера.

**Metrics:**

```prometheus
# Основные счетчики
kafka_writer_write_count
kafka_writer_message_count
kafka_writer_message_bytes_count
kafka_writer_retries_count
kafka_writer_error_count

# Временные показатели
kafka_writer_avg_batch_time_duration_ms
kafka_writer_min_batch_time_duration_ms
kafka_writer_max_batch_time_duration_ms
kafka_writer_avg_batch_queue_time_duration_ms
kafka_writer_min_batch_queue_time_duration_ms
kafka_writer_max_batch_queue_time_duration_ms
kafka_writer_avg_write_time_duration_ms
kafka_writer_min_write_time_duration_ms
kafka_writer_max_write_time_duration_ms
kafka_writer_avg_wait_time_duration_ms
kafka_writer_min_wait_time_duration_ms
kafka_writer_max_wait_time_duration_ms

# Размеры батчей
kafka_writer_avg_batch_size_count
kafka_writer_min_batch_size_count
kafka_writer_max_batch_size_count
kafka_writer_avg_batch_bytes_count
kafka_writer_min_batch_bytes_count
kafka_writer_max_batch_bytes_count
```
