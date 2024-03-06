package constants

type Role string

// Master-slave states
const (
	Master  Role = "master"
	Slave   Role = "slave"
	Unknown Role = "unknown"
)

// UDP commands
const (
	Keep_alive string = "keep_alive"
)

type FromMSE struct {
	Role  Role
	IP string
	IPAddressMap map[string]int
}

type ToMSE struct {
	LocalIP string
	IPAddressMap map[string]int
}

const (
	UDP_PORT = 23456
)
