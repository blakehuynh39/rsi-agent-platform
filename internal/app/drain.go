package app

import (
	"os"
	"os/signal"
	"sync"
	"sync/atomic"
	"syscall"
)

var (
	draining          atomic.Bool
	drainNotifyMu     sync.Mutex
	drainStarted      = make(chan struct{})
	signalDrainMu     sync.Mutex
	signalDrainOnce   sync.Once
	signalDrainNotify chan os.Signal
	signalDrainStop   chan struct{}
)

func StartDrain() {
	drainNotifyMu.Lock()
	defer drainNotifyMu.Unlock()
	if !draining.CompareAndSwap(false, true) {
		return
	}
	close(drainStarted)
}

func StopDrainForTest() {
	signalDrainMu.Lock()
	defer signalDrainMu.Unlock()
	if signalDrainNotify != nil {
		signal.Stop(signalDrainNotify)
		// Drain any buffered signals to prevent race
		for {
			select {
			case <-signalDrainNotify:
			default:
				goto drained
			}
		}
	drained:
		signalDrainNotify = nil
	}
	if signalDrainStop != nil {
		close(signalDrainStop)
		signalDrainStop = nil
	}
	signalDrainOnce = sync.Once{}
	drainNotifyMu.Lock()
	draining.Store(false)
	drainStarted = make(chan struct{})
	drainNotifyMu.Unlock()
}

func IsDraining() bool {
	return draining.Load()
}

func DrainStarted() <-chan struct{} {
	drainNotifyMu.Lock()
	defer drainNotifyMu.Unlock()
	return drainStarted
}

func InstallSignalDrain() {
	signalDrainMu.Lock()
	once := &signalDrainOnce
	signalDrainMu.Unlock()
	once.Do(func() {
		signalDrainMu.Lock()
		notify := make(chan os.Signal, 2)
		stop := make(chan struct{})
		signalDrainNotify = notify
		signalDrainStop = stop
		signal.Notify(notify, syscall.SIGINT, syscall.SIGTERM)
		signalDrainMu.Unlock()
		go func() {
			for {
				select {
				case <-notify:
					StartDrain()
				case <-stop:
					return
				}
			}
		}()
	})
}
