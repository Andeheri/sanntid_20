package slave

import (
	"fmt"
	"net"
	"time"
	"project/mscomm"
	"project/slave/elevio"
	"project/slave/fsm"
	"project/slave/mastercom"
)

func Start(initialMasterAddress string, masterAddress <-chan string) {
	numFloors := 4

	elevio.Init("localhost:15657", numFloors)

	drvButtons := make(chan elevio.ButtonEvent)
	drvFloors := make(chan int)
	drvObstr := make(chan bool)

	go elevio.PollButtons(drvButtons)
	go elevio.PollFloorSensor(drvFloors)
	go elevio.PollObstructionSwitch(drvObstr)

	doorTimer := time.NewTimer(-1)

	buttonPress := make(chan elevio.ButtonEvent)
	clearRequest := make(chan elevio.ButtonEvent)
	assignedRequests := make(chan mscomm.AssignedRequests)
	requestedState := make(chan struct{})
	sender := make(chan interface{})
	hallLights := make(chan mscomm.Lights)

	masterChans := mastercom.MasterChannels{
		ButtonPress:    buttonPress,
		ClearRequest:   clearRequest,
		AssignedRequests: assignedRequests,
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

	go mastercom.MasterCommunication(TCPAddr, &masterChans, stopMaster)


	watchDogTime := 3*time.Second
	watchDog := time.AfterFunc(watchDogTime, func() {
		elevio.SetMotorDirection(elevio.MD_Stop)
		panic("Watchdog timeout on slave")
	})
	defer watchDog.Stop()

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
			sender = make(chan interface{})
			masterChans.Sender = sender
			go mastercom.MasterCommunication(TCPAddr, &masterChans, stopMaster)

		//moved these from mastercom.go as they are involved with current state
		case a := <-masterChans.AssignedRequests:
			fmt.Println(a, "mottat master request melding")
			fsm.RequestsClearAll()
			fsm.RequestsSetAll(a, doorTimer, masterChans.ClearRequest)

		case a := <-masterChans.RequestedState:
			fmt.Println(a, "sender state til master")
			mastercom.SendState(sender)

		case a := <-masterChans.HallLights:
			fmt.Println(a, "mottat hall lights melding")
			fsm.Elev.HallLights = a
			fsm.SetAllLights(&fsm.Elev)

		case <- time.After(watchDogTime/10):
		}	
		watchDog.Reset(watchDogTime)
	}
}
