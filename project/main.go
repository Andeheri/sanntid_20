package main

import (
	. "elevator/constants"
	"elevator/scout"
	"elevator/udp_commands"
	. "fmt"
	"strings"
	"time"
)

func setElevatorRole(elevator_role *Role) {
	*elevator_role = Unknown
}

func main() {
	// Variables
	// var master_port string = "1861"  // Civil war
	delta_t_keep_alive := 5000 * time.Millisecond
	num_keep_alive := 5 // Number of missed keep-alive messages missed before assumed offline

	var elevator_role Role = Unknown
	ip_address_map := make(map[string]int)

	Printf("Starting elevator ...\n\n")

	local_IP, err := scout.LocalIP()
	if err != nil {
		// Should maybe become master
		Printf("Error when getting local IP. Probably disconnected.\n")
	} else {
		Printf("Local IP: %s\n", local_IP)
	}

	// Channels
	send_udp_channel := make(chan string)
	recieve_udp_channel := make(chan string)
	mse_filtered_udp_channel := make(chan map[string] int)
	role_channel := make(chan Role)

	// UDP
	go scout.ListenForInfo(recieve_udp_channel)
	go scout.BroadcastInfo(local_IP, send_udp_channel)
	go scout.SendKeepAliveMessage(local_IP, delta_t_keep_alive)

	// Master-slave election
	go udp_commands.MasterSlaveElection(local_IP, role_channel, mse_filtered_udp_channel)

	setElevatorRole(&elevator_role)

	Printf("Elevator role: %s\n\n", elevator_role)

	for {
		select {
			case recieved_message := <- recieve_udp_channel:
				splitted_string := strings.Split(recieved_message, ": ")
				IP_Addr_sender := splitted_string[0]
				message := splitted_string[1]
				//Printf("%s: %s\n", IP_Addr_sender, message)

				if message == Master_slave_election {
					// Send start-signal to master-slave election thread
				}
				if message == Keep_alive {
					// Update IP-address-list and see if new master should be elected
					if (IP_Addr_sender != local_IP){
						_, exists := ip_address_map[IP_Addr_sender]
						ip_address_map[IP_Addr_sender] = num_keep_alive
						if (!exists){
							Println("Current IP-adress list:", ip_address_map)
							// Create deep copy
							ip_address_map_copy := make(map[string]int)
							// Manually copy elements from the original map to the new map
							for key, value := range ip_address_map {
								ip_address_map_copy[key] = value
							}
							mse_filtered_udp_channel <- ip_address_map_copy
						}
					}
				}
			case elevator_role = <- role_channel:
				// Do something based on it being slave or master
		}
	}
}
