package fsm

import (
	"fmt"
	"project/commontypes"
	"project/slave/elevator"
	"project/slave/elevio"
	"project/slave/iodevice"
	"project/slave/requests"
	"time"
)

var Elev elevator.Elevator = elevator.Initialize()
var outputDevice iodevice.ElevOutputDevice


func Init(){
    outputDevice = iodevice.Elevio_getOutputDevice()

    // Code for fixing starting position between floors
    outputDevice.MotorDirection(elevio.MD_Down);
    Elev.Dirn = elevio.MD_Down;
    Elev.Behaviour = elevator.EB_Moving;
}

func SetAllLights(es elevator.Elevator){
    for floor := 0; floor < iodevice.N_FLOORS; floor++{
        for btn := elevio.ButtonType(0); btn < iodevice.N_BUTTONS; btn++{
            outputDevice.RequestButtonLight(btn, floor, es.HallLights[floor][btn]!=0);
        }
    }
    for floor := 0; floor < iodevice.N_FLOORS; floor++{
        outputDevice.RequestButtonLight(elevio.BT_Cab, floor, es.Requests[floor][elevio.BT_Cab]!=0);
    }
}


func OnRequestButtonPress(btn_floor int, btn_type elevio.ButtonType, door_timer *time.Timer, clear_request chan elevio.ButtonEvent){
    fmt.Printf("\n(%d, %s)\n", btn_floor, iodevice.Elevio_button_toString(btn_type))
    Elev.Print()
    
    switch(Elev.Behaviour){
    case elevator.EB_DoorOpen:
        if requests.ShouldClearImmediately(Elev, btn_floor, btn_type){
            door_timer.Reset(Elev.Config.DoorOpenDuration_s)
        } else {
            Elev.Requests[btn_floor][btn_type] = 1
        }


    case elevator.EB_Moving:
        Elev.Requests[btn_floor][btn_type] = 1

        
    case elevator.EB_Idle:    
        Elev.Requests[btn_floor][btn_type] = 1
        pair := requests.ChooseDirection(Elev)
        Elev.Dirn = pair.Dirn;
        Elev.Behaviour = pair.Behaviour;
        switch(pair.Behaviour){
        case elevator.EB_DoorOpen:
            outputDevice.DoorLight(true);
            door_timer.Reset(Elev.Config.DoorOpenDuration_s)
            Elev = requests.ClearAtCurrentFloor(Elev, clear_request)
        case elevator.EB_Moving:
            outputDevice.MotorDirection(Elev.Dirn)
        case elevator.EB_Idle:

        }

    }
    
    SetAllLights(Elev);
    
    fmt.Println("\nNew state:")
    Elev.Print();
}




func OnFloorArrival(newFloor int, door_timer *time.Timer, clear_request chan elevio.ButtonEvent){
    fmt.Printf("\n(newfloor: %d)\n",newFloor)
    Elev.Print();

    Elev.Floor = newFloor;
    
    outputDevice.FloorIndicator(Elev.Floor);
    
    switch(Elev.Behaviour){
    case elevator.EB_Moving:
        if requests.ShouldStop(Elev){
            fmt.Println("Opening door")
            outputDevice.MotorDirection(elevio.MD_Stop);
            outputDevice.DoorLight(true);
            Elev = requests.ClearAtCurrentFloor(Elev, clear_request);
            door_timer.Reset(Elev.Config.DoorOpenDuration_s)
            fmt.Println(Elev.Config.DoorOpenDuration_s)
            SetAllLights(Elev);
            Elev.Behaviour = elevator.EB_DoorOpen;
        }
    default:
    }
    
    fmt.Println("\nNew state:")
    Elev.Print();
}




func OnDoorTimeout(door_timer *time.Timer, clear_request chan elevio.ButtonEvent){
    fmt.Println("\n(doorTimeout)")
    
    Elev.Print();
    
    if (Elev.Behaviour == elevator.EB_DoorOpen){

        if (Elev.Obstructed){
            door_timer.Reset(Elev.Config.DoorOpenDuration_s)
            return
        }
        pair := requests.ChooseDirection(Elev);
        Elev.Dirn = pair.Dirn;
        Elev.Behaviour = pair.Behaviour;
        
        switch(Elev.Behaviour){
        case elevator.EB_DoorOpen:
            door_timer.Reset(Elev.Config.DoorOpenDuration_s)
            Elev = requests.ClearAtCurrentFloor(Elev, clear_request);
            SetAllLights(Elev);

        case elevator.EB_Moving:
            outputDevice.DoorLight(false);
            outputDevice.MotorDirection(elevio.MotorDirection(Elev.Dirn));

        case elevator.EB_Idle:
            outputDevice.DoorLight(false);
            outputDevice.MotorDirection(elevio.MotorDirection(Elev.Dirn));
        }
    }
    
    fmt.Println("\nNew state:")
    Elev.Print();
}

func OnObstruction(is_obstructed bool){
    Elev.Obstructed = is_obstructed
}


// clear all requests when receiving restructured list of orders from master.?
func RequestsClearAll(){
    for btn := 0; btn < 2; btn++{
        for floor := 0; floor < iodevice.N_FLOORS; floor++{
            Elev.Requests[floor][btn] = 0
        }    
    }
}

// call fsm button press for the restructured list of orders from master.?
func RequestsSetAll(masterRequests commontypes.AssignedRequests, door_timer *time.Timer, clear_request chan elevio.ButtonEvent) {
    //fsm on butonpress for loop
    for btn := 0; btn < 2; btn++{
        for floor := 0; floor < iodevice.N_FLOORS; floor++{
            if masterRequests[floor][btn] {
                OnRequestButtonPress(floor, elevio.ButtonType(btn), door_timer, clear_request)
            }
        }    
    }  
}

func GetCabRequests()[]bool{
    cabRequests := make([]bool,iodevice.N_FLOORS)
    var btn elevio.ButtonType = elevio.BT_Cab
    for floor := 0; floor < iodevice.N_FLOORS; floor++{
        if Elev.Requests[floor][btn] == 1 {
            cabRequests[floor] = true
        } else{
            cabRequests[floor] = false
        }
    }    
    return cabRequests
}













