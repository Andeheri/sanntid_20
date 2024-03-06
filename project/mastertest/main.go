package main

import (
	"project/master"
	"project/master/slavecomm"
	"project/mscomm"
)

func main() {

	// state := assigner.CommunityState{}
	// state.HallRequests = make(mscomm.HallRequests, 4)
	// state.States = make(map[string]mscomm.ElevatorState)
	// state.States["slave1"] = mscomm.ElevatorState{
	// 	Behavior:    "idle",
	// 	Floor:       0,
	// 	Direction:   "up",
	// 	CabRequests: []bool{false, false, false, false},
	// }
	// assigned_requests, err := assigner.Assign(&state)
	// if err != nil {
	// 	panic(err)
	// }
	// fmt.Println(*assigned_requests)

	masterCh := make(chan mscomm.Package)
	connEventCh := make(chan mscomm.ConnectionEvent)
	go slavecomm.Listener(12221, masterCh, connEventCh)
	go master.Run(masterCh, connEventCh)

	select {}
}
