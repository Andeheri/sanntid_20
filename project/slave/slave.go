package slave

import (
	"project/mscomm"
	"project/rblog"
	"project/slave/elevator"
	"project/slave/elevio"
	"project/slave/fsm"
	"project/slave/mastercom"
	"time"
)

func Start(masterAddressCh <-chan string) {
	elevio.Init("localhost:15657", elevator.N_FLOORS)

	drvButtonsCh := make(chan elevio.ButtonEvent)
	drvFloorsCh := make(chan int)
	drvObstrCh := make(chan bool)

	go elevio.PollButtons(drvButtonsCh)
	go elevio.PollFloorSensor(drvFloorsCh)
	go elevio.PollObstructionSwitch(drvObstrCh)

	senderCh := make(chan interface{}, 1)
	fromMasterCh := make(chan mscomm.Package)
	go mastercom.ConnManager(masterAddressCh, senderCh, fromMasterCh)

	doorTimer := time.NewTimer(fsm.Elev.Config.DoorOpenDuration)
	inbetweenFloorsTimer := time.NewTimer(fsm.Elev.Config.InbetweenFloorsDuration)
	inbetweenFloorsTimer.Stop()
	fsm.Init(doorTimer, inbetweenFloorsTimer, senderCh)

	const watchDogTimeout = 500 * time.Millisecond
	watchDog := time.AfterFunc(watchDogTimeout, func() {
		elevio.SetMotorDirection(elevio.MD_Stop)
		panic("Watchdog timeout on slave")
	})

	rblog.White.Println("Slave started")

	for {
		select {
		case a := <-drvButtonsCh:
			rblog.Green.Printf("Buttons: %+v\n", a)

			if a.Button == elevio.BT_Cab {
				fsm.OnNewRequest(a.Floor, a.Button, doorTimer, inbetweenFloorsTimer, senderCh)
				mastercom.Send(senderCh, fsm.GetState())
			} else {
				pressed := mscomm.ButtonPressed{Floor: a.Floor, Button: int(a.Button)}
				mastercom.Send(senderCh, pressed)
			}

		case a := <-drvFloorsCh:
			rblog.Green.Printf("Floor: %+v\n", a)

			fsm.OnFloorArrival(a, doorTimer, inbetweenFloorsTimer, senderCh)

		case a := <-drvObstrCh:
			rblog.Yellow.Printf("Obstruction: %+v\n", a)
			fsm.OnObstruction(a)
			mastercom.Send(senderCh, fsm.GetState())

		case <-doorTimer.C:
			rblog.Green.Println("Door timeout")

			fsm.OnDoorTimeout(doorTimer, inbetweenFloorsTimer, senderCh)

		case <-inbetweenFloorsTimer.C:
			rblog.Red.Println("No floor arrival, setting Elev.Floor = -1")
			fsm.Elev.Floor = -1
			mastercom.Send(senderCh, fsm.GetState())

		case a := <-fromMasterCh:
			rblog.Green.Printf("From master: %+v\n", a)

			mastercom.HandleMessage(a.Payload, senderCh, doorTimer, inbetweenFloorsTimer)

		case <-time.After(watchDogTimeout / 3):
		}
		watchDog.Reset(watchDogTimeout)
	}
}
