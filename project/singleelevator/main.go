package main

import (
	"fmt"
	// "single/elevator"
	"single/elevio"
	"single/fsm"
	// "single/iodevice"
	"single/timer"
)

func main(){
    numFloors := 4

    elevio.Init("localhost:15657", numFloors)
    
    var d elevio.MotorDirection = elevio.MD_Up
    // elevio.SetMotorDirection(d)
    
    drv_buttons := make(chan elevio.ButtonEvent)
    drv_floors  := make(chan int)
    drv_obstr   := make(chan bool)
    drv_stop    := make(chan bool)    
    
    go elevio.PollButtons(drv_buttons)
    go elevio.PollFloorSensor(drv_floors)
    go elevio.PollObstructionSwitch(drv_obstr)
    go elevio.PollStopButton(drv_stop)

    // input := iodevice.Elevio_getInputDevice()
    fsm.Fsm_init()
    
    for {
        select {
        case a := <- drv_buttons:
            fmt.Printf("%+v\n", a)
            fsm.Fsm_onRequestButtonPress(a.Floor, a.Button)
            
        case a := <- drv_floors:
            fmt.Printf("%+v\n", a)
            fsm.Fsm_onFloorArrival(a)
            
            
        case a := <- drv_obstr:
            fmt.Printf("%+v\n", a)
            if a {
                elevio.SetMotorDirection(elevio.MD_Stop)
            } else {
                elevio.SetMotorDirection(d)
            }
            
        case a := <- drv_stop:
            fmt.Printf("%+v\n", a)
            for f := 0; f < numFloors; f++ {
                for b := elevio.ButtonType(0); b < 3; b++ {
                    elevio.SetButtonLamp(b, f, false)
                }
            }
        }

        if timer.Timer_timedOut() {
            timer.Timer_stop()
            fsm.Fsm_onDoorTimeout()
        }
    }    
}