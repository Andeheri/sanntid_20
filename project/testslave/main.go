package main

import (
	"fmt"
	"project/slave"
)

func main() {
	fmt.Println("Hello, World!")
	initialMasterAddress := "127.0.0.1:11000"
	masterAddress := make(chan string)

	slave.Start(initialMasterAddress, masterAddress)
}