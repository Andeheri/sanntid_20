package main

import (
	"fmt"
	"project/slave"
)

func main() {
	fmt.Println("Hello, World!")
	initialMasterAddress := "127.0.0.1:11000"
	masterAddress := make(chan string)
	slaveQuit := make(chan struct{})
	go slave.Start(initialMasterAddress, masterAddress, slaveQuit)

	select {
		case <-slaveQuit:
			fmt.Println("Slave quit")
	}
}