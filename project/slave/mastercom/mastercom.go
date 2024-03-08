package mastercom

import (
	"fmt"
	"log"
	"net"
	"project/mscomm"
	"project/slave/elevio"
	"project/slave/fsm"
	"reflect"
	"time"
)

var hallRequests mscomm.HallRequests = mscomm.HallRequests{{false, false}, {false, false}, {false, false}, {false, false}}

func StartUp(address string, senderCh <-chan interface{}, fromMasterCh chan<- mscomm.Package) *net.TCPConn {

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
			panic("Error connecting to new master:")
		}
	}

	allowedTypes := [...]reflect.Type{
		reflect.TypeOf(mscomm.RequestState{}),
		reflect.TypeOf(mscomm.RequestHallRequests{}),
		reflect.TypeOf(mscomm.SyncRequests{}),
		reflect.TypeOf(mscomm.Lights{}),
		reflect.TypeOf(mscomm.AssignedRequests{}),
	}

	go mscomm.TCPSender(masterConn, senderCh)
	go mscomm.TCPReader(masterConn, fromMasterCh, nil, allowedTypes[:]...)

	return masterConn
}

func HandleMessage(payload interface{}, senderCh chan<- interface{}, doorTimer *time.Timer) {

	switch reflect.TypeOf(payload) {
	case reflect.TypeOf(mscomm.RequestState{}):
		select {
		case senderCh <- fsm.GetState():
		case <-time.After(10 * time.Millisecond):
			log.Println("Sending state timed out")
		}

	case reflect.TypeOf(mscomm.RequestHallRequests{}):
		select {
		case senderCh <- hallRequests:
		case <-time.After(10 * time.Millisecond):
			log.Println("Sending hallrequests timed out")
		}

	case reflect.TypeOf(mscomm.SyncRequests{}):
		hallRequests = payload.(mscomm.SyncRequests).Requests
		select {
		case senderCh <- mscomm.SyncOK{Id: payload.(mscomm.SyncRequests).Id}:
		case <-time.After(10 * time.Millisecond):
			log.Println("Sending SyncOK timed out")
		}

	case reflect.TypeOf(mscomm.Lights{}):
		fsm.Elev.HallLights = payload.(mscomm.Lights)
		fsm.SetAllLights(&fsm.Elev)
		//should also set hallrequests as lights are higher rank
		hallRequests = mscomm.HallRequests(payload.(mscomm.Lights))

	case reflect.TypeOf(mscomm.AssignedRequests{}):
		fsm.RequestsClearAll()
		fsm.RequestsSetAll(payload.(mscomm.AssignedRequests), doorTimer, senderCh)

	default:
		fmt.Println("received invalid type on fromMasterCh", payload)
	}
}
