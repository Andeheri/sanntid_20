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
	Master_slave_election string = "master_slave_election"
	Keep_alive string            = "keep_alive"
)