// Based on https://github.com/TTK4145/Project-resources/tree/master/elev_algo
package requests

import (
	"project/mscomm"
	"project/rblog"
	"project/slave/cabfile"
	"project/slave/elevator"
	"project/slave/elevio"
	"project/slave/iodevice"
	"time"
)

type DirnBehaviourPair struct {
	Dirn      elevio.MotorDirection
	Behaviour elevator.ElevatorBehaviour
}

func requestsAbove(e elevator.Elevator) bool {
	if e.Floor == -1 {
		return false
	}
	for f := e.Floor + 1; f < iodevice.N_FLOORS; f++ {
		for btn := 0; btn < iodevice.N_BUTTONS; btn++ {
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
		for btn := 0; btn < iodevice.N_BUTTONS; btn++ {
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
	for btn := 0; btn < iodevice.N_BUTTONS; btn++ {
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
	switch e.Config.ClearRequestVariant {
	case elevator.CV_All:
		return e.Floor == btn_floor
	case elevator.CV_InDirn:
		return e.Floor == btn_floor && (e.Dirn == elevio.MD_Up && btn_type == elevio.BT_HallUp || e.Dirn == elevio.MD_Down && btn_type == elevio.BT_HallDown || e.Dirn == elevio.MD_Stop || btn_type == elevio.BT_Cab)
	default:
		return false
	}
}

func ClearAtCurrentFloor(e elevator.Elevator, clearRequestCh chan<- interface{}) elevator.Elevator {

	switch e.Config.ClearRequestVariant {
	case elevator.CV_All:
		for btn := 0; btn < iodevice.N_BUTTONS; btn++ {
			e = Clear(e, e.Floor, elevio.ButtonType(btn), clearRequestCh)
		}

	case elevator.CV_InDirn:
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

	default:

	}
	return e
}

func Clear(e elevator.Elevator, floor int, btnType elevio.ButtonType, clearRequestCh chan<- interface{}) elevator.Elevator {
	if btnType == elevio.BT_Cab {
		return e
	}
	e.Requests[floor][btnType] = 0
	e.HallLights[floor][btnType] = false
	select {
	case clearRequestCh <- mscomm.OrderComplete{Floor: floor, Button: int(btnType)}:
	case <-time.After(10 * time.Millisecond):
		rblog.Yellow.Println("Sending ordercomplete timed out")
	}
	return e
}
