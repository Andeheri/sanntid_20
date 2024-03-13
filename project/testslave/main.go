package main

import (
	"project/slave"
	"time"
)

func main() {
	initialMasterAddress := "127.0.0.1:11000"
	masterAddressCh := make(chan string)

	go slave.Start(masterAddressCh)

	time.Sleep(10 * time.Second)
	masterAddressCh <- initialMasterAddress
	time.Sleep(10 * time.Second)
	masterAddressCh <- "127.0.0.1:11001"
	select {}
}
