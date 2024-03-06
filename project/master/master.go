package master

import (
	"fmt"
	"math/rand"
	"project/master/assigner"
	"project/mscomm"
	"reflect"
	"time"
)

func Run(fromSlaveCh chan mscomm.Package, slaveConnEventCh chan mscomm.ConnectionEvent) {

	const floorCount int = 4
	communityState := assigner.CommunityState{}
	communityState.HallRequests = make(mscomm.HallRequests, floorCount)
	communityState.States = make(map[string]mscomm.ElevatorState)

	//init watchdog
	const (
		watchdogTimeout     time.Duration = 1 * time.Second
		watchdogResetPeriod time.Duration = 300 * time.Millisecond
	)
	vaktbikkje := time.AfterFunc(watchdogTimeout, func() {
		panic("Vaktbikkje sier voff! - master deadlock?")
	})
	defer vaktbikkje.Stop()

	slaveChans := make(map[string]chan interface{})
	//Careful when sending to slaveChans. If a slave is disconnected, noone will read from the channel, and it will block forever :(
	//TODO: deal with that

	//Should we have a separate map for hiredSlaves so that we do not attempt to sync and so on with slaves that are still in the application process?
	//Maybe make a map containing structs that have a channel and a bool for hired or not?

	applicantSlaves := make(map[string]*time.Timer)
	applicationTimeoutCh := make(chan string)
	const applicationTimeout = 1 * time.Second //Is a timeout actually needed?

	currentSyncId := -1 //-1 means not syncing
	syncButton := mscomm.ButtonPressed{}
	syncPending := make(map[string]struct{})
	syncTimeoutCh := make(chan int)

	for {
		select {
		case slave := <-slaveConnEventCh:
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
				delete(slaveChans, slave.Addr)
				delete(communityState.States, slave.Addr)
			}
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
				slaveHallRequests := message.Payload.(mscomm.HallRequests)
				communityState.HallRequests.Merge(&slaveHallRequests)
				delete(applicantSlaves, message.Addr) // is this necessary? timeout should handle this
				//TODO: Should also make sure that the slave receives the hall requests from the master
				//This will be handled by starting assignment process when the slave is hired???

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
					syncPending[addr] = struct{}{}
					ch <- syncRequests
				}

				//TEST. REMOVE AFTER TESTING
				assignedOrder := make(mscomm.AssignedRequests, floorCount)
				assignedOrder[buttonPressed.Floor][buttonPressed.Button] = true
				slaveChans[message.Addr] <- assignedOrder

				//TODO: create timeout

			case mscomm.ElevatorState:
				communityState.States[message.Addr] = message.Payload.(mscomm.ElevatorState)

			case mscomm.OrderComplete:
				orderComplete := message.Payload.(mscomm.OrderComplete)
				communityState.HallRequests[orderComplete.Floor][orderComplete.Button] = false
				syncRequests := mscomm.SyncRequests{
					Requests: communityState.HallRequests,
					Id:       -1, //Unsafe sync
				}
				for _, ch := range slaveChans {
					ch <- syncRequests
					//Also update lights???
				}

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
					//Assign!!
				}

			default:
				fmt.Println("master received unknown message type from", message.Addr)

			}
		case <-time.After(watchdogResetPeriod):
			//unblock select to reset watchdog
		}

		vaktbikkje.Reset(watchdogTimeout)
	}
}
