package constants

import (
	"time"
)

type Role string

// Master-slave states
const (
	Master  Role = "master"
	Slave   Role = "slave"
	Unknown Role = "unknown"
)

type FromMSE struct {
	ElevatorRole  Role
	MasterIP string
	CurrentIPAddressMap map[string]int
}

type ToMSE struct {
	LocalIP string
	IPAddressMap map[string]int
}

// Variables to tweak system
const (
	UDPPort                 int           = 23456
	MasterPort              string        = "1861"
	LoopbackIp              string        = "127.0.0.1"
	NumKeepAlive            int           = 5 // Number of missed keep-alive messages missed before assumed offline
	DeltaTKeepAlive         time.Duration = 50 * time.Millisecond
	DeltaTSamplingKeepAlive time.Duration = 100 * time.Millisecond
)

// Colored text
const (
	ColorReset = "\033[0m"
	ColorYellow = "\033[33m"
	ColorCyan  = "\033[36m"
	ColorGreen = "\033[32m"
	ColorRed   = "\033[31m"
)