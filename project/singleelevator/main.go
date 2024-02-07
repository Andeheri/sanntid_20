package main

import (
	"fmt"
	//"single/elevator"
	"single/elevio"
	"single/fsm"

	// "single/iodevice"
	//"single/timer"
	"time"
)


func main(){
    numFloors := 4

    elevio.Init("localhost:15657", numFloors)
    
    
    drv_buttons := make(chan elevio.ButtonEvent)
    drv_floors  := make(chan int)
    drv_obstr   := make(chan bool)

    door_timer := time.NewTimer(-1)


    go elevio.PollButtons(drv_buttons)
    go elevio.PollFloorSensor(drv_floors)
    go elevio.PollObstructionSwitch(drv_obstr)

    fsm.Fsm_init()

    for {
        select {
        case a := <- drv_buttons:
            fmt.Printf("%+v\n", a)
            fsm.Fsm_onRequestButtonPress(a.Floor, a.Button, door_timer)
            
        case a := <- drv_floors:
            fmt.Printf("%+v\n", a)
            fsm.Fsm_onFloorArrival(a, door_timer)
            
        case a := <- drv_obstr:
            fmt.Printf("%+v\n", a)
            fsm.OnObstruction(a)

        case a := <- door_timer.C:
            fmt.Printf("%+v\n", a)
            fsm.Fsm_onDoorTimeout(door_timer)
        }
    }
}