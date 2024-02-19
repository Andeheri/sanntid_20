package elevator_constants

type Role string

const (
	Master  Role = "master"
	Slave   Role = "slave"
	Unknown Role = "unknown"
)