package kafkax

type Observer interface {
	ClientReady()
	ClientError(err error)
	ConsumerError(consumer Consumer, err error)
	PublisherError(publisher *Publisher, err error)
	PublishingFlow(publisher *Publisher, flow bool)
	ShutdownStarted()
	ShutdownDone()
}

type NoopObserver struct {
}

func (n NoopObserver) ClientReady() {

}

func (n NoopObserver) ClientError(err error) {

}

func (n NoopObserver) ConsumerError(consumer Consumer, err error) {

}

func (n NoopObserver) ShutdownStarted() {
}

func (n NoopObserver) ShutdownDone() {

}

func (n NoopObserver) PublisherError(publisher *Publisher, err error) {

}

func (n NoopObserver) PublishingFlow(publisher *Publisher, flow bool) {

}
