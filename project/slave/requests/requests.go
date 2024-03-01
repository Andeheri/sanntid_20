package requests

import (
	"slave/elevator"
	"slave/elevio"
	"slave/iodevice"
)

type DirnBehaviourPair struct{
	Dirn elevio.MotorDirection
	Behaviour elevator.ElevatorBehaviour
}


func above(e elevator.Elevator) bool{
    for f := e.Floor+1; f < iodevice.N_FLOORS; f++ {
        for btn := 0; btn < iodevice.N_BUTTONS; btn++{
            if(e.Requests[f][btn] != 0){
                return true
            }
        }
    }
    return false
}

func below(e elevator.Elevator) bool{
    for f := 0; f < e.Floor; f++ {
        for btn := 0; btn < iodevice.N_BUTTONS; btn++{
            if(e.Requests[f][btn] != 0){
                return true
            }
        }
    }
    return false
}

func here(e elevator.Elevator) bool{
    for btn := 0; btn < iodevice.N_BUTTONS; btn++ {
		if(e.Requests[e.Floor][btn] != 0){
			return true
		}
    }
    return false
}

func ChooseDirection(e elevator.Elevator) DirnBehaviourPair {
    switch e.Dirn {
    case elevio.MD_Up:
        if above(e) {
            return DirnBehaviourPair{elevio.MD_Up, elevator.EB_Moving}
        } else if here(e) {
            return DirnBehaviourPair{elevio.MD_Down, elevator.EB_DoorOpen}
        } else if below(e) {
            return DirnBehaviourPair{elevio.MD_Down, elevator.EB_Moving}
        } else {
            return DirnBehaviourPair{elevio.MD_Stop, elevator.EB_Idle}
        }

    case elevio.MD_Down:
        if below(e) {
            return DirnBehaviourPair{elevio.MD_Down, elevator.EB_Moving}
        } else if here(e) {
            return DirnBehaviourPair{elevio.MD_Up, elevator.EB_DoorOpen}
        } else if above(e) {
            return DirnBehaviourPair{elevio.MD_Up, elevator.EB_Moving}
        } else {
            return DirnBehaviourPair{elevio.MD_Stop, elevator.EB_Idle}
        }

    case elevio.MD_Stop:
        if here(e) {
            return DirnBehaviourPair{elevio.MD_Stop, elevator.EB_DoorOpen}
        } else if above(e) {
            return DirnBehaviourPair{elevio.MD_Up, elevator.EB_Moving}
        } else if below(e) {
            return DirnBehaviourPair{elevio.MD_Down, elevator.EB_Moving}
        } else {
            return DirnBehaviourPair{elevio.MD_Stop, elevator.EB_Idle}
        }

    default:
        return DirnBehaviourPair{elevio.MD_Stop, elevator.EB_Idle}
    }
}

func ShouldStop(e elevator.Elevator) bool{
    switch(e.Dirn){
    case elevio.MD_Down:
        return e.Requests[e.Floor][elevio.BT_HallDown] != 0 || e.Requests[e.Floor][elevio.BT_Cab] != 0 || !below(e);
    case elevio.MD_Up:
        return e.Requests[e.Floor][elevio.BT_HallUp] != 0 || e.Requests[e.Floor][elevio.BT_Cab] != 0 || !above(e);
    case elevio.MD_Stop:
    default:
        return true
    }
	return true
}

func ShouldClearImmediately(e elevator.Elevator, btn_floor int, btn_type elevio.ButtonType) bool{
    switch(e.Config.ClearRequestVariant){
    case elevator.CV_All:
        return e.Floor == btn_floor
    case elevator.CV_InDirn:
        return e.Floor == btn_floor && (e.Dirn == elevio.MD_Up && btn_type == elevio.BT_HallUp || e.Dirn == elevio.MD_Down && btn_type == elevio.BT_HallDown || e.Dirn == elevio.MD_Stop || btn_type == elevio.BT_Cab)
    default:
        return false
    }
}

func ClearAtCurrentFloor(e elevator.Elevator, clear_request chan<- elevio.ButtonEvent) elevator.Elevator{
        
    switch(e.Config.ClearRequestVariant){
    case elevator.CV_All:
        for btn := 0; btn < iodevice.N_BUTTONS; btn++{
            e.Requests[e.Floor][btn] = 0
        }
 
        
    case elevator.CV_InDirn:
        e.Requests[e.Floor][elevio.BT_Cab] = 0;
        switch(e.Dirn){
        case elevio.MD_Up:
            if(!above(e) && e.Requests[e.Floor][elevio.BT_HallUp] == 0){
                e.Requests[e.Floor][elevio.BT_HallDown] = 0;
                clear_request <- elevio.ButtonEvent{Floor: e.Floor, Button: elevio.BT_HallDown}
            }
            e.Requests[e.Floor][elevio.BT_HallUp] = 0;
            clear_request <- elevio.ButtonEvent{Floor: e.Floor, Button: elevio.BT_HallUp}
 
            
        case elevio.MD_Down:
            if(!below(e) && e.Requests[e.Floor][elevio.BT_HallDown] == 0){
                e.Requests[e.Floor][elevio.BT_HallUp] = 0;
                clear_request <- elevio.ButtonEvent{Floor: e.Floor, Button: elevio.BT_HallUp}
            }
            e.Requests[e.Floor][elevio.BT_HallDown] = 0;
            clear_request <- elevio.ButtonEvent{Floor: e.Floor, Button: elevio.BT_HallDown}

            
        case elevio.MD_Stop:
        default:
            e.Requests[e.Floor][elevio.BT_HallUp] = 0;
            e.Requests[e.Floor][elevio.BT_HallDown] = 0;
            clear_request <- elevio.ButtonEvent{Floor: e.Floor, Button: elevio.BT_HallUp}
            clear_request <- elevio.ButtonEvent{Floor: e.Floor, Button: elevio.BT_HallDown}

        }
 
    default:

    }    
    return e
}

















