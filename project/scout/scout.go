package scout

import (
	"fmt"
	"net"
	. "project/constants"
	"project/rblog"
	"project/scout/conn"
	"strconv"
	"strings"
	"time"
)

type FromMSE struct {
	ElevatorRole        Role
	MasterIP            string
	CurrentIPAddressMap map[string]int
}

type ToMSE struct {
	LocalIP      string
	IPAddressMap map[string]int
}

var localIP string = LoopbackIp

func LocalIP() (string, error) {
	if localIP == LoopbackIp {
		dialer := net.Dialer{Timeout: 500 * time.Millisecond} // Timeout duration
		conn, err := dialer.Dial("tcp4", "8.8.8.8:53")
		if err != nil {
			return LoopbackIp, err
		}
		defer conn.Close()
		localIP = strings.Split(conn.LocalAddr().String(), ":")[0]
	}
	return localIP, nil
}

func ListenUDP(recieve_broadcast_channel chan<- string) {
	const bufSize = 1024

	conn := conn.DialBroadcastUDP(UDPPort)
	defer conn.Close()

	buff := make([]byte, bufSize)
	for {
		n, _, e := conn.ReadFrom(buff)
		if e != nil {
			rblog.Red.Printf("Error when recieving with UDP on port %d: \"%+v\"\n", UDPPort, e)
			continue
		}
		recieved_message := string(buff[:n])
		recieve_broadcast_channel <- recieved_message
	}
}

func SendKeepAliveMessage(deltaT time.Duration) {
	// Sends keep-alive messages, updating all elevators that it is active, and maybe trigger a reelection of master
	bcastConn := conn.DialBroadcastUDP(UDPPort)
	bcastAddr, _ := net.ResolveUDPAddr("udp4", fmt.Sprintf("%s:%d", BroadcastAddr, UDPPort))
	defer bcastConn.Close()

	for {
		localIP, err := LocalIP()
		if err == nil {
			_, e := bcastConn.WriteTo([]byte(localIP), bcastAddr)
			if e != nil {
				rblog.Red.Printf("Error when broadcasting: %+v", e)
			}
		}
		time.Sleep(deltaT)
	}
}

func TrackMissedKeepAliveMessagesAndMSE(deltaT time.Duration, numKeepAlive int, UDPRecieveChannel <-chan string, mseUpdatedIPMapChannel chan<- ToMSE) {
	knownMap := make(map[string]int)      // Known IP-addresses with number keeping track of 'aliveness'
	aliveMap := make(map[string]struct{}) // IP-addresses that sent keep-alive over UDP
	timer := time.NewTicker(deltaT)       // Timer to check for keep-alive messages
	hasChanged := false
	defer timer.Stop()
	for {
		select {
		case IPAddr := <-UDPRecieveChannel:
			// Received keep alive message
			aliveMap[IPAddr] = struct{}{}
			_, exists := knownMap[IPAddr]
			if !exists {
				knownMap[IPAddr] = numKeepAlive
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
			if hasChanged {
				localIP, _ := LocalIP()
				// Run master slave election
				copyKnownMap := MakeDeepCopyMap(knownMap)
				mseUpdatedIPMapChannel <- ToMSE{LocalIP: localIP, IPAddressMap: copyKnownMap}
			}
			hasChanged = false
		}
	}
}

func MasterSlaveElection(mseCh chan<- FromMSE, updatedIPAddressCh <-chan ToMSE) {
	var highestIPInt int = 0
	var highestIPString string = "0.0.0.0"
	var lastMasterIP string = LoopbackIp
	var lastRole Role = Unknown

	for mseData := range updatedIPAddressCh {
		localIP := mseData.LocalIP
		IPAddressMap := mseData.IPAddressMap
		rblog.Yellow.Printf("Current active IP's: %+v\n", IPAddressMap)

		role := Unknown
		highestIPInt = 0
		if len(IPAddressMap) <= 1 { // Elevator is disconnected or alone
			rblog.Cyan.Println("--- Master Slave Election ---")
			lastRole = Master
			lastMasterIP = LoopbackIp // (Always smaller than a regular IP)
			mseCh <- FromMSE{ElevatorRole: Master, MasterIP: lastMasterIP, CurrentIPAddressMap: IPAddressMap}
			continue
		}

		for ipAddress := range IPAddressMap {
			ipAddressInt := IPToNum(ipAddress)
			if ipAddressInt > highestIPInt {
				highestIPString = ipAddress
				highestIPInt = ipAddressInt
			}
		}

		if highestIPString != lastMasterIP {
			rblog.Cyan.Println("--- Master Slave Election ---")
			if highestIPString == localIP {
				role = Master
			} else {
				role = Slave
			}
			// Check if a change in roles needs to take place
			if lastRole != role || lastMasterIP != highestIPString {
				lastRole = role
				lastMasterIP = highestIPString
				mseCh <- FromMSE{ElevatorRole: role, MasterIP: highestIPString, CurrentIPAddressMap: IPAddressMap}
			}
		}
	}
}

func IPToNum(ipAddress string) int {
	ip_as_string := strings.Join(strings.Split(ipAddress, "."), "")
	ip_as_num, err := strconv.Atoi(ip_as_string)
	if err != nil {
		rblog.Red.Printf("Error when casting IP address to num.\n")
	}
	return ip_as_num
}

func MakeDeepCopyMap[K comparable, V any](current_map map[K]V) map[K]V {
	// Create deep copy
	map_copy := make(map[K]V)
	// Manually copy elements from the original map to the new map
	for key, value := range current_map {
		map_copy[key] = value
	}
	return map_copy
}

func SendMapToChannel[K comparable, V any](current_map map[K]V, channel chan<- map[K]V) {
	map_copy := MakeDeepCopyMap[K, V](current_map)
	// Passes updated list to see if new master should be elected
	channel <- map_copy
}
