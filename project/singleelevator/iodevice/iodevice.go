package iodevice

import (
	"single/elevio"
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
        return "D_Up"
    case elevio.MD_Down:
        return "D_Down"
    case elevio.MD_Stop:
        return "D_Stop"
    default:
        return "D_UNDEFINED"
    }
}


func Elevio_button_toString(b elevio.ButtonType) string {
    switch b {
    case elevio.BT_HallUp:
        return "B_HallUp"
    case elevio.BT_HallDown:
        return "B_HallDown"
    case elevio.BT_Cab:
        return "B_Cab"
    default:
        return "B_UNDEFINED"
    }
}