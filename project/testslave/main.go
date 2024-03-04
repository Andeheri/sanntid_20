package main

import (
	"fmt"
	"project/slave"
)

func main() {
	fmt.Println("Hello, World!")
	initialMasterAddress := "10.100.23.192:12221"
	masterAddress := make(chan string)
	slaveQuit := make(chan struct{})
	go slave.Start(initialMasterAddress, masterAddress, slaveQuit)

	select {
		case <-slaveQuit:
			fmt.Println("Slave quit")
	}
}