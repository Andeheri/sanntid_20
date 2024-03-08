package main

import (
	"project/master"
	"project/rblog"
)

func main() {

	rblog.Rainbow.Print("Promoted to master")
	rblog.Magenta.Print("Master started")

	go master.Run(12221, make(chan struct{}))

	select {}
}
