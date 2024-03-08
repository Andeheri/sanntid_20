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

func Start(masterAddressCh <-chan string) {
	numFloors := 4

	elevio.Init("localhost:15657", numFloors)

	drvButtonsCh := make(chan elevio.ButtonEvent)
	drvFloorsCh := make(chan int)
	drvObstrCh := make(chan bool)

	go elevio.PollButtons(drvButtonsCh)
	go elevio.PollFloorSensor(drvFloorsCh)
	go elevio.PollObstructionSwitch(drvObstrCh)

	doorTimer := time.NewTimer(-1) 
	                           
	senderCh := make(chan interface{})
	fromMasterCh := make(chan mscomm.Package)

	fsm.Init(doorTimer, senderCh)

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
		case a := <-drvButtonsCh:
			if a.Button == elevio.BT_Cab{
				fsm.OnRequestButtonPress(a.Floor, a.Button, doorTimer, senderCh)
			} else{
				pressed := mscomm.ButtonPressed{Floor: a.Floor, Button: int(a.Button)}
				select {
				case senderCh <- pressed:
				case <-time.After(10 * time.Millisecond):
					log.Println("Timed out on sending button press")
				}
			}

		case a := <-drvFloorsCh:
			fsm.OnFloorArrival(a, doorTimer, senderCh)

		case a := <-drvObstrCh:
			fmt.Printf("Obstruction: %+v\n", a)
			fsm.OnObstruction(a)

		case <-doorTimer.C:
			fsm.OnDoorTimeout(doorTimer, senderCh)

		case a := <-masterAddressCh:
			fmt.Println("mottat ny master addresse:", a)
			if masterConn != nil {
				masterConn.Close()
			}
			masterConn = mastercom.StartUp(a, senderCh, fromMasterCh)

		case a := <-fromMasterCh:
			mastercom.HandleMessage(a.Payload, senderCh, doorTimer)

		case <- time.After(watchDogTime/5):
		}	

		watchDog.Reset(watchDogTime)
	}
}
