package requests

import(
	"slave/elevator"
	"slave/iodevice"
	"slave/elevio"
)

type DirnBehaviourPair struct{
	Dirn elevio.MotorDirection
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
    case elevio.MD_Up:
        if requests_above(e) {
            return DirnBehaviourPair{elevio.MD_Up, elevator.EB_Moving}
        } else if requests_here(e) {
            return DirnBehaviourPair{elevio.MD_Down, elevator.EB_DoorOpen}
        } else if requests_below(e) {
            return DirnBehaviourPair{elevio.MD_Down, elevator.EB_Moving}
        } else {
            return DirnBehaviourPair{elevio.MD_Stop, elevator.EB_Idle}
        }

    case elevio.MD_Down:
        if requests_below(e) {
            return DirnBehaviourPair{elevio.MD_Down, elevator.EB_Moving}
        } else if requests_here(e) {
            return DirnBehaviourPair{elevio.MD_Up, elevator.EB_DoorOpen}
        } else if requests_above(e) {
            return DirnBehaviourPair{elevio.MD_Up, elevator.EB_Moving}
        } else {
            return DirnBehaviourPair{elevio.MD_Stop, elevator.EB_Idle}
        }

    case elevio.MD_Stop:
        if requests_here(e) {
            return DirnBehaviourPair{elevio.MD_Stop, elevator.EB_DoorOpen}
        } else if requests_above(e) {
            return DirnBehaviourPair{elevio.MD_Up, elevator.EB_Moving}
        } else if requests_below(e) {
            return DirnBehaviourPair{elevio.MD_Down, elevator.EB_Moving}
        } else {
            return DirnBehaviourPair{elevio.MD_Stop, elevator.EB_Idle}
        }

    default:
        return DirnBehaviourPair{elevio.MD_Stop, elevator.EB_Idle}
    }
}

func Requests_shouldStop(e elevator.Elevator) bool{
    switch(e.Dirn){
    case elevio.MD_Down:
        return e.Requests[e.Floor][elevio.BT_HallDown] != 0 || e.Requests[e.Floor][elevio.BT_Cab] != 0 || !requests_below(e);
    case elevio.MD_Up:
        return e.Requests[e.Floor][elevio.BT_HallUp] != 0 || e.Requests[e.Floor][elevio.BT_Cab] != 0 || !requests_above(e);
    case elevio.MD_Stop:
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
        return e.Floor == btn_floor && (e.Dirn == elevio.MD_Up && btn_type == elevio.BT_HallUp || e.Dirn == elevio.MD_Down && btn_type == elevio.BT_HallDown || e.Dirn == elevio.MD_Stop || btn_type == elevio.BT_Cab)
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
 
        
    case elevator.CV_InDirn:
        e.Requests[e.Floor][elevio.BT_Cab] = 0;
        switch(e.Dirn){
        case elevio.MD_Up:
            if(!requests_above(e) && e.Requests[e.Floor][elevio.BT_HallUp] == 0){
                e.Requests[e.Floor][elevio.BT_HallDown] = 0;
            }
            e.Requests[e.Floor][elevio.BT_HallUp] = 0;
 
            
        case elevio.MD_Down:
            if(!requests_below(e) && e.Requests[e.Floor][elevio.BT_HallDown] == 0){
                e.Requests[e.Floor][elevio.BT_HallUp] = 0;
            }
            e.Requests[e.Floor][elevio.BT_HallDown] = 0;

            
        case elevio.MD_Stop:
        default:
            e.Requests[e.Floor][elevio.BT_HallUp] = 0;
            e.Requests[e.Floor][elevio.BT_HallDown] = 0;

        }
 
    default:

    }    
    return e
}


// clear all requests when receiving restructured list of orders from master.?
func Requests_clearAll(e elevator.Elevator) elevator.Elevator{
    for btn := 0; btn < iodevice.N_BUTTONS; btn++{
        for floor := 0; floor < iodevice.N_FLOORS; floor++{
            e.Requests[e.Floor][btn] = 0
        }    
    }  
    return e
}


// call fsm button press for the restructured list of orders from master.?
func Requests_setAll(e elevator.Elevator) elevator.Elevator{
    //fsm on butonpress for loop
    return e
}











