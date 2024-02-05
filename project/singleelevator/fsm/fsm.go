package fsm

import (
	"fmt"
	"single/elevator"
	"single/elevio"
	"single/iodevice"
	"single/requests"
	"time"
)

var Elev elevator.Elevator = elevator.Elevator_uninitialized()//had to use elev instead of elevator because of name conflict
var outputDevice iodevice.ElevOutputDevice


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
    Elev.Dirn = elevio.MD_Down;
    Elev.Behaviour = elevator.EB_Moving;
}


func Fsm_onRequestButtonPress(btn_floor int, btn_type elevio.ButtonType, door_timer *time.Timer){
    fmt.Printf("\n(%d, %s)\n", btn_floor, iodevice.Elevio_button_toString(btn_type))
    elevator.ElevatorPrint(Elev);
    
    switch(Elev.Behaviour){
    case elevator.EB_DoorOpen:
        if requests.Requests_shouldClearImmediately(Elev, btn_floor, btn_type){
            //timer.Timer_start(elev.Config.DoorOpenDuration_s);
            door_timer.Reset(Elev.Config.DoorOpenDuration_s)
        } else {
            Elev.Requests[btn_floor][btn_type] = 1;
        }


    case elevator.EB_Moving:
        Elev.Requests[btn_floor][btn_type] = 1;

        
    case elevator.EB_Idle:    
        Elev.Requests[btn_floor][btn_type] = 1
        pair := requests.Requests_chooseDirection(Elev)
        Elev.Dirn = pair.Dirn;
        Elev.Behaviour = pair.Behaviour;
        switch(pair.Behaviour){
        case elevator.EB_DoorOpen:
            outputDevice.DoorLight(true);
            //timer.Timer_start(elev.Config.DoorOpenDuration_s);
            door_timer.Reset(Elev.Config.DoorOpenDuration_s)
            Elev = requests.Requests_clearAtCurrentFloor(Elev);
        case elevator.EB_Moving:
            outputDevice.MotorDirection(Elev.Dirn);
        case elevator.EB_Idle:

        }

    }
    
    SetAllLights(Elev);
    
    fmt.Println("\nNew state:")
    elevator.ElevatorPrint(Elev);
}




func Fsm_onFloorArrival(newFloor int, door_timer *time.Timer){
    fmt.Printf("\n(newfloor: %d)\n",newFloor)
    elevator.ElevatorPrint(Elev);

    Elev.Floor = newFloor;
    
    outputDevice.FloorIndicator(Elev.Floor);
    
    switch(Elev.Behaviour){
    case elevator.EB_Moving:
        if requests.Requests_shouldStop(Elev){
            fmt.Println("Opening door")
            outputDevice.MotorDirection(elevio.MD_Stop);
            outputDevice.DoorLight(true);
            Elev = requests.Requests_clearAtCurrentFloor(Elev);
            //timer.Timer_start(elev.Config.DoorOpenDuration_s);
            door_timer.Reset(Elev.Config.DoorOpenDuration_s)
            fmt.Println(Elev.Config.DoorOpenDuration_s)
            SetAllLights(Elev);
            Elev.Behaviour = elevator.EB_DoorOpen;
        }
    default:
    }
    
    fmt.Println("\nNew state:")
    elevator.ElevatorPrint(Elev);
}




func Fsm_onDoorTimeout(door_timer *time.Timer){
    fmt.Println("\n(doorTimeout)")
    
    elevator.ElevatorPrint(Elev);
    
    switch(Elev.Behaviour){
    case elevator.EB_DoorOpen:
        pair := requests.Requests_chooseDirection(Elev);
        Elev.Dirn = pair.Dirn;
        Elev.Behaviour = pair.Behaviour;
        
        switch(Elev.Behaviour){
        case elevator.EB_DoorOpen:
            //timer.Timer_start(elev.Config.DoorOpenDuration_s);
            door_timer.Reset(Elev.Config.DoorOpenDuration_s)
            Elev = requests.Requests_clearAtCurrentFloor(Elev);
            SetAllLights(Elev);

        case elevator.EB_Moving:
            outputDevice.DoorLight(false);
            outputDevice.MotorDirection(elevio.MotorDirection(Elev.Dirn));

        case elevator.EB_Idle:
            outputDevice.DoorLight(false);
            outputDevice.MotorDirection(elevio.MotorDirection(Elev.Dirn));
        }
        
    default:
    }
    
    fmt.Println("\nNew state:")
    elevator.ElevatorPrint(Elev);
}














