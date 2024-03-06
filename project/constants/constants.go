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
	Role  Role
	IP string
	IPAddressMap map[string]int
}

type ToMSE struct {
	LocalIP string
	IPAddressMap map[string]int
}

const (
	UDPPort               int           = 23456
	MasterPort            string        = "1861"
	DeltaTKeepAlive       time.Duration = 100 * time.Millisecond
	DeltaTMissedKeepAlive time.Duration = 50 * time.Millisecond
	NumKeepAlive          int           = 5 // Number of missed keep-alive messages missed before assumed offline
	LoopbackIp            string        = "127.0.0.1"
)
