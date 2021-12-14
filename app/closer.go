package app

type Closer interface {
	Close() error
}

type CloserFunc func() error

func (c CloserFunc) Close() error {
	return c()
}
