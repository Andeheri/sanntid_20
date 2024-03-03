package master

import (
	"math/rand"
	"project/commontypes"
	"project/master/assigner"
	"project/master/slavecomm"
	"time"
)

var MASTER_PORT = 42752

func Run(fromSlaveCh chan slavecomm.SlaveMessage, slaveConnEventCh chan slavecomm.ConnectionEvent) {

	communityState := assigner.CommunityState{}

	slaveChans := make(map[string]chan interface{})

	applicantSlaves := make(map[string]*time.Timer)
	applicationTimeoutCh := make(chan string)
	const applicationTimeout = 1 * time.Second //Is a timeout actually needed?

	currentSyncId := -1 //-1 means not syncing
	syncButton := commontypes.ButtonPressed{}
	syncPending := make(map[string]struct{})
	syncTimeoutCh := make(chan int)

	for {
		select {
		case slave := <-slaveConnEventCh:
			if slave.Connected {
				slaveChans[slave.Addr] = slave.Ch

				//reset timer if already exists??
				applicantSlaves[slave.Addr] = time.AfterFunc(applicationTimeout, func() {
					applicationTimeoutCh <- slave.Addr
				})
				slave.Ch <- commontypes.RequestHallRequests{}
			} else {
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

			switch message.Payload.(type) {

			case commontypes.HallRequests:
				slaveHallRequests := message.Payload.(commontypes.HallRequests)
				communityState.HallRequests.Merge(&slaveHallRequests)
				delete(applicantSlaves, message.Addr) // is this necessary? timeout should handle this

			case commontypes.ButtonPressed:
				buttonPressed := message.Payload.(commontypes.ButtonPressed)
				if buttonPressed.Button >= 2 {
					continue // ignore cab requests
				}

				currentSyncId = rand.Int()
				syncButton = buttonPressed

				syncRequests := commontypes.SyncRequests{
					Requests: communityState.HallRequests,
					Id:       currentSyncId,
				}
				syncRequests.Requests[buttonPressed.Floor][buttonPressed.Button] = true

				for addr, ch := range slaveChans {
					syncPending[addr] = struct{}{}
					ch <- syncRequests
				}

				//TODO: create timeout

			case commontypes.ElevatorState:
				communityState.States[message.Addr] = message.Payload.(commontypes.ElevatorState)

			case commontypes.OrderComplete:
				orderComplete := message.Payload.(commontypes.OrderComplete)
				communityState.HallRequests[orderComplete.Floor][orderComplete.Button] = false
				syncRequests := commontypes.SyncRequests{
					Requests: communityState.HallRequests,
					Id:       -1, //Unsafe sync
				}
				for _, ch := range slaveChans {
					ch <- syncRequests
				}

			case commontypes.SyncOK:
				syncId := message.Payload.(commontypes.SyncOK).Id
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

			}
		}
	}
}
