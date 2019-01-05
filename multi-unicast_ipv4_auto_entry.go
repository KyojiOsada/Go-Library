package main

import (
    "flag"
    "fmt"
    "net"
    "strings"
)

var DefaultIp string = "192.0.2.1" // node2 は 192.0.2.2
var DefaultPort string = ":56789"
var DefaultAddr string = DefaultIp + DefaultPort
var ReceptorPort string = ":56788"
var ReceptorAddr string = DefaultIp + ReceptorPort
var ApplyingToIp string = ""
var ApplyingToAddr string = ""
var ApplyingMessage string = "applying"
var EntryPrefix string = "entry:"
var MultiUnicastAddresses []string
var BufferByte int = 64

func main() {

    // Start receptor
    receptor_addr_byte, err := net.ResolveUDPAddr("udp", ReceptorAddr)
    Error(err)

    // Listen receptor
    receptor, err := net.ListenUDP("udp", receptor_addr_byte)
    Error(err)
    defer receptor.Close()
    fmt.Printf("Listened receptor *:* > %s\n", ReceptorAddr)

    receptor_buffer := make([]byte, BufferByte)
    go func() {
        for {
            // Receive applying message
            length, applying_addr_byte, err := receptor.ReadFrom(receptor_buffer)
            Error(err)
            applying_message := string(receptor_buffer[:length])
            applying_addr := applying_addr_byte.(*net.UDPAddr).String()

            applying_addr_parts := strings.Split(applying_addr, ":")
            entry_ip := applying_addr_parts[0]
            entry_addr := entry_ip + DefaultPort

            if applying_message != ApplyingMessage {
                continue
            }
            fmt.Printf("Reveived applying %s > %s as “%s”\n", applying_addr, ReceptorAddr, applying_message)

            // Check duplicated addresses
            entered_flag := false
            for _, addr := range MultiUnicastAddresses {
                if addr == entry_addr {
                    entered_flag = true
                    fmt.Printf("Duplicated entry address in multi-unicast addresses: %s\n", entry_addr)
                    break
                }
            }

            if !entered_flag {
                // Append entry address
                MultiUnicastAddresses = append(MultiUnicastAddresses, entry_addr)
                fmt.Printf("Appended entry address into multi-unicast addresses: %s\n", entry_addr)

                // Send entry message
                entry_message := EntryPrefix + strings.Join(MultiUnicastAddresses, ",")
                SendMultiUnicast(entry_message)
            }
        }
    }()

    // Apply own address to multi-unicast group
    flag.Parse()
    ApplyingToIp = flag.Arg(0)
    ApplyingToAddr = ApplyingToIp + ReceptorPort
    Apply()

    // Start inbound
    inbound_to_addr_byte, err := net.ResolveUDPAddr("udp", DefaultAddr)
    Error(err)

    // Listen inbound
    inbound, err := net.ListenUDP("udp", inbound_to_addr_byte)
    Error(err)
    defer inbound.Close()
    fmt.Printf("Listened inbound *:* > %s\n", DefaultAddr)

    inbound_buffer := make([]byte, BufferByte)
    for {
        // Receive inbound message
        length, inbound_from_addr, err := inbound.ReadFrom(inbound_buffer)
        Error(err)
        inbound_message := string(inbound_buffer[:length])

        // Receive entry message
        if 0 == strings.Index(inbound_message, EntryPrefix) {
            fmt.Printf("Reveived entry message %s > %s as “%s”\n", inbound_from_addr, DefaultAddr, inbound_message)

            // Extract entry address
            entry_messages := strings.Split(inbound_message, EntryPrefix)
            entry_address := entry_messages[1]
            entry_addresses := strings.Split(entry_address, ",")

            for _, entry_addr := range entry_addresses {
                entered_flag := false
                for _, addr := range MultiUnicastAddresses {
                    if addr == entry_addr {
                        entered_flag = true
                        fmt.Printf("Duplicated entry address in multi-unicast addresses: %s\n", entry_addr)
                        break
                    }
                }

                if !entered_flag {
                    MultiUnicastAddresses = append(MultiUnicastAddresses, entry_addr)
                    fmt.Printf("Appended entry address into multi-unicast addresses: %s\n", entry_addr)
                }
            }
        }

        // Receive others message
        if 0 == strings.Index(inbound_message, "others:") {
            // Process others
        }
    }
}

func Error(_err error) {
    if _err != nil {
        panic(_err)
    }
}

func Apply() {
    // Start applying
    applying_to_addr, err := net.ResolveUDPAddr("udp", ApplyingToAddr)
    applying, err := net.DialUDP("udp", nil, applying_to_addr)
    Error(err)
    defer applying.Close()
    fmt.Printf("Connected applying > %s\n", ApplyingToAddr)

    // Outbound applying message
    applying.Write([]byte(ApplyingMessage))
    fmt.Printf("Outbound applying > %s as “%s”\n", ApplyingToAddr, ApplyingMessage)
}

func SendMultiUnicast(_message string) {
    // Start multi-unicast
    for _, to_addr := range MultiUnicastAddresses {

        go func(_to_addr string) {
            outbound_to_addr, err := net.ResolveUDPAddr("udp", _to_addr)
            multi_unicast, err := net.DialUDP("udp", nil, outbound_to_addr)
            Error(err)

            defer multi_unicast.Close()
            fmt.Printf("Connected multi-unicast > %s\n", outbound_to_addr)

            // Multi-unicast message
            multi_unicast.Write([]byte(_message))
            fmt.Printf("Multi-unicast > %v as “%s”\n", outbound_to_addr, _message)
        }(to_addr)
    }
}
