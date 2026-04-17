// Package shutdown provides utilities for handling process termination signals.
package shutdown

import (
	"os"
	"os/signal"
	"syscall"
)

// On registers a callback function to be executed when the process receives
// SIGINT or SIGTERM signals. It returns a channel that receives the signal
// when it occurs. The callback runs in a separate goroutine upon signal receipt.
//
// Example usage:
//
//	shutdown.On(func() {
//	    log.Println("Shutting down...")
//	})
func On(do func()) chan os.Signal {
	ch := make(chan os.Signal, 1)
	signal.Notify(ch, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-ch
		do()
	}()
	return ch
}
