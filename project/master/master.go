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
	ch                 chan interface{}
	quitCh             chan struct{}
	employmentStatus   slaveEmploymentStatus
	statePending       bool
	assignedRequests   mscomm.AssignedRequests
	orderCompleteTimer *time.Timer
	elevatorState      mscomm.ElevatorState
}

type syncAttemptType struct {
	button        mscomm.ButtonPressed
	pendingSlaves map[string]struct{}
}

type master struct {
	slaves                 map[string]*slaveType
	hallRequests           mscomm.HallRequests
	syncAttempts           map[int]syncAttemptType
	applicationTimeoutCh   chan string
	syncTimeoutCh          chan int
	orderCompleteTimeoutCh chan string
	sickLeaveTimeoutCh     chan string
	collectStateTimer      *time.Timer
	retryAssignmentTimer   *time.Timer
}

const (
	collectStateTimeout  time.Duration = 500 * time.Millisecond
	applicationTimeout   time.Duration = 500 * time.Millisecond
	syncTimeout          time.Duration = 500 * time.Millisecond
	watchdogTimeout      time.Duration = 1000 * time.Millisecond
	watchdogResetPeriod  time.Duration = 300 * time.Millisecond
	orderCompleteTimeout time.Duration = 15 * time.Second
	sickLeaveDuration    time.Duration = 30 * time.Second
	retryAssignmentDelay time.Duration = 200 * time.Millisecond
	terminationDelay     time.Duration = 100 * time.Millisecond
	floorCount           int           = 4
)

type slaveEmploymentStatus string

const (
	sesHired       slaveEmploymentStatus = "hired"
	sesApplicant   slaveEmploymentStatus = "applicant"
	sesOnSickLeave slaveEmploymentStatus = "onSickLeave"
)

