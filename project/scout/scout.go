// Responsible for overlooking UDP, and reelecting master
// Based on https://github.com/TTK4145/Network-go/tree/master/network/conn
package scout

import (
	"bytes"
	"fmt"
	"net"
	"project/rblog"
	"project/scout/conn"
	"strconv"
	"strings"
	"time"
)

type FromMSE struct {
	ElevatorRole        Role
	MasterIP            string
}

type ToMSE struct {
	LocalIP      string
	IPAddressMap map[string]int
}

type Role string

const (
	Master  Role = "master"
	Slave   Role = "slave"
)

const (
	udpPort                 int           = 23456
	LoopbackIP              string        = "127.0.0.1"
	broadcastAddr           string        = "255.255.255.255"
	numKeepAlive            int           = 5 // Number of missed keep-alive messages missed before assumed offline
	deltaTKeepAlive         time.Duration = 50 * time.Millisecond
	deltaTSamplingKeepAlive time.Duration = 100 * time.Millisecond
)

var localIP string = LoopbackIP

func LocalIP() (string, error) {
	if localIP == LoopbackIP {
		dialer := net.Dialer{Timeout: 500 * time.Millisecond}
		conn, err := dialer.Dial("tcp4", "8.8.8.8:53")
		if err != nil {
			return LoopbackIP, err
		}
		defer conn.Close()
		localIP = strings.Split(conn.LocalAddr().String(), ":")[0]
	}
	return localIP, nil
}

// Intended to run as a go-routine. Returns when conn is closed.
func RecieveUDP(recieve_broadcast_channel chan<- string) {
	const bufSize = 1024

	conn := conn.DialBroadcastUDP(udpPort)
	defer conn.Close()

	buff := make([]byte, bufSize)
	for {
		n, _, e := conn.ReadFrom(buff)
		if e != nil {
			rblog.Red.Printf("Error when recieving with UDP on port %d: \"%+v\"\n", udpPort, e)
			continue
		}
		recieved_message := string(buff[:n])
		recieve_broadcast_channel <- recieved_message
	}
}

// Intended to run as a go-routine. Returns when bcastConn is closed.
//
// Sends keep-alive messages, and maybe trigger a reelection of master
func SendKeepAliveMessage(deltaT time.Duration) {
	bcastConn := conn.DialBroadcastUDP(udpPort)
	bcastAddr, _ := net.ResolveUDPAddr("udp4", fmt.Sprintf("%s:%d", broadcastAddr, udpPort))
	failedLastBroadcast := false
	defer bcastConn.Close()

	for {
		localIP, err := LocalIP()
		if err == nil {
			_, e := bcastConn.WriteTo([]byte(localIP), bcastAddr)
			if e != nil {
				if !failedLastBroadcast {
					rblog.Red.Printf("Error when broadcasting: %+v", e)
					failedLastBroadcast = true
				}
			} else { 
				failedLastBroadcast = false
			}
		}
		time.Sleep(deltaT)
	}
}

// Intended to run as a go-routine. Called from main.
//
// Responsible for keeping track of which elevators are alive, and elect master-slave configuration.
func Start(fromMSEChannel chan<- FromMSE) {
	// Initilization of scout
	UDPRecieveChannel := make(chan string)

	go RecieveUDP(UDPRecieveChannel)
	go SendKeepAliveMessage(deltaTKeepAlive)
	
	knownIPMap := make(map[string]int) // Known IP-addresses with number keeping track of 'aliveness'
	timer := time.NewTicker(deltaTSamplingKeepAlive) // Timer to check for keep-alive messages
	hasChanged := false
	defer timer.Stop()

	// Check if connected to internet
	localIP, err := LocalIP()
	if err != nil {
		rblog.Red.Println("Error when getting local IP. Probably disconnected.")
		fromMSEData, _ := masterSlaveElection(map[string]int{LoopbackIP: numKeepAlive})
		fromMSEChannel <- fromMSEData
	} else {
		rblog.Green.Printf("Local IP: %s", localIP)
	}

	for {
		select {
		case IPAddr := <-UDPRecieveChannel:
			// Received keep alive message
			if net.ParseIP(IPAddr) != nil {  // Ensures that the string recieved is a valid IP address
				_, exists := knownIPMap[IPAddr]
				knownIPMap[IPAddr] = numKeepAlive
				if !exists {
					hasChanged = true
				}
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
				fromMSEData, isNewMaster := masterSlaveElection(knownIPMap)
				if isNewMaster {
					fromMSEChannel <- fromMSEData
				}
			}
			hasChanged = false
		}
	}
}

var lastMasterIP string

func masterSlaveElection(IPAddressMap map[string]int) (FromMSE, bool){
	rblog.Yellow.Printf("Current active IP's: %+v\n", IPAddressMap)

	role := Slave
	masterIP := getMasterIP(IPAddressMap)

	if masterIP != lastMasterIP {
		rblog.Cyan.Println("--- Master Slave Election ---")
		if masterIP == LoopbackIP {
			role = Master
		} else {
			role = Slave
		}
		lastMasterIP = masterIP
		return FromMSE{ElevatorRole: role, MasterIP: masterIP}, true
	} else {
		return FromMSE{}, false
	}
}

// Finds masterIP by choosing the IP with the highest numerical value
func getMasterIP(IPAddressMap map[string]int) string{
	rblog.Yellow.Printf("Current active IP's: %+v\n", IPAddressMap)
	var highestIPInt []byte = []byte{0, 0, 0, 0}
	var highestIPString string
	if len(IPAddressMap) <= 1 { // Elevator is disconnected or alone
		return LoopbackIP
	} else{
		for ipAddress := range IPAddressMap {
			ipAddressInt := net.ParseIP(ipAddress)
			if bytes.Compare(ipAddressInt, highestIPInt) > 0{
				highestIPString = ipAddress
				highestIPInt = ipAddressInt
			}
		}
		localIP, _ := LocalIP()
		if highestIPString == localIP {  // To reduce number of reelections needed
			return LoopbackIP
		} else{
			return highestIPString
		}
	}
}
