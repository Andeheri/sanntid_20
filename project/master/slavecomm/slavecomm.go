package slavecomm

import (
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"reflect"
	"time"
)

type SlaveMessage struct {
	Addr    string
	Payload interface{}
}

type sendRequest struct {
	msg     SlaveMessage
	errorCh chan error
}

type getSlavesRequest struct {
	returnCh chan []string
}

type addSlaveRequest struct {
	addr string
	ch   chan interface{}
}
type removeSlaveRequest struct {
	addr string
}

type tcpPackage struct {
	Datatype  reflect.Type
	jsonBytes []byte
}

var managerCh chan interface{}

func Send(addr string, data interface{}, timeoutms int) error {

	request := sendRequest{
		msg: SlaveMessage{
			Addr:    addr,
			Payload: data,
		},
		errorCh: make(chan error),
	}

	managerCh <- request
	select {
	case err := <-request.errorCh:
		if err != nil {
			return err
		}
		return nil

	case <-time.After(time.Duration(timeoutms) * time.Millisecond):
		return errors.New("slavecomm.Send() Timeout")
	}

}

func Slaves(timeoutms int) ([]string, error) {

	request := getSlavesRequest{
		returnCh: make(chan []string),
	}

	managerCh <- request

	select {
	case slaves := <-request.returnCh:
		return slaves, nil
	case <-time.After(time.Duration(timeoutms) * time.Millisecond):
		return nil, errors.New("slavecomm.Slaves() Timeout")
	}
}

func Manager() {
	addrToCh := make(map[string]chan interface{})
	managerCh = make(chan interface{})

	for request := range managerCh {
		switch req := request.(type) {
		case sendRequest:
			addrToCh[req.msg.Addr] <- req.msg.Payload
			req.errorCh <- nil
		case getSlavesRequest:
			slaves := make([]string, 0, len(addrToCh))
			for addr := range addrToCh {
				slaves = append(slaves, addr)
			}
			req.returnCh <- slaves
		case removeSlaveRequest:
			delete(addrToCh, req.addr)
		case addSlaveRequest:
			addrToCh[req.addr] = req.ch
		default:
			fmt.Println("ERROR: Unknown manager request")
		}
	}
}

func Listener(port int, fromSlaveCh chan<- interface{}) {

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

		go handleSlave(slaveConn, fromSlaveCh)
	}

}

func handleSlave(slaveConn *net.TCPConn, fromSlaveCh chan<- interface{}) {

	fmt.Println("Connected to", slaveConn.RemoteAddr().String())

	slaveAddr := slaveConn.RemoteAddr().String()

	tcpReadCh := make(chan []byte)
	quitCh := make(chan struct{})
	toSlaveCh := make(chan interface{})

	managerCh <- addSlaveRequest{addr: slaveAddr, ch: toSlaveCh}

	defer func() {
		slaveConn.Close()
		managerCh <- removeSlaveRequest{addr: slaveAddr}
	}()

	go tcpReader(slaveConn, tcpReadCh, quitCh)

	for {
		select {
		case <-quitCh:
			fmt.Println("Closing connection to", slaveConn.RemoteAddr().String())
			return
		case data := <-toSlaveCh:
			jsonBytesPayload, err := json.Marshal(data)
			if err != nil {
				fmt.Println("json.Marshal error: ", err)
				continue
			}
			tcpPackage := tcpPackage{
				Datatype:  reflect.TypeOf(data),
				jsonBytes: jsonBytesPayload,
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
			var tcpPack tcpPackage

			//this unmarshaling no work :(
			err := json.Unmarshal(msg, &tcpPack)
			if err != nil || tcpPack.Datatype == nil {
				fmt.Println("received invalid TCP Package ", err)
				continue
			}

			data := reflect.New(tcpPack.Datatype).Interface()
			err = json.Unmarshal(tcpPack.jsonBytes, &data)
			if err != nil {
				fmt.Println("payload (jsonBytes) of TCP Package is invalid", err)
				continue
			}

			if data != nil {
				fromSlaveCh <- SlaveMessage{
					Addr:    slaveAddr,
					Payload: data,
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
			fmt.Println("net.Read Error:", err)
			quitCh <- struct{}{}
			return
		}
		ch <- buffer[:n]
	}

}

func stringToType(s string) reflect.Type {
	switch s {
	case "ElevatorState":
		return reflect.TypeOf(ElevatorState{})
	}
}
