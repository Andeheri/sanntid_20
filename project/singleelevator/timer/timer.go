package timer

import(
	"time"
    "fmt"
)

/*

func Get_wall_time() float64{
    now := time.Now();
    return float64(now.Unix()) //+ float64(now.Nanosecond())/1e9
}

var timerEndTime float64;
var timerActive bool;

func Timer_start(duration float64){
    
    timerEndTime    = Get_wall_time() + duration;
    timerActive     = true;

    fmt.Println("Timer start")
}

func Timer_stop(){
    timerActive = false;
    
    fmt.Println("Timer stop")
}

func Timer_timedOut() bool{
    fmt.Println("Timer timed out")
    return timerActive && (Get_wall_time() > timerEndTime)
}
*/


var timerEndTime time.Time
var timerActive bool

func Timer_start(duration float64){
    
    timerEndTime = time.Now().Add(time.Duration(duration)*time.Second)
    timerActive     = true;

    fmt.Println("Timer start")
}

func Timer_stop(){
    timerActive = false;
    
    fmt.Println("Timer stop")
}

func Timer_timedOut(tim chan<- bool) bool{
    for{
        if timerActive && time.Now().After(timerEndTime){
            tim <- true
            Timer_stop()
        }
        time.Sleep(time.Millisecond*3)
    }
}

