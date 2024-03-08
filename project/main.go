package main

import (
	. "fmt"
	. "project/constants"
	"project/rblog"
	"project/scout"
)

func startMaster(masterPort string, ipAddressMap map[string]int) {
	// Should set up TCP-connection to each ip in ipAddressMap
}

func startSlave(masterPort string, masterIP string) {
	// Should set up
}

func main() {
	// Variables
	elevatorRole := Unknown
	masterIP := LoopbackIp // Default is loopback address
	var ipAddressMap map[string]int

	Printf("Starting elevator ...\n")

	// Channels
	recieveUDPChannel := make(chan string)
	toMSEChannel := make(chan ToMSE)
	fromMSEChannel := make(chan FromMSE)

	// Start all go-threads
	go scout.ListenForInfo(recieveUDPChannel)
	go scout.SendKeepAliveMessage(DeltaTKeepAlive)
	go scout.TrackMissedKeepAliveMessagesAndMSE(DeltaTSamplingKeepAlive, NumKeepAlive, recieveUDPChannel, toMSEChannel)
	go scout.MasterSlaveElection(fromMSEChannel, toMSEChannel)

	localIP, err := scout.LocalIP()
	if err != nil {
		// Should maybe become master
		rblog.Red.Println("Error when getting local IP. Probably disconnected.")
		toMSEChannel <- ToMSE{LocalIP: localIP, IPAddressMap: map[string]int{localIP: NumKeepAlive}}
	} else {
		rblog.Green.Printf("Local IP: %s\n\n", localIP)
	}

	for mseData := range fromMSEChannel {
		// Data recieved from Master Slave Election
		elevatorRole = mseData.ElevatorRole
		masterIP = mseData.MasterIP
		ipAddressMap = mseData.CurrentIPAddressMap
		rblog.Cyan.Printf("\nElevator role: %s\nMaster IP: %s\n\n", elevatorRole, masterIP)
		if elevatorRole == Master {
			// Start master protocol
			startMaster(MasterPort, ipAddressMap)
		} else if elevatorRole == Slave {
			// Start slave protocol
			startSlave(MasterPort, masterIP)
		}
	}
}
