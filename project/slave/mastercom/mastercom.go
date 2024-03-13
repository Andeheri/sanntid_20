package mastercom

import (
	"net"
	"project/mscomm"
	"project/rblog"
	"project/slave/elevio"
	"project/slave/fsm"
	"reflect"
	"time"
)

var hallRequests mscomm.HallRequests = mscomm.HallRequests{{false, false}, {false, false}, {false, false}, {false, false}}

var TCPSendTimeOut time.Duration = 100 * time.Millisecond

func ConnManager(masterAddressCh <-chan string, senderCh chan interface{}, fromMasterCh chan<- mscomm.Package) {
	var currentMasterAddress string
	var masterConn *net.TCPConn
	var attempts int = 0
	masterDisconnectCh := make(chan mscomm.ConnectionEvent)
	senderQuitCh := make(chan struct{})

	const reconnectDelay time.Duration = 50 * time.Millisecond
	const dialTimeout time.Duration = 100 * time.Millisecond

	connAttempt := time.NewTimer(reconnectDelay)
	connAttempt.Stop()

	for {
		select {
		case newMasterAddress := <-masterAddressCh:
			if newMasterAddress != currentMasterAddress {
				rblog.White.Println("Master address changed to:", newMasterAddress)
				currentMasterAddress = newMasterAddress
				if masterConn != nil {
					masterConn.Close()
				}
				connAttempt.Reset(0)
			}

		case disconnect := <-masterDisconnectCh:
			rblog.White.Println("Disconnected from master:", disconnect.Addr)
			if disconnect.Addr == currentMasterAddress {
				rblog.Yellow.Println("Disconnected from master attempting reconnect", currentMasterAddress)
				connAttempt.Reset(0)
			}

		case <-connAttempt.C:
			rblog.White.Println("Attempt at dialing to connect to master")
			conn, err := net.DialTimeout("tcp4", currentMasterAddress, dialTimeout)
			if err == nil {
				rblog.White.Println("Connected to master")
				close(senderQuitCh)
				senderQuitCh = make(chan struct{})
				masterConn = conn.(*net.TCPConn)
				StartUp(masterConn, senderCh, fromMasterCh, masterDisconnectCh, senderQuitCh)
				attempts = 0
			} else {
				connAttempt.Reset(reconnectDelay)
				attempts += 1
				if attempts > 20 {
					elevio.SetMotorDirection(elevio.MD_Stop)
					panic("Failed to connect to master")
				}
			}
		}
	}
}

func StartUp(masterConn *net.TCPConn, senderCh <-chan interface{}, fromMasterCh chan<- mscomm.Package, masterDisconnect chan<- mscomm.ConnectionEvent, senderQuitCh chan struct{}) {

	allowedTypes := [...]reflect.Type{
		reflect.TypeOf(mscomm.RequestState{}),
		reflect.TypeOf(mscomm.RequestHallRequests{}),
		reflect.TypeOf(mscomm.SyncRequests{}),
		reflect.TypeOf(mscomm.Lights{}),
		reflect.TypeOf(mscomm.AssignedRequests{}),
	}

	go mscomm.TCPSender(masterConn, senderCh, senderQuitCh, &TCPSendTimeOut)
	go mscomm.TCPReader(masterConn, fromMasterCh, masterDisconnect, allowedTypes[:]...)
}

func HandleMessage(payload interface{}, senderCh chan<- interface{}, doorTimer *time.Timer, inbetweenFloorsTimer *time.Timer) {

	switch payload := payload.(type) {
	case mscomm.RequestState:
		senderCh <- fsm.GetState()

	case mscomm.RequestHallRequests:
		senderCh <- hallRequests
		rblog.White.Println("Sent hallrequests")
		
	case mscomm.SyncRequests:
		hallRequests = payload.Requests
		senderCh <- mscomm.SyncOK{Id: payload.Id}
		
	case mscomm.Lights:
		fsm.Elev.HallLights = payload
		fsm.SetAllLights(&fsm.Elev)
		//should also set hallrequests as lights are higher rank
		hallRequests = mscomm.HallRequests(payload)

	case mscomm.AssignedRequests:
		fsm.RequestsClearAll()
		fsm.RequestsSetAll(payload, doorTimer, inbetweenFloorsTimer, senderCh)

	default:
		rblog.Red.Println("Slave received invalid type on fromMasterCh", payload)
	}
}
