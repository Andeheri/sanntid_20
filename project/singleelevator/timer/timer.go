package timer

import(
	"time"
)

func Get_wall_time() float64{
    now := time.Now();
    return float64(now.Unix()) + float64(now.Nanosecond())/1e9
}

var timerEndTime float64;
var timerActive bool;

func Timer_start(duration float64){
    timerEndTime    = Get_wall_time() + duration;
    timerActive     = true;
}

func Timer_stop(){
    timerActive = false;
}

func Timer_timedOut() bool{
    return timerActive && (Get_wall_time() > timerEndTime)
}