package main

import (
	"project/master"
	"project/master/slavecomm"
	"project/mscomm"
)

func main() {

	masterCh := make(chan mscomm.Package)
	connEventCh := make(chan mscomm.ConnectionEvent)
	go slavecomm.Listener(12221, masterCh, connEventCh)
	go master.Run(masterCh, connEventCh, make(chan struct{}))

	select {}
}
