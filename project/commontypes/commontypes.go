package commontypes

import (
	"encoding/json"
	"fmt"
	"reflect"
)

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

func (ttj *TypeTaggedJSON) ToObject(allowedTypes ...reflect.Type) (interface{}, error) {

	var dataType reflect.Type

	for _, allowedType := range allowedTypes {
		if ttj.TypeId == allowedType.Name() {
			dataType = allowedType
			break
		}
	}

	if dataType == nil {
		return nil, fmt.Errorf("TypeTaggedJSON.TypeId %s does not match any specified types", ttj.TypeId)
	}

	v := reflect.New(dataType)
	err := json.Unmarshal(ttj.JSON, v.Interface())
	if err != nil {
		return nil, err
	}

	object := reflect.Indirect(v).Interface()

	return object, nil
}
