package main

import (
	"fmt"
	"slave/elevio"
	"slave/fsm"
	"slave/mastercom"
	"slave/testclear"
	"time"
	"slave/iodevice"
)

func main() {
	numFloors := 4
	var master_req = [4][3]int{{0,0,1},{1,1,0},{1,0,1},{0,0,1}}

	elevio.Init("localhost:15657", numFloors)

	drv_buttons := make(chan elevio.ButtonEvent)
	drv_floors := make(chan int)
	drv_obstr := make(chan bool)

	door_timer := time.NewTimer(-1)
	master_test := make(chan bool)

	go elevio.PollButtons(drv_buttons)
	go elevio.PollFloorSensor(drv_floors)
	go elevio.PollObstructionSwitch(drv_obstr)
	go testclear.Set_master_test(master_test)


	button_press := make(chan elevio.ButtonEvent)
	clear_request := make(chan elevio.ButtonEvent)
	master_requests := make(chan [iodevice.N_FLOORS][iodevice.N_BUTTONS] int)

	master_chans := mastercom.Master_channels{
		Button_press: button_press,
		Clear_request: clear_request,
		Master_requests: master_requests,
	}

	go mastercom.Master_communication(master_chans, door_timer)

	fsm.Fsm_init()

	for {
		select {
		case a := <-drv_buttons:
			fmt.Printf("%+v\n", a)
			fsm.Fsm_onRequestButtonPress(a.Floor, a.Button, door_timer)

		case a := <-drv_floors:
			fmt.Printf("%+v\n", a)
			fsm.Fsm_onFloorArrival(a, door_timer)

		case a := <-drv_obstr:
			fmt.Printf("%+v\n", a)
			fsm.OnObstruction(a)

		case a := <-door_timer.C:
			fmt.Printf("%+v\n", a)
			fsm.Fsm_onDoorTimeout(door_timer)

		//test for clearing and setting new requests from master
		case a := <-master_test:
			fmt.Printf("%+v\n", a)
			fsm.Requests_clearAll()
			fsm.Requests_setAll(master_req, door_timer)
		}
	}
}
