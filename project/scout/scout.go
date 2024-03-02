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
	if (localIP == "") {
		conn, err := net.DialTCP("tcp4", nil, &net.TCPAddr{IP: []byte{8, 8, 8, 8}, Port: 53})
		if err != nil {
			return "", err
		}
		defer conn.Close()
		localIP = strings.Split(conn.LocalAddr().String(), ":")[0]
	}
	return localIP, nil
}

var broadcastAddr *net.UDPAddr

func getBroadcastAddr(localIP string, UDP_PORT int) *net.UDPAddr{
	if broadcastAddr == nil {
		interfaces, err := net.Interfaces()
		if err != nil {
			return nil
		}

		for _, iface := range interfaces {
			addrs, err := iface.Addrs()
			if err != nil {
				continue
			}
			for _, addr := range addrs {
				ipnet, ok := addr.(*net.IPNet)
				if !ok {
					continue
				}
				if ipnet.IP.String() == localIP {
					// Check if this interface has a broadcast address
					if iface.Flags&net.FlagBroadcast != 0 {
						// Extract and return the broadcast address
						parts := strings.Split(ipnet.IP.String(), ".")
						parts[3] = "255" // Set last octet to 255 for broadcast
						broadcastAddr := strings.Join(parts, ".")
						addr, _ := net.ResolveUDPAddr("udp4", fmt.Sprintf("%s:%d", broadcastAddr, UDP_PORT))
						return addr
					}
				}
			}
		}
		return nil
	}else{
		return broadcastAddr
	}
}

func BroadcastInfo(localIP string, send_broadcast_channel <-chan string) {
	// May be used later. Isn't used atm
	bcastConn := conn.DialBroadcastUDP(UDP_PORT)
	defer bcastConn.Close()

	var bcastAddr *net.UDPAddr

	for broadcast_message := range send_broadcast_channel {
		bcastAddr = getBroadcastAddr(localIP, UDP_PORT)
		if bcastAddr != nil {
			bcastConn.WriteTo([]byte(localIP+": "+broadcast_message), bcastAddr)
		}
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

	var bcastAddr *net.UDPAddr
	for{
		bcastAddr = getBroadcastAddr(localIP, UDP_PORT)
		if bcastAddr != nil {
			bcastConn.WriteTo([]byte(local_IP+": "+Keep_alive), bcastAddr)
		}
		time.Sleep(delta_t)
	}
}

func TrackMissedKeepAliveMessages(delta_t time.Duration, num_keep_alive int, keep_alive_receive_channel <-chan string, keep_alive_transmit_channel chan<- string) {
	knownMap := make(map[string]int) // Known IP-addresses with number keeping track of 'aliveness'
	aliveMap := make(map[string]struct{})  // IP-addresses that sent keep-alive over UDP
	timer := time.NewTicker(delta_t)  // Timer to check for keep-alive messages
	defer timer.Stop()

	for {
		select {
		case ip_addr := <-keep_alive_receive_channel:
			// Received keep alive message
			aliveMap[ip_addr] = struct{}{}
			knownMap[ip_addr] = num_keep_alive

		case <-timer.C:
			// Timer fired due to delta_t duration passing
			// Compute the difference.
			not_responding := []string{}
			for ip, count := range knownMap {
				if _, found := aliveMap[ip]; !found {
					count -= 1
					knownMap[ip] = count
					if count <= 0 {
						not_responding = append(not_responding, ip)
						delete(knownMap, ip) // Remove from knownMap as it reached the limit
					}
				} else {
					knownMap[ip] = num_keep_alive // Reset back to num_keep_alive since it's alive
				}
			}
			// Check if disconnected
			if len(knownMap) == len(not_responding) {
			}
			for _, ip := range not_responding {
				keep_alive_transmit_channel <- ip  // Transmit dead IP's to main thread
			}

			// Clear the aliveMap for the next interval
			for ip := range aliveMap {
				delete(aliveMap, ip)
			}
		}
	}
}