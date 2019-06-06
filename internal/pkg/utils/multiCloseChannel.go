package utils;

import (
	"syscall"
	"os"
	"os/signal"
	"sync"
)

//MultiCloseChannel can be close several times
type MultiCloseChannel struct {
	C    chan os.Signal
	once sync.Once
}

//NewMultiCloseChannel creates new channel
func NewMultiCloseChannel() *MultiCloseChannel {
	return &MultiCloseChannel{C: make(chan os.Signal)}
}

//NewSignalChannel returns new channel that listens for system interupts
func NewSignalChannel() *MultiCloseChannel {
	fc := NewMultiCloseChannel()
	signal.Notify(fc.C, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)
	return fc
}

//Close closes channel if not closed
func (mc *MultiCloseChannel) Close() {
	mc.once.Do(func() {
		close(mc.C)
	})
}