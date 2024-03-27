package utils

import (
	"os"
	"os/signal"
	"syscall"
)

func OnShutdown(cb func()) {
	c := make(chan os.Signal, 1)
	signal.Notify(c, syscall.SIGINT, syscall.SIGTERM)
	<-c
	cb()
}
