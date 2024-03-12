package iodevice

import (
	"project/slave/elevio"
)

const N_FLOORS = 4
const N_BUTTONS = 3

type ElevOutputDevice struct {
	FloorIndicator     func(int)
	RequestButtonLight func(elevio.ButtonType, int, bool)
	DoorLight          func(bool)
	StopButtonLight    func(bool)
	MotorDirection     func(elevio.MotorDirection)
}

func ElevioGetOutputDevice() ElevOutputDevice {
	return ElevOutputDevice{
		FloorIndicator:     elevio.SetFloorIndicator,
		RequestButtonLight: elevio.SetButtonLamp,
		DoorLight:          elevio.SetDoorOpenLamp,
		StopButtonLight:    elevio.SetStopLamp,
		MotorDirection:     elevio.SetMotorDirection,
	}
}

func ElevioDirnToString(d elevio.MotorDirection) string {
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
