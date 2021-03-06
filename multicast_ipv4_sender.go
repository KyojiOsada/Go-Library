package main

import (
	"fmt"
	"net"
	"strconv"
	"time"
)

func main() {
	ipv4 := "224.0.0.1"
	port := ":9999"
	addr := ipv4 + port
	wait_time := 1
	format := "Message "

	fmt.Println("Sender:", addr)
	conn, err := net.Dial("udp", addr)
	_error(err)
	defer conn.Close()

	c := 0
	for {
		time.Sleep(time.Duration(wait_time) * time.Second)
		message := format + strconv.Itoa(c)
		conn.Write([]byte(message))
		fmt.Printf("%s\n", message)
		c++
	}
}

func _error(_err error) {
	if _err != nil {
		panic(_err)
	}
}
