package main

import (
	. "elevator/constants"
	"elevator/scout"
	"elevator/master_slave"
	. "fmt"
	"time"
)

func startMaster(master_port string, ip_address_map map[string]int){
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
	ip_address_map := make(map[string]int)
	var master_ip string

	Printf("Starting elevator ...\n\n")

	local_IP, err := scout.LocalIP()
	if err != nil {
		// Should maybe become master
		Printf("Error when getting local IP. Probably disconnected.\n")
	} else {
		Printf("Local IP: %s\n", local_IP)
	}

	// Channels
	recieveUDPChannel            := make(chan string)
	MSEUpdatedIPChannel       := make(chan ToMSE)
	MSEChannel                    := make(chan FromMSE)

	// UDP
	go scout.ListenForInfo(recieveUDPChannel)
	go scout.SendKeepAliveMessage(delta_t_keep_alive)
	go scout.TrackMissedKeepAliveMessagesAndMSE(delta_t_missed_keep_alive, num_keep_alive, recieveUDPChannel, MSEUpdatedIPChannel)

	// Master-slave election
	go master_slave.Election(MSEChannel, MSEUpdatedIPChannel)

	Printf("Elevator role: %s\n\n", elevator_role)

	for {
		select {
			case mse_data := <- MSEChannel:
				// Data recieved from Master Slave Election
				elevator_role = mse_data.Role
				master_ip = mse_data.IP
				ip_address_map = mse_data.IPAddressMap
				Printf("Elevator role: %s\nMaster IP: %s\n\n", elevator_role, master_ip)
				if (elevator_role == Master){
					// Start master protocol
					startMaster(master_port, ip_address_map)
				}else if (elevator_role == Slave){
					// Start slave protocol
					startSlave(master_port, master_ip)
				}
		}
	}
}
