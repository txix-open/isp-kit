// Package kafka_metrics provides Prometheus metric collectors for Kafka producer and consumer operations.
// It tracks message publish/consume latencies, body sizes, commit counts, retry counts, and error counts.
//
// Example usage for Kafka producer:
//
//	publisherStorage := kafka_metrics.NewPublisherStorage(reg)
//	publisherStorage.ObservePublishDuration(topic, duration)
//	publisherStorage.IncPublishError(topic)
//
// Example usage for Kafka consumer:
//
//	consumerStorage := kafka_metrics.NewConsumerStorage(reg)
//	consumerStorage.ObserveConsumeDuration(consumerGroup, topic, duration)
//	consumerStorage.IncCommitCount(consumerGroup, topic)
package kafka_metrics
