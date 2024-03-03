package assigner

import (
	"encoding/json"
	"fmt"
	"os/exec"
	"path/filepath"
	"project/commontypes"
	"runtime"
)

type CommunityState struct {
	HallRequests commontypes.HallRequests             `json:"hallRequests"`
	States       map[string]commontypes.ElevatorState `json:"states"`
}

var assignerExecutable string = ""

// Based on https://github.com/TTK4145/Project-resources/blob/master/cost_fns/usage_examples/example.go
// hall_request_assigner from https://github.com/TTK4145/Project-resources/releases/tag/v1.1.1
func Assign(state *CommunityState) *map[string]commontypes.AssignedRequests {

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

	output := new(map[string]commontypes.AssignedRequests)
	err = json.Unmarshal(ret, &output)
	if err != nil {
		fmt.Println("json.Unmarshal error: ", err)
		return nil
	}

	return output
}
