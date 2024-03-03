package commontypes

type ElevatorState struct {
	Behavior    string `json:"behaviour"`
	Floor       int    `json:"floor"`
	Direction   string `json:"direction"`
	CabRequests []bool `json:"cabRequests"`
}

type ButtonPressed struct {
	Floor  int
	Button int
}

type OrderComplete ButtonPressed

type Lights [][2]bool
type HallRequests [][2]bool
type AssignedRequests [][2]bool

type SyncOK struct{}

type RequestState struct{}
type RequestHallRequests struct{}

type MISOChBundle struct {
	HallRequests  chan HallRequests
	ElevatorState chan ElevatorState
	ButtonPressed chan ButtonPressed
	OrderComplete chan OrderComplete
	SyncOK        chan SyncOK
}

type MOSIChBundle struct {
	RequestHallRequests chan RequestHallRequests
	RequestState        chan RequestState
	UpdateOrders        chan HallRequests
	UpdateLights        chan Lights
	AssignedRequests    chan AssignedRequests
}

type TypeTaggedJSON struct {
	TypeId string
	JSON   []byte
}