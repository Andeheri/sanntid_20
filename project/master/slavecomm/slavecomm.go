package slavecomm

import (
	"encoding/json"
	"fmt"
	"net"
	"project/commontypes"
	"reflect"
)

type SlaveMessage struct {
	Addr    string
	Payload interface{}
}

type ConnectionEvent struct {
	Connected bool
	Addr      string
	Ch        chan interface{}
}

func Listener(port int, fromSlaveCh chan SlaveMessage, connectionEventCh chan ConnectionEvent) {

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

		go handleSlave(slaveConn, fromSlaveCh, connectionEventCh)
	}

}

func handleSlave(slaveConn *net.TCPConn, fromSlaveCh chan<- SlaveMessage, connectionEventCh chan<- ConnectionEvent) {

	fmt.Println("Connected to", slaveConn.RemoteAddr().String())

	slaveAddr := slaveConn.RemoteAddr().String()

	tcpReadCh := make(chan []byte)
	quitCh := make(chan struct{})
	toSlaveCh := make(chan interface{})

	defer func() {
		slaveConn.Close()
	}()

	go tcpReader(slaveConn, tcpReadCh, quitCh)

	for {
		select {
		case <-quitCh:
			fmt.Println("Closing connection to", slaveConn.RemoteAddr().String())
			return
		case data := <-toSlaveCh:
			fmt.Println("ready to send data to slave")
			jsonBytesPayload, err := json.Marshal(data)
			if err != nil {
				fmt.Println("json.Marshal error: ", err)
				continue
			}
			tcpPackage := commontypes.TypeTaggedJSON{
				TypeId: reflect.TypeOf(data).Name(),
				JSON:   jsonBytesPayload,
			}

			jsonBytesPackage, err := json.Marshal(tcpPackage)
			if err != nil {
				fmt.Println("json.Marshal error: ", err)
				continue
			}

			_, err = slaveConn.Write(jsonBytesPackage)
			if err != nil {
				fmt.Println("net.Write Error:", err)
				quitCh <- struct{}{}
				return
			}
		case msg := <-tcpReadCh:
			var ttj commontypes.TypeTaggedJSON

			err := json.Unmarshal(msg, &ttj)
			if err != nil {
				fmt.Println("received invalid TCP Package ", err)
				continue
			}

			object, err := ttj.ToObject(
				reflect.TypeOf(commontypes.ElevatorState{}),
				reflect.TypeOf(commontypes.ButtonPressed{}),
				reflect.TypeOf(commontypes.OrderComplete{}),
				reflect.TypeOf(commontypes.SyncOK{}),
				reflect.TypeOf(commontypes.HallRequests{}),
			)

			if err != nil {
				fmt.Println("ttj.ToValuePtr error: ", err)
				continue
			}

			fromSlaveCh <- SlaveMessage{
				Addr:    slaveAddr,
				Payload: object,
			}

		}
	}
}

func tcpReader(slaveConn *net.TCPConn, ch chan<- []byte, quitCh chan<- struct{}) {
	buffer := make([]byte, 1024)

	for {
		// Read data from the client
		n, err := slaveConn.Read(buffer)
		if err != nil {
			fmt.Println("net.Read Error:", err)
			quitCh <- struct{}{}
			return
		}
		ch <- buffer[:n]
	}

}
