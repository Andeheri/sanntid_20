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

// Variables to tweak system
const (
	UDPPort                 int           = 23456
	MasterPort              int           = 1861
	LoopbackIp              string        = "127.0.0.1"
	NumKeepAlive            int           = 5 // Number of missed keep-alive messages missed before assumed offline
	DeltaTKeepAlive         time.Duration = 50 * time.Millisecond
	DeltaTSamplingKeepAlive time.Duration = 100 * time.Millisecond
)
