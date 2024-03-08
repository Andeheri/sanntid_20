package slavecomm

import (
	"fmt"
	"net"
	"project/mscomm"
	"reflect"
)

func Listener(port int, fromSlaveCh chan mscomm.Package, connectionEventCh chan mscomm.ConnectionEvent) {

	localAddr, err := net.ResolveTCPAddr("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		fmt.Println("Error:", err)
		return
	}

	listener, err := net.ListenTCP("tcp", localAddr)
	if err != nil {
		fmt.Println("Error:", err)
		return
	}
	defer listener.Close()

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
			fmt.Println("Error:", err)
			continue
		}

		slaveAddr := slaveConn.RemoteAddr().String()

		toSlaveCh := make(chan interface{})
		go mscomm.TCPSender(slaveConn, toSlaveCh)
		go mscomm.TCPReader(slaveConn, fromSlaveCh, connectionEventCh, allowedTypes[:]...)

		connectionEventCh <- mscomm.ConnectionEvent{
			Connected: true,
			Addr:      slaveAddr,
			Ch:        toSlaveCh,
		}
	}

}
