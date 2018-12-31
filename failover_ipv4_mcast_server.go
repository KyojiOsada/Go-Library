package main

import (
    "fmt"
    "net"
    "time"
)

func main() {
    mcast_ip4 := "224.0.0.1"
    mcast_port := ":9999"
    mcast_addr := mcast_ip4 + mcast_port
    wait_time := 1
    message := "master"

    fmt.Printf("Send to multicast address: %s\n", mcast_addr)
    conn, err := net.Dial("udp", mcast_addr)
    _Error(err)
    defer conn.Close()

    for {
        time.Sleep(time.Duration(wait_time) * time.Second)
        conn.Write([]byte(message))
        fmt.Printf("%s\n", message)
    }
}

func _Error(_err error) {
    if _err != nil {
        panic(_err)
    }
}
