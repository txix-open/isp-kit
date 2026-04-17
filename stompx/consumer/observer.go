package consumer

// Observer defines the interface for observing consumer lifecycle events.
type Observer interface {
	// Error is called when a consumer encounters an error.
	Error(c *Consumer, err error)
	// CloseStart is called when shutdown begins.
	CloseStart(c *Consumer)
	// CloseDone is called when shutdown completes.
	CloseDone(c *Consumer)
	// BeginConsuming is called when message consumption starts.
	BeginConsuming(c *Consumer)
}

// NoopObserver is a no-op implementation of the Observer interface.
type NoopObserver struct {
}

// CloseStart performs no action.
func (n NoopObserver) CloseStart(c *Consumer) {
}

// CloseDone performs no action.
func (n NoopObserver) CloseDone(c *Consumer) {
}

// BeginConsuming performs no action.
func (n NoopObserver) BeginConsuming(c *Consumer) {
}

// Error performs no action.
func (n NoopObserver) Error(c *Consumer, err error) {
}
