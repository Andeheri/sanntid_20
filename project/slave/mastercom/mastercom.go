package mastercom

import(
	"fmt"
	"slave/elevio"
	"slave/iodevice"
	"slave/fsm"
	"time"
)

type Master_channels struct {
	Button_press chan elevio.ButtonEvent
	Clear_request chan elevio.ButtonEvent

	Master_requests chan [iodevice.N_FLOORS][iodevice.N_BUTTONS] int

}

func Master_communication(chans Master_channels, door_timer *time.Timer){
	for {
		select {
		case a := <- chans.Button_press:
			fmt.Println(a, "sender button press melding")
			//send message over TCP
		case a := <- chans.Clear_request:
			fmt.Println(a, "sender clear request melding")
			//send message over TCP
		case a := <- chans.Master_requests:
			fmt.Println(a, "mottat master request melding")
			//mottatt melding over TCP
			fsm.Requests_clearAll()
			fsm.Requests_setAll(a, door_timer)
		}
	}	
}