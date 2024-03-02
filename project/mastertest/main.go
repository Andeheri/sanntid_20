package main

import (
	"fmt"
	"project/master/assigner"
	"project/master/community"
)

func main() {
	input := community.CommunityState{
		HallRequests: [][2]bool{{false, false}, {true, false}, {false, false}, {false, true}},
		States: map[string]community.ElevatorState{
			"one": {
				Behavior:    "moving",
				Floor:       2,
				Direction:   "up",
				CabRequests: []bool{false, false, false, true},
			},
			"two": {
				Behavior:    "idle",
				Floor:       0,
				Direction:   "stop",
				CabRequests: []bool{false, false, false, false},
			},
		},
	}

	fmt.Print(assigner.Assign(&input))
}
