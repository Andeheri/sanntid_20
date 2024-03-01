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
	RequestedState  chan bool

	Log chan bool

	Sender chan json.Marshaler
}

func Master_communication(chans *Master_channels) {
	for {
		select {
		case a := <-chans.ButtonPress:
			fmt.Println(a, "sender button press melding")
			//send message over TCP
		case a := <-chans.ClearRequest:
			fmt.Println(a, "sender clear request melding")
			//send message over TCP
		case a := <-chans.Log:
			fmt.Println(a, "lagrer data fra master")
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

func Receiver(masterConn *net.TCPConn, chans *Master_channels) {

	buffer := make([]byte, 1024)

	for {
		// Read data from the master
		n, err := masterConn.Read(buffer)
		if err != nil {
			fmt.Println("Error:", err)
			return
		}

		var requests [iodevice.N_FLOORS][iodevice.N_BUTTONS]int
		var requestedState bool

		if err = json.Unmarshal(buffer[:n], &requests); err == nil {
			chans.MasterRequests <- requests
		} else if err = json.Unmarshal(buffer[:n], &requestedState); err == nil {
			chans.RequestedState <- requestedState
		} else {
			fmt.Println("json.Unmarshal error (no matching data types) : ", err)
			return
		}
	}

}

func Sender(masterConn *net.TCPConn, ch <-chan json.Marshaler) {
	for {
		data := <-ch
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
	}
}
