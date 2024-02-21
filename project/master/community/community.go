package community

type ElevatorState struct{
	Behavior    string      `json:"behaviour"`
    Floor       int         `json:"floor"` 
    Direction   string      `json:"direction"`
    CabRequests []bool      `json:"cabRequests"`
}
type CommunityState struct {
	HallRequests    [][2]bool                   `json:"hallRequests"`
    States          map[string]ElevatorState     `json:"states"`
}


