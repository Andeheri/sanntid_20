package main

import (
	"project/master"
	"project/mscomm"
	"project/rblog"
)

func main() {

	rblog.Rainbow.Print("Promoted to master")
	rblog.Magenta.Print("Master started")

	masterCh := make(chan mscomm.Package)
	connEventCh := make(chan mscomm.ConnectionEvent)
	go master.Run(masterCh, connEventCh, make(chan struct{}))

	select {}
}
