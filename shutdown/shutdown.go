package shutdown

import (
	"os"
	"os/signal"
	"syscall"
)

func On(do func()) chan os.Signal {
	ch := make(chan os.Signal, 1)
	signal.Notify(ch, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-ch
		do()
	}()
	return ch
}
