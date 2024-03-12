package master

import (
	"fmt"
	"log"
	"math/rand"
	"os"
	"project/master/assigner"
	"project/master/server"
	"project/mscomm"
	"project/rblog"
	"reflect"
	"time"
)

type slaveType struct {
	ch           chan interface{}
	quitCh       chan struct{}
	hired        bool
	statePending bool
}

type syncAttemptType struct {
	button        mscomm.ButtonPressed
	pendingSlaves map[string]struct{}
}

type master struct {
	slaves               map[string]*slaveType
	communityState       assigner.CommunityState
	syncAttempts         map[int]syncAttemptType
	applicationTimeoutCh chan string
	syncTimeoutCh        chan int
	collectStateTimer    *time.Timer
}

const (
	collectStateTimeout time.Duration = 500 * time.Millisecond
	applicationTimeout  time.Duration = 500 * time.Millisecond
	syncTimeout         time.Duration = 500 * time.Millisecond
	watchdogTimeout     time.Duration = 1000 * time.Millisecond
	watchdogResetPeriod time.Duration = 300 * time.Millisecond
	terminationDelay    time.Duration = 100 * time.Millisecond
	floorCount          int           = 4
)

var logfile, _ = os.OpenFile("masterlog.txt", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
var flog = log.New(logfile, "master: ", log.Ltime|log.Lmicroseconds|log.Lshortfile)

// Run as a goroutine. Will start to quit after something is sent on quitCh or if it closes
func Run(masterPort int, quitCh chan struct{}) {

	rblog.Magenta.Print("--- Starting master ---")
	flog.Println("[INFO] Starting master")

	terminateTimer := time.NewTimer(terminationDelay)
	terminateTimer.Stop()
	m := master{}
	m.init()

	//init watchdog
	watchdog := time.AfterFunc(watchdogTimeout, func() {
		flog.Println("[ERROR] master deadlock")
		panic("Vaktbikkje sier voff!🐕‍🦺 - master deadlock?")
	})
	defer watchdog.Stop()

	//start listening for slaves
	fromSlaveCh := make(chan mscomm.Package)
	slaveConnEventCh := make(chan mscomm.ConnectionEvent)

	listener, err := server.Listen(masterPort)
	if err != nil {
		flog.Print("[ERROR] Could not get listener:", err)
		panic(fmt.Sprint("Could not get listener", err))
	}
	defer listener.Close()

	go server.Acceptor(listener, fromSlaveCh, slaveConnEventCh)

	for {
		select {
		case slave := <-slaveConnEventCh:
			if slave.Connected {
				flog.Println("[INFO] slave connected: ", slave.Addr)
				m.slaves[slave.Addr] = &slaveType{
					ch:     slave.Ch,
					quitCh: slave.QuitCh,
				}

				go func() {
					time.Sleep(applicationTimeout)
					m.applicationTimeoutCh <- slave.Addr
				}()

				slave.Ch <- mscomm.RequestHallRequests{}

			} else { //Slave disconnected
				if _, exists := m.slaves[slave.Addr]; exists {
					rblog.Magenta.Println("slave resigned:", slave.Addr)
				} else {
					rblog.Magenta.Println("slave dismissed:", slave.Addr)
				}
				flog.Println("[INFO] slave disconnected: ", slave.Addr)
				m.dismiss(slave.Addr)
				m.collectStates()
			}
		case addr := <-m.applicationTimeoutCh:
			if _, exists := m.slaves[addr]; !exists {
				continue
			}
			if !m.slaves[addr].hired {
				rblog.Yellow.Println("slave did not meet application deadline:", addr)
				flog.Println("[WARNING] slave did not meet application deadline:", addr)
				m.dismiss(addr)
			}
		case syncId := <-m.syncTimeoutCh:
			if _, exists := m.syncAttempts[syncId]; exists {
				flog.Println("[WARNING] sync attempt timed out", m.syncAttempts[syncId].pendingSlaves, "did not respond")
				for addr := range m.syncAttempts[syncId].pendingSlaves {
					if _, stillThere := m.slaves[addr]; stillThere {
						m.dismiss(addr)
						rblog.Yellow.Println("slave did not acknowledge sync attempt:", addr)
					}
				}
				delete(m.syncAttempts, syncId)
				m.shareLights() //Send lights to the slaves still connected
			}
		case <-m.collectStateTimer.C:
			for addr, slave := range m.slaves {
				if slave.statePending {
					rblog.Yellow.Println("slave did not respond to state request:", addr)
					flog.Println("[WARNING] slave did not respond to state request:", addr)
					m.dismiss(addr)
				}
			}
			m.assignHallRequests()
		case <-quitCh:
			quitCh = nil //avoid endless loop if quitCh is closed
			flog.Println("[INFO] master ready to quit")
			listener.Close()
			for addr := range m.slaves {
				m.dismiss(addr)
			}
			//some delay to clear channels before terminating
			terminateTimer.Reset(terminationDelay)

		case <-terminateTimer.C:
			rblog.Magenta.Print("--- Terminating master ---")
			flog.Println("[INFO] Terminating master")
			return
		case message := <-fromSlaveCh:
			flog.Println("[INFO] received message", message)
			if _, exists := m.slaves[message.Addr]; !exists {
				rblog.Red.Println("master received message from unknown slave", message.Addr)
				flog.Println("[ERROR] master received message from unknown slave")
				continue
			}

			switch message.Payload.(type) {

			case mscomm.HallRequests:
				//Slave is hired!
				slaveHallRequests := message.Payload.(mscomm.HallRequests)
				m.communityState.HallRequests.Merge(&slaveHallRequests)
				m.slaves[message.Addr].hired = true
				m.slaves[message.Addr].ch <- mscomm.Lights(m.communityState.HallRequests)
				rblog.Magenta.Println("slave hired:", message.Addr)
				flog.Println("[INFO] slave hired:", message.Addr)

				m.collectStates()

			case mscomm.ButtonPressed:
				flog.Println("[INFO] button pressed:", message.Payload.(mscomm.ButtonPressed))
				buttonPressed := message.Payload.(mscomm.ButtonPressed)
				if buttonPressed.Button >= 2 {
					continue // ignore cab requests
				}

				syncId := rand.Int()

				anyoneHired := false
				for _, slave := range m.slaves {
					if slave.hired {
						anyoneHired = true
						break
					}
				}
				if !anyoneHired {
					rblog.Yellow.Println("Noone hired, cannot sync")
					continue
				}

				m.syncAttempts[syncId] = syncAttemptType{
					button:        buttonPressed,
					pendingSlaves: make(map[string]struct{}),
				}

				syncRequests := mscomm.SyncRequests{
					Requests: m.communityState.HallRequests, //TODO: copy
					Id:       syncId,
				}
				syncRequests.Requests[buttonPressed.Floor][buttonPressed.Button] = true

				for addr, slave := range m.slaves {
					if slave.hired {
						flog.Println("[INFO] syncing requests to", addr)
						m.syncAttempts[syncId].pendingSlaves[addr] = struct{}{}
						slave.ch <- syncRequests
					}
				}

				go func() {
					time.Sleep(syncTimeout)
					m.syncTimeoutCh <- syncId
				}()

			case mscomm.ElevatorState:
				flog.Println("[INFO] received elevatorstate from", message.Addr)
				elevState := message.Payload.(mscomm.ElevatorState)
				if elevState.Floor < 0 || elevState.Behavior == "blocked" {
					delete(m.communityState.States, message.Addr)
				} else {
					m.communityState.States[message.Addr] = elevState
				}
				m.slaves[message.Addr].statePending = false

				anyonePending := false
				for _, slave := range m.slaves {
					if slave.statePending {
						anyonePending = true
						break
					}
				}

				if !anyonePending {
					m.collectStateTimer.Stop()
					m.assignHallRequests()
				}

			case mscomm.OrderComplete:
				orderComplete := message.Payload.(mscomm.OrderComplete)
				flog.Println("[INFO]", message.Addr, "completed order:", orderComplete)
				m.communityState.HallRequests[orderComplete.Floor][orderComplete.Button] = false
				m.shareLights()

				//Do not need to assign here, right?

			case mscomm.SyncOK:
				syncId := message.Payload.(mscomm.SyncOK).Id
				if _, exists := m.syncAttempts[syncId]; !exists {
					continue //ignore
				}
				flog.Println("[INFO] ", message.Addr, "synced successfully")
				delete(m.syncAttempts[syncId].pendingSlaves, message.Addr)
				if len(m.syncAttempts[syncId].pendingSlaves) == 0 {
					//sync successful
					flog.Println("[INFO] sync was successful")
					m.communityState.HallRequests.Set(m.syncAttempts[syncId].button)
					m.shareLights()
					delete(m.syncAttempts, syncId)
					m.collectStates()
				}

			default:
				rblog.Red.Println("master received unknown message type", reflect.TypeOf(message.Payload).Name(), "from", message.Addr)
				flog.Println("[ERROR] received unknown message type", reflect.TypeOf(message.Payload).Name(), "from", message.Addr)

			}
		case <-time.After(watchdogResetPeriod):
			//unblock select to reset watchdog
		}

		watchdog.Reset(watchdogTimeout) //flink bisk🐶
	}
}

func (m *master) init() {
	m.slaves = make(map[string]*slaveType)
	m.communityState.HallRequests = make(mscomm.HallRequests, floorCount)
	m.communityState.States = make(map[string]mscomm.ElevatorState)
	m.syncAttempts = make(map[int]syncAttemptType)

	m.applicationTimeoutCh = make(chan string)
	m.syncTimeoutCh = make(chan int)

	m.collectStateTimer = time.NewTimer(collectStateTimeout)
	m.collectStateTimer.Stop()
}

func (m *master) dismiss(addr string) {
	flog.Println("[INFO] Dismissing", addr)
	if _, exists := m.slaves[addr]; exists {
		m.slaves[addr].quitCh <- struct{}{}
		delete(m.slaves, addr)
	}
	delete(m.communityState.States, addr)
}

func (m *master) collectStates() {
	flog.Println("[INFO] Collect elevator states before assigning")
	for addr, slave := range m.slaves {
		if slave.hired {
			flog.Println("[INFO] Requesting state from", addr)
			slave.ch <- mscomm.RequestState{}
			slave.statePending = true
		}
	}
	m.collectStateTimer.Reset(collectStateTimeout)
}

func (m *master) shareLights() {
	for addr, slave := range m.slaves {
		if slave.hired {
			flog.Println("[INFO] distributing lights to", addr)
			slave.ch <- mscomm.Lights(m.communityState.HallRequests)
		}
	}
}

func (m *master) assignHallRequests() {
	flog.Println("[INFO] ready to assign")
	if len(m.communityState.States) == 0 {
		rblog.Yellow.Println("Noone to assign to")
		flog.Println("[WARNING] Noone to assign to")
		return
	}
	//TODO: fix deadlock that occurs right about here - when running assigner executable
	flog.Println("[INFO] starting assigner executable")
	assignedRequests, err := assigner.Assign(&m.communityState)
	if err != nil {
		rblog.Red.Println("assigner error:", err)
		flog.Println("assigner error:", err)
		return
	}
	flog.Println("[INFO] assigner ran sucessfully:", assignedRequests)

	for addr, requests := range *assignedRequests {
		flog.Println("[INFO]", addr, "got assigned", requests)
		m.slaves[addr].ch <- requests
		//TODO: timeout if slave does not respond
	}
}
