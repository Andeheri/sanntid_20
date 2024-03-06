package main

import (
	"fmt"
	"project/slave"
)

func main() {
	fmt.Println("Hello, World!")
	// initialMasterAddress := "10.100.23.192:12221"
	initialMasterAddress := "localhost:11000"
	masterAddress := make(chan string)
	go slave.Start(initialMasterAddress, masterAddress)

	select {}
}