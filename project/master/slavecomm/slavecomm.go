package slavecomm

import (
	"encoding/json"
	"fmt"
	"net"
	"project/master/community"
)

func Receiver(slaveConn *net.TCPConn, ch chan<- community.SlaveMessage) {
	remoteIP := slaveConn.RemoteAddr().String()
	buffer := make([]byte, 1024)

	for {
		// Read data from the client
		n, err := slaveConn.Read(buffer)
		if err != nil {
			fmt.Println("Error:", err)
			return
		}

		//attempt to unmarshal to different formats:
		var data interface{}
		var elevatorState community.ElevatorState
		var buttonEvent community.ButtonEvent
		var orderComplete community.OrderComplete

		if err = json.Unmarshal(buffer[:n], &elevatorState); err == nil {
			data = elevatorState
		} else if err = json.Unmarshal(buffer[:n], &buttonEvent); err == nil {
			data = buttonEvent
		} else if err = json.Unmarshal(buffer[:n], &orderComplete); err == nil {
			data = orderComplete
		} else {
			fmt.Println("json.Unmarshal error (no matching data types) : ", err)
			return
		}

		if data != nil {
			ch <- community.SlaveMessage{
				SenderIP: remoteIP,
				Payload:  data,
			}
		}

	}

}

func Sender(slaveConn *net.TCPConn, ch <-chan json.Marshaler) {

	for {
		data := <-ch
		jsonBytes, err := json.Marshal(data)
		if err != nil {
			fmt.Println("json.Marshal error: ", err)
			return
		}

		_, err = slaveConn.Write(jsonBytes)
		if err != nil {
			fmt.Println("Error:", err)
			return
		}

	}

}
