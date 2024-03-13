package requests

import (
	"project/mscomm"
	"project/slave/cabfile"
	"project/slave/elevator"
	"project/slave/elevio"
)

type DirnBehaviourPair struct {
	Dirn      elevio.MotorDirection
	Behaviour elevator.ElevatorBehaviour
}

func requestsAbove(e elevator.Elevator) bool {
	if e.Floor == -1 {
		return false
	}
	for f := e.Floor + 1; f < elevator.N_FLOORS; f++ {
		for btn := 0; btn < elevator.N_BUTTONS; btn++ {
			if e.Requests[f][btn] != 0 {
				return true
			}
		}
	}
	return false
}

func requestsBelow(e elevator.Elevator) bool {
	if e.Floor == -1 {
		return false
	}
	for f := 0; f < e.Floor; f++ {
		for btn := 0; btn < elevator.N_BUTTONS; btn++ {
			if e.Requests[f][btn] != 0 {
				return true
			}
		}
	}
	return false
}

func requestsHere(e elevator.Elevator) bool {
	if e.Floor == -1 {
		return false
	}
	for btn := 0; btn < elevator.N_BUTTONS; btn++ {
		if e.Requests[e.Floor][btn] != 0 {
			return true
		}
	}
	return false
}

func ChooseDirection(e elevator.Elevator) DirnBehaviourPair {
	switch e.Dirn {
	case elevio.MD_Up:
		if requestsAbove(e) {
			return DirnBehaviourPair{elevio.MD_Up, elevator.EB_Moving}
		} else if requestsHere(e) {
			return DirnBehaviourPair{elevio.MD_Down, elevator.EB_DoorOpen}
		} else if requestsBelow(e) {
			return DirnBehaviourPair{elevio.MD_Down, elevator.EB_Moving}
		} else {
			return DirnBehaviourPair{elevio.MD_Stop, elevator.EB_Idle}
		}

	case elevio.MD_Down:
		if requestsBelow(e) {
			return DirnBehaviourPair{elevio.MD_Down, elevator.EB_Moving}
		} else if requestsHere(e) {
			return DirnBehaviourPair{elevio.MD_Up, elevator.EB_DoorOpen}
		} else if requestsAbove(e) {
			return DirnBehaviourPair{elevio.MD_Up, elevator.EB_Moving}
		} else {
			return DirnBehaviourPair{elevio.MD_Stop, elevator.EB_Idle}
		}

	case elevio.MD_Stop:
		if requestsHere(e) {
			return DirnBehaviourPair{elevio.MD_Stop, elevator.EB_DoorOpen}
		} else if requestsAbove(e) {
			return DirnBehaviourPair{elevio.MD_Up, elevator.EB_Moving}
		} else if requestsBelow(e) {
			return DirnBehaviourPair{elevio.MD_Down, elevator.EB_Moving}
		} else {
			return DirnBehaviourPair{elevio.MD_Stop, elevator.EB_Idle}
		}

	default:
		return DirnBehaviourPair{elevio.MD_Stop, elevator.EB_Idle}
	}
}

func ShouldStop(e elevator.Elevator) bool {
	switch e.Dirn {
	case elevio.MD_Down:
		return e.Requests[e.Floor][elevio.BT_HallDown] != 0 || e.Requests[e.Floor][elevio.BT_Cab] != 0 || !requestsBelow(e)
	case elevio.MD_Up:
		return e.Requests[e.Floor][elevio.BT_HallUp] != 0 || e.Requests[e.Floor][elevio.BT_Cab] != 0 || !requestsAbove(e)
	case elevio.MD_Stop:
	default:
		return true
	}
	return true
}

func ShouldClearImmediately(e elevator.Elevator, btn_floor int, btn_type elevio.ButtonType) bool {
	return e.Floor == btn_floor && (e.Dirn == elevio.MD_Up && btn_type == elevio.BT_HallUp || e.Dirn == elevio.MD_Down && btn_type == elevio.BT_HallDown || e.Dirn == elevio.MD_Stop || btn_type == elevio.BT_Cab)
}

func ClearAtCurrentFloor(e elevator.Elevator, clearRequestCh chan<- interface{}) elevator.Elevator {

	e.Requests[e.Floor][elevio.BT_Cab] = 0
	cabfile.Clear(e.Floor)

	switch e.Dirn {
	case elevio.MD_Up:
		if !requestsAbove(e) && e.Requests[e.Floor][elevio.BT_HallUp] == 0 {
			e = Clear(e, e.Floor, elevio.BT_HallDown, clearRequestCh)
		}
		e = Clear(e, e.Floor, elevio.BT_HallUp, clearRequestCh)

	case elevio.MD_Down:
		if !requestsBelow(e) && e.Requests[e.Floor][elevio.BT_HallDown] == 0 {
			e = Clear(e, e.Floor, elevio.BT_HallUp, clearRequestCh)
		}
		e = Clear(e, e.Floor, elevio.BT_HallDown, clearRequestCh)

	case elevio.MD_Stop:
		e = Clear(e, e.Floor, elevio.BT_HallUp, clearRequestCh)
		e = Clear(e, e.Floor, elevio.BT_HallDown, clearRequestCh)
	default:
		e = Clear(e, e.Floor, elevio.BT_HallUp, clearRequestCh)
		e = Clear(e, e.Floor, elevio.BT_HallDown, clearRequestCh)
	}

	return e
}

func Clear(e elevator.Elevator, floor int, btnType elevio.ButtonType, clearRequestCh chan<- interface{}) elevator.Elevator {
	if btnType == elevio.BT_Cab {
		return e
	}
	e.Requests[floor][btnType] = 0
	e.HallLights[floor][btnType] = false
	clearRequestCh <- mscomm.OrderComplete{Floor: floor, Button: int(btnType)}
	return e
}
