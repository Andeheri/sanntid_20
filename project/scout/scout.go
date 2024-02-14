package scout

import (
	"encoding/binary"
	"fmt"
	"net"
	"strings"
	"syscall"
	"os"
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

func BroadcastInfo(send_broadcast_channel <- chan string) {
	bcastConn := DialBroadcastUDP(UDP_PORT)
	defer bcastConn.Close()
	addr, _ := net.ResolveUDPAddr("udp4", fmt.Sprintf("255.255.255.255:%d", UDP_PORT))

	localAddr, err := LocalIP()
	if err != nil {
		fmt.Println("Error finding local address:", err)
		return
	}

	// Attempt to get broadcast port
	// localConn := bcastConn.if ()LocalAddr().(*net.UDPAddr)  // For local testing (all units have same IP but different ports)
	// addrString := fmt.Sprint(localAddr, ":", localConn.Port)

	for {
		select{
		case broadcast_message := <- send_broadcast_channel:
			bcastConn.WriteTo([]byte(localAddr + ": " + broadcast_message), addr)
		}
	}
}

func ListenForInfo(recieve_broadcast_channel chan <- string) {
	listenAddr, err := net.ResolveUDPAddr("udp4", fmt.Sprint(":", UDP_PORT))
	if err != nil {
		fmt.Println("Error resolving listenAddr:", err)
		return
	}

	listenConn, err := net.ListenUDP("udp4", listenAddr)
	if err != nil {
		fmt.Println("Error listening to listenAddr:", err)
		return
	}
	defer listenConn.Close()

	var recieved_message string
	for {
		buff := make([]byte, 1024)
		listenConn.ReadFromUDP(buff)  // Dumps recieved message into buff

		recieved_message = string(buff)
		recieve_broadcast_channel <- recieved_message
	}

}
