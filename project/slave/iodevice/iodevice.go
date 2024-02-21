package iodevice

import (
	"slave/elevio"
)

const N_FLOORS = 4
const N_BUTTONS = 3


type ElevOutputDevice struct {
	FloorIndicator      func(int)
	RequestButtonLight func(elevio.ButtonType, int, bool)
	DoorLight           func(bool)
	StopButtonLight     func(bool)
	MotorDirection      func(elevio.MotorDirection)
}


func Elevio_getOutputDevice() ElevOutputDevice{
    return ElevOutputDevice{
        FloorIndicator:     elevio.SetFloorIndicator,
        RequestButtonLight: elevio.SetButtonLamp,
        DoorLight:          elevio.SetDoorOpenLamp,
        StopButtonLight:    elevio.SetStopLamp,
        MotorDirection:     elevio.SetMotorDirection,
    }
}


func Elevio_dirn_toString(d elevio.MotorDirection) string{
    switch d {
    case elevio.MD_Up:
        return "up"
    case elevio.MD_Down:
        return "down"
    case elevio.MD_Stop:
        return "stop"
    default:
        return "UNDEFINED"
    }
}



func Elevio_button_toString(b elevio.ButtonType) string {
    switch b {
    case elevio.BT_HallUp:
        return "HallUp"
    case elevio.BT_HallDown:
        return "HallDown"
    case elevio.BT_Cab:
        return "Cab"
    default:
        return "UNDEFINED"
    }
}