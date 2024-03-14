package main

import (
	"fmt"
	"project/master"
	"project/rblog"
	"project/scout"
	"project/slave"
	"time"
)

const (
	masterPort            int           = 1861
	watchdogResetPeriod   time.Duration = 1 * time.Second
	watchdogTimeoutPeriod time.Duration = 2 * time.Second
)

var welcomeMessage string = "\033[41m \033[0m\033[41m \033[0m\033[43m \033[0m\033[43m \033[0m\033[42m \033[0m\033[42m \033[0m\033[46m \033[0m \033[44m \033[0m\033[44m \033[0m      \033[42m \033[0m\033[42m \033[0m\033[46m \033[0m\033[46m \033[0m\033[44m \033[0m\033[44m \033[0m\033[45m \033[0m \033[41m \033[0m\033[41m \033[0m    \033[46m \033[0m\033[46m \033[0m  \033[45m \033[0m\033[45m \033[0m\033[41m \033[0m\033[41m \033[0m\033[43m \033[0m  \033[42m \033[0m\033[46m \033[0m\033[46m \033[0m\033[44m \033[0m\033[44m \033[0m\033[45m \033[0m\033[45m \033[0m\033[41m \033[0m  \033[43m \033[0m\033[42m \033[0m\033[42m \033[0m\033[46m \033[0m\033[46m \033[0m\033[44m \033[0m  \033[45m \033[0m\033[41m \033[0m\033[41m \033[0m\033[43m \033[0m\033[43m \033[0m\033[42m \033[0m\n\033[41m \033[0m\033[43m \033[0m      \033[44m \033[0m\033[45m \033[0m      \033[42m \033[0m\033[46m \033[0m      \033[41m \033[0m\033[43m \033[0m    \033[46m \033[0m\033[44m \033[0m \033[45m \033[0m\033[45m \033[0m   \033[43m \033[0m\033[42m \033[0m    \033[44m \033[0m\033[45m \033[0m    \033[43m \033[0m\033[42m \033[0m    \033[44m \033[0m\033[45m \033[0m \033[41m \033[0m\033[41m \033[0m   \033[42m \033[0m\033[46m \033[0m\n\033[43m \033[0m\033[43m \033[0m\033[42m \033[0m\033[42m \033[0m\033[46m \033[0m   \033[45m \033[0m\033[45m \033[0m      \033[46m \033[0m\033[46m \033[0m\033[44m \033[0m\033[44m \033[0m\033[45m \033[0m   \033[43m \033[0m\033[43m \033[0m    \033[44m \033[0m\033[44m \033[0m \033[45m \033[0m\033[41m \033[0m\033[41m \033[0m\033[43m \033[0m\033[43m \033[0m\033[42m \033[0m\033[42m \033[0m    \033[45m \033[0m\033[45m \033[0m    \033[42m \033[0m\033[42m \033[0m    \033[45m \033[0m\033[45m \033[0m \033[41m \033[0m\033[43m \033[0m\033[43m \033[0m\033[42m \033[0m\033[42m \033[0m\033[46m \033[0m\n\033[43m \033[0m\033[42m \033[0m      \033[45m \033[0m\033[41m \033[0m      \033[46m \033[0m\033[44m \033[0m       \033[42m \033[0m\033[42m \033[0m  \033[44m \033[0m\033[44m \033[0m  \033[41m \033[0m\033[41m \033[0m   \033[42m \033[0m\033[46m \033[0m    \033[45m \033[0m\033[41m \033[0m    \033[42m \033[0m\033[46m \033[0m    \033[45m \033[0m\033[41m \033[0m \033[43m \033[0m\033[43m \033[0m   \033[46m \033[0m\033[44m \033[0m\n\033[42m \033[0m\033[42m \033[0m\033[46m \033[0m\033[46m \033[0m\033[44m \033[0m\033[44m \033[0m\033[45m \033[0m \033[41m \033[0m\033[41m \033[0m\033[43m \033[0m\033[43m \033[0m\033[42m \033[0m\033[42m \033[0m\033[46m \033[0m \033[44m \033[0m\033[44m \033[0m\033[45m \033[0m\033[45m \033[0m\033[41m \033[0m\033[41m \033[0m\033[43m \033[0m   \033[46m \033[0m\033[46m \033[0m\033[44m \033[0m\033[44m \033[0m   \033[41m \033[0m\033[43m \033[0m   \033[46m \033[0m\033[46m \033[0m    \033[41m \033[0m\033[41m \033[0m     \033[46m \033[0m\033[44m \033[0m\033[44m \033[0m\033[45m \033[0m\033[45m \033[0m\033[41m \033[0m  \033[43m \033[0m\033[42m \033[0m   \033[44m \033[0m\033[44m \033[0m"

func main() {
	fmt.Println(welcomeMessage)

	// Watchdog timer
	watchdog := time.AfterFunc(watchdogTimeoutPeriod, func() {
		rblog.Red.Println("[ERROR] main froze")
		panic("Vaktbikkje sier voff! - main froze. Resets program.")
	})

	// Variables
	elevatorRole := scout.Slave
	masterIP := scout.LoopbackIP // Default is loopback address

	// Channels
	fromMSEChannel := make(chan scout.FromMSE)
	masterAddressChannel := make(chan string)
	masterQuitChannel := make(chan struct{})

	go scout.Start(fromMSEChannel)
	go slave.Start(masterAddressChannel)

	lastRole := scout.Slave
	for {
		select {
		case mseData := <-fromMSEChannel:
			// Data recieved from Master Slave Election
			elevatorRole = mseData.ElevatorRole
			masterIP = mseData.MasterIP
			rblog.Cyan.Printf("Elevator role: %s | Master IP: %s", elevatorRole, masterIP)

			if elevatorRole == scout.Master && lastRole == scout.Slave {
				// Start master protocol
				rblog.Rainbow.Println("Promoted to Master")
				go master.Start(masterPort, masterQuitChannel)
			}
			if elevatorRole == scout.Slave && lastRole == scout.Master {
				// Stop master protocol
				rblog.Cyan.Println("Demoted to Slave😥")
				masterQuitChannel <- struct{}{}
			}

			lastRole = elevatorRole

			// Update master IP-address
			masterAddressChannel <- fmt.Sprintf("%s:%d", masterIP, masterPort)
		case <-time.After(watchdogResetPeriod):
			//unblock select to reset watchdog

		}
		watchdog.Reset(watchdogTimeoutPeriod)
	}
}
