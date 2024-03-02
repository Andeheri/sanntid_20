package slave

import (
	"fmt"
	"net"
	"project/commontypes"
	"project/slave/elevio"
	"project/slave/fsm"
	"project/slave/iodevice"
	"project/slave/mastercom"
	"time"
)

func start(initalMasterAddress string, masterAddress <-chan string) {
	numFloors := 4

	elevio.Init("localhost:15657", numFloors)

	drvButtons := make(chan elevio.ButtonEvent)
	drvFloors := make(chan int)
	drvObstr := make(chan bool)

	doorTimer := time.NewTimer(-1)

	go elevio.PollButtons(drvButtons)
	go elevio.PollFloorSensor(drvFloors)
	go elevio.PollObstructionSwitch(drvObstr)

	button_press := make(chan elevio.ButtonEvent)
	clear_request := make(chan elevio.ButtonEvent)
	master_requests := make(chan commontypes.AssignedRequests)
	requestedState := make(chan bool)
	sender := make(chan interface{})
	allLights := make(chan [iodevice.N_FLOORS][iodevice.N_BUTTONS]int)

	masterChans := mastercom.Master_channels{
		ButtonPress:    button_press,
		ClearRequest:   clear_request,
		MasterRequests: master_requests,
		RequestedState: requestedState,
		Sender:         sender,
		AllLights:   allLights,
	}

	fsm.Init()

	stopMaster := make(chan bool)
	TCPAddr, err := net.ResolveTCPAddr("tcp", initalMasterAddress)
	if err != nil {
		fmt.Println("Error resolving TCP address from master:", err)
	}
	go mastercom.Master_communication(TCPAddr, &masterChans, stopMaster)

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
			stopMaster <- true
			fmt.Println("mottat master addresse:", a)
			TCPAddr, err := net.ResolveTCPAddr("tcp", a)
			if err != nil {
				fmt.Println("Error resolving TCP address from master:", err)
			}
			go mastercom.Master_communication(TCPAddr, &masterChans, stopMaster)

		//moved these from mastercom.go as they are involved with current state
		case a := <-masterChans.MasterRequests:
			fmt.Println(a, "mottat master request melding")
			fsm.RequestsClearAll()
			fsm.RequestsSetAll(a, doorTimer, masterChans.ClearRequest)

		case a := <-masterChans.RequestedState:
			fmt.Println(a, "sender state til master")
			mastercom.SendState(sender)

		case a := <-masterChans.AllLights:
			fmt.Println(a, "mottat all lights melding")
			fsm.Elev.HallLights = a
			fsm.SetAllLights(fsm.Elev)
		}
	}
}
