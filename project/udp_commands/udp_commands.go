package udp_commands

import (
	. "elevator/constants"
	. "fmt"
	"strconv"
	"strings"
)

func IPToNum(ip_address string) int {
	ip_as_string := strings.Join(strings.Split(ip_address, "."), "")
	ip_as_num, err := strconv.Atoi(ip_as_string)
	if err != nil {
		Printf("Error when casting IP address to num.\n")
	}
	return ip_as_num
}

func MasterSlaveElection(local_IP string, role_channel chan<- Role, filtered_udp_recieve_channel <-chan map[string]int) {
	for ip_address_map := range filtered_udp_recieve_channel {
		Printf("Master slave election:\n")
		for ip_address := range ip_address_map {
			Printf("%d\n", IPToNum(ip_address))
		}
	}
}

func SendMapToChannel[K comparable, V any](current_map map[K]V, channel chan<- map[K]V){
	// Create deep copy
	map_copy := make(map[K]V)
	// Manually copy elements from the original map to the new map
	for key, value := range current_map {
		map_copy[key] = value
	}
	// Passes updated list to see if new master should be elected
	channel <- map_copy 
}
