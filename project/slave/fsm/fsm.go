package fsm

import (
	"project/mscomm"
	"project/slave/cabfile"
	"project/slave/elevator"
	"project/slave/elevio"
	"project/slave/requests"
	"time"
)

var Elev elevator.Elevator = elevator.Initialize()

func Init(doorTimer *time.Timer, inbetweenFloorsTimer *time.Timer, clearRequestCh chan<- interface{}) {

	elevio.SetMotorDirection(elevio.MD_Stop)
	Elev.Dirn = elevio.MD_Stop
	Elev.Behaviour = elevator.EB_Idle

	floor := elevio.GetFloor()
	Elev.Floor = floor
	// Code for fixing starting position between floors
	if floor == -1 {
		elevio.SetMotorDirection(elevio.MD_Down)
		Elev.Dirn = elevio.MD_Down
		Elev.Behaviour = elevator.EB_Moving
	}

	cabRequests := cabfile.Read()
	for floor := 0; floor < elevator.N_FLOORS; floor++ {
		if cabRequests[floor] != 0 {
			OnRequestOrder(floor, elevio.BT_Cab, doorTimer, inbetweenFloorsTimer, clearRequestCh)
		}
	}
}

func SetAllLights(es *elevator.Elevator) {
	for floor := 0; floor < elevator.N_FLOORS; floor++ {
		for btn := elevio.ButtonType(0); btn < 2; btn++ {
			elevio.SetButtonLamp(btn, floor, es.HallLights[floor][btn])
		}
	}
	for floor := 0; floor < elevator.N_FLOORS; floor++ {
		elevio.SetButtonLamp(elevio.BT_Cab, floor, es.Requests[floor][elevio.BT_Cab] != 0)
	}
}

func OnRequestOrder(btnFloor int, btnType elevio.ButtonType, doorTimer *time.Timer, inbetweenFloorsTimer *time.Timer, clearRequestCh chan<- interface{}) {

	switch Elev.Behaviour {
	case elevator.EB_DoorOpen:
		if requests.ShouldClearImmediately(Elev, btnFloor, btnType) {
			Elev = requests.Clear(Elev, btnFloor, btnType, clearRequestCh)
			doorTimer.Reset(Elev.Config.DoorOpenDuration)

		} else {
			if btnType == elevio.BT_Cab {
				setCabFile(btnFloor)
			}
			Elev.Requests[btnFloor][btnType] = 1
		}

	case elevator.EB_Moving:
		if btnType == elevio.BT_Cab {
			setCabFile(btnFloor)
		}
		Elev.Requests[btnFloor][btnType] = 1

	case elevator.EB_Idle:
		if btnType == elevio.BT_Cab {
			setCabFile(btnFloor)
		}
		Elev.Requests[btnFloor][btnType] = 1

		pair := requests.ChooseDirection(Elev)
		Elev.Dirn = pair.Dirn
		Elev.Behaviour = pair.Behaviour
		switch pair.Behaviour {
		case elevator.EB_DoorOpen:
			elevio.SetDoorOpenLamp(true)
			doorTimer.Reset(Elev.Config.DoorOpenDuration)
			Elev = requests.ClearAtCurrentFloor(Elev, clearRequestCh)
		case elevator.EB_Moving:
			elevio.SetMotorDirection(Elev.Dirn)
			if Elev.Dirn != elevio.MD_Stop {
				inbetweenFloorsTimer.Reset(Elev.Config.InbetweenFloorsDuration)
			}
		case elevator.EB_Idle:

		}

	}
	SetAllLights(&Elev)
}

