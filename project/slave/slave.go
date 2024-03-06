package slave

import (
	"fmt"
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
	state := make(chan mscomm.ElevatorState)
	fromMasterCh := make(chan mscomm.Package)

	masterChans := mastercom.MasterChannels{
		ButtonPress:    buttonPress,
		ClearRequest:   clearRequest,
		AssignedRequests: assignedRequests,
		RequestedState: requestedState,
		Sender:         sender,
		HallLights:   hallLights,
		State: state,
		FromMasterCh: fromMasterCh,
	}

	fsm.Init(doorTimer, sender)

	masterConn := mastercom.StartUp(initialMasterAddress, &masterChans)

	watchDogTime := 3*time.Second
	watchDog := time.AfterFunc(watchDogTime, func() {
		elevio.SetMotorDirection(elevio.MD_Stop)
		panic("Watchdog timeout on slave")
	})
	defer watchDog.Stop()
	fmt.Println("slave startet")
	for {
		select {
		case a := <-drvButtons:
			fmt.Printf("Buttons: %+v\n", a)
			if a.Button == elevio.BT_Cab{
				fsm.OnRequestButtonPress(a.Floor, a.Button, doorTimer, sender)
			} else{
				fmt.Println(a, "sender button press melding")
				pressed := mscomm.ButtonPressed{Floor: a.Floor, Button: int(a.Button)}
				select {
				case sender <- pressed:
				case <-time.After(10 * time.Millisecond):
				}
			}

		case a := <-drvFloors:
			fmt.Printf("Floor: %+v\n", a)
			fsm.OnFloorArrival(a, doorTimer, sender)

		case a := <-drvObstr:
			fmt.Printf("Obstruction: %+v\n", a)
			fsm.OnObstruction(a)

		case a := <-doorTimer.C:
			fmt.Printf("Doortimer: %+v\n", a)
			fsm.OnDoorTimeout(doorTimer, sender)

		case a := <-masterAddress:
			fmt.Println("mottat ny master addresse:", a)
			masterConn.Close()
		
			mastercom.StartUp(a, &masterChans)

		case a := <-fromMasterCh:
			fmt.Println(a, "mottat melding fra master")
			mastercom.HandleMessage(a.Payload, &masterChans, doorTimer)

		case a := <-clearRequest:
			fmt.Println(a, "sender clear request melding")
			completedOrder := mscomm.OrderComplete{Floor: a.Floor, Button: int(a.Button)}
			select {
			case sender <- completedOrder:
			case <-time.After(10 * time.Millisecond):
			}

		case <- time.After(watchDogTime/5):
		}	

		watchDog.Reset(watchDogTime)
	}
}
