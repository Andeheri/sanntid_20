package elevator

import (
	"fmt"
	"slave/elevio"
	"slave/iodevice"
	"time"
)

type ElevatorBehaviour string
const (
    EB_Idle ElevatorBehaviour = "idle"
    EB_DoorOpen ElevatorBehaviour = "doorOpen"
    EB_Moving ElevatorBehaviour = "moving"
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
	Obstructed bool
	Door_timer time.Timer
    Requests[iodevice.N_FLOORS][iodevice.N_BUTTONS] int
	All_lights[iodevice.N_FLOORS][iodevice.N_BUTTONS] int
    Behaviour ElevatorBehaviour
	Config Config 
}
type Config struct{
	ClearRequestVariant ClearRequestVariant
	DoorOpenDuration_s time.Duration           
} 

func (es Elevator)Print(){
	fmt.Println("  +--------------------+")
	fmt.Printf(
		"  |floor = %-2d          |\n"+
			"  |dirn  = %-12.12s|\n"+
			"  |behav = %-12.12s|\n",
		es.Floor,
		iodevice.Elevio_dirn_toString(es.Dirn),
		es.Behaviour,
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


func Initialize() Elevator{
    return Elevator{
        Floor: -1,
        Dirn: elevio.MD_Stop,
		Obstructed: false,
		Door_timer: *time.NewTimer(-1),
        Behaviour: EB_Idle,
        Config: Config {
            ClearRequestVariant: CV_InDirn,
            DoorOpenDuration_s: 3*time.Second,
        },
    }
}

