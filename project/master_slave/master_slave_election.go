package master_slave

import (
	. "elevator/constants"
	"fmt"
	. "fmt"
	"strconv"
	"strings"
)

func IPToNum(ipAddress string) int {
	ip_as_string := strings.Join(strings.Split(ipAddress, "."), "")
	ip_as_num, err := strconv.Atoi(ip_as_string)
	if err != nil {
		Printf("Error when casting IP address to num.\n")
	}
	return ip_as_num
}

func Election(mseCh chan<- FromMSE, updatedIPAddressCh <-chan ToMSE) {
	var highestIPInt int = 0
	var highestIPString string = "0.0.0.0"
	lastHighestIP := ""
	lastRole := Unknown

	for mseData := range updatedIPAddressCh {
		localIP := mseData.LocalIP
		IPAddressMap := mseData.IPAddressMap
		fmt.Printf("Master-slave election: %v\n", IPAddressMap)
		role := Unknown
		highestIPInt = 0
		if len(IPAddressMap) == 0 || localIP == "" { // Elevator is disconnected
			lastRole = Master
			lastHighestIP = "127.0.0.1" // Loopback address (Always smaller than a regular IP)
			mseCh <- FromMSE{Role: Master, IP: "127.0.0.1", IPAddressMap: IPAddressMap}
			continue
		}

		for ipAddress := range IPAddressMap {
			ipAddressInt := IPToNum(ipAddress)
			if ipAddressInt > highestIPInt {
				highestIPString = ipAddress
				highestIPInt = ipAddressInt
			}
		}

		if highestIPString != lastHighestIP {
			if highestIPString == localIP {
				role = Master
			} else {
				role = Slave
			}
			// Check if a change in roles needs to take place
			if lastRole != role || lastHighestIP != highestIPString {
				lastRole = role
				lastHighestIP = highestIPString
				mseCh <- FromMSE{Role: role, IP: highestIPString, IPAddressMap: IPAddressMap}
			}
		}
	}
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
