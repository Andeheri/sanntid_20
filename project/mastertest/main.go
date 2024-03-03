package main

import (
	"fmt"
	"project/master/slavecomm"
	"reflect"
)

func main() {
	go slavecomm.Manager()

	testch := make(chan slavecomm.SlaveMessage)
	go slavecomm.Listener(12221, testch)

	for {
		data := <-testch
		fmt.Println(reflect.TypeOf(data.Payload), data.Payload)
	}

	select {}
}
