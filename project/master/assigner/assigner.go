package assigner

import (
	"encoding/json"
	"fmt"
	"project/master/community"
	"os/exec"
	"path/filepath"
	"runtime"
)

var assignerExecutable string = ""

// Based on https://github.com/TTK4145/Project-resources/blob/master/cost_fns/usage_examples/example.go
// hall_request_assigner from https://github.com/TTK4145/Project-resources/releases/tag/v1.1.1
func Assign(state *community.CommunityState) *map[string][][2]bool {

	if assignerExecutable == "" {
		_, filename, _, ok := runtime.Caller(0)
		if !ok {
			panic("Could not find the path to assigner executable")
		}
		dir := filepath.Dir(filename)

		switch runtime.GOOS {
		case "linux":
			assignerExecutable = filepath.Join(dir, "hall_request_assigner")
		case "windows":
			assignerExecutable = filepath.Join(dir, "hall_request_assigner.exe")
		default:
			panic("OS not supported")
		}
	}

	jsonBytes, err := json.Marshal(state)
	if err != nil {
		fmt.Println("json.Marshal error: ", err)
		return nil
	}

	ret, err := exec.Command(assignerExecutable, "-i", string(jsonBytes)).CombinedOutput()
	if err != nil {
		fmt.Println("exec.Command error: ", err)
		fmt.Println(string(ret))
		return nil
	}

	output := new(map[string][][2]bool)
	err = json.Unmarshal(ret, &output)
	if err != nil {
		fmt.Println("json.Unmarshal error: ", err)
		return nil
	}

	return output
}

var syncState struct {
	syncing bool
	button  int
	floor   int
	pending map[string]struct{} //Only using the keys. values are empty
}

var communityState community.CommunityState
var ipToSendCh map[string]chan interface{}

func Start(newSlaveCh chan string, readSlaveCh chan community.SlaveMessage) {
	ipToSendCh = make(map[string]chan interface{})
	syncState.pending = make(map[string]struct{})

	for {
		//TODO: handle new slave, and remove slave, and quit?

		message := <-readSlaveCh

		switch message.Payload.(type) {
		case community.ElevatorState:
			_, sendChExists := ipToSendCh[message.SenderIP]
			if !sendChExists {
				ipToSendCh[message.SenderIP] = make(chan interface{})
			}

			communityState.States[message.SenderIP] = message.Payload.(community.ElevatorState)

			delete(syncState.pending, message.SenderIP)

		case community.ButtonEvent:
			//if hall button:
			buttonEvent := message.Payload.(community.ButtonEvent)
			//start syncing:
			syncState.syncing = true
			syncState.pending = make(map[string]struct{})
			syncState.button = buttonEvent.Button
			syncState.floor = buttonEvent.Floor

			//sync button first

			//then request info from all slaves
			//mark all slaves as pending
			for IP := range ipToSendCh {
				syncState.pending[IP] = struct{}{}
			}

		case community.OrderComplete:
			orderComplete := message.Payload.(community.OrderComplete)
			communityState.HallRequests[orderComplete.Floor][orderComplete.Button] = false
		}

		if syncState.syncing && len(syncState.pending) == 0 {
			//button light syncing complete
			syncState.syncing = false

			communityState.HallRequests[syncState.floor][syncState.button] = true

			orders := *Assign(&communityState)
			for IP, order := range orders {
				//TODO: update button light information
				ipToSendCh[IP] <- order
			}

		}

	}
}
