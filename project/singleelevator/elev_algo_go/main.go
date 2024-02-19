package main

import (
    "fmt"
    "single/elevio"
)


func main(void) int{
    fmt.Printf("Started!\n");
    
    inputPollRate_ms := 25;

    elevio.Init();  // Initilizes driver. Sets up connection to elevator
    
    ifelevio.GetFloor() == -1{
        // outputDevice.motorDirection(D_Down);
        // elevator.dirn = D_Down;
        // elevator.behaviour = EB_Moving;
    }
        
    while(1){
        { // Request button
            static int prev[N_FLOORS][N_BUTTONS];
            for(int f = 0; f < N_FLOORS; f++){
                for(int b = 0; b < N_BUTTONS; b++){
                    int v = input.requestButton(f, b);
                    if(v  &&  v != prev[f][b]){
                        fsm_onRequestButtonPress(f, b);
                    }
                    prev[f][b] = v;
                }
            }
        }
        
        { // Floor sensor
            static int prev = -1;
            int f = input.floorSensor();
            if(f != -1  &&  f != prev){
                fsm_onFloorArrival(f);
            }
            prev = f;
        }
        
        
        { // Timer
            if(timer_timedOut()){
                timer_stop();
                fsm_onDoorTimeout();
            }
        }
        
        usleep(inputPollRate_ms*1000);
    }
}









