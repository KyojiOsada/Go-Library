package main

import (
    "fmt"
    "net"
)

func main() {
    addresses, err := net.InterfaceAddrs()
    Error(err)

    for _, address := range addresses {
        if ipnet, ok := address.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
            if ipnet.IP.To4() != nil {
                fmt.Printf("%s\n", ipnet.IP.String())
            }
        }
    }
}

func Error(_err error) {
    if _err != nil {
        panic(_err)
    }
}
