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

const (
	collectStateTimeout time.Duration = 500 * time.Millisecond
	applicationTimeout  time.Duration = 500 * time.Millisecond
	syncTimeout         time.Duration = 500 * time.Millisecond
	watchdogTimeout     time.Duration = 1000 * time.Millisecond
	watchdogResetPeriod time.Duration = 300 * time.Millisecond
	floorCount          int           = 4
)

var collectStateTimer *time.Timer

var logfile, _ = os.OpenFile("masterlog.txt", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
var flog = log.New(logfile, "master: ", log.Ltime|log.Lmicroseconds|log.Lshortfile)

var communityState assigner.CommunityState

var applicationTimeoutCh chan string

var slaves = make(map[string]*slaveType)

func Run(masterPort int, quitCh chan struct{}) {

	rblog.Magenta.Print("--- Starting master ---")
	flog.Println("[INFO] Starting master")

	communityState.HallRequests = make(mscomm.HallRequests, floorCount)
	communityState.States = make(map[string]mscomm.ElevatorState)

	syncAttempts := make(map[int]syncAttemptType)

	//init watchdog
	watchdog := time.AfterFunc(watchdogTimeout, func() {
		flog.Println("[ERROR] master deadlock")
		panic("Vaktbikkje sier voff!🐕‍🦺 - master deadlock?")
	})
	defer watchdog.Stop()

	applicationTimeoutCh = make(chan string)
	syncTimeoutCh := make(chan int)

	terminateTimer := time.NewTimer(100 * time.Millisecond)
	terminateTimer.Stop()

	collectStateTimer = time.NewTimer(collectStateTimeout)
	collectStateTimer.Stop()

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
				slaves[slave.Addr] = &slaveType{
					ch:     slave.Ch,
					quitCh: slave.QuitCh,
				}

				go func() {
					time.Sleep(applicationTimeout)
					applicationTimeoutCh <- slave.Addr
				}()

				slave.Ch <- mscomm.RequestHallRequests{}

			} else { //Slave disconnected
				if _, exists := slaves[slave.Addr]; exists{
					rblog.Magenta.Println("slave resigned:", slave.Addr)
				}else{
					rblog.Magenta.Println("slave dismissed:", slave.Addr)
				}
				flog.Println("[INFO] slave disconnected: ", slave.Addr)
				dismiss(slave.Addr)
				collectStates()
			}
		case addr := <-applicationTimeoutCh:
			if _, exists := slaves[addr]; !exists {
				continue
			}
			if !slaves[addr].hired {
				rblog.Yellow.Println("slave did not meet application deadline:", addr)
				flog.Println("[WARNING] slave did not meet application deadline:", addr)
				dismiss(addr)
			}
		case syncId := <-syncTimeoutCh:
			if _, exists := syncAttempts[syncId]; exists {
				flog.Println("[WARNING] sync attempt timed out", syncAttempts[syncId].pendingSlaves, "did not respond")
				for addr := range syncAttempts[syncId].pendingSlaves {
					if _, stillThere := slaves[addr]; stillThere{
						dismiss(addr)
						rblog.Yellow.Println("slave did not acknowledge sync attempt:", addr)
					}
				}
				delete(syncAttempts, syncId)
				shareLights() //Send lights to the slaves still connected
			}
		case <-collectStateTimer.C:
			for addr, slave := range slaves {
				if slave.statePending {
					rblog.Yellow.Println("slave did not respond to state request:", addr)
					flog.Println("[WARNING] slave did not respond to state request:", addr)
					dismiss(addr)
				}
			}
			assignHallRequests()
		case <-quitCh:
			flog.Println("[INFO] master ready to quit")
			listener.Close()
			for addr := range slaves {
				dismiss(addr)
			}
			//some delay to clear channels before terminating
			terminateTimer.Reset(100 * time.Millisecond)

		case <-terminateTimer.C:
			rblog.Magenta.Print("--- Terminating master ---")
			flog.Println("[INFO] Terminating master")
			return
		case message := <-fromSlaveCh:
			flog.Println("[INFO] received message", message)
			if _, exists := slaves[message.Addr]; !exists {
				rblog.Red.Println("master received message from unknown slave", message.Addr)
				flog.Println("[ERROR] master received message from unknown slave")
				//worthy of panic?
				//all messages received should come from registered slaves, or else something went wrong
				continue
			}

			switch message.Payload.(type) {

			case mscomm.HallRequests:
				//Slave is hired!
				slaveHallRequests := message.Payload.(mscomm.HallRequests)
				communityState.HallRequests.Merge(&slaveHallRequests)
				slaves[message.Addr].hired = true
				slaves[message.Addr].ch <- mscomm.Lights(communityState.HallRequests)
				slaves[message.Addr].ch <- mscomm.SyncRequests{
					Requests: communityState.HallRequests,
					Id:       -1, //Unsafe sync
				}
				rblog.Magenta.Println("slave hired:", message.Addr)
				flog.Println("[INFO] slave hired:", message.Addr)

				collectStates()

			case mscomm.ButtonPressed:
				flog.Println("[INFO] button pressed:", message.Payload.(mscomm.ButtonPressed))
				buttonPressed := message.Payload.(mscomm.ButtonPressed)
				if buttonPressed.Button >= 2 {
					continue // ignore cab requests
				}

				syncId := rand.Int()

				anyoneHired := false
				for _, slave := range slaves {
					if slave.hired {
						anyoneHired = true
						break
					}
				}

				if !anyoneHired {
					rblog.Yellow.Println("Noone hired, cannot sync")
					continue
				}

				syncAttempts[syncId] = syncAttemptType{
					button:        buttonPressed,
					pendingSlaves: make(map[string]struct{}),
				}

				syncRequests := mscomm.SyncRequests{
					Requests: communityState.HallRequests,
					Id:       syncId,
				}
				syncRequests.Requests[buttonPressed.Floor][buttonPressed.Button] = true

				for addr, slave := range slaves {
					if slave.hired {
						flog.Println("[INFO] syncing requests to", addr)
						syncAttempts[syncId].pendingSlaves[addr] = struct{}{}
						slave.ch <- syncRequests
					}
				}

				go func() {
					time.Sleep(syncTimeout)
					syncTimeoutCh <- syncId
				}()

			case mscomm.ElevatorState:
				flog.Println("[INFO] received elevatorstate from", message.Addr)
				elevState := message.Payload.(mscomm.ElevatorState)
				if elevState.Floor < 0 || elevState.Behavior == "blocked" {
					delete(communityState.States, message.Addr)
				} else {
					communityState.States[message.Addr] = elevState
				}
				slaves[message.Addr].statePending = false

				anyonePending := false
				for _, slave := range slaves {
					if slave.statePending {
						anyonePending = true
						break
					}
				}

				if !anyonePending {
					collectStateTimer.Stop()
					assignHallRequests()
				}

			case mscomm.OrderComplete:
				orderComplete := message.Payload.(mscomm.OrderComplete)
				flog.Println("[INFO]", message.Addr, "completed order:", orderComplete)
				communityState.HallRequests[orderComplete.Floor][orderComplete.Button] = false
				syncRequests := mscomm.SyncRequests{
					Requests: communityState.HallRequests,
					Id:       -1, //Unsafe sync
				}
				for addr, slave := range slaves {
					if slave.hired {
						flog.Println("[INFO] syncing cleared order to", addr)
						slave.ch <- syncRequests
						slave.ch <- mscomm.Lights(communityState.HallRequests)
					}
				}

				//Do not need to assign here, right?

			case mscomm.SyncOK:
				syncId := message.Payload.(mscomm.SyncOK).Id
				if _, exists := syncAttempts[syncId]; !exists {
					continue //ignore
				}
				flog.Println("[INFO] ", message.Addr, "synced successfully")
				delete(syncAttempts[syncId].pendingSlaves, message.Addr)
				if len(syncAttempts[syncId].pendingSlaves) == 0 {
					//sync successful
					flog.Println("[INFO] sync was successful")
					communityState.HallRequests.Set(syncAttempts[syncId].button)
					shareLights()
					delete(syncAttempts, syncId)
					collectStates()
				}

			default:
				rblog.Red.Println("master received unknown message type", reflect.TypeOf(message.Payload).Name(), "from", message.Addr)
				flog.Println("[ERROR] received unknown message type", reflect.TypeOf(message.Payload).Name(), "from", message.Addr)

			}
		case <-time.After(watchdogResetPeriod):
			//unblock select to reset watchdog
		}

		watchdog.Reset(watchdogTimeout)
	}
}

