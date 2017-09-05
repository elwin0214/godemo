package sig

import (
	"os"
	"os/signal"
	"syscall"
)

func RegisterStopSignal(f func()) {
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGQUIT, syscall.SIGKILL)
	go func() {
		<-sig
		f()
	}()
}
