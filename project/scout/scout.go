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

func TrackMissedKeepAliveMessagesAndMSE(deltaT time.Duration, numKeepAlive int, UDPRecieveChannel <-chan string, fromMSEChannel chan<- FromMSE) {
	knownIPMap := make(map[string]int)      // Known IP-addresses with number keeping track of 'aliveness'
	timer := time.NewTicker(deltaT)       // Timer to check for keep-alive messages
	hasChanged := false
	defer timer.Stop()

	// Check if connected to internet
	localIP, err := LocalIP()
	if err != nil {
		rblog.Red.Println("Error when getting local IP. Probably disconnected.")
		MasterSlaveElection(fromMSEChannel, map[string]int{LoopbackIp: NumKeepAlive})
	} else {
		rblog.Green.Printf("Local IP: %s", localIP)
	}

	for {
		select {
		case IPAddr := <-UDPRecieveChannel:
			// Received keep alive message
			_, exists := knownIPMap[IPAddr]
			knownIPMap[IPAddr] = numKeepAlive
			if !exists {
				hasChanged = true
			}
		case <-timer.C:
			for ip, count := range knownIPMap {
				count -= 1
				if count == 0 {
					delete(knownIPMap, ip)
					hasChanged = true
				} else{
					knownIPMap[ip] = count
				}
			}
			// Check if master-slave-configuration needs to be updated
			if hasChanged {
				MasterSlaveElection(fromMSEChannel, knownIPMap)
			}
			hasChanged = false
		}
	}
}

var lastMasterIP string

func MasterSlaveElection(mseCh chan<- FromMSE, IPAddressMap map[string]int) {
	rblog.Yellow.Printf("Current active IP's: %+v\n", IPAddressMap)

	role := Slave
	masterIP := getMasterIP(IPAddressMap)

	if masterIP != lastMasterIP {
		rblog.Cyan.Println("--- Master Slave Election ---")
		if masterIP == localIP {
			role = Master
		} else {
			role = Slave
		}
		lastMasterIP = masterIP
		mseCh <- FromMSE{ElevatorRole: role, MasterIP: masterIP, CurrentIPAddressMap: IPAddressMap}
	}
}

// Finds masterIP by choosing the IP with the higest numerical value
func getMasterIP(IPAddressMap map[string]int) string{
	rblog.Yellow.Printf("Current active IP's: %+v\n", IPAddressMap)
	var highestIPInt int = 0
	var highestIPString string
	if len(IPAddressMap) <= 1 { // Elevator is disconnected or alone
		return LoopbackIp
	} else{
		for ipAddress := range IPAddressMap {
			ipAddressInt := IPToNum(ipAddress)
			if ipAddressInt > highestIPInt {
				highestIPString = ipAddress
				highestIPInt = ipAddressInt
			}
		}
		localIP, _ := LocalIP()
		if highestIPString == localIP {  // To reduce number of reelections needed
			return LoopbackIp
		} else{
			return highestIPString
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