var logfile, _ = os.OpenFile("masterlog.txt", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
var flog = log.New(logfile, "master: ", log.Ltime|log.Lmicroseconds|log.Lshortfile)

// Run as a goroutine. Will start to quit after something is sent on quitCh or if it closes
func Start(masterPort int, quitCh chan struct{}) {

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
		case event := <-slaveConnEventCh:
			if event.Connected {
				flog.Println("[INFO] slave connected: ", event.Addr)
				m.slaves[event.Addr] = &slaveType{
					ch:               event.Ch,
					quitCh:           event.QuitCh,
					employmentStatus: sesApplicant,
					assignedRequests: make(mscomm.AssignedRequests, floorCount),
					orderCompleteTimer: time.AfterFunc(orderCompleteTimeout, func() {
						m.orderCompleteTimeoutCh <- event.Addr
					}),
				}
				m.slaves[event.Addr].orderCompleteTimer.Stop()

				//assume blocked until slave reports otherwise
				m.slaves[event.Addr].elevatorState.Behavior = "blocked"

				go func() {
					time.Sleep(applicationTimeout)
					m.applicationTimeoutCh <- event.Addr
				}()

				event.Ch <- mscomm.RequestHallRequests{}

			} else { //Slave disconnected
				if _, exists := m.slaves[event.Addr]; exists {
					rblog.Magenta.Println("slave resigned:", event.Addr)
				} else {
					rblog.Magenta.Println("slave dismissed:", event.Addr)
				}
				flog.Println("[INFO] slave disconnected: ", event.Addr)
				m.dismiss(event.Addr)
				m.collectStates()
			}
		case addr := <-m.applicationTimeoutCh:
			if _, exists := m.slaves[addr]; !exists {
				continue
			}
			if m.slaves[addr].employmentStatus == sesApplicant {
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
		case addr := <-m.orderCompleteTimeoutCh:
			if slave, exists := m.slaves[addr]; exists {
				if slave.elevatorState.Behavior == "blocked" {
					//slave has reported a reason why it did not complete the order. No sick leave for you
					continue
				}
				rblog.Yellow.Println("slave did not complete order in time:", addr)
				flog.Println("[WARNING] slave did not complete order in time:", addr)
				slave.employmentStatus = sesOnSickLeave
				go func() {
					time.Sleep(sickLeaveDuration)
					m.sickLeaveTimeoutCh <- addr
				}()
			}
		case addr := <-m.sickLeaveTimeoutCh:
<<<<<<< Updated upstream
			if slave, exists := m.slaves[addr]; exists {
				rblog.Magenta.Println("slave back from sick leave:", addr)
				slave.employmentStatus = sesHired
				m.collectStates()
			}
=======
			m.onSickLeaveTimeout(addr)
		case <-m.retryAssignmentTimer.C:
			m.collectStates()

>>>>>>> Stashed changes
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
				m.hallRequests.Merge(&slaveHallRequests)
				m.slaves[message.Addr].employmentStatus = sesHired
				m.slaves[message.Addr].ch <- mscomm.Lights(m.hallRequests)
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

				anyoneToSyncTo := false
				for _, slave := range m.slaves {
					if slave.employmentStatus == sesHired || slave.employmentStatus == sesOnSickLeave {
						anyoneToSyncTo = true
						break
					}
				}

				if !anyoneToSyncTo {
					rblog.Yellow.Println("Noone hired, cannot sync")
					continue
				}

				m.syncAttempts[syncId] = syncAttemptType{
					button:        buttonPressed,
					pendingSlaves: make(map[string]struct{}),
				}

				syncRequests := mscomm.SyncRequests{
					Requests: make(mscomm.HallRequests, floorCount),
					Id:       syncId,
				}
				for i := range m.hallRequests {
					syncRequests.Requests[i][0] = m.hallRequests[i][0]
					syncRequests.Requests[i][1] = m.hallRequests[i][1]
				}
				syncRequests.Requests[buttonPressed.Floor][buttonPressed.Button] = true

				for addr, slave := range m.slaves {
					if slave.employmentStatus == sesHired || slave.employmentStatus == sesOnSickLeave {
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
				m.slaves[message.Addr].elevatorState = elevState
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
				m.hallRequests[orderComplete.Floor][orderComplete.Button] = false
				m.shareLights()

				m.slaves[message.Addr].assignedRequests[orderComplete.Floor][orderComplete.Button] = false

				allOrdersComplete := true
			RequestLoop:
				for _, floor := range m.slaves[message.Addr].assignedRequests {
					for _, request := range floor {
						if request {
							allOrdersComplete = false
							break RequestLoop
						}
					}
				}
				if allOrdersComplete {
					m.slaves[message.Addr].orderCompleteTimer.Stop()
				} else {
					m.slaves[message.Addr].orderCompleteTimer.Reset(orderCompleteTimeout)
				}

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
					m.hallRequests[m.syncAttempts[syncId].button.Floor][m.syncAttempts[syncId].button.Button] = true
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
	m.hallRequests = make(mscomm.HallRequests, floorCount)
	m.syncAttempts = make(map[int]syncAttemptType)

	m.applicationTimeoutCh = make(chan string)
	m.syncTimeoutCh = make(chan int)

	m.orderCompleteTimeoutCh = make(chan string)
	m.sickLeaveTimeoutCh = make(chan string)

	m.collectStateTimer = time.NewTimer(collectStateTimeout)
	m.collectStateTimer.Stop()

	m.retryAssignmentTimer = time.NewTimer(retryAssignmentDelay)
	m.retryAssignmentTimer.Stop()
}

func (m *master) dismiss(addr string) {
	flog.Println("[INFO] Dismissing", addr)
	if slave, exists := m.slaves[addr]; exists {
		slave.quitCh <- struct{}{}
		slave.orderCompleteTimer.Stop()
		delete(m.slaves, addr)
	}
}

func (m *master) collectStates() {
	flog.Println("[INFO] Collect elevator states before assigning")
	for addr, slave := range m.slaves {
		if slave.employmentStatus == sesHired {
			flog.Println("[INFO] Requesting state from", addr)
			slave.ch <- mscomm.RequestState{}
			slave.statePending = true
		}
	}
	m.collectStateTimer.Reset(collectStateTimeout)
}

func (m *master) shareLights() {
	for addr, slave := range m.slaves {
		if slave.employmentStatus == sesHired || slave.employmentStatus == sesOnSickLeave {
			flog.Println("[INFO] distributing lights to", addr)
			slave.ch <- mscomm.Lights(m.hallRequests)
		}
	}
}

func (m *master) assignHallRequests() {

	flog.Println("[INFO] ready to assign")

	assignerInput := assigner.AssignerInput{
		HallRequests: m.hallRequests,
		States:       make(map[string]mscomm.ElevatorState),
	}

	for addr, slave := range m.slaves {
		if slave.employmentStatus == sesHired && slave.elevatorState.Behavior != "blocked" {
			assignerInput.States[addr] = slave.elevatorState
		}
	}

	if len(assignerInput.States) == 0 {
		rblog.Yellow.Println("Noone to assign to")
		flog.Println("[WARNING] Noone to assign to")
		m.retryAssignmentTimer.Reset(retryAssignmentDelay)
		return
	}
	m.retryAssignmentTimer.Stop()

	flog.Println("[INFO] starting assigner executable")
	assignedRequests, err := assigner.Assign(&assignerInput)
	if err != nil {
		rblog.Red.Println("assigner error:", err)
		flog.Println("assigner error:", err)
		return
	}
	flog.Println("[INFO] assigner ran sucessfully:", assignedRequests)

	for addr, requests := range *assignedRequests {
		flog.Println("[INFO]", addr, "got assigned", requests)
		m.slaves[addr].ch <- requests
		m.slaves[addr].assignedRequests = requests
		if requests.IsEmpty() {
			m.slaves[addr].orderCompleteTimer.Stop()
		} else {
			m.slaves[addr].orderCompleteTimer.Reset(orderCompleteTimeout)
		}

	}
}
