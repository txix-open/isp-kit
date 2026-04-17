// Package rabbitmq_metrics provides Prometheus metric collectors for RabbitMQ publisher and consumer operations.
// It tracks message publish/consume latencies, body sizes, and various operation counts including success,
// retry, requeue, and dead letter queue (DLQ) counts.
//
// Example usage for RabbitMQ publisher:
//
//	publisherStorage := rabbitmq_metrics.NewPublisherStorage(reg)
//	publisherStorage.ObservePublishDuration(exchange, routingKey, duration)
//	publisherStorage.IncPublishError(exchange, routingKey)
//
// Example usage for RabbitMQ consumer:
//
//	consumerStorage := rabbitmq_metrics.NewConsumerStorage(reg)
//	consumerStorage.ObserveConsumeDuration(exchange, routingKey, duration)
//	consumerStorage.IncSuccessCount(exchange, routingKey)
package rabbitmq_metrics
