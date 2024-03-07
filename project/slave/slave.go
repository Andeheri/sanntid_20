package slave

import (
	"fmt"
	"log"
	"net"
	"project/mscomm"
	"project/slave/elevio"
	"project/slave/fsm"
	"project/slave/mastercom"
	"time"
)

func Start(masterAddress <-chan string) {
	numFloors := 4

	elevio.Init("localhost:15657", numFloors)

	drvButtons := make(chan elevio.ButtonEvent)
	drvFloors := make(chan int)
	drvObstr := make(chan bool)

	go elevio.PollButtons(drvButtons)
	go elevio.PollFloorSensor(drvFloors)
	go elevio.PollObstructionSwitch(drvObstr)

	doorTimer := time.NewTimer(-1) 
	                           
	sender := make(chan interface{})
	fromMasterCh := make(chan mscomm.Package)

	fsm.Init(doorTimer, sender)

	var masterConn *net.TCPConn

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
					log.Println("Timed out on sending button press")
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
			if masterConn != nil {
				masterConn.Close()
			}
			masterConn = mastercom.StartUp(a, sender, fromMasterCh)

		case a := <-fromMasterCh:
			fmt.Println(a, "mottat melding fra master")
			mastercom.HandleMessage(a.Payload, sender, doorTimer)

		case <- time.After(watchDogTime/5):
		}	

		watchDog.Reset(watchDogTime)
	}
}
