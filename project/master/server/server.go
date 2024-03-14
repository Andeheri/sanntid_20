package server

import (
	"fmt"
	"net"
	"project/mscomm"
	"reflect"
)

func Listen(port int) (*net.TCPListener, error) {
	localAddr, err := net.ResolveTCPAddr("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		return nil, err
	}

	listener, err := net.ListenTCP("tcp", localAddr)
	if err != nil {
		listener.Close()
		return nil, err
	}
	return listener, nil
}

// Intended to run as a goroutine
// Returns when listener is closed
func Acceptor(listener *net.TCPListener, fromSlaveCh chan mscomm.Package, connectionEventCh chan mscomm.ConnectionEvent) {

	allowedTypes := [...]reflect.Type{
		reflect.TypeOf(mscomm.ElevatorState{}),
		reflect.TypeOf(mscomm.ButtonPressed{}),
		reflect.TypeOf(mscomm.OrderComplete{}),
		reflect.TypeOf(mscomm.SyncOK{}),
		reflect.TypeOf(mscomm.HallRequests{}),
	}

	for {

		slaveConn, err := listener.AcceptTCP()
		if err != nil {
			//listener closed. Terminating
			return
		}

		slaveAddr := slaveConn.RemoteAddr().String()

		toSlaveCh := make(chan interface{})
		slaveQuitCh := make(chan struct{})
		go mscomm.TCPSender(slaveConn, toSlaveCh, slaveQuitCh, nil)
		go mscomm.TCPReader(slaveConn, fromSlaveCh, connectionEventCh, allowedTypes[:]...)

		connectionEventCh <- mscomm.ConnectionEvent{
			Connected: true,
			Addr:      slaveAddr,
			Ch:        toSlaveCh,
			QuitCh:    slaveQuitCh,
		}
	}

}
