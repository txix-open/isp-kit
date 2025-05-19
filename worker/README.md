# Package `worker`

Пакет `worker` предоставляет настраиваемый воркер для периодического запуска задач с контролем параллелизма и возможностью корректного завершения.

## Types

### Job

Интерфейс задачи, которую выполняет воркер.

```go
type Job interface {
	Do(ctx context.Context)
}
```

### Worker

Структура воркера, выполняющего задачу с заданным интервалом и уровнем параллелизма.

**Fields (настраиваемые через опции):**

- `interval time.Duration` — интервал между запусками задачи (по умолчанию 1 секунда).
- `concurrency int` — количество параллельных воркеров (по умолчанию 1).

**Methods:**

#### `New(job Job, opts ...Option) *Worker`

Создаёт воркера с опциями и задачей для выполнения.

#### `Run(ctx context.Context)`

Запускает воркер(ы). Операция неблокирующая.

#### `Shutdown()`

Останавливает все воркеры и ожидает их завершения.

**Опции для конфигурации воркера:**

#### `WithInterval(interval time.Duration) Option`

Установить интервал между запусками.

#### `WithConcurrency(concurrency int) Option`

Установить количество параллельных воркеров.

## Usage

```go
package main

import (
	"context"
	"fmt"
	"time"

	"github.com/txix-open/worker"
)

// Пример задачи
type PrintJob struct{}

func (p PrintJob) Do(ctx context.Context) {
	fmt.Println("Job executed at", time.Now())
}

func main() {
	job := PrintJob{}

	// Создаём воркер с интервалом 2 секунды и 3 параллельными воркерами
	w := worker.New(job,
		worker.WithInterval(2*time.Second),
		worker.WithConcurrency(3),
	)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Запускаем воркер
	w.Run(ctx)

	// Работаем 7 секунд
	time.Sleep(7 * time.Second)

	// Останавливаем воркеры и ждём завершения
	w.Shutdown()

	fmt.Println("Worker shutdown complete")
}
```
