package master

import (
	"fmt"
	"math/rand"
	"project/commontypes"
	"project/master/assigner"
	"project/master/slavecomm"
	"reflect"
	"time"
)

func Run(fromSlaveCh chan slavecomm.SlaveMessage, slaveConnEventCh chan slavecomm.ConnectionEvent) {

	const floorCount int = 4
	communityState := assigner.CommunityState{}
	communityState.HallRequests = make([][2]bool, floorCount)

	slaveChans := make(map[string]chan interface{})
	//Careful when sending to slaveChans. If a slave is disconnected, noone will read from the channel, and it will block forever :(
	//TODO: deal with that

	//Should we have a separate map for hiredSlaves so that we do not attempt to sync and so on with slaves that are still in the application process?
	//Maybe make a map containing structs that have a channel and a bool for hired or not?

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
			fmt.Println("received", reflect.TypeOf(message.Payload), "from", message.Addr)
			fmt.Println(message.Payload)

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

			default:
				fmt.Println("master received unknown message type from", message.Addr)

			}
		}
	}
}
