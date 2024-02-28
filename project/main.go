package main

import (
	. "elevator/constants"
	"elevator/scout"
	"elevator/master_slave"
	. "fmt"
	"strings"
	"time"
)

func startMaster(master_port string, ip_address_map map[string]struct{}){
	// Should set up TCP-connection to each ip in ip_address_map
}

func startSlave(master_port string, master_ip string){
	// Should set up
}

func main() {
	// Parameters
	master_port               := "1861"  // Civil war
	delta_t_keep_alive        := 100 * time.Millisecond
	delta_t_missed_keep_alive := 50 * time.Millisecond
	num_keep_alive            := 5 // Number of missed keep-alive messages missed before assumed offline

	// Variables
	elevator_role := Unknown
	ip_address_map := make(map[string]struct{})

	Printf("Starting elevator ...\n\n")

	local_IP, err := scout.LocalIP()
	if err != nil {
		// Should maybe become master
		Printf("Error when getting local IP. Probably disconnected.\n")
	} else {
		Printf("Local IP: %s\n", local_IP)
	}

	// Channels
	send_udp_channel               := make(chan string)
	recieve_udp_channel            := make(chan string)
	mse_filtered_udp_channel       := make(chan map[string]struct{})
	mse_channel                    := make(chan MSE_type)
	keep_alive_tracker_reciever    := make(chan string)
	keep_alive_tracker_transmitter := make(chan string)

	// UDP
	go scout.ListenForInfo(recieve_udp_channel)
	go scout.BroadcastInfo(local_IP, send_udp_channel)
	go scout.SendKeepAliveMessage(local_IP, delta_t_keep_alive)
	go scout.TrackMissedKeepAliveMessages(delta_t_missed_keep_alive, num_keep_alive, keep_alive_tracker_reciever, keep_alive_tracker_transmitter)

	// Master-slave election
	go master_slave.Election(local_IP, mse_channel, mse_filtered_udp_channel)

	Printf("Elevator role: %s\n\n", elevator_role)

	for {
		select {
			case recieved_message := <- recieve_udp_channel:
				splitted_string := strings.Split(recieved_message, ": ")
				IP_Addr_sender := splitted_string[0]
				message := splitted_string[1]
				//Printf("%s: %s\n", IP_Addr_sender, message)
				if message == Keep_alive {
					// Update IP-address-list and see if new master should be elected
					if (IP_Addr_sender != local_IP){  // if elevator is not itself
						// Send IP-address to keep-alive-tracker
						keep_alive_tracker_reciever <- IP_Addr_sender

						_, exists := ip_address_map[IP_Addr_sender]
						ip_address_map[IP_Addr_sender] = struct{}{}
						if (!exists){  // If it is a new elevator
							Println("Current IP-adress list:", ip_address_map)
							master_slave.SendMapToChannel[string, struct{}](ip_address_map, mse_filtered_udp_channel)
						}
					}
				}
			case mse_data := <- mse_channel:
				// Data recieved from Master Slave Election
				elevator_role = mse_data.Role
				master_ip := mse_data.IP
				Printf("Elevator role: %s\nMaster IP: %s\n\n", elevator_role, master_ip)
				if (elevator_role == Master){
					// Start master protocol
					startMaster(master_port, ip_address_map)
				}else if (elevator_role == Slave){
					// Start slave protocol
					startSlave(master_port, master_ip)
				}

			case not_alive_ip_address := <- keep_alive_tracker_transmitter:
				delete(ip_address_map, not_alive_ip_address)
				Println("Current IP-adress list:", ip_address_map)
				master_slave.SendMapToChannel[string, struct{}](ip_address_map, mse_filtered_udp_channel)
		}
	}
}
