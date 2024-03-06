package mastercom

import (
	"fmt"
	"net"
	"project/mscomm"
	"project/slave/elevio"
	"project/slave/fsm"
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

	FromMasterCh chan mscomm.Package
	Sender chan interface{}
}

var hallRequests mscomm.HallRequests = mscomm.HallRequests{{false, false}, {false, false}, {false, false}, {false, false}}

func StartUp(address string, chans *MasterChannels) *net.TCPConn {

	masterAddress, err := net.ResolveTCPAddr("tcp", address)
	if err != nil {
		fmt.Println("Error resolving TCP address from master:", err)
	}

	masterConn, err := net.DialTCP("tcp", nil, masterAddress)
	if err != nil {
		fmt.Println("Error connecting to new master:", err)
		time.Sleep(10 * time.Millisecond)
		masterConn, err = net.DialTCP("tcp", nil, masterAddress)
		if err != nil {
			elevio.SetMotorDirection(elevio.MD_Stop)
			panic("Error connecting to new master")
		}
	}

	allowedTypes := [...]reflect.Type{
		reflect.TypeOf(mscomm.RequestState{}),
		reflect.TypeOf(mscomm.RequestHallRequests{}),
		reflect.TypeOf(mscomm.SyncRequests{}),
		reflect.TypeOf(mscomm.Lights{}),
		reflect.TypeOf(mscomm.AssignedRequests{}),
	}

	go mscomm.TCPSender(masterConn, chans.Sender)
	go mscomm.TCPReader(masterConn, chans.FromMasterCh, nil, allowedTypes[:]...)

	return masterConn
}

func HandleMessage(payload interface{}, chans *MasterChannels, doorTimer *time.Timer) {
	switch reflect.TypeOf(payload) {
	case reflect.TypeOf(mscomm.RequestState{}):
		fmt.Println("State requested by master")
		select {
		case chans.Sender <- fsm.GetState():
		case <-time.After(10 * time.Millisecond):
		}
	case reflect.TypeOf(mscomm.RequestHallRequests{}):
		fmt.Println("Master requested Hallrequests")
		select {
		case chans.Sender <- hallRequests:
		case <-time.After(10 * time.Millisecond):
		}
	case reflect.TypeOf(mscomm.SyncRequests{}):
		fmt.Println("Received Syncrequests")
		hallRequests = payload.(mscomm.SyncRequests).Requests
		select {
		case chans.Sender <- mscomm.SyncOK{Id: payload.(mscomm.SyncRequests).Id}:
		case <-time.After(10 * time.Millisecond):
		}
	case reflect.TypeOf(mscomm.Lights{}):
		fmt.Println("Received halllights", payload.(mscomm.Lights))
		fsm.Elev.HallLights = payload.(mscomm.Lights)
		fsm.SetAllLights(&fsm.Elev)
		//should also set hallrequests as lights are higher rank
		hallRequests = mscomm.HallRequests(payload.(mscomm.Lights))
	case reflect.TypeOf(mscomm.AssignedRequests{}):
		fmt.Println("Received assigned requests")
		fsm.RequestsClearAll()
		fsm.RequestsSetAll(payload.(mscomm.AssignedRequests), doorTimer, chans.Sender)
		
	default:
		fmt.Println("received invalid type on fromMasterCh", payload)
	}
}
