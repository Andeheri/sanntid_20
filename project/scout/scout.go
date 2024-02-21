package scout

import (
	"encoding/binary"
	"fmt"
	"net"
	"strings"
	"syscall"
	"os"
	"time"
	. "elevator/constants"
)

const (
	UDP_PORT = 23456
)

var localIP string

func LocalIP() (string, error) {
	if localIP == "" {
		conn, err := net.DialTCP("tcp4", nil, &net.TCPAddr{IP: []byte{8, 8, 8, 8}, Port: 53})
		if err != nil {
			return "", err
		}
		defer conn.Close()
		localIP = strings.Split(conn.LocalAddr().String(), ":")[0]
	}
	return localIP, nil
}

func DialBroadcastUDP(port int) net.PacketConn {
	s, err := syscall.Socket(syscall.AF_INET, syscall.SOCK_DGRAM, syscall.IPPROTO_UDP)
	if err != nil { fmt.Println("Error: Socket:", err) }
	syscall.SetsockoptInt(s, syscall.SOL_SOCKET, syscall.SO_REUSEADDR, 1)
	if err != nil { fmt.Println("Error: SetSockOpt REUSEADDR:", err) }
	syscall.SetsockoptInt(s, syscall.SOL_SOCKET, syscall.SO_BROADCAST, 1)
	if err != nil { fmt.Println("Error: SetSockOpt BROADCAST:", err) }
	syscall.Bind(s, &syscall.SockaddrInet4{Port: port})
	if err != nil { fmt.Println("Error: Bind:", err) }

	f := os.NewFile(uintptr(s), "")
	conn, err := net.FilePacketConn(f)
	if err != nil { fmt.Println("Error: FilePacketConn:", err) }
	f.Close()

	return conn
}


func uint64ToBytes(num uint64) []byte {
	bytes := make([]byte, 8)
	binary.BigEndian.PutUint64(bytes, num)
	return bytes
}

func BroadcastInfo(localAddr string, send_broadcast_channel <- chan string) {
	bcastConn := DialBroadcastUDP(UDP_PORT)
	defer bcastConn.Close()
	addr, _ := net.ResolveUDPAddr("udp4", fmt.Sprintf("255.255.255.255:%d", UDP_PORT))

	time.Sleep(250 * time.Millisecond) // To ensure the listen-thread also has a connection

	for {
		select{
		case broadcast_message := <- send_broadcast_channel:
			bcastConn.WriteTo([]byte(localAddr + ": " + broadcast_message), addr)
		}
	}
}

func ListenForInfo(recieve_broadcast_channel chan <- string) {
	const bufSize = 1024

	conn := DialBroadcastUDP(UDP_PORT)
	defer conn.Close()

	for {
		buff := make([]byte, bufSize)
		_, _, e := conn.ReadFrom(buff)
		if e != nil {
			fmt.Printf("Error when recieving with UDP on port %d: \"%+v\"\n", UDP_PORT, e)
		}else{
			recieved_message := string(buff)
			recieve_broadcast_channel <- recieved_message
		}
	}
}

func SendKeepAliveMessage(local_IP string, delta_t time.Duration){
	// Sends keep-alive messages, updating all elevators that it is active, and maybe trigger a reelection of master
	bcastConn := DialBroadcastUDP(UDP_PORT)
	defer bcastConn.Close()
	addr, _ := net.ResolveUDPAddr("udp4", fmt.Sprintf("255.255.255.255:%d", UDP_PORT))

	for {
		bcastConn.WriteTo([]byte(local_IP + ": " + Keep_alive), addr)
		time.Sleep(delta_t)
	}
}
