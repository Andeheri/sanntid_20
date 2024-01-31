package iodevice

import (
	"single/elevio"
)

const N_FLOORS = 4
const N_BUTTONS = 3

type Dirn int
const ( 
    D_Down  Dirn = -1
    D_Stop  Dirn = 0
    D_Up Dirn = 1
)

type Button int
const (
    B_HallUp Button = 0
    B_HallDown Button = 1
    B_Cab Button = 2
)


type ElevInputDevice struct {
	FloorSensor    func() int
	RequestButton  func(elevio.ButtonType, int) bool
	StopButton     func() bool
	Obstruction    func() bool
}

type ElevOutputDevice struct {
	FloorIndicator      func(int)
	RequestButtonLight func(elevio.ButtonType, int, bool)
	DoorLight           func(bool)
	StopButtonLight     func(bool)
	MotorDirection      func(elevio.MotorDirection)
}

/*
kan hende vi må wrappe funksjonenne....

static void __attribute__((constructor)) elev_init(void){
    elevator_hardware_init();
}

static int _wrap_requestButton(int f, Button b){
    return elevator_hardware_get_button_signal(b, f);
}
static void _wrap_requestButtonLight(int f, Button b, int v){
    elevator_hardware_set_button_lamp(b, f, v);
}
static void _wrap_motorDirection(Dirn d){
    elevator_hardware_set_motor_direction(d);
}
*/

func elevio_getInputDevice() ElevInputDevice {
	return ElevInputDevice{
        FloorSensor: elevio.GetFloor,
        RequestButton: elevio.GetButton,
        StopButton: elevio.GetStop,
        Obstruction: elevio.GetObstruction,
    }
}

func elevio_getOutputDevice() ElevOutputDevice{
    return ElevOutputDevice{
        FloorIndicator:     elevio.SetFloorIndicator,
        RequestButtonLight: elevio.SetButtonLamp,
        DoorLight:          elevio.SetDoorOpenLamp,
        StopButtonLight:    elevio.SetStopLamp,
        MotorDirection:     elevio.SetMotorDirection,
    }
}


func elevio_dirn_toString(d Dirn) string{
    switch d {
    case D_Up:
        return "D_Up"
    case D_Down:
        return "D_Down"
    case D_Stop:
        return "D_Stop"
    default:
        return "D_UNDEFINED"
    }
}


func elevio_button_toString(b Button) string {
    switch b {
    case B_HallUp:
        return "B_HallUp"
    case B_HallDown:
        return "B_HallDown"
    case B_Cab:
        return "B_Cab"
    default:
        return "B_UNDEFINED"
    }
}