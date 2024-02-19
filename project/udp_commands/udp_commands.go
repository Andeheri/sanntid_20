package udp_commands


type role string

// Master-slave states
const (
	master  role = "master"
	slave   role = "slave"
	unknown role = "unknown"
)

func MasterSlaveElection(elevator_role *role) {
	*elevator_role = master;
}