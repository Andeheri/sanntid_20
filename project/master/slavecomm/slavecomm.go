package slavecomm

import (
	"encoding/json"
	"fmt"
	"net"
	"project/master/community"
)

func Listener(port int) {

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

		handleSlave(slaveConn, nil, nil)
	}

}

func handleSlave(slaveConn *net.TCPConn, toSlaveCh <-chan json.Marshaler, toMasterCh chan<- interface{}) {
	defer slaveConn.Close()

	remoteIP := slaveConn.RemoteAddr().String()

	tcpReadCh := make(chan []byte)
	quitCh := make(chan struct{})

	go tcpReader(slaveConn, tcpReadCh, quitCh)

	for {
		select {
		case <-quitCh:
			toMasterCh <- community.SlaveMessage{
				SenderIP: remoteIP,
				Payload:  "disconnected",
			}
			return
		case data := <-toSlaveCh:
			jsonBytes, err := json.Marshal(data)
			if err != nil {
				fmt.Println("json.Marshal error: ", err)
				continue
			}

			_, err = slaveConn.Write(jsonBytes)
			if err != nil {
				fmt.Println("Error:", err)
				fmt.Println("Closing connection to", slaveConn.RemoteAddr().String())
				quitCh <- struct{}{}
				return
			}
		case msg := <-tcpReadCh:
			var data interface{}
			var elevatorState community.ElevatorState
			var buttonEvent community.ButtonEvent
			var orderComplete community.OrderComplete

			if err := json.Unmarshal(msg, &elevatorState); err == nil {
				data = elevatorState
			} else if err = json.Unmarshal(msg, &buttonEvent); err == nil {
				data = buttonEvent
			} else if err = json.Unmarshal(msg, &orderComplete); err == nil {
				data = orderComplete
			} else {
				fmt.Println("json.Unmarshal error (no matching data types) : ", err)
				continue
			}

			if data != nil {
				toMasterCh <- community.SlaveMessage{
					SenderIP: remoteIP,
					Payload:  data,
				}
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
			fmt.Println("Error:", err)
			fmt.Println("Closing connection to", slaveConn.RemoteAddr().String())
			quitCh <- struct{}{}
			return
		}
		ch <- buffer[:n]
	}

}


