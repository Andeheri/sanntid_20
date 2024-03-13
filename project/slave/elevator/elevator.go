package elevator

import (
	"project/mscomm"
	"project/rblog"
	"project/slave/elevio"
	"time"
)

const N_FLOORS = 4
const N_BUTTONS = 3

type ElevatorBehaviour string

const (
	EB_Idle     ElevatorBehaviour = "idle"
	EB_DoorOpen ElevatorBehaviour = "doorOpen"
	EB_Moving   ElevatorBehaviour = "moving"
)

type Elevator struct {
	Floor      int
	Dirn       elevio.MotorDirection
	Obstructed bool
	DoorTimer  time.Timer
	Requests   [N_FLOORS][N_BUTTONS]int
	HallLights mscomm.Lights
	Behaviour  ElevatorBehaviour
	Config     Config
}

type Config struct {
	DoorOpenDuration        time.Duration
	InbetweenFloorsDuration time.Duration
}

func Initialize() Elevator {
	return Elevator{
		Floor:      -1,
		Dirn:       elevio.MD_Stop,
		Obstructed: false,
		DoorTimer:  *time.NewTimer(-1),
		Behaviour:  EB_Idle,
		Config: Config{
			DoorOpenDuration:        3 * time.Second,
			InbetweenFloorsDuration: 10 * time.Second,
		},
		HallLights: mscomm.Lights{{false, false}, {false, false}, {false, false}, {false, false}},
	}
}

func (es *Elevator) Print() {
	rblog.White.Println("  +--------------------+")
	rblog.White.Printf(
		"  |floor = %-2d          |\n"+
			"  |dirn  = %-12.12s|\n"+
			"  |behav = %-12.12s|\n",
		es.Floor,
		dirnToString(es.Dirn),
		es.Behaviour,
	)
	rblog.White.Println("  +--------------------+")
	rblog.White.Println("  |  | up  | dn  | cab |")
	for f := N_FLOORS - 1; f >= 0; f-- {
		rblog.White.Printf("  | %d", f)
		for btn := 0; btn < N_BUTTONS; btn++ {
			if (f == N_FLOORS-1 && btn == int(elevio.BT_HallUp)) ||
				(f == 0 && btn == int(elevio.BT_HallDown)) {
				rblog.White.Print("|     ")
			} else {
				rblog.White.Print(func() string {
					if es.Requests[f][btn] != 0 {
						return "|  #  "
					}
					return "|  -  "
				}())
			}
		}
		rblog.White.Println("|")
	}
	rblog.White.Println("  +--------------------+")
}

func dirnToString(d elevio.MotorDirection) string {
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
