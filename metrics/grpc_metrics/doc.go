// Package grpc_metrics provides Prometheus metric collectors for gRPC server and client operations.
// It tracks request latencies, body sizes, and status code counts for both incoming gRPC requests
// and outgoing gRPC client calls.
//
// Example usage for gRPC server:
//
//	serverStorage := grpc_metrics.NewServerStorage(reg)
//	serverStorage.ObserveDuration(endpoint, duration)
//	serverStorage.CountStatusCode(endpoint, code)
//
// Example usage for gRPC client:
//
//	clientStorage := grpc_metrics.NewClientStorage(reg)
//	clientStorage.ObserveDuration(endpoint, duration)
package grpc_metrics
