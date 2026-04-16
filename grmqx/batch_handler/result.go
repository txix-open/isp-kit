package batch_handler

// Result represents the outcome of batch message processing.
type Result struct {
	// Ack indicates the message should be acknowledged (successfully processed).
	Ack bool
	// Retry indicates the message should be retried using the retry policy.
	Retry bool
	// MoveToDlq indicates the message should be moved to the dead letter queue.
	MoveToDlq bool
	// Err contains the error that occurred during processing.
	Err error
}
