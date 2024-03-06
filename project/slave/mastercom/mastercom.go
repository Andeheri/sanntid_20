package mastercom

import (
	"fmt"
	"net"
	"project/mscomm"
	"project/slave/elevio"
	"reflect"
	"time"
)

type MasterChannels struct {
	ButtonPress  chan elevio.ButtonEvent
	ClearRequest chan elevio.ButtonEvent

	AssignedRequests chan mscomm.AssignedRequests
	HallLights       chan mscomm.Lights
	RequestedState   chan struct{}

	State chan mscomm.ElevatorState

	Sender chan interface{}
}

var hallRequests mscomm.HallRequests = mscomm.HallRequests{{false, false}, {false, false}, {false, false}, {false, false}}

func MasterCommunication(masterAddress *net.TCPAddr, chans *MasterChannels, stopch <-chan struct{}) {

	masterConn, err := net.DialTCP("tcp", nil, masterAddress)
	if err != nil {
		fmt.Println("Error connecting to new master:", err)
		time.Sleep(10 * time.Millisecond)
		masterConn, err = net.DialTCP("tcp", nil, masterAddress)
		if err != nil {
			// elevio.SetMotorDirection(elevio.MD_Stop)
			panic("Error connecting to new master")
		}
	}
	defer func() {
		masterConn.Close()
	}()

	allowedTypes := [...]reflect.Type{
		reflect.TypeOf(mscomm.RequestState{}),
		reflect.TypeOf(mscomm.RequestHallRequests{}),
		reflect.TypeOf(mscomm.SyncRequests{}),
		reflect.TypeOf(mscomm.Lights{}),
		reflect.TypeOf(mscomm.AssignedRequests{}),
	}

	fromMasterCh := make(chan mscomm.Package)

	go mscomm.TCPSender(masterConn, chans.Sender)
	go mscomm.TCPReader(masterConn, fromMasterCh, nil, allowedTypes[:]...)

	for {
		select {
		case a := <-chans.ButtonPress:
			fmt.Println(a, "sender button press melding")
			pressed := mscomm.ButtonPressed{Floor: a.Floor, Button: int(a.Button)}
			select {
			case chans.Sender <- pressed:
			case <-time.After(10 * time.Millisecond):
			}

		case a := <-chans.ClearRequest:
			fmt.Println(a, "sender clear request melding")
			completedOrder := mscomm.OrderComplete{Floor: a.Floor, Button: int(a.Button)}
			select {
			case chans.Sender <- completedOrder:
			case <-time.After(10 * time.Millisecond):
			}

		case a := <-chans.State:
			fmt.Println("Sending state message to master")
			select {
			case chans.Sender <- a:
			case <-time.After(10 * time.Millisecond):
			}

		case a := <-fromMasterCh:
			switch reflect.TypeOf(a.Payload) {
			case reflect.TypeOf(mscomm.RequestState{}):
				fmt.Println("State requested by master")
				chans.RequestedState <- struct{}{}
			case reflect.TypeOf(mscomm.RequestHallRequests{}):
				fmt.Println("Master requested Hallrequests")
				select {
				case chans.Sender <- hallRequests:
				case <-time.After(10 * time.Millisecond):
				}
			case reflect.TypeOf(mscomm.SyncRequests{}):
				fmt.Println("Received Syncrequests")
				hallRequests = a.Payload.(mscomm.SyncRequests).Requests
				select {
				case chans.Sender <- mscomm.SyncOK{Id: a.Payload.(mscomm.SyncRequests).Id}:
				case <-time.After(10 * time.Millisecond):
				}
			case reflect.TypeOf(mscomm.Lights{}):
				fmt.Println("Received halllights")
				chans.HallLights <- a.Payload.(mscomm.Lights)
				//should also set hallrequests as lights are higher rank
				hallRequests = mscomm.HallRequests(a.Payload.(mscomm.Lights))
			case reflect.TypeOf(mscomm.AssignedRequests{}):
				fmt.Println("Received assigned requests")
				chans.AssignedRequests <- a.Payload.(mscomm.AssignedRequests)
			default:
				fmt.Println("received invalid type on fromMasterCh", a)
				continue
			}
		case <-stopch:
			close(chans.Sender)
			return
		}
	}
}
