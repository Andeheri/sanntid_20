package main

import (
	"project/master/slavecomm"
)

func main() {
	go slavecomm.Manager()

	testch := make(chan interface{})
	go slavecomm.Listener(12221, testch)

	select {}
}
