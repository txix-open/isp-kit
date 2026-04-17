// Package http_metrics provides Prometheus metric collectors for HTTP server and client operations.
// It tracks request latencies, body sizes, status codes, and error counts for both incoming HTTP
// requests and outgoing HTTP client calls.
//
// Example usage for HTTP server:
//
//	reg := metrics.NewRegistry()
//	serverStorage := http_metrics.NewServerStorage(reg)
//
//	// In your HTTP handler:
//	start := time.Now()
//	// ... handle request ...
//	serverStorage.ObserveDuration(method, endpoint, time.Since(start))
//	serverStorage.CountStatusCode(method, endpoint, statusCode)
//
// Example usage for HTTP client:
//
//	clientStorage := http_metrics.NewClientStorage(reg)
//	clientStorage.ObserveDuration(endpoint, duration)
//	clientStorage.CountError(endpoint, err)
package http_metrics
