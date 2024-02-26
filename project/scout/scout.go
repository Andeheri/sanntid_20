package scout

import (
	"bytes"
	. "elevator/constants"
	"elevator/scout/conn"
	"fmt"
	"net"
	"strings"
	"time"
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

func BroadcastInfo(localAddr string, send_broadcast_channel <-chan string) {
	bcastConn := conn.DialBroadcastUDP(UDP_PORT)
	defer bcastConn.Close()
	addr, _ := net.ResolveUDPAddr("udp4", fmt.Sprintf("255.255.255.255:%d", UDP_PORT))

	time.Sleep(250 * time.Millisecond) // To ensure the listen-thread also has a connection

	for broadcast_message := range send_broadcast_channel {
		bcastConn.WriteTo([]byte(localAddr+": "+broadcast_message), addr)
	}
}

func ListenForInfo(recieve_broadcast_channel chan<- string) {
	const bufSize = 1024

	conn := conn.DialBroadcastUDP(UDP_PORT)
	defer conn.Close()

	for {
		buff := make([]byte, bufSize)
		_, _, e := conn.ReadFrom(buff)
		if e != nil {
			fmt.Printf("Error when recieving with UDP on port %d: \"%+v\"\n", UDP_PORT, e)
		} else {
			// Trim trailing \0
			index := bytes.IndexByte(buff, 0)
			if index == -1 {
				// Handle error: null terminator not found
				// Fallback option could be to use the whole buffer,
				// but this could lead to incorrect behavior if the buffer is not cleaned properly
				index = len(buff)
			}

			recieved_message := string(buff[:index])
			recieve_broadcast_channel <- recieved_message
		}
	}
}

func SendKeepAliveMessage(local_IP string, delta_t time.Duration) {
	// Sends keep-alive messages, updating all elevators that it is active, and maybe trigger a reelection of master
	bcastConn := conn.DialBroadcastUDP(UDP_PORT)
	defer bcastConn.Close()
	addr, _ := net.ResolveUDPAddr("udp4", fmt.Sprintf("255.255.255.255:%d", UDP_PORT))

	for {
		bcastConn.WriteTo([]byte(local_IP+": "+Keep_alive), addr)
		time.Sleep(delta_t)
	}
}

func TrackMissedKeepAliveMessages(delta_t time.Duration, num_keep_alive int, keep_alive_recieve_channel <-chan string, keep_alive_transmit_channel chan<- string) {
	knownMap := make(map[string]int)
	aliveMap := make(map[string]struct{})
	for {
		select {
		case ip_addr := <-keep_alive_recieve_channel:
			aliveMap[ip_addr] = struct{}{}
			knownMap[ip_addr] = num_keep_alive
		default:
			for ip := range aliveMap {
				fmt.Printf("Alive: %s\n", ip)
			}
			// Compute the difference.
			not_responding := []string{}
			for ip := range knownMap {
				fmt.Printf("Known: %s\n", ip)
				if _, found := aliveMap[ip]; !found {
					knownMap[ip] -= 1
					fmt.Printf("%s not responding\nknownmap: %d\n", ip, knownMap[ip])
					if knownMap[ip] == 0 {
						not_responding = append(not_responding, ip)
					}
				}
			}

			for _, not_responded := range not_responding {
				keep_alive_transmit_channel <- not_responded
			}

			// Clear the aliveMap by deleting all keys:
			for key := range aliveMap {
				delete(aliveMap, key)
			}
			time.Sleep(delta_t) // Twice because Nyquist sampling theorem
		}
	}
}
