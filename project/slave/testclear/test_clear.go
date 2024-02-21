package testclear

import(
	"time"
)


func Set_master_test(receiver chan<- bool){
	for{
		time.Sleep(time.Duration(10)*time.Second)
		receiver <- true
	}
}