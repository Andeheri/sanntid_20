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

func ConnManager(masterAddressCh <-chan string, senderCh chan interface{}, fromMasterCh chan<- mscomm.Package) {
	var currentMasterAddress string
	var masterTCPaddr *net.TCPAddr
	var masterConn *net.TCPConn
	var err error
	var attempts int = 0
	masterDisconnectCh := make(chan mscomm.ConnectionEvent)
	retry := time.NewTimer(-1)
	retry.Stop()

	for {
		select {
			case newMasterAddress := <-masterAddressCh:
				if newMasterAddress != currentMasterAddress {
					rblog.White.Println("Master address changed to:", newMasterAddress)
					currentMasterAddress = newMasterAddress
					if masterConn != nil {
						masterConn.Close()
						//send something to close TCPSender
						select {
						case senderCh <- struct{}{}:
						case <-time.After(10 * time.Millisecond):
							rblog.Yellow.Println("Timed out on sending close of TCPSender")
						}
					}
					masterTCPaddr, err = net.ResolveTCPAddr("tcp", currentMasterAddress)
					if err != nil {
						rblog.Red.Println("Error resolving TCP address from master:", err)
					}
					retry.Reset(0)
				}
			
			case disconnect := <-masterDisconnectCh:
				rblog.White.Println("Disconnected from master:", disconnect.Addr)
				rblog.White.Println("Attempting reconnect",currentMasterAddress)
				select {
				case senderCh <- struct{}{}:
				case <-time.After(10 * time.Millisecond):
					rblog.Yellow.Println("Timed out on sending close of TCPSender")
				}
				if disconnect.Addr == currentMasterAddress {
					rblog.Yellow.Println("Disconnected from master attempting reconnect")
					retry.Reset(0)
				}

			case <-retry.C:
				rblog.White.Println("Attempting dialing to connect to master")
				conn, err := net.DialTCP("tcp", nil, masterTCPaddr)
				if err == nil {
					masterConn = conn
					StartUp(masterConn, senderCh, fromMasterCh, masterDisconnectCh)
					attempts = 0
				} else {
					retry.Reset(50 * time.Millisecond)
					attempts += 1
					if attempts > 20 {
						panic("Failed to connect to master")
					}
				}
		}
	}
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
