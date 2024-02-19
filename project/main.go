package main

import (
	"elevator/scout"
	"elevator/udp_commands"
	. "fmt"
	"strings"
	"udp_commands"
)

type role string

// Master-slave states
const (
	master  role = "master"
	slave   role = "slave"
	unknown role = "unknown"
)

// Commands
const (
	master_slave_election string = "master_slave_election"
)

func setElevatorRole(elevator_role *role){
	*elevator_role = unknown
}

func main() {
	var elevator_role role = unknown
	// IP_address, err := scout.LocalIP()
	// if (err != nil){
	// 	Printf("Error when getting local IP:\n%s\n", err)
	// }
	// var master_port string = "1861"  // Civil war

	send_udp_channel    := make(chan string)
	recieve_udp_channel := make(chan string)

	go scout.BroadcastInfo(send_udp_channel)
	go scout.ListenForInfo(recieve_udp_channel)

	send_udp_channel <- master_slave_election  // Starts master-slave election

	setElevatorRole(&elevator_role)

	Printf("Elevator role: %s\n", elevator_role)

	select {
		case recieved_message := <- recieve_udp_channel:
			Printf("%s\n", recieved_message)

			splitted_string := strings.Split(recieved_message, ": ")
			IP_Addr_sender, message := splitted_string[0], splitted_string[1]

			if (message == master_slave_election){
				// Compare IP-addresses
				udp_commands.MasterSlaveElection(&elevator_role)
			}
	}
}
