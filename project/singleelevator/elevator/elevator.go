package elevator

import (
	// "single/elevio"
	"fmt"
	"single/elevio"
	"single/iodevice"
	"time"
)

type ElevatorBehaviour int
const (
    EB_Idle ElevatorBehaviour = 0
    EB_DoorOpen ElevatorBehaviour = 1
    EB_Moving ElevatorBehaviour = 2
)

type ClearRequestVariant int
const (
    // Assume everyone waiting for the elevator gets on the elevator, even if 
    // they will be traveling in the "wrong" direction for a while
    CV_All ClearRequestVariant = 0
    
    // Assume that only those that want to travel in the current direction 
    // enter the elevator, and keep waiting outside otherwise
    CV_InDirn ClearRequestVariant = 1
)

type Elevator struct {
    Floor int
    Dirn elevio.MotorDirection
    Requests[iodevice.N_FLOORS][iodevice.N_BUTTONS] int
    Behaviour ElevatorBehaviour
	Config Config 
}
type Config struct{
	ClearRequestVariant ClearRequestVariant
	DoorOpenDuration_s time.Duration           
} 

func Eb_toString(eb ElevatorBehaviour) string{
    switch eb {
    case EB_Idle:
        return "EB_Idle"
    case EB_DoorOpen:
        return "EB_DoorOpen"
    case EB_Moving:
        return "EB_Moving"
    default:
        return "EB_UNDEFINED"
    }
}

func ElevatorPrint(es Elevator) {
	fmt.Println("  +--------------------+")
	fmt.Printf(
		"  |floor = %-2d          |\n"+
			"  |dirn  = %-12.12s|\n"+
			"  |behav = %-12.12s|\n",
		es.Floor,
		iodevice.Elevio_dirn_toString(es.Dirn),
		Eb_toString(es.Behaviour),
	)
	fmt.Println("  +--------------------+")
	fmt.Println("  |  | up  | dn  | cab |")
	for f := iodevice.N_FLOORS - 1; f >= 0; f-- {
		fmt.Printf("  | %d", f)
		for btn := 0; btn < iodevice.N_BUTTONS; btn++ {
			if (f == iodevice.N_FLOORS-1 && btn == int(elevio.BT_HallUp)) ||
				(f == 0 && btn == elevio.BT_HallDown) {
				fmt.Print("|     ")
			} else {
				fmt.Print(func() string {
					if es.Requests[f][btn] != 0 {
						return "|  #  "
					}
					return "|  -  "
				}())
			}
		}
		fmt.Println("|")
	}
	fmt.Println("  +--------------------+")
}


func Elevator_uninitialized() Elevator{
    return Elevator{
        Floor: -1,
        Dirn: elevio.MD_Stop,
        Behaviour: EB_Idle,
        Config: Config {
            ClearRequestVariant: CV_All,
            DoorOpenDuration_s: 3*time.Second,
        },
    }
}

