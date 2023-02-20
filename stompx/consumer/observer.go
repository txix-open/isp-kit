package consumer

type Observer interface {
	Error(c *Consumer, err error)
	CloseStart(c *Consumer)
	CloseDone(c *Consumer)
	BeginConsuming(c *Consumer)
}

type NoopObserver struct {
}

func (n NoopObserver) CloseStart(c *Consumer) {
}

func (n NoopObserver) CloseDone(c *Consumer) {
}

func (n NoopObserver) BeginConsuming(c *Consumer) {
}

func (n NoopObserver) Error(c *Consumer, err error) {
}
