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
