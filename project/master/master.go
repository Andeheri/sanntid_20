package master

import (
	"fmt"
	"math/rand"
	"project/master/assigner"
	"project/mscomm"
	"reflect"
	"time"
)

var communityState assigner.CommunityState
var slaveChans map[string]chan interface{}

var applicantSlaves map[string]*time.Timer

const applicationTimeout = 1 * time.Second //Is a timeout actually needed?
var applicationTimeoutCh chan string

var statePending map[string]struct{}

func Run(fromSlaveCh chan mscomm.Package, slaveConnEventCh chan mscomm.ConnectionEvent) {

	const floorCount int = 4

	communityState.HallRequests = make(mscomm.HallRequests, floorCount)
	communityState.States = make(map[string]mscomm.ElevatorState)

	slaveChans = make(map[string]chan interface{})

	//init watchdog
	const (
		watchdogTimeout     time.Duration = 1 * time.Second
		watchdogResetPeriod time.Duration = 300 * time.Millisecond
	)
	watchdog := time.AfterFunc(watchdogTimeout, func() {
		panic("Vaktbikkje sier voff! - master deadlock?")
	})
	defer watchdog.Stop()

	//Careful when sending to slaveChans. If a slave is disconnected, noone will read from the channel, and it will block forever :(
	//TODO: deal with that

	//Should we have a separate map for hiredSlaves so that we do not attempt to sync and so on with slaves that are still in the application process?
	//Maybe make a map containing structs that have a channel and a bool for hired or not?

	applicantSlaves = make(map[string]*time.Timer)
	applicationTimeoutCh = make(chan string)

	currentSyncId := -1 //-1 means not syncing
	syncButton := mscomm.ButtonPressed{}
	syncPending := make(map[string]struct{})
	syncTimeoutCh := make(chan int)
	const syncTimeout = 500 * time.Millisecond

	statePending = make(map[string]struct{})

	for {
		select {
		case slave := <-slaveConnEventCh:
			onConnectionEvent(&slave)
		case addr := <-applicationTimeoutCh:
			delete(applicantSlaves, addr)
			//Force disconnect???
		case syncId := <-syncTimeoutCh:
			if syncId == currentSyncId {
				//sync failed
				currentSyncId = -1
				syncPending = make(map[string]struct{})
			}
		case message := <-fromSlaveCh:
			fmt.Println("received", reflect.TypeOf(message.Payload), "from", message.Addr)
			fmt.Println(message.Payload)

			switch message.Payload.(type) {

			case mscomm.HallRequests:
				//Slave is hired!
				slaveHallRequests := message.Payload.(mscomm.HallRequests)
				communityState.HallRequests.Merge(&slaveHallRequests)
				delete(applicantSlaves, message.Addr) // is this necessary? timeout should handle this
				//TODO: Should also make sure that the slave receives the hall requests from the master
				//This will be handled by starting assignment process when the slave is hired???

				beginAssignment()

			case mscomm.ButtonPressed:
				buttonPressed := message.Payload.(mscomm.ButtonPressed)
				if buttonPressed.Button >= 2 {
					continue // ignore cab requests
				}

				currentSyncId = rand.Int()
				syncButton = buttonPressed

				syncRequests := mscomm.SyncRequests{
					Requests: communityState.HallRequests,
					Id:       currentSyncId,
				}
				syncRequests.Requests[buttonPressed.Floor][buttonPressed.Button] = true

				for addr, ch := range slaveChans {
					if _, isApplicant := applicantSlaves[addr]; isApplicant {
						continue //skip applicants
					}
					syncPending[addr] = struct{}{}
					ch <- syncRequests
				}

				//TEST. REMOVE AFTER TESTING
				// assignedOrder := make(mscomm.AssignedRequests, floorCount)
				// assignedOrder[buttonPressed.Floor][buttonPressed.Button] = true
				// slaveChans[message.Addr] <- assignedOrder

				go func() {
					time.Sleep(syncTimeout)
					syncTimeoutCh <- currentSyncId
				}()

			case mscomm.ElevatorState:
				communityState.States[message.Addr] = message.Payload.(mscomm.ElevatorState)
				delete(statePending, message.Addr)
				if len(statePending) == 0 {
					//all states received. ready to assign
					assignedRequests, err := assigner.Assign(&communityState)
					if err != nil {
						fmt.Println("assigner error:", err)
						continue
					}
					for addr, requests := range *assignedRequests {
						slaveChans[addr] <- requests
					}
				}

			case mscomm.OrderComplete:
				orderComplete := message.Payload.(mscomm.OrderComplete)
				communityState.HallRequests[orderComplete.Floor][orderComplete.Button] = false
				syncRequests := mscomm.SyncRequests{
					Requests: communityState.HallRequests,
					Id:       -1, //Unsafe sync
				}
				for addr, ch := range slaveChans {
					if _, isApplicant := applicantSlaves[addr]; isApplicant {
						continue
					}
					ch <- syncRequests
					ch <- communityState.HallRequests.ToLights()
				}

				//Do not need to assign here, right?

			case mscomm.SyncOK:
				syncId := message.Payload.(mscomm.SyncOK).Id
				if syncId != currentSyncId {
					continue //ignore
				}
				delete(syncPending, message.Addr)
				if len(syncPending) == 0 {
					//sync successful
					currentSyncId = -1
					communityState.HallRequests[syncButton.Floor][syncButton.Button] = true
					for addr, ch := range slaveChans {
						if _, isApplicant := applicantSlaves[addr]; isApplicant {
							continue
						}
						ch <- communityState.HallRequests.ToLights()
					}
					beginAssignment()
				}

			default:
				fmt.Println("master received unknown message type from", message.Addr)

			}
		case <-time.After(watchdogResetPeriod):
			//unblock select to reset watchdog
		}

		watchdog.Reset(watchdogTimeout)
	}
}

func onConnectionEvent(slave *mscomm.ConnectionEvent) {
	if slave.Connected {
		fmt.Println("slave connected:", slave.Addr)
		slaveChans[slave.Addr] = slave.Ch

		//reset timer if already exists??
		applicantSlaves[slave.Addr] = time.AfterFunc(applicationTimeout, func() {
			applicationTimeoutCh <- slave.Addr
		})
		slave.Ch <- mscomm.RequestHallRequests{}
	} else {
		fmt.Println("slave disconnected:", slave.Addr)
		close(slaveChans[slave.Addr])
		delete(slaveChans, slave.Addr)
		delete(communityState.States, slave.Addr)
		beginAssignment()
	}
}

func beginAssignment() {
	for addr, ch := range slaveChans {
		if _, isApplicant := applicantSlaves[addr]; isApplicant {
			continue
		}
		ch <- mscomm.RequestState{}
		statePending[addr] = struct{}{}
	}
}