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

func Election(local_IP string, mse_channel chan<- MSE_type, filtered_udp_recieve_channel <-chan map[string]struct{}) {
	var highest_ip_int int
	var highest_ip_string string
	last_highest_ip := ""
	last_role := Unknown

	for ip_address_map := range filtered_udp_recieve_channel {
		// Possibly need to update role or master to connect to
		highest_ip_string = local_IP
		highest_ip_int = IPToNum(local_IP)
		for ip_address := range ip_address_map {
			ip_address_int := IPToNum(ip_address)
			if (ip_address_int > highest_ip_int){
				highest_ip_string = ip_address
				highest_ip_int = ip_address_int
			}
		}
		if (highest_ip_string != last_highest_ip){
			role := Unknown
			if (highest_ip_string == local_IP){
				role = Master
			}else{
				role = Slave
			}
			if (last_role != role || last_highest_ip != highest_ip_string){
				last_role = role
				last_highest_ip = highest_ip_string
				mse_channel <- MSE_type{Role: role, IP: highest_ip_string}
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
