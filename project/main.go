package main

import (
	"elevator/scout"
	//"elevator/udp_commands"
	. "fmt"
	"strings"
	. "elevator/constants"
	"elevator/udp_commands"
	"time"
)

func setElevatorRole(elevator_role *Role){
	*elevator_role = Unknown
}

func main() {
	Printf("Starting elevator ...\n\n")
	var elevator_role Role = Unknown
	IP_address, err := scout.LocalIP()
	if (err != nil){
		// Should maybe become master
		Printf("Error when getting local IP. Probably disconnected.\n")
	} else{
		Printf("Local IP: %s\n", IP_address)
	}

	// Variables
	// var master_port string = "1861"  // Civil war
	delta_t_keep_alive := 5 *  time.Second

	// Channels
	send_udp_channel    	          := make(chan string)
	recieve_udp_channel 	          := make(chan string)
	mse_filtered_udp_channel          := make(chan string)
	role_channel				      := make(chan Role)

	// UDP
	go scout.ListenForInfo(recieve_udp_channel)
	go scout.BroadcastInfo(IP_address, send_udp_channel)
	go scout.SendKeepAliveMessage(IP_address, delta_t_keep_alive)

	// Master-slave election
	go udp_commands.MasterSlaveElection(role_channel, mse_filtered_udp_channel)

	send_udp_channel <- Master_slave_election

	setElevatorRole(&elevator_role)

	Printf("Elevator role: %s\n\n", elevator_role)

	for {
		select {
			case recieved_message := <- recieve_udp_channel:
				splitted_string := strings.Split(recieved_message, ": ")
				IP_Addr_sender, message := splitted_string[0], splitted_string[1]
				Printf("%s: %s\n", IP_Addr_sender, message)
				if (message == Master_slave_election){
					// Send start-signal to master-slave election thread
				}
				if (message == Keep_alive){
					// Update IP-address-list and see if new master should be elected
				}
		}
	}
}
