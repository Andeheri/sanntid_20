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

func main() {
	fmt.Printf("Starting elevator ...\n")
	time.Sleep(100 * time.Millisecond) // To give elevatorserver time to boot

	// Watchdog timer
	watchdog := time.AfterFunc(WatchdogTimeoutPeriod, func() {
		rblog.Red.Println("[ERROR] main froze")
		panic("Vaktbikkje sier voff! - main froze. Resets program.")
	})

	// Variables
	elevatorRole := Slave
	masterIP := LoopbackIp // Default is loopback address

	// Channels
	recieveUDPChannel := make(chan string)
	fromMSEChannel := make(chan scout.FromMSE)
	masterAddressChannel := make(chan string)
	masterQuitChannel := make(chan struct{})

	// Start all go-threads
	go scout.ListenUDP(recieveUDPChannel)
	go scout.SendKeepAliveMessage(DeltaTKeepAlive)
	go scout.TrackMissedKeepAliveMessagesAndMSE(DeltaTSamplingKeepAlive, NumKeepAlive, recieveUDPChannel, fromMSEChannel)

	go slave.Start(masterAddressChannel)

	lastRole := Slave
	for {
		select {
		case mseData := <-fromMSEChannel:
			// Data recieved from Master Slave Election
			elevatorRole = mseData.ElevatorRole
			masterIP = mseData.MasterIP
			_ = mseData.CurrentIPAddressMap
			rblog.Cyan.Printf("Elevator role: %s | Master IP: %s", elevatorRole, masterIP)

			if elevatorRole == Master && lastRole == Slave {
				// Start master protocol
				rblog.Rainbow.Println("Promoted to Master")
				go master.Run(MasterPort, masterQuitChannel)
			}
			if elevatorRole == Slave && lastRole == Master {
				// Stop master protocol
				rblog.Cyan.Println("Demoted to Slave😥")
				masterQuitChannel <- struct{}{}
			}

			lastRole = elevatorRole

			// Update master IP-address
			masterAddressChannel <- fmt.Sprintf("%s:%d", masterIP, MasterPort)
		case <-time.After(WatchdogResetPeriod):
			//unblock select to reset watchdog

		}
		watchdog.Reset(WatchdogTimeoutPeriod)
	}
}
