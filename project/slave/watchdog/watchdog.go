package watchdog

import (
	"time"
)

func Start(timeout time.Duration, fromWatchDog chan<- struct{}, toWatchDog <-chan struct{}) {
	timer := time.NewTimer(timeout)
	go send(timeout/2, fromWatchDog)
	for {
		select {
		case <-toWatchDog:
			timer.Reset(timeout)
		case <-timer.C:
			panic("Watchdog timeout, slave was hanging for a full second. Exiting.")
		}
	}
	
}

func send(interval time.Duration, fromWatchDog chan<- struct{}) {
	for {
		time.Sleep(interval)
		fromWatchDog <- struct{}{}
	}
}