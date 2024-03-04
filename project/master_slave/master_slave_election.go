package master_slave

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

func Election(localIP string, mseCh chan<- MSE_type, updatedIPAddressCh <-chan map[string]struct{}) {
	var highest_ip_int int = 0
	var highest_ip_string string = "0.0.0.0"
	last_highest_ip := ""
	last_role := Unknown

	for ip_address_map := range updatedIPAddressCh {
		role := Unknown
		if len(ip_address_map) == 0 {  // Elevator is disconnected
			last_role = Master
			last_highest_ip = "127.0.0.1" // Loopback address (Always smaller than a regular IP)
			mseCh <- MSE_type{Role: Master, IP: "127.0.0.1"}
			continue
		}
		// Possibly need to update role or master to connect to
		for ip_address := range ip_address_map {
			ip_address_int := IPToNum(ip_address)
			if (ip_address_int > highest_ip_int){
				highest_ip_string = ip_address
				highest_ip_int = ip_address_int
			}
		}

		if (highest_ip_string != last_highest_ip){
			if (highest_ip_string == localIP){
				role = Master
			}else{
				role = Slave
			}
			// Check if a change in roles needs to take place
			if (last_role != role || last_highest_ip != highest_ip_string){
				last_role = role
				last_highest_ip = highest_ip_string
				mseCh <- MSE_type{Role: role, IP: highest_ip_string}
			}
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
