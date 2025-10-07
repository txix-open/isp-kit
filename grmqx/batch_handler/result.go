package batch_handler

type Result struct {
	Ack       bool
	Retry     bool
	MoveToDlq bool
	Err       error
}
