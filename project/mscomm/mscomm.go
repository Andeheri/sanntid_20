// Master-Slave communication
package mscomm

import (
	"encoding/json"
	"fmt"
	"log"
	"net"
	"reflect"
	"time"
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
type AssignedRequests [][2]bool
type HallRequests [][2]bool

func (hr1 *HallRequests) Merge(hr2 *HallRequests) error {
	if len(*hr1) != len(*hr2) {
		return fmt.Errorf("HallRequests length mismatch")
	}
	for i := range *hr1 {
		(*hr1)[i][0] = (*hr1)[i][0] || (*hr2)[i][0]
		(*hr1)[i][1] = (*hr1)[i][1] || (*hr2)[i][1]
	}
	return nil
}

type SyncRequests struct {
	Requests HallRequests
	Id       int
}

type SyncOK struct {
	Id int
}

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

func NewTypeTaggedJSON(object interface{}) (*TypeTaggedJSON, error) {
	jsonBytes, err := json.Marshal(object)
	if err != nil {
		return nil, err
	}
	tcpPackage := TypeTaggedJSON{
		TypeId: reflect.TypeOf(object).Name(),
		JSON:   jsonBytes,
	}
	return &tcpPackage, nil
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

type Package struct {
	Addr    string
	Payload interface{}
}

type ConnectionEvent struct {
	Connected bool
	Addr      string
	Ch        chan interface{}
}

func TCPReader(conn *net.TCPConn, ch chan<- Package, disconnectEventCh chan<- ConnectionEvent, allowedTypes ...reflect.Type) {
	defer conn.Close()
	addr := conn.RemoteAddr().String()

	decoder := json.NewDecoder(conn)
	for {
		ttj := TypeTaggedJSON{}
		if err := decoder.Decode(&ttj); err != nil {
			//Probably disconnected
			if disconnectEventCh != nil {
				connEvent := ConnectionEvent{
					Connected: false,
					Addr:      addr,
				}
				select {
				case disconnectEventCh <- connEvent:
				case <-time.After(1 * time.Second):
					log.Println("Noone reading from connEventCh")
				}

			}
			log.Println("Reader routine offline")
			return
		}

		object, err := ttj.ToObject(allowedTypes...)

		if err != nil {
			fmt.Println("ttj.ToObject error: ", err)
			continue
		}

		ch <- Package{
			Addr:    addr,
			Payload: object,
		}

	}

}

// is a sender thread necessary??? can't master just send directly?
func TCPSender(conn *net.TCPConn, ch <-chan interface{}) {
	defer conn.Close()

	encoder := json.NewEncoder(conn)

	for {
		data, isOpen := <-ch
		if !isOpen {
			log.Println("Channel closed")
			return
		}

		ttj, err := NewTypeTaggedJSON(data)
		if err != nil {
			log.Println(err)
			continue
		}

		if err := encoder.Encode(ttj); err != nil {
			log.Println(err)
			log.Println("Sender routine offline")
			return
		}
	}

}
