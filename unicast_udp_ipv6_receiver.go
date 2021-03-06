package main

import (
	"fmt"
	"net"
)

func main() {
	ipv6 := "[fe80::4550:8d0c:3ad8:6fb5]"
	port := ":9999"
	addr := ipv6 + port
	fmt.Println("Receiver:", addr)
	udp_addr, err := net.ResolveUDPAddr("udp", addr)
	_error(err)

	listener, err := net.ListenMulticastUDP("udp", nil, udp_addr)
	_error(err)
	defer listener.Close()

	buffer := make([]byte, 1240)
	for {
		length, remoteAddress, err := listener.ReadFrom(buffer)
		_error(err)

		fmt.Printf("Sender: %v\n", remoteAddress)
		fmt.Printf("Contents: %s\n", string(buffer[:length]))
	}
}

func _error(_err error) {
	if _err != nil {
		panic(_err)
	}
}
