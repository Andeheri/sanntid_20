package elevatoriotypes

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
	RequestButton  func(int, Button) int
	StopButton     func() int
	Obstruction    func() int
}

type ElevOutputDevice struct {
	FloorIndicator      func(int)
	RequestButtonLight func(int, Button, int)
	DoorLight           func(int)
	StopButtonLight     func(int)
	MotorDirection      func(Dirn)
}