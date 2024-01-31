package elevator

import(
	// "single/elevio"
	"single/iodevice"
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
    Dirn iodevice.Dirn
    Requests[iodevice.N_FLOORS][iodevice.N_BUTTONS] int
    Behaviour ElevatorBehaviour
	Config config 
}
type config struct{
	ClearRequestVariant ClearRequestVariant
	DoorOpenDuration_s float32            
} 