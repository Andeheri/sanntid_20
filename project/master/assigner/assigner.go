package assigner

import (
	"encoding/json"
	"fmt"
	"os/exec"
	"path/filepath"
	"project/mscomm"
	"runtime"
)

type CommunityState struct {
	HallRequests mscomm.HallRequests             `json:"hallRequests"`
	States       map[string]mscomm.ElevatorState `json:"states"`
}

var assignerExecutable string = ""

// Based on https://github.com/TTK4145/Project-resources/blob/master/cost_fns/usage_examples/example.go
// hall_request_assigner from https://github.com/TTK4145/Project-resources/releases/tag/v1.1.1
func Assign(state *CommunityState) (*map[string]mscomm.AssignedRequests, error) {

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

	jsonBytes, err := json.Marshal(*state)
	if err != nil {
		return nil, fmt.Errorf("assigner could not marshal json: %v", err)
	}

	ret, err := exec.Command(assignerExecutable, "-i", string(jsonBytes)).CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("assigner executable returned error: %v return value: %+v", err, string(ret))
	}

	output := new(map[string]mscomm.AssignedRequests)
	err = json.Unmarshal(ret, &output)
	if err != nil {
		return nil, fmt.Errorf("assigner could not unmarshal json: %v", err)
	}

	return output, nil
}