func OnFloorArrival(newFloor int, doorTimer *time.Timer, inbetweenFloorsTimer *time.Timer, clearRequestCh chan<- interface{}) {
	Elev.Floor = newFloor

	elevio.SetFloorIndicator(Elev.Floor)

	switch Elev.Behaviour {
	case elevator.EB_Moving:
		if requests.ShouldStop(Elev) {
			elevio.SetMotorDirection(elevio.MD_Stop)
			elevio.SetDoorOpenLamp(true)
			Elev = requests.ClearAtCurrentFloor(Elev, clearRequestCh)
			doorTimer.Reset(Elev.Config.DoorOpenDuration)
			SetAllLights(&Elev)
			Elev.Behaviour = elevator.EB_DoorOpen
			inbetweenFloorsTimer.Stop()
		}
	default:
	}
}

func OnDoorTimeout(doorTimer *time.Timer, inbetweenFloorsTimer *time.Timer, clearRequestCh chan<- interface{}) {

	if Elev.Behaviour == elevator.EB_DoorOpen {

		if Elev.Obstructed {
			doorTimer.Reset(Elev.Config.DoorOpenDuration)
			return
		}
		pair := requests.ChooseDirection(Elev)
		Elev.Dirn = pair.Dirn
		Elev.Behaviour = pair.Behaviour

		switch Elev.Behaviour {
		case elevator.EB_DoorOpen:
			doorTimer.Reset(Elev.Config.DoorOpenDuration)
			Elev = requests.ClearAtCurrentFloor(Elev, clearRequestCh)
			SetAllLights(&Elev)

		case elevator.EB_Moving:
			elevio.SetDoorOpenLamp(false)
			elevio.SetMotorDirection(elevio.MotorDirection(Elev.Dirn))
			if Elev.Dirn != elevio.MD_Stop {
				inbetweenFloorsTimer.Reset(Elev.Config.InbetweenFloorsDuration)
			}

		case elevator.EB_Idle:
			elevio.SetDoorOpenLamp(false)
			elevio.SetMotorDirection(elevio.MotorDirection(Elev.Dirn))
			if Elev.Dirn != elevio.MD_Stop {
				inbetweenFloorsTimer.Reset(Elev.Config.InbetweenFloorsDuration)
			}

		}
	}
}

func OnObstruction(isObstructed bool) {
	Elev.Obstructed = isObstructed
}

func RequestsClearAll() {
	for btn := 0; btn < 2; btn++ {
		for floor := 0; floor < elevator.N_FLOORS; floor++ {
			Elev.Requests[floor][btn] = 0
		}
	}
}

func RequestsSetAll(masterRequests mscomm.AssignedRequests, doorTimer *time.Timer, inbetweenFloorsTimer *time.Timer, clearRequestCh chan<- interface{}) {
	for btn := 0; btn < 2; btn++ {
		for floor := 0; floor < elevator.N_FLOORS; floor++ {
			if masterRequests[floor][btn] {
				OnRequestOrder(floor, elevio.ButtonType(btn), doorTimer, inbetweenFloorsTimer, clearRequestCh)
			}
		}
	}
}

func getCabRequests() []bool {
	cabRequests := make([]bool, elevator.N_FLOORS)
	var btn elevio.ButtonType = elevio.BT_Cab
	for floor := 0; floor < elevator.N_FLOORS; floor++ {
		if Elev.Requests[floor][btn] == 1 {
			cabRequests[floor] = true
		} else {
			cabRequests[floor] = false
		}
	}
	return cabRequests
}

func GetState() mscomm.ElevatorState {
	behavior := string(Elev.Behaviour)
	//should not be part of hall assignment when blocked
	if Elev.Obstructed || Elev.Floor == -1 {
		behavior = "blocked"
	}
	state := mscomm.ElevatorState{
		Behavior:    behavior,
		Floor:       Elev.Floor,
		Direction:   elevio.Elevio_dirn_toString(Elev.Dirn),
		CabRequests: getCabRequests(),
	}
	return state
}

func setCabFile(btnFloor int) {
	err := cabfile.Set(btnFloor)
	if err != nil {
		err = cabfile.Set(btnFloor)
	}
	if err != nil {
		elevio.SetMotorDirection(elevio.MD_Stop)
		panic("Cab data could not be set to file")
	}
}
