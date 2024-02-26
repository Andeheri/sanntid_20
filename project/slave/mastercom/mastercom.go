package mastercom

import(
	"fmt"
	"slave/elevio"
	"slave/iodevice"
	"slave/fsm"
	"time"
	// "encoding/json"
)

type Master_channels struct {
	Button_press chan elevio.ButtonEvent
	Clear_request chan elevio.ButtonEvent

	Master_requests chan [iodevice.N_FLOORS][iodevice.N_BUTTONS] int
	RequestedState chan bool

	
}



func Master_communication(chans Master_channels, door_timer *time.Timer){
	for {
		select {
		case a := <- chans.Button_press:
			fmt.Println(a, "sender button press melding")
			//send message over TCP
			//hvis det ikke er noen forbindelse "fsm on button press" direkte
		case a := <- chans.Clear_request:
			fmt.Println(a, "sender clear request melding")
			//send message over TCP
		case a := <- chans.Master_requests:
			fmt.Println(a, "mottat master request meldin g")
			//mottatt melding over TCP
			fsm.Requests_clearAll()
			fsm.Requests_setAll(a, door_timer)
		
		case a := <- chans.RequestedState:
			fmt.Println(a, "sender state til master")
			SendState()
		}	
	}
}

type HRAElevState struct {
    Behavior    string      `json:"behaviour"`
    Floor       int         `json:"floor"` 
    Direction   string      `json:"direction"`
    CabRequests []bool      `json:"cabRequests"`
}

func SendState(){
	state := HRAElevState{
		Behavior: string(fsm.Elev.Behaviour),
		Floor: fsm.Elev.Floor,
		Direction: elevio.Elevio_dirn_toString(fsm.Elev.Dirn),
		CabRequests: fsm.GetCabRequests(),
	}
	
	fmt.Println(state)
	// jsonBytes, err := json.Marshal(state)
	//send the state over tcp...

}
