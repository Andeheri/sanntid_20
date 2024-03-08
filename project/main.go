package main

import (
	. "elevator/constants"
	"elevator/rblog"
	"elevator/scout"
	. "fmt"
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
	ipAddressMap := make(map[string]int)
	masterIP := LoopbackIp // Default is loopback address

	Printf("Starting elevator ...\n")

	// Channels
	recieveUDPChannel := make(chan string)
	tofromMSEChannel := make(chan ToMSE)
	fromMSEChannel := make(chan FromMSE)

	// Start all go-threads
	go scout.ListenForInfo(recieveUDPChannel)
	go scout.SendKeepAliveMessage(DeltaTKeepAlive)
	go scout.TrackMissedKeepAliveMessagesAndMSE(DeltaTSamplingKeepAlive, NumKeepAlive, recieveUDPChannel, tofromMSEChannel)
	go scout.MasterSlaveElection(fromMSEChannel, tofromMSEChannel)

	localIP, err := scout.LocalIP()
	if err != nil {
		// Should maybe become master
		rblog.Red.Println("Error when getting local IP. Probably disconnected.")
		tofromMSEChannel <- ToMSE{LocalIP: localIP, IPAddressMap: map[string]int{localIP: NumKeepAlive}}
	} else {
		rblog.Green.Printf("Local IP: %s\n\n", localIP)
	}

	for {
		select {
		case mseData := <-fromMSEChannel:
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
}
