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
