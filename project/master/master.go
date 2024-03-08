package master

import (
	"fmt"
	"math/rand"
	"project/master/assigner"
	"project/master/server"
	"project/mscomm"
	"project/rblog"
	"reflect"
	"time"
)

type slaveType struct {
	ch           chan interface{}
	hired        bool
	statePending bool
}

type syncAttemptType struct {
	button        mscomm.ButtonPressed
	pendingSlaves map[string]struct{}
}

const (
	applicationTimeout  time.Duration = 500 * time.Second //Is a timeout actually needed?
	syncTimeout         time.Duration = 500 * time.Millisecond
	watchdogTimeout     time.Duration = 1 * time.Second
	watchdogResetPeriod time.Duration = 300 * time.Millisecond
	floorCount          int           = 4
)

var communityState assigner.CommunityState

var applicationTimeoutCh chan string

var slaves = make(map[string]*slaveType)

func Run(masterPort int, quitCh chan struct{}) {

	rblog.Magenta.Print("--- Starting master ---")

	communityState.HallRequests = make(mscomm.HallRequests, floorCount)
	communityState.States = make(map[string]mscomm.ElevatorState)

	syncAttempts := make(map[int]syncAttemptType)

	//init watchdog
	watchdog := time.AfterFunc(watchdogTimeout, func() {
		panic("Vaktbikkje sier voff! - master deadlock?")
	})
	defer watchdog.Stop()

	//Careful when sending to slaveChans. If a slave is disconnected, noone will read from the channel, and it will block forever :(
	//TODO: deal with that

	//Should we have a separate map for hiredSlaves so that we do not attempt to sync and so on with slaves that are still in the application process?
	//Maybe make a map containing structs that have a channel and a bool for hired or not?

	applicationTimeoutCh = make(chan string)
	syncTimeoutCh := make(chan int)
	terminateCh := make(chan struct{})

	fromSlaveCh := make(chan mscomm.Package)
	slaveConnEventCh := make(chan mscomm.ConnectionEvent)

	listener, err := server.Listen(masterPort)
	if err != nil {
		panic(fmt.Sprint("Could not get listener", err))
	}
	defer listener.Close()

	go server.Acceptor(listener, fromSlaveCh, slaveConnEventCh)

	for {
		select {
		case slave := <-slaveConnEventCh:
			if slave.Connected {
				rblog.Magenta.Println("slave connected: ", slave.Addr)
				slaves[slave.Addr] = &slaveType{
					ch: slave.Ch,
				}

				go func() {
					time.Sleep(applicationTimeout)
					applicationTimeoutCh <- slave.Addr
				}()

				slave.Ch <- mscomm.RequestHallRequests{}

			} else {
				rblog.Magenta.Println("slave disconnected: ", slave.Addr)
				dismiss(slave.Addr)
				beginAssignment()
			}
		case addr := <-applicationTimeoutCh:
			if !slaves[addr].hired {
				rblog.Yellow.Println("slave did not meet application deadline:", addr)
				dismiss(addr)
			}
		case syncId := <-syncTimeoutCh:
			if _, exists := syncAttempts[syncId]; exists {
				rblog.Yellow.Println("sync attempt timed out")
				//dismiss pending slaves??
				delete(syncAttempts, syncId)
			}
		case <-quitCh:
			listener.Close()
			for addr := range slaves {
				dismiss(addr)
			}
			//some delay to clear channels before terminating
			go func() {
				time.Sleep(100 * time.Millisecond)
				terminateCh <- struct{}{}
			}()

		case <-terminateCh:
			rblog.Magenta.Print("--- Terminating master ---")
			return
		case message := <-fromSlaveCh:
			if _, exists := slaves[message.Addr]; !exists {
				rblog.Red.Println("master received message from unknown slave", message.Addr)
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
				//TODO: Should also make sure that the slave receives the hall requests from the master
				//This will be handled by starting assignment process when the slave is hired???

				beginAssignment()

			case mscomm.ButtonPressed:
				buttonPressed := message.Payload.(mscomm.ButtonPressed)
				if buttonPressed.Button >= 2 {
					continue // ignore cab requests
				}

				syncId := rand.Int()

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
						syncAttempts[syncId].pendingSlaves[addr] = struct{}{}
						slave.ch <- syncRequests
					}

				}

				go func() {
					time.Sleep(syncTimeout)
					syncTimeoutCh <- syncId
				}()

			case mscomm.ElevatorState:
				communityState.States[message.Addr] = message.Payload.(mscomm.ElevatorState)
				slaves[message.Addr].statePending = false
				anyonePending := false

				for _, slave := range slaves {
					if slave.statePending {
						anyonePending = true
						break
					}
				}

				if !anyonePending {
					//all states received. ready to assign
					assignedRequests, err := assigner.Assign(&communityState)
					if err != nil {
						rblog.Red.Println("assigner error:", err)
						continue
					}
					for addr, requests := range *assignedRequests {
						slaves[addr].ch <- requests
					}
				}

			case mscomm.OrderComplete:
				orderComplete := message.Payload.(mscomm.OrderComplete)
				communityState.HallRequests[orderComplete.Floor][orderComplete.Button] = false
				syncRequests := mscomm.SyncRequests{
					Requests: communityState.HallRequests,
					Id:       -1, //Unsafe sync
				}
				for _, slave := range slaves {
					if slave.hired {
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
				delete(syncAttempts[syncId].pendingSlaves, message.Addr)
				if len(syncAttempts[syncId].pendingSlaves) == 0 {
					//sync successful
					communityState.HallRequests.Set(syncAttempts[syncId].button)
					for _, slave := range slaves {
						if slave.hired {
							slave.ch <- mscomm.Lights(communityState.HallRequests)
						}
					}
					delete(syncAttempts, syncId)
					beginAssignment()
				}

			default:
				rblog.Red.Println("master received unknown message type", reflect.TypeOf(message.Payload).Name(), "from", message.Addr)

			}
		case <-time.After(watchdogResetPeriod):
			//unblock select to reset watchdog
		}

		watchdog.Reset(watchdogTimeout)
	}
}
func dismiss(addr string) {
	if _, exists := slaves[addr]; exists {
		close(slaves[addr].ch)
		delete(slaves, addr)
	}
	delete(communityState.States, addr)
}

func beginAssignment() {
	for _, slave := range slaves {
		if slave.hired {
			slave.ch <- mscomm.RequestState{}
			slave.statePending = true
		}
	}
}
