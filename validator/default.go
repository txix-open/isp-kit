package validator

// Default is a globally available Adapter instance for quick validation without
// explicit initialization. It is safe for concurrent use.
//
// Example usage:
//
//	ok, details := validator.Default.Validate(myStruct)
var (
	Default = New()
)
