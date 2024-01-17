package main

import (
	"fmt"
	"net"
	"time"
)

func findServer() net.IP{
	my_ip, _ := net.ResolveUDPAddr("udp4", "0.0.0.0:30000")
	conn, _ := net.ListenUDP("udp4", my_ip);
	read_buf := make([]byte, 1024);
	_, source_addr, _ := conn.ReadFromUDP(read_buf);
	// fmt.Printf("%s is saying %s\n", source_addr.IP.String(), read_buf);
	return source_addr.IP

}

func UDPDump() {
	my_ip, _ := net.ResolveUDPAddr("udp4", "0.0.0.0:30000")
	conn, _ := net.ListenUDP("udp4", my_ip);
	read_buf := make([]byte, 1024);
	for{
		_, source_addr, _ := conn.ReadFromUDP(read_buf);
		fmt.Printf("%s is saying %s\n", source_addr.IP.String(), read_buf);
	}

}

func SendUDP() {
	server_ip, _ := net.ResolveUDPAddr("udp4", "10.100.23.129:20020");
	my_ip, _ := net.ResolveUDPAddr("udp4","10.100.23.30:20020");
	server_conn, _ := net.DialUDP("udp4", nil, server_ip);

	read_conn, _ := net.ListenUDP("udp4", my_ip);

	message := []byte("Halla balla");
	server_conn.Write(message);

	time.Sleep(10* time.Second)
	

	read_buf := make([]byte, 1024);
	_, source_addr, _ := read_conn.ReadFromUDP(read_buf);
	fmt.Printf("%s is saying %s\n", source_addr.IP.String(), read_buf);


}

func TCP_Sender(conn *net.TCPConn){
	for i := 0; true; i++ {
		conn.Write([]byte("Hei \000"))
		time.Sleep(1 * time.Second)
		
	}
}

func TCP_Receiver(conn *net.TCPConn){
	for {
		reply := make([]byte, 1024)
		conn.Read(reply)
		println(string(reply))
	}
}



func main(){
	server_ip, _ := net.ResolveTCPAddr("tcp", "10.100.23.129:34933");
	//my_ip, _ := net.ResolveTCPAddr("tcp","10.100.23.30:20021");
	server_conn, _ := net.DialTCP("tcp", nil, server_ip);

	go TCP_Sender(server_conn)
	go TCP_Receiver(server_conn)
	select{}
}