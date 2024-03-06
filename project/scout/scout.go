package scout

import (
	"bytes"
	. "elevator/constants"
	"elevator/scout/conn"
	"elevator/master_slave"
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

func getBroadcastAddr(localIP string, UDP_PORT int) *net.UDPAddr {
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
						// Calculate the broadcast address
						ip := ipnet.IP.To4()
						if ip == nil {
							continue // Not an ipv4 address
						}
						mask := net.IP(ipnet.Mask).To4()
						broadcast := net.IP(make([]byte, 4))
						for i := range ip {
							broadcast[i] = ip[i] | ^mask[i]
						}
						broadcastAddr := broadcast.String()
						addr, _ := net.ResolveUDPAddr("udp4", fmt.Sprintf("%s:%d", broadcastAddr, UDP_PORT))
						return addr
					}
				}
			}
		}
		return nil
	} else {
		return broadcastAddr
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

func SendKeepAliveMessage(deltaT time.Duration) {
	// Sends keep-alive messages, updating all elevators that it is active, and maybe trigger a reelection of master
	bcastConn := conn.DialBroadcastUDP(UDP_PORT)
	defer bcastConn.Close()
	localIP, _ := LocalIP()
	var bcastAddr *net.UDPAddr
	for{
		bcastAddr = getBroadcastAddr(localIP, UDP_PORT)
		if bcastAddr != nil {
			_, e := bcastConn.WriteTo([]byte(localIP), bcastAddr)
			if e != nil {
				fmt.Printf("Error when broadcasting: %+v", e)
			}
		}
		time.Sleep(deltaT)
	}
}

func TrackMissedKeepAliveMessagesAndMSE(deltaT time.Duration, numKeepAlive int, UDPRecieveChannel <-chan string, mseUpdatedIPMapChannel chan<- ToMSE) {
	knownMap := make(map[string]int) // Known IP-addresses with number keeping track of 'aliveness'
	aliveMap := make(map[string]struct{})  // IP-addresses that sent keep-alive over UDP
	timer := time.NewTicker(deltaT)  // Timer to check for keep-alive messages
	hasChanged := false
	defer timer.Stop()
	
	for {
		select {
		case IPAddr := <- UDPRecieveChannel:
			// Received keep alive message
			aliveMap[IPAddr] = struct{}{}
			_, exists := knownMap[IPAddr]
			knownMap[IPAddr] = numKeepAlive
			if !exists {
				hasChanged = true
			}
		case <-timer.C:
			// Timer fired due to deltaT duration passing
			// Compute the difference.
			for ip, count := range knownMap {
				if _, found := aliveMap[ip]; !found {
					count -= 1
					knownMap[ip] = count
					if count <= 0 {
						hasChanged = true
						delete(knownMap, ip) // Remove from knownMap as it reached the limit
					}
				} else {
					knownMap[ip] = numKeepAlive // Reset back to numKeepAlive since it's alive
				}
			}

			// Clear the aliveMap for the next interval
			for ip := range aliveMap {
				delete(aliveMap, ip)
			}

			// Check if master-slave-configuration needs to be updated
			if hasChanged{
				localIP, _ := LocalIP()
				// Run master slave election
				copyKnownMap := master_slave.MakeDeepCopyMap[string, int](knownMap)
				mseUpdatedIPMapChannel <- ToMSE{LocalIP: localIP, IPAddressMap: copyKnownMap}
			}
			hasChanged = false
		}
	}
}