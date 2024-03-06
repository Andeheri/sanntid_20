package slave

import (
	"fmt"
	"net"
	"time"
	"project/commontypes"
	"project/slave/elevio"
	"project/slave/fsm"
	"project/slave/mastercom"
	"project/slave/watchdog"
)

func Start(initialMasterAddress string, masterAddress <-chan string, quit chan struct{}) {
	numFloors := 4

	elevio.Init("localhost:15657", numFloors)

	drvButtons := make(chan elevio.ButtonEvent)
	drvFloors := make(chan int)
	drvObstr := make(chan bool)

	doorTimer := time.NewTimer(-1)

	go elevio.PollButtons(drvButtons)
	go elevio.PollFloorSensor(drvFloors)
	go elevio.PollObstructionSwitch(drvObstr)

	buttonPress := make(chan elevio.ButtonEvent)
	clearRequest := make(chan elevio.ButtonEvent)
	masterRequests := make(chan commontypes.AssignedRequests)
	requestedState := make(chan struct{})
	sender := make(chan interface{})
	hallLights := make(chan commontypes.Lights)

	masterChans := mastercom.MasterChannels{
		ButtonPress:    buttonPress,
		ClearRequest:   clearRequest,
		MasterRequests: masterRequests,
		RequestedState: requestedState,
		Sender:         sender,
		HallLights:   hallLights,
	}

	fsm.Init(doorTimer, masterChans.ClearRequest)

	stopMaster := make(chan struct{})
	TCPAddr, err := net.ResolveTCPAddr("tcp", initialMasterAddress)
	if err != nil {
		fmt.Println("Error resolving TCP address from master:", err)
	}
	quitSlave := make(chan bool)
	go mastercom.MasterCommunication(TCPAddr, &masterChans, stopMaster, quitSlave)
	a := <- quitSlave
	fmt.Println(a, "failed connection to master")
	if a {
		quit <- struct{}{}
		return
	}

	toWatchDog := make(chan struct{})
	fromWatchDog := make(chan struct{})
	go watchdog.Start(1*time.Second, toWatchDog, fromWatchDog)

	for {
		select {
		case a := <-drvButtons:
			fmt.Printf("%+v\n", a)
			if a.Button == elevio.BT_Cab{
				fsm.OnRequestButtonPress(a.Floor, a.Button, doorTimer, masterChans.ClearRequest)
			} else {
				masterChans.ButtonPress <- a
			}
			
		case a := <-drvFloors:
			fmt.Printf("%+v\n", a)
			fsm.OnFloorArrival(a, doorTimer, masterChans.ClearRequest)

		case a := <-drvObstr:
			fmt.Printf("%+v\n", a)
			fsm.OnObstruction(a)

		case a := <-doorTimer.C:
			fmt.Printf("%+v\n", a)
			fsm.OnDoorTimeout(doorTimer, masterChans.ClearRequest)

		case a := <-masterAddress:
			stopMaster <- struct{}{}
			fmt.Println("mottat master addresse:", a)
			TCPAddr, err := net.ResolveTCPAddr("tcp", a)
			if err != nil {
				fmt.Println("Error resolving TCP address from master:", err)
			}
			go mastercom.MasterCommunication(TCPAddr, &masterChans, stopMaster, quitSlave)

		//moved these from mastercom.go as they are involved with current state
		case a := <-masterChans.MasterRequests:
			fmt.Println(a, "mottat master request melding")
			fsm.RequestsClearAll()
			fsm.RequestsSetAll(a, doorTimer, masterChans.ClearRequest)

		case a := <-masterChans.RequestedState:
			fmt.Println(a, "sender state til master")
			mastercom.SendState(sender)

		case a := <-masterChans.HallLights:
			fmt.Println(a, "mottat all lights melding")
			fsm.Elev.HallLights = a
			fsm.SetAllLights(fsm.Elev)
			
		case a := <-quitSlave:
			fmt.Println(a, "quit slave melding")
			quit <- struct{}{}
			return

		case <-fromWatchDog:
			toWatchDog <- struct{}{}
		}	
	}
}
