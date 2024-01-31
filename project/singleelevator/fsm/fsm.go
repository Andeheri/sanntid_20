package fsm

import(
	"fmt"
	"single/elevator"
	"single/iodevice"
	"single/requests"
	"single/timer"
	"single/elevio"
)

var elev elevator.Elevator //had to use elev instead of elevator because of name conflict
var outputDevice iodevice.ElevOutputDevice


/*
static void __attribute__((constructor)) fsm_init(){
    elevator = elevator_uninitialized();
    
    con_load("elev.con",
        con_val("doorOpenDuration_s", &elev.config.doorOpenDuration_s, "%lf")
        con_enum("clearRequestVariant", &elev.config.clearRequestVariant,
            con_match(CV_All)
            con_match(CV_InDirn)
        )
    )
    
    outputDevice = elevio_getOutputDevice();
}
*/

func Fsm_init(){
    outputDevice = iodevice.Elevio_getOutputDevice()
}

func SetAllLights(es elevator.Elevator){
    for floor := 0; floor < iodevice.N_FLOORS; floor++{
        for btn := elevio.ButtonType(0); btn < iodevice.N_BUTTONS; btn++{
            outputDevice.RequestButtonLight(btn, floor, es.Requests[floor][btn]!=0);
        }
    }
}

func Fsm_onInitBetweenFloors(){
    outputDevice.MotorDirection(elevio.MD_Down);
    elev.Dirn = iodevice.D_Down;
    elev.Behaviour = elevator.EB_Moving;
}


func Fsm_onRequestButtonPress(btn_floor int, btn_type elevio.ButtonType){
    fmt.Printf("\n(%d, %s)\n", btn_floor, iodevice.Elevio_button_toString(btn_type))
    elevator.ElevatorPrint(elev);
    
    switch(elev.Behaviour){
    case elevator.EB_DoorOpen:
        if requests.Requests_shouldClearImmediately(elev, btn_floor, btn_type){
            timer.Timer_start(elev.Config.DoorOpenDuration_s);
        } else {
            elev.Requests[btn_floor][btn_type] = 1;
        }
        break;

    case elevator.EB_Moving:
        elev.Requests[btn_floor][btn_type] = 1;
        break;
        
    case elevator.EB_Idle:    
        elev.Requests[btn_floor][btn_type] = 1
        pair := requests.Requests_chooseDirection(elev)
        elev.Dirn = pair.Dirn;
        elev.Behaviour = pair.Behaviour;
        switch(pair.Behaviour){
        case elevator.EB_DoorOpen:
            outputDevice.DoorLight(true);
            timer.Timer_start(elev.Config.DoorOpenDuration_s);
            elev = requests.Requests_clearAtCurrentFloor(elev);
            break;

        case elevator.EB_Moving:
			//workaround to satisfy input to motorDirection
            outputDevice.MotorDirection(elevio.MotorDirection(elev.Dirn));
            break;
            
        case elevator.EB_Idle:
            break;
        }
        break;
    }
    
    SetAllLights(elev);
    
    fmt.Println("\nNew state:")
    elevator.ElevatorPrint(elev);
}




func Fsm_onFloorArrival(newFloor int){
    fmt.Printf("\n(newfloor: %d)\n",newFloor)
    elevator.ElevatorPrint(elev);

    elev.Floor = newFloor;
    
    outputDevice.FloorIndicator(elev.Floor);
    
    switch(elev.Behaviour){
    case elevator.EB_Moving:
        if requests.Requests_shouldStop(elev){
            outputDevice.MotorDirection(elevio.MD_Stop);
            outputDevice.DoorLight(true);
            elev = requests.Requests_clearAtCurrentFloor(elev);
            timer.Timer_start(elev.Config.DoorOpenDuration_s);
            SetAllLights(elev);
            elev.Behaviour = elevator.EB_DoorOpen;
        }
        break;
    default:
        break;
    }
    
    fmt.Println("\nNew state:")
    elevator.ElevatorPrint(elev);
}




func Fsm_onDoorTimeout(){
    fmt.Println("\n(doorTimeout)")
    
    elevator.ElevatorPrint(elev);
    
    switch(elev.Behaviour){
    case elevator.EB_DoorOpen:;
        pair := requests.Requests_chooseDirection(elev);
        elev.Dirn = pair.Dirn;
        elev.Behaviour = pair.Behaviour;
        
        switch(elev.Behaviour){
        case elevator.EB_DoorOpen:
            timer.Timer_start(elev.Config.DoorOpenDuration_s);
            elev = requests.Requests_clearAtCurrentFloor(elev);
            SetAllLights(elev);
            break;
        case elevator.EB_Moving:
        case elevator.EB_Idle:
            outputDevice.DoorLight(false);
            outputDevice.MotorDirection(elevio.MotorDirection(elev.Dirn));
            break;
        }
        
        break;
    default:
        break;
    }
    
    fmt.Println("\nNew state:")
    elevator.ElevatorPrint(elev);
}














