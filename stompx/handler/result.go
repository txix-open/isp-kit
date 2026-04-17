package handler

// Result represents the outcome of message processing.
type Result struct {
	// Ack indicates the message should be acknowledged.
	Ack bool
	// Requeue indicates the message should be requeued (negatively acknowledged).
	Requeue bool
	// Err contains an error if the processing failed.
	Err error
}

// Ack returns a Result indicating successful processing.
func Ack() Result {
	return Result{Ack: true}
}

// Requeue returns a Result indicating the message should be requeued with an error.
func Requeue(err error) Result {
	return Result{Requeue: true, Err: err}
}
