package mastercom

import (
	"encoding/json"
	"fmt"
	"net"
	"project/commontypes"
	"project/slave/elevio"
	"project/slave/fsm"
	"reflect"
	"time"
)

type MasterChannels struct {
	ButtonPress  chan elevio.ButtonEvent
	ClearRequest chan elevio.ButtonEvent

	AssignedRequests chan commontypes.AssignedRequests
	HallLights     chan commontypes.Lights
	RequestedState chan struct{}

	Sender chan interface{}
}

var hallRequests commontypes.HallRequests = commontypes.HallRequests{{false, false}, {false, false}, {false, false}, {false, false}}

func MasterCommunication(masterAddress *net.TCPAddr, chans *MasterChannels, stopch <-chan struct{}) {

	masterConn, err := net.DialTCP("tcp", nil, masterAddress)
	if err != nil {
		fmt.Println("Error connecting to master:", err)
		time.Sleep(10 * time.Millisecond)
		masterConn, err = net.DialTCP("tcp", nil, masterAddress)
		if err != nil {
			return
		}
	}

	defer masterConn.Close()

	go Receiver(masterConn, chans, stopch)
	go Sender(masterConn, chans.Sender, stopch)

	for {
		select {
		case a := <-chans.ButtonPress:
			fmt.Println(a, "sender button press melding")
			pressed := commontypes.ButtonPressed{Floor: a.Floor, Button: int(a.Button)}
			chans.Sender <- pressed
		case a := <-chans.ClearRequest:
			fmt.Println(a, "sender clear request melding")
			completedOrder := commontypes.OrderComplete{Floor: a.Floor, Button: int(a.Button)}
			chans.Sender <- completedOrder
		case <-stopch:
			return
		}
	}
}

func SendState(sender chan<- interface{}) {
	state := commontypes.ElevatorState{
		Behavior:    string(fsm.Elev.Behaviour),
		Floor:       fsm.Elev.Floor,
		Direction:   elevio.Elevio_dirn_toString(fsm.Elev.Dirn),
		CabRequests: fsm.GetCabRequests(),
	}

	fmt.Println("Sender state", state)

	sender <- state
}

func Receiver(masterConn *net.TCPConn, chans *MasterChannels, stopch <-chan struct{}) {

	buffer := make([]byte, 1024)

	for {
		select {
		case <-stopch:
			return
		default:
			// Read data from the master
			n, err := masterConn.Read(buffer)
			if err != nil {
				fmt.Println("Error:", err)
				return
			}

			var ttj commontypes.TypeTaggedJSON

			err = json.Unmarshal(buffer[:n], &ttj)
			if err != nil {
				fmt.Println("received invalid TCP Package ", err)
				continue
			}

			object, err := ttj.ToObject(
				reflect.TypeOf(commontypes.RequestState{}),
				reflect.TypeOf(commontypes.RequestHallRequests{}),
				reflect.TypeOf(commontypes.HallRequests{}),
				reflect.TypeOf(commontypes.SyncRequests{}),
				reflect.TypeOf(commontypes.Lights{}),
				reflect.TypeOf(commontypes.AssignedRequests{}),
			)

			if err != nil {
				fmt.Println("ttj.ToValuePtr error: ", err)
				continue
			}

			switch reflect.TypeOf(object) {
			case reflect.TypeOf(commontypes.RequestState{}):
				fmt.Println("State requested by master")
				chans.RequestedState <- struct{}{}
			case reflect.TypeOf(commontypes.RequestHallRequests{}):
				fmt.Println("Master requested Hallrequests")
				chans.Sender <- hallRequests
			case reflect.TypeOf(commontypes.SyncRequests{}):
				fmt.Println("Received Syncrequests")
				hallRequests = object.(commontypes.SyncRequests).Requests
				chans.Sender <- commontypes.SyncOK{Id: object.(commontypes.SyncRequests).Id}
			case reflect.TypeOf(commontypes.Lights{}):
				fmt.Println("Received halllights")
				chans.HallLights <- object.(commontypes.Lights)
			case reflect.TypeOf(commontypes.AssignedRequests{}):
				fmt.Println("Received assigned requests")
				chans.AssignedRequests <- object.(commontypes.AssignedRequests)
			default:
				fmt.Println("received invalid TypeTaggedJSON.TypeId ", ttj.TypeId)
				continue
			}
		}
	}
}

func Sender(masterConn *net.TCPConn, ch <-chan interface{}, stopch <-chan struct{}) {
	for {
		select {
		case data := <-ch:
			jsonBytesPayload, err := json.Marshal(data)
			if err != nil {
				fmt.Println("json.Marshal error: ", err)
				continue
			}
			tcpPackage := commontypes.TypeTaggedJSON{
				TypeId: reflect.TypeOf(data).Name(),
				JSON:   jsonBytesPayload,
			}

			jsonBytesPackage, err := json.Marshal(tcpPackage)
			if err != nil {
				fmt.Println("json.Marshal error: ", err)
				continue
			}
			_, err = masterConn.Write(jsonBytesPackage)
			if err != nil {
				fmt.Println("net.Write Error:", err)
				return
			}
		case <-stopch:
			return
		}
	}
}
