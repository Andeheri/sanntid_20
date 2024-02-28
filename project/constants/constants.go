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

type MSE_type struct {
	Role  Role
	IP string
}

const (
	UDP_PORT = 23456
)
