package main

import (
	"fmt"
	"slave/elevio"
	"slave/fsm"
	"slave/mastercom"
	// "slave/testclear"
	"time"
	"slave/iodevice"
)

func main() {
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
	master_requests := make(chan [iodevice.N_FLOORS][iodevice.N_BUTTONS] int)
	requestedState := make(chan bool)
	log := make(chan bool)
	sender := make(chan interface{})

	masterChans := mastercom.Master_channels{
		ButtonPress: button_press,
		ClearRequest: clear_request,
		MasterRequests: master_requests,
		RequestedState: requestedState,
		Log: log,
		Sender: sender,
	}

	// need tcp connection to master
	// go mastercom.Master_communication(&masterChans)

	fsm.Init()

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


		//moved these from mastercom.go as they are involved with current state
		case a := <- masterChans.MasterRequests:
			fmt.Println(a, "mottat master request melding")
			fsm.Requests_clearAll()
			fsm.Requests_setAll(a, doorTimer, masterChans.ClearRequest)
			
		case a := <- masterChans.RequestedState:
			fmt.Println(a, "sender state til master")
			mastercom.SendState(sender)
		}
	}
}
