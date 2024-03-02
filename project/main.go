package main

import (
	"fmt"
	"project/scout"
)

type role string

// Master-slave states
const (
	master  role = "master"
	slave   role = "slave"
	unknown role = "unknown"
)

func setElevatorRole(elevator_role *role) {
	*elevator_role = slave
}

func main() {
	var elevator_role role = unknown
	//var udp_port string = "1861"  // Civil war

	setElevatorRole(&elevator_role)

	fmt.Printf("Elevator role: %s\n", elevator_role)

	go scout.BroadcastInfo()
	go scout.ListenForInfo()

	select {}
}
