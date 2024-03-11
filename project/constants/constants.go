package constants

import (
	"time"
)

type Role string

// Master-slave states
const (
	Master  Role = "master"
	Slave   Role = "slave"
)

// Variables to tweak system
const (
	UDPPort                 int           = 23456
	MasterPort              int           = 1861
	LoopbackIp              string        = "127.0.0.1"
	BroadcastAddr           string        = "255.255.255.255"
	NumKeepAlive            int           = 5 // Number of missed keep-alive messages missed before assumed offline
	DeltaTKeepAlive         time.Duration = 50 * time.Millisecond
	DeltaTSamplingKeepAlive time.Duration = 100 * time.Millisecond
	WatchdogResetPeriod     time.Duration = 1 * time.Second
	WatchdogTimeoutPeriod   time.Duration = 2 * time.Second
)
