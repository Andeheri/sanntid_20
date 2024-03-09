package mastercom

import (
	"net"
	"project/mscomm"
	"project/rblog"
	"project/slave/fsm"
	"reflect"
	"time"
)

var hallRequests mscomm.HallRequests = mscomm.HallRequests{{false, false}, {false, false}, {false, false}, {false, false}}

func EstablishTCPConnection(address string, connCh chan<- *net.TCPConn) {

	masterAddress, err := net.ResolveTCPAddr("tcp", address)
	if err != nil {
		rblog.Red.Println("Error resolving TCP address from master:", err)
	}

	attempts := 5
	for i := 0; i < attempts; i++ {
		masterConn, err := net.DialTCP("tcp", nil, masterAddress)
		if err == nil {
			connCh <- masterConn
			return
		} else {
			time.Sleep(50 * time.Millisecond)
		}

	}
	connCh <- nil
}

func StartUp(masterConn *net.TCPConn, senderCh <-chan interface{}, fromMasterCh chan<- mscomm.Package, masterDisconnect chan<- mscomm.ConnectionEvent) {

	allowedTypes := [...]reflect.Type{
		reflect.TypeOf(mscomm.RequestState{}),
		reflect.TypeOf(mscomm.RequestHallRequests{}),
		reflect.TypeOf(mscomm.SyncRequests{}),
		reflect.TypeOf(mscomm.Lights{}),
		reflect.TypeOf(mscomm.AssignedRequests{}),
	}

	go mscomm.TCPSender(masterConn, senderCh)
	go mscomm.TCPReader(masterConn, fromMasterCh, masterDisconnect, allowedTypes[:]...)
}

func HandleMessage(payload interface{}, senderCh chan<- interface{}, doorTimer *time.Timer, inbetweenFloorsTimer *time.Timer) {

	switch reflect.TypeOf(payload) {
	case reflect.TypeOf(mscomm.RequestState{}):
		select {
		case senderCh <- fsm.GetState():
		case <-time.After(10 * time.Millisecond):
			rblog.Yellow.Println("Sending state timed out")
		}

	case reflect.TypeOf(mscomm.RequestHallRequests{}):
		select {
		case senderCh <- hallRequests:
			rblog.White.Println("Sending hallrequests")
		case <-time.After(10 * time.Millisecond):
			rblog.Yellow.Println("Sending hallrequests timed out")
		}

	case reflect.TypeOf(mscomm.SyncRequests{}):
		hallRequests = payload.(mscomm.SyncRequests).Requests
		select {
		case senderCh <- mscomm.SyncOK{Id: payload.(mscomm.SyncRequests).Id}:
		case <-time.After(10 * time.Millisecond):
			rblog.Yellow.Println("Sending SyncOK timed out")
		}

	case reflect.TypeOf(mscomm.Lights{}):
		fsm.Elev.HallLights = payload.(mscomm.Lights)
		fsm.SetAllLights(&fsm.Elev)
		//should also set hallrequests as lights are higher rank
		hallRequests = mscomm.HallRequests(payload.(mscomm.Lights))

	case reflect.TypeOf(mscomm.AssignedRequests{}):
		fsm.RequestsClearAll()
		fsm.RequestsSetAll(payload.(mscomm.AssignedRequests), doorTimer, inbetweenFloorsTimer, senderCh)

	default:
		rblog.Red.Println("Slave received invalid type on fromMasterCh", payload)
	}
}