func dismiss(addr string) {
	flog.Println("[INFO] Dismissing", addr)
	if _, exists := slaves[addr]; exists {
		slaves[addr].quitCh <- struct{}{}
		delete(slaves, addr)
	}
	delete(communityState.States, addr)
}

func collectStates() {
	flog.Println("[INFO] Collect elevator states before assigning")
	collectStateTimer.Reset(collectStateTimeout)
	for addr, slave := range slaves {
		if slave.hired {
			flog.Println("[INFO] Requesting state from", addr)
			slave.ch <- mscomm.RequestState{}
			slave.statePending = true
		}
	}
}

func shareLights() {
	for addr, slave := range slaves {
		if slave.hired {
			flog.Println("[INFO] distributing lights to", addr)
			slave.ch <- mscomm.Lights(communityState.HallRequests)
		}
	}
}

func assignHallRequests() {
	flog.Println("[INFO] ready to assign")
	if len(communityState.States) == 0 {
		rblog.Yellow.Println("Noone to assign to")
		flog.Println("[WARNING] Noone to assign to")
		return
	}
	//TODO: fix deadlock that occurs right about here - when running assigner executable
	flog.Println("[INFO] starting assigner executable")
	assignedRequests, err := assigner.Assign(&communityState)
	if err != nil {
		rblog.Red.Println("assigner error:", err)
		flog.Println("assigner error:", err)
		return
	}
	flog.Println("[INFO] assigner ran sucessfully:", assignedRequests)

	for addr, requests := range *assignedRequests {
		flog.Println("[INFO]", addr, "got assigned", requests)
		slaves[addr].ch <- requests
		//TODO: timeout if slave does not respond
	}
}
