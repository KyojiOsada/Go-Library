package main

import (
    "fmt"
    "net"
    "os"
    "time"
)

var InboundToAddr string = "192.0.2.255:56789"
var OutboundFromAddr string = "192.0.2.1:56789"
var OutboundToAddr string = "192.0.2.255/56789"
var BufferByte int = 64
var IntervalSeconds int = 1

func main() {

    // Start inbound
    go func() {
        inbound_to_addr_byte, err := net.ResolveUDPAddr("udp", InboundToAddr)
        Error(err)

        inbound, err := net.ListenUDP("udp", inbound_to_addr_byte)
        Error(err)
        defer inbound.Close()
        fmt.Printf("Listened *:* > %s\n", InboundToAddr)

        buffer := make([]byte, BufferByte)
        for {
            // Inbound message
            length, inbound_from_addr_byte, err := inbound.ReadFrom(buffer)
            Error(err)
            inbound_message := string(buffer[:length])

            inbound_from_addr := inbound_from_addr_byte.(*net.UDPAddr).String()

            fmt.Printf("Inbound %v > %v as “%s”\n", inbound_from_addr, InboundToAddr, inbound_message)
        }
    }()

    // Start outbound
    outbound_from_addr, err := net.ResolveUDPAddr("udp", OutboundFromAddr)
    outbound_to_addr, err := net.ResolveUDPAddr("udp", OutboundToAddr)
    outbound, err := net.DialUDP("udp", outbound_from_addr, outbound_to_addr)
    Error(err)
    defer outbound.Close()
    fmt.Printf("Connected %s > %s\n", OutboundFromAddr, OutboundToAddr)

    // Get hostname
    outbound_message, err := os.Hostname()
    Error(err)

    for {
        time.Sleep(time.Duration(IntervalSeconds) * time.Second)

        // Outbound message
        outbound.Write([]byte(outbound_message))
        fmt.Printf("Outbound %v > %v as “%s”\n", outbound_from_addr, outbound_to_addr, outbound_message)
    }

}

func Error(_err error) {
    if _err != nil {
        panic(_err)
    }
}
