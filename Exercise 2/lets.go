package main

import (
	"fmt"
	"net"
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
	server_conn, _ := net.DialUDP("udp4", nil, server_ip)

	my_ip, _ := net.ResolveUDPAddr("udp4",":20020")
	read_conn, _ := net.ListenUDP("udp4", my_ip);

	message := []byte("Halla balla");
	server_conn.Write(message);

	read_buf := make([]byte, 1024);
	_, source_addr, _ := read_conn.ReadFromUDP(read_buf);
	fmt.Printf("%s is saying %s\n", source_addr.IP.String(), read_buf);

}



func main(){
	SendUDP();
}