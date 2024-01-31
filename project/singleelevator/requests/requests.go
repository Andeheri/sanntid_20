package requests

import(
	"single/elevator"
	"single/iodevice"
	"single/elevio"
)

type DirnBehaviourPair struct{
	Dirn iodevice.Dirn
	Behaviour elevator.ElevatorBehaviour
}


func requests_above(e elevator.Elevator) bool{
    for f := e.Floor+1; f < iodevice.N_FLOORS; f++ {
        for btn := 0; btn < iodevice.N_BUTTONS; btn++{
            if(e.Requests[f][btn] != 0){
                return true
            }
        }
    }
    return false
}

func requests_below(e elevator.Elevator) bool{
    for f := 0; f < e.Floor; f++ {
        for btn := 0; btn < iodevice.N_BUTTONS; btn++{
            if(e.Requests[f][btn] != 0){
                return true
            }
        }
    }
    return false
}

func requests_here(e elevator.Elevator) bool{
    for btn := 0; btn < iodevice.N_BUTTONS; btn++ {
		if(e.Requests[e.Floor][btn] != 0){
			return true
		}
    }
    return false
}

func Requests_chooseDirection(e elevator.Elevator) DirnBehaviourPair {
    switch e.Dirn {
    case iodevice.D_Up:
        if requests_above(e) {
            return DirnBehaviourPair{iodevice.D_Up, elevator.EB_Moving}
        } else if requests_here(e) {
            return DirnBehaviourPair{iodevice.D_Down, elevator.EB_DoorOpen}
        } else if requests_below(e) {
            return DirnBehaviourPair{iodevice.D_Down, elevator.EB_Moving}
        } else {
            return DirnBehaviourPair{iodevice.D_Stop, elevator.EB_Idle}
        }

    case iodevice.D_Down:
        if requests_below(e) {
            return DirnBehaviourPair{iodevice.D_Down, elevator.EB_Moving}
        } else if requests_here(e) {
            return DirnBehaviourPair{iodevice.D_Up, elevator.EB_DoorOpen}
        } else if requests_above(e) {
            return DirnBehaviourPair{iodevice.D_Up, elevator.EB_Moving}
        } else {
            return DirnBehaviourPair{iodevice.D_Stop, elevator.EB_Idle}
        }

    case iodevice.D_Stop:
        if requests_here(e) {
            return DirnBehaviourPair{iodevice.D_Stop, elevator.EB_DoorOpen}
        } else if requests_above(e) {
            return DirnBehaviourPair{iodevice.D_Up, elevator.EB_Moving}
        } else if requests_below(e) {
            return DirnBehaviourPair{iodevice.D_Down, elevator.EB_Moving}
        } else {
            return DirnBehaviourPair{iodevice.D_Stop, elevator.EB_Idle}
        }

    default:
        return DirnBehaviourPair{iodevice.D_Stop, elevator.EB_Idle}
    }
}

func Requests_shouldStop(e elevator.Elevator) bool{
    switch(e.Dirn){
    case iodevice.D_Down:
        return e.Requests[e.Floor][elevio.BT_HallDown] != 0 || e.Requests[e.Floor][elevio.BT_Cab] != 0 || !requests_below(e);
    case iodevice.D_Up:
        return e.Requests[e.Floor][elevio.BT_HallUp] != 0 || e.Requests[e.Floor][elevio.BT_Cab] != 0 || !requests_above(e);
    case iodevice.D_Stop:
    default:
        return true
    }
	return true
}

func Requests_shouldClearImmediately(e elevator.Elevator, btn_floor int, btn_type elevio.ButtonType) bool{
    switch(e.Config.ClearRequestVariant){
    case elevator.CV_All:
        return e.Floor == btn_floor
    case elevator.CV_InDirn:
        return e.Floor == btn_floor && e.Dirn == iodevice.D_Up && btn_type == elevio.BT_HallUp || e.Dirn == iodevice.D_Down && btn_type == elevio.BT_HallDown || e.Dirn == iodevice.D_Stop || btn_type == elevio.BT_Cab
    default:
        return false
    }
}

func Requests_clearAtCurrentFloor(e elevator.Elevator) elevator.Elevator{
        
    switch(e.Config.ClearRequestVariant){
    case elevator.CV_All:
        for btn := 0; btn < iodevice.N_BUTTONS; btn++{
            e.Requests[e.Floor][btn] = 0
        }
        break
        
    case elevator.CV_InDirn:
        e.Requests[e.Floor][elevio.BT_Cab] = 0;
        switch(e.Dirn){
        case iodevice.D_Up:
            if(!requests_above(e) && e.Requests[e.Floor][elevio.BT_HallUp] == 0){
                e.Requests[e.Floor][elevio.BT_HallDown] = 0;
            }
            e.Requests[e.Floor][elevio.BT_HallUp] = 0;
            break;
            
        case iodevice.D_Down:
            if(!requests_below(e) && e.Requests[e.Floor][elevio.BT_HallDown] == 0){
                e.Requests[e.Floor][elevio.BT_HallUp] = 0;
            }
            e.Requests[e.Floor][elevio.BT_HallDown] = 0;
            break
            
        case iodevice.D_Stop:
        default:
            e.Requests[e.Floor][elevio.BT_HallUp] = 0;
            e.Requests[e.Floor][elevio.BT_HallDown] = 0;
            break
        }
        break   
    default:
        break; 
    }    
    return e
}











