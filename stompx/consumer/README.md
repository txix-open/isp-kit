# Package `consumer`

Пакет `consumer` предназначен для организации надёжного потребления сообщений из очереди через протокол STOMP с автоматическим переподключением, поддержкой конкурентной обработки и отслеживанием состояния.

## Types

### Config

Структура `Config` содержит конфигурацию подключения и работы потребителя: адрес сервера, очередь, параметры подключения и подписки, конкуренция, middlewares, обработчик сообщений и наблюдатель событий.

**Fields:**

#### `Address string`

Адрес брокера (обязательное).

#### `Queue string`

Имя очереди (обязательное).

#### `ConnOpts []ConnOption`

Параметры подключения

#### `Concurrency int`

Количество обработчиков (по умолчанию 1).

#### `Middlewares []Middleware`

Список мидлвар.

#### `SubscriptionOpts []SubscriptionOption`

Параметры подписки.

#### `Observer Observer`

Наблюдатель за событиями жизненного цикла потребителя (ошибки, запуск, остановка). Реализация интерфейса `Observer`.

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

### Observer

Интерфейс `Observer` для слежения за ЖЦ потребителя.

**Methods:**

#### `Error(c *Consumer, err error)`

Уведомление о возникшей ошибке.

#### `CloseStart(c *Consumer)`

Уведомление о начале закрытия потребителя.

#### `CloseDone(c *Consumer)`

Уведомление о завершении закрытия потребителя.

#### `BeginConsuming(c *Consumer)`

Уведомление о начале работы потребителя.

### Handler

Интерфейс `Handler` описывает обработчик сообщений из очереди.

### Middleware

Тип `Middleware` — функция, оборачивающая обработчик для расширения функциональности (логирование, ретрай, и т.п.).

### Watcher

Структура `Watcher` реализует высокоуровневый наблюдатель за процессом потребления сообщений. Отвечает за управление жизненным циклом потребителя, автоматический повтор подключения и обработку ошибок.

**Methods:**

#### `(w *Watcher) Run(ctx context.Context) error`

Запускает процесс наблюдения и потребления сообщений с ожиданием первой сессии. Блокирующая операция, возвращает ошибку при неудачном первом подключении или завершении контекста.

#### `(w *Watcher) Serve(ctx context.Context)`

Запускает процесс наблюдения в отдельной горутине. Не блокирует вызывающий поток.

#### `(w *Watcher) Shutdown()`

Выполняет корректное завершение работы `Watcher`, останавливая внутренние горутины.

#### `(w *Watcher) Healthcheck(ctx context.Context) error`

Проверить работоспособность соединения потребителя. Возвращает ошибку при проблемах с подключением.

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
