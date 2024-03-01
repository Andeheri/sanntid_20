package master

import (
    "fmt"
    "net"
)

var MASTER_PORT = 42752

func Start(){
	go listener()
}

func listener() {
	listener, err := net.Listen("tcp", fmt.Sprint(MASTER_PORT))
	if err != nil {
        fmt.Println("Error:", err)
        return
    }
	defer listener.Close()

	for {

        slaveConn, err := listener.Accept()
        if err != nil {
            fmt.Println("Error:", err)
            continue
        }

        _ = slaveConn
        //add slave
    }

}

