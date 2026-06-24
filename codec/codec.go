// Package codec provides abstractions and implementations for payload
// encoding and decoding.
//
// The package defines common interfaces for streaming and byte-based codecs
// that can be used by transports such as HTTP, gRPC, and AMQP. It also
// provides a reusable Zstandard (zstd) implementation with encoder pooling
// for efficient compression.
//
// Codecs support both streaming APIs (io.Reader/io.Writer) and in-memory
// byte slice operations, allowing the same implementation to be reused
// across different transports and middleware.
//
// Example usage:
//
//	codec := codec.Default
//
//	// Encode bytes.
//	encoded, err := codec.EncodeBytes(data)
//	if err != nil {
//		log.Fatal(err)
//	}
//
//	// Decode bytes.
//	plain, err := codec.DecodeBytes(encoded)
//	if err != nil {
//		log.Fatal(err)
//	}
package codec
