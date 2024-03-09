package main

import (
	"fmt"
	. "project/constants"
	"project/master"
	"project/rblog"
	"project/scout"
	"project/slave"
	"time"
)

func startMaster(masterPort string, ipAddressMap map[string]int) {
	// Should set up TCP-connection to each ip in ipAddressMap
}

func main() {
	// Variables
	elevatorRole := Unknown
	masterIP := LoopbackIp // Default is loopback address
	//var ipAddressMap map[string]int

	fmt.Printf("Starting elevator ...\n")
	time.Sleep(100 * time.Millisecond) // To give elevatorserver time to boot

	// Channels
	recieveUDPChannel    := make(chan string)
	toMSEChannel         := make(chan scout.ToMSE)
	fromMSEChannel       := make(chan scout.FromMSE)
	masterAddressChannel := make(chan string)
	masterQuitChannel    := make(chan struct{})

	// Start all go-threads
	go scout.ListenUDP(recieveUDPChannel)
	go scout.SendKeepAliveMessage(DeltaTKeepAlive)
	go scout.TrackMissedKeepAliveMessagesAndMSE(DeltaTSamplingKeepAlive, NumKeepAlive, recieveUDPChannel, toMSEChannel)
	go scout.MasterSlaveElection(fromMSEChannel, toMSEChannel)

	go slave.Start(masterAddressChannel)

	localIP, err := scout.LocalIP()
	if err != nil {
		// Should maybe become master
		rblog.Red.Println("Error when getting local IP. Probably disconnected.")
		toMSEChannel <- scout.ToMSE{LocalIP: localIP, IPAddressMap: map[string]int{localIP: NumKeepAlive}}
	} else {
		rblog.Green.Printf("Local IP: %s\n\n", localIP)
	}

	for mseData := range fromMSEChannel {
		// Data recieved from Master Slave Election
		elevatorRole = mseData.ElevatorRole
		masterIP = mseData.MasterIP
		_ = mseData.CurrentIPAddressMap
		rblog.Cyan.Printf("\nElevator role: %s\nMaster IP: %s\n\n", elevatorRole, masterIP)

		if elevatorRole == Master {
			// Start master protocol
			rblog.Rainbow.Println("Promoted to Master")
			go master.Run(MasterPort, masterQuitChannel)
		}

		// Update master IP-address
		masterAddressChannel <- fmt.Sprintf("%s:%d", masterIP, MasterPort)
	}
}
