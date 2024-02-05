package scout

import (
	"encoding/binary"
	"fmt"
	"net"
	"time"
)

const (
	UDP_PORT = 23456
)

func uint64ToBytes(num uint64) []byte {
	bytes := make([]byte, 8)
	binary.BigEndian.PutUint64(bytes, num)
	return bytes
}

func BroadcastInfo() {

	bcastAddr, err := net.ResolveUDPAddr("udp4", fmt.Sprint("10.100.23.255:", UDP_PORT))
	if err != nil {
		fmt.Println("Error resolving broadcast address:", err)
		return
	}

	bcastConn, err := net.DialUDP("udp4", nil, bcastAddr)
	if err != nil {
		fmt.Println("Error dialing broadcast address:", err)
		return
	}
	defer bcastConn.Close()

	localAddr := bcastConn.LocalAddr().(*net.UDPAddr)

	addrString := fmt.Sprint(localAddr.IP, ":", localAddr.Port)

	for {
		bcastConn.Write([]byte(addrString))
		time.Sleep(1 * time.Second)
	}
}

func ListenForInfo() {

	listenAddr, err := net.ResolveUDPAddr("udp4", fmt.Sprint(":", UDP_PORT))
	if err != nil {
		fmt.Println("Error resolving broadcast address:", err)
		return
	}

	listenConn, err := net.ListenUDP("udp4", listenAddr)
	if err != nil {
		fmt.Println("Error dialing broadcast address:", err)
		return
	}
	defer listenConn.Close()

	for {
		buff := make([]byte, 1024)
		listenConn.ReadFromUDP(buff)
		fmt.Println(string(buff))
	}

}
