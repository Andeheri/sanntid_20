package main

import (
	"project/master"
	"project/master/slavecomm"
)

func main() {

	masterCh := make(chan slavecomm.SlaveMessage)
	connEventCh := make(chan slavecomm.ConnectionEvent)
	go slavecomm.Listener(12221, masterCh, connEventCh)
	go master.Run(masterCh, connEventCh)

	select {}
}
