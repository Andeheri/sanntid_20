package slave

import (
	"project/mscomm"
	"project/rblog"
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

	senderCh := make(chan interface{})
	fromMasterCh := make(chan mscomm.Package)
	go mastercom.ConnManager(masterAddressCh, senderCh, fromMasterCh)

	//TODO: set to something other than -1
	doorTimer := time.NewTimer(-1)
	inbetweenFloorsTimer := time.NewTimer(-1)
	inbetweenFloorsTimer.Stop()

	fsm.Init(doorTimer, inbetweenFloorsTimer, senderCh)

	watchDogTime := 3 * time.Second
	watchDog := time.AfterFunc(watchDogTime, func() {
		elevio.SetMotorDirection(elevio.MD_Stop)
		panic("Watchdog timeout on slave")
	})
	defer watchDog.Stop()

	rblog.White.Println("Slave started")

	for {
		select {
		case a := <-drvButtonsCh:
			if a.Button == elevio.BT_Cab {
				fsm.OnRequestButtonPress(a.Floor, a.Button, doorTimer, inbetweenFloorsTimer, senderCh)
				senderCh <- fsm.GetState()
			} else {
				pressed := mscomm.ButtonPressed{Floor: a.Floor, Button: int(a.Button)}
				senderCh <- pressed
			}

		case a := <-drvFloorsCh:
			fsm.OnFloorArrival(a, doorTimer, inbetweenFloorsTimer, senderCh)

		case a := <-drvObstrCh:
			rblog.Yellow.Printf("Obstruction: %+v\n", a)
			fsm.OnObstruction(a)
			senderCh <- fsm.GetState()

		case <-doorTimer.C:
			fsm.OnDoorTimeout(doorTimer, inbetweenFloorsTimer, senderCh)

		case <-inbetweenFloorsTimer.C:
			rblog.Red.Println("No floor arrival, setting Elev.Floor = -1")
			fsm.Elev.Floor = -1
			senderCh <- fsm.GetState()

		case a := <-fromMasterCh:
			mastercom.HandleMessage(a.Payload, senderCh, doorTimer, inbetweenFloorsTimer)

		case <-time.After(watchDogTime / 5):
		}

		watchDog.Reset(watchDogTime)
	}
}
