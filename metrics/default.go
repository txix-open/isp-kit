package metrics

// nolint:gochecknoglobals
var (
	// DefaultRegistry is a globally accessible registry instance for quick access
	// without explicit initialization. Suitable for simple applications or testing.
	DefaultRegistry = NewRegistry()
)
