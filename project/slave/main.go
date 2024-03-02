package slave

import (
	"fmt"
	"net"
	"slave/elevio"
	"slave/fsm"
	"slave/iodevice"
	"slave/mastercom"
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
	master_requests := make(chan [iodevice.N_FLOORS][iodevice.N_BUTTONS]int)
	requestedState := make(chan bool)
	log := make(chan bool)
	sender := make(chan interface{})

	masterChans := mastercom.Master_channels{
		ButtonPress:    button_press,
		ClearRequest:   clear_request,
		MasterRequests: master_requests,
		RequestedState: requestedState,
		Log:            log,
		Sender:         sender,
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
			// fsm.Fsm_onRequestButtonPress(a.Floor, a.Button, door_timer)
			masterChans.ButtonPress <- a

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
			fsm.Requests_clearAll()
			fsm.Requests_setAll(a, doorTimer, masterChans.ClearRequest)

		case a := <-masterChans.RequestedState:
			fmt.Println(a, "sender state til master")
			mastercom.SendState(sender)
		}
	}
}
