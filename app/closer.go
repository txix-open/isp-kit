package app

// Closer represents a resource that needs to be cleaned up.
// Implementations should release resources and return an error if cleanup fails.
type Closer interface {
	Close() error
}

// CloserFunc is a function adapter that implements the Closer interface.
// It allows using simple functions as closers without defining a new type.
//
// Example usage:
//
//	closer := app.CloserFunc(func() error {
//		// close database connection
//		return nil
//	})
type CloserFunc func() error

// Close executes the underlying function to release resources.
func (c CloserFunc) Close() error {
	return c()
}
