package slavecomm

import (
	"encoding/json"
	"fmt"
	"log"
	"net"
	"project/commontypes"
	"reflect"
	"time"
)

type Package struct {
	Addr    string
	Payload interface{}
}

type ConnectionEvent struct {
	Connected bool
	Addr      string
	Ch        chan interface{}
}

func Listener(port int, fromSlaveCh chan Package, connectionEventCh chan ConnectionEvent) {

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

	for {

		slaveConn, err := listener.AcceptTCP()
		if err != nil {
			fmt.Println("Error:", err)
			continue
		}

		slaveAddr := slaveConn.RemoteAddr().String()

		toSlaveCh := make(chan interface{})
		go tcpSender(slaveConn, toSlaveCh)
		go tcpReader(slaveConn, fromSlaveCh, connectionEventCh)

		connectionEventCh <- ConnectionEvent{
			Connected: true,
			Addr:      slaveAddr,
			Ch:        toSlaveCh,
		}
	}

}

func tcpReader(conn *net.TCPConn, ch chan<- Package, connEventCh chan<- ConnectionEvent) {
	defer conn.Close()
	addr := conn.RemoteAddr().String()

	decoder := json.NewDecoder(conn)
	for {
		ttj := commontypes.TypeTaggedJSON{}
		if err := decoder.Decode(&ttj); err != nil {
			//Probably disconnected
			if connEventCh != nil {
				connEvent := ConnectionEvent{
					Connected: false,
					Addr:      addr,
				}
				select {
				case connEventCh <- connEvent:
				case <-time.After(1 * time.Second):
					log.Println("Noone reading from connEventCh")
				}

			}
			return
		}

		object, err := ttj.ToObject(
			reflect.TypeOf(commontypes.ElevatorState{}),
			reflect.TypeOf(commontypes.ButtonPressed{}),
			reflect.TypeOf(commontypes.OrderComplete{}),
			reflect.TypeOf(commontypes.SyncOK{}),
			reflect.TypeOf(commontypes.HallRequests{}),
		)

		if err != nil {
			fmt.Println("ttj.ToObject error: ", err)
			continue
		}

		ch <- Package{
			Addr:    addr,
			Payload: object,
		}

	}

}

func tcpSender(conn *net.TCPConn, ch <-chan interface{}) {
	defer conn.Close()

	encoder := json.NewEncoder(conn)

	for {
		data, isOpen := <-ch
		if !isOpen {
			log.Println("Channel closed")
			return
		}

		ttj, err := commontypes.NewTypeTaggedJSON(data)
		if err != nil {
			log.Println(err)
			continue
		}

		if err := encoder.Encode(ttj); err != nil {
			log.Println(err)
			return
		}
	}

}
