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

		//test for clearing and setting new requests from master
		// case a := <-master_test:
		// 	fmt.Printf("%+v\n", a)
		// 	fsm.Requests_clearAll()
		// 	fsm.Requests_setAll(master_req, door_timer)
