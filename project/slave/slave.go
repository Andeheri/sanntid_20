package slave

import (
	"project/mscomm"
	"project/rblog"
	"project/slave/elevio"
	"project/slave/fsm"
	"project/slave/iodevice"
	"project/slave/mastercom"
	"time"
)

//To be run as a goroutine
func Start(masterAddressCh <-chan string) {
	elevio.Init("localhost:15657", iodevice.N_FLOORS)

	drvButtonsCh := make(chan elevio.ButtonEvent)
	drvFloorsCh := make(chan int)
	drvObstrCh := make(chan bool)

	go elevio.PollButtons(drvButtonsCh)
	go elevio.PollFloorSensor(drvFloorsCh)
	go elevio.PollObstructionSwitch(drvObstrCh)

	senderCh := make(chan interface{})
	fromMasterCh := make(chan mscomm.Package)
	go mastercom.ConnManager(masterAddressCh, senderCh, fromMasterCh)

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

	rblog.White.Println("slave startet")

	for {
		select {
		case a := <-drvButtonsCh:
			if a.Button == elevio.BT_Cab {
				fsm.OnNewRequest(a.Floor, a.Button, doorTimer, inbetweenFloorsTimer, senderCh)
				select {
				case senderCh <- fsm.GetState():
				case <-time.After(10 * time.Millisecond):
					rblog.Yellow.Println("Timed out on sending state after cab button press")
				}
			} else {
				pressed := mscomm.ButtonPressed{Floor: a.Floor, Button: int(a.Button)}
				select {
				case senderCh <- pressed:
				case <-time.After(10 * time.Millisecond):
					rblog.Yellow.Println("Timed out on sending button press")
				}
			}

		case a := <-drvFloorsCh:
			fsm.OnFloorArrival(a, doorTimer, inbetweenFloorsTimer, senderCh)

		case a := <-drvObstrCh:
			rblog.Yellow.Printf("Obstruction: %+v\n", a)
			fsm.OnObstruction(a)
			select {
			case senderCh <- fsm.GetState():
			case <-time.After(10 * time.Millisecond):
				rblog.Yellow.Println("Timed out on sending state after change in obstruction:", a)
			}

		case <-doorTimer.C:
			fsm.OnDoorTimeout(doorTimer, inbetweenFloorsTimer, senderCh)

		case <-inbetweenFloorsTimer.C:
			rblog.Red.Println("No floor arrival, setting Elev.Floor = -1")
			fsm.Elev.Floor = -1
			select {
			case senderCh <- fsm.GetState():
			case <-time.After(10 * time.Millisecond):
				rblog.Yellow.Println("Timed out on sending state after inbetweenFloorsTimer timeout")
			}

		case a := <-fromMasterCh:
			mastercom.HandleMessage(a.Payload, senderCh, doorTimer, inbetweenFloorsTimer)

		case <-time.After(watchDogTime / 5):
		}

		watchDog.Reset(watchDogTime)
	}
}
