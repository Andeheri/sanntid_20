package main

import (
	"master/assigner"
	"master/community"
	"fmt"
)

func main() {
	 input := community.CommunityState{
		HallRequests: [][2]bool{{false, false}, {true, false}, {false, false}, {false, true}},
		States: map[string]community.ElevatorState{
			"one": community.ElevatorState{
				Behavior:    "moving",
				Floor:       2,
				Direction:   "up",
				CabRequests: []bool{false, false, false, true},
			},
			"two": community.ElevatorState{
				Behavior:    "idle",
				Floor:       0,
				Direction:   "stop",
				CabRequests: []bool{false, false, false, false},
			},
		},
	}

	fmt.Print(assigner.Assign(&input))
}
