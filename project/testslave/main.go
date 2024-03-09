package main

import (
	"project/slave"
	"time"
)

func main() {
	// initialMasterAddress := "10.100.23.192:12221"
	initialMasterAddress := "localhost:11000"
	masterAddressCh := make(chan string)

	go slave.Start(masterAddressCh)

	time.Sleep(10 * time.Second)
	masterAddressCh <- initialMasterAddress
	time.Sleep(30 * time.Second)
	masterAddressCh <- "localhost:11001"
	select {}
}
