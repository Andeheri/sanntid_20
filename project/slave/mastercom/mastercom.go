package mastercom

import (
	"encoding/json"
	"fmt"
	"net"
	"slave/community"
	"slave/elevio"
	"slave/fsm"
	"slave/iodevice"
)

type Master_channels struct {
	ButtonPress  chan elevio.ButtonEvent
	ClearRequest chan elevio.ButtonEvent
	MasterRequests chan [iodevice.N_FLOORS][iodevice.N_BUTTONS]int
	AllLights chan [iodevice.N_FLOORS][iodevice.N_BUTTONS]int
	RequestedState chan bool
	Sender chan interface{}
}

var communityState community.CommunityState

func Master_communication(masterAddress *net.TCPAddr, chans *Master_channels, stopch <-chan bool) {

	masterConn, err := net.DialTCP("tcp", nil, masterAddress)
	if err != nil {
		fmt.Println("Error connecting to master:", err)
		return
	}

	go Receiver(masterConn, chans, stopch)
	go Sender(masterConn, chans.Sender, stopch)

	for {
		select {
		case a := <-chans.ButtonPress:
			fmt.Println(a, "sender button press melding")
			//cab requests
			pressed := community.ButtonEvent{Floor: a.Floor, Button: int(a.Button)}
			chans.Sender <- pressed
		case a := <-chans.ClearRequest:
			fmt.Println(a, "sender clear request melding")
			completedOrder := community.OrderComplete{Floor: a.Floor, Button: int(a.Button)}
			chans.Sender <- completedOrder
		case <-stopch:
			return
		}
	}
}

func SendState(sender chan<- interface{}) {
	state := community.ElevatorState{
		Behavior:    string(fsm.Elev.Behaviour),
		Floor:       fsm.Elev.Floor,
		Direction:   elevio.Elevio_dirn_toString(fsm.Elev.Dirn),
		CabRequests: fsm.GetCabRequests(),
	}

	fmt.Println("Sender state", state)

	sender <- state
}

func Receiver(masterConn *net.TCPConn, chans *Master_channels, stopch <-chan bool) {

	buffer := make([]byte, 1024)

	for {
		select {
		case <-stopch:
			return
		default:
		}

		// Read data from the master
		n, err := masterConn.Read(buffer)
		if err != nil {
			fmt.Println("Error:", err)
			return
		}

		var requests [iodevice.N_FLOORS][iodevice.N_BUTTONS]int
		var requestedState bool
		var requestedCommunityState bool
		var potentialCommunityState community.CommunityState
		var allLights [iodevice.N_FLOORS][iodevice.N_BUTTONS]int
		

		if err = json.Unmarshal(buffer[:n], &requests); err == nil {
			chans.MasterRequests <- requests
		} else if err = json.Unmarshal(buffer[:n], &requestedState); err == nil {
			chans.RequestedState <- requestedState
		} else if err = json.Unmarshal(buffer[:n], &potentialCommunityState); err == nil {
			communityState = potentialCommunityState
		} else if err = json.Unmarshal(buffer[:n], &requestedCommunityState); err == nil {
			chans.Sender <- communityState
		} else if err = json.Unmarshal(buffer[:n], &allLights); err == nil {
			chans.AllLights <- allLights
		} else {
			fmt.Println("json.Unmarshal error (no matching data types) : ", err)
			return
		}
	}
}

func Sender(masterConn *net.TCPConn, ch <-chan interface{}, stopch <-chan bool) {
	for {
		select {
		case data := <-ch:
			jsonBytes, err := json.Marshal(data)
			if err != nil {
				fmt.Println("json.Marshal error: ", err)
				return
			}

			_, err = masterConn.Write(jsonBytes)
			if err != nil {
				fmt.Println("Error:", err)
				return

			}
		case <-stopch:
			return
		}
	}
}
