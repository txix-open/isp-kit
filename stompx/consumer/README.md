# Package `consumer`

Пакет `consumer` предназначен для организации надёжного потребления сообщений из очереди через протокол STOMP с автоматическим переподключением, поддержкой конкурентной обработки и отслеживанием состояния.

## Types

### Watcher

Структура `Watcher` реализует высокоуровневый наблюдатель за процессом потребления сообщений. Отвечает за управление жизненным циклом потребителя, автоматический повтор подключения и обработку ошибок.

**Methods:**

#### `(w *Watcher) Run(ctx context.Context) error`

Запускает процесс наблюдения и потребления сообщений с ожиданием первой сессии. Блокирующая операция, возвращает ошибку при неудачном первом подключении или завершении контекста.

#### `(w *Watcher) Serve(ctx context.Context)`

Запускает процесс наблюдения в отдельной горутине. Не блокирует вызывающий поток.

#### `(w *Watcher) Shutdown()`

Выполняет корректное завершение работы `Watcher`, останавливая внутренние горутины.

### Consumer

Структура `Consumer` инкапсулирует логику подключения к STOMP-серверу, подписки на очередь и конкурентной обработки поступающих сообщений с подтверждениями (ack/nack).

**Methods:**

#### `(c *Consumer) Run() error`

Запускает обработку сообщений с учётом заданной конкуренции. Блокирующая операция, возвращает ошибку при неустранимой ошибке подключения или подписки.

#### `(c *Consumer) Close() error`

Выполняет корректное завершение работы, включая отписку от очереди, ожидание завершения обработки сообщений и отключение от сервера.

### Delivery

Структура `Delivery` представляет собой сообщение, доставленное из очереди, с возможностью подтверждения успешной обработки (`Ack`) или отказа (`Nack`).

**Methods:**

#### `(d *Delivery) Ack() error`

Подтверждает успешную обработку сообщения.

#### `(d *Delivery) Nack() error`

Отмечает сообщение как неуспешно обработанное.

### Handler

Интерфейс `Handler` описывает обработчик сообщений из очереди.

### Middleware

Тип `Middleware` — функция, оборачивающая обработчик для расширения функциональности (логирование, ретрай, и т.п.).

### Config

Структура `Config` содержит конфигурацию подключения и работы потребителя: адрес сервера, очередь, параметры подключения и подписки, конкуренция, middlewares, обработчик сообщений и наблюдатель событий.

## Usage

### Default usage flow

```go
package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/txix-open/isp-kit/consumer"
)

func main() {
	handler := consumer.HandlerFunc(func(delivery consumer.Delivery) error {
		log.Printf("received message: %s", delivery.Body())
		// Обработка сообщения...
		return delivery.Ack()
	})

	cfg := consumer.NewConfig(
		"tcp://localhost:61613",
		"/queue/example",
		handler,
		consumer.WithConcurrency(5),
	)

	watcher := consumer.NewWatcher(cfg)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go func() {
		if err := watcher.Run(ctx); err != nil {
			log.Fatalf("consumer watcher error: %v", err)
		}
	}()

	// Ожидаем системных сигналов для завершения
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	<-sigCh

	log.Println("shutting down consumer...")
	watcher.Shutdown()
	log.Println("consumer stopped")
}

```
