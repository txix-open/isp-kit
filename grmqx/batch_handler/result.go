package batch_handler

const (
	Unknown = iota
	Ack
	Retry
	MoveToDlq
)

// Result represents the outcome of batch message processing.
type Result struct {
	status int8
	Err    error
}
