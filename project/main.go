package main

import (
	. "elevator/constants"
	"elevator/scout"
	. "fmt"
)

func startMaster(masterPort string, ipAddressMap map[string]int){
	// Should set up TCP-connection to each ip in ipAddressMap
}

func startSlave(masterPort string, masterIP string){
	// Should set up
}

func main() {
	// Variables
	elevatorRole := Unknown
	ipAddressMap := make(map[string]int)
	masterIP     := LoopbackIp  // Default is loopback address

	Printf("Starting elevator ...\n\n")

	// Channels
	recieveUDPChannel   := make(chan string)
	MSEUpdatedIPChannel := make(chan ToMSE)
	MSEChannel          := make(chan FromMSE)

	// Start all go-threads
	go scout.ListenForInfo(recieveUDPChannel)
	go scout.SendKeepAliveMessage(DeltaTKeepAlive)
	go scout.TrackMissedKeepAliveMessagesAndMSE(DeltaTMissedKeepAlive, NumKeepAlive, recieveUDPChannel, MSEUpdatedIPChannel)
	go scout.MasterSlaveElection(MSEChannel, MSEUpdatedIPChannel)

	localIP, err := scout.LocalIP()
	if err != nil {
		// Should maybe become master
		Printf("Error when getting local IP. Probably disconnected.\n")
		MSEUpdatedIPChannel <- ToMSE{LocalIP: localIP, IPAddressMap: map[string]int{localIP: NumKeepAlive}}
	} else {
		Printf("Local IP: %s\n", localIP)
	}

	for {
		select {
			case mseData := <- MSEChannel:
				// Data recieved from Master Slave Election
				elevatorRole = mseData.Role
				masterIP = mseData.IP
				ipAddressMap = mseData.IPAddressMap
				Printf("Elevator role: %s\nMaster IP: %s\n\n", elevatorRole, masterIP)
				if (elevatorRole == Master){
					// Start master protocol
					startMaster(MasterPort, ipAddressMap)
				}else if (elevatorRole == Slave){
					// Start slave protocol
					startSlave(MasterPort, masterIP)
				}
		}
	}
}
