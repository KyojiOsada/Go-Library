package main

import (
    "bufio"
    "encoding/json"
    "flag"
    "fmt"
    "net"
    "os"
    "os/exec"
    "sort"
    "strconv"
    "strings"
    "time"
)

var DefaultIp string = "192.0.2.1" // node2 は 192.0.2.2、node3 は 192.0.2.3
var DefaultPort string = "56789"
var DefaultAddr string = DefaultIp + ":" + DefaultPort
var ReceptorPort string = "56788"
var ReceptorAddr string = DefaultIp + ":" + ReceptorPort
var ApplyingToIp string = ""
var ApplyingToAddr string = ""
var ApplyingMessage string = "applying"
var EntryPrefix string = "entry:"
var BufferByte int = 1024
var NearcastCheckInterval int = 10
var Addresses AddressList

type AddressList []AddressMaps
type AddressMaps struct {
    Address string
    Neary   float64
    Group   int
}

func main() {

    // Start applying receptor
    receptor_addr_byte, err := net.ResolveUDPAddr("udp", ReceptorAddr)
    Error(err)

    // Listen applying receptor
    receptor, err := net.ListenUDP("udp", receptor_addr_byte)
    Error(err)
    defer receptor.Close()
    fmt.Printf("Listened applying receptor *:* > %s\n", ReceptorAddr)

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
            entry_addr := entry_ip + ":" + DefaultPort

            if applying_message != ApplyingMessage {
                continue
            }
            fmt.Printf("Reveived applying %s > %s as “%s”\n", applying_addr, ReceptorAddr, applying_message)

            // Check duplicated addresses
            entered_flag := false
            for _, addr := range Addresses {
                if entry_addr == addr.Address {
                    entered_flag = true
                    fmt.Printf("Duplicated entry address in multi-unicast addresses: %s\n", entry_addr)
                    break
                }
            }

            if !entered_flag {
                // Append entry address
                entry := AddressMaps{Address: entry_addr, Neary: 0, Group: 0}
                Addresses = append(Addresses, entry)
                fmt.Printf("Appended entry address into multi-unicast addresses: %s\n", entry_addr)

                // Send entry message
                addresses_json, err := json.Marshal(Addresses)
                Error(err)
                entry_message := EntryPrefix + string(addresses_json)
                MultiUnicast(entry_message)
            }
        }
    }()

    // Apply own address to address group
    flag.Parse()
    ApplyingToIp = flag.Arg(0)
    ApplyingToAddr = ApplyingToIp + ":" + ReceptorPort
    Apply()

    // Check nearcast
    go CheckNearcast()

    // Start standard input
    go Stdin()

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

            // Extract entry addresses
            entry_messages := strings.Split(inbound_message, EntryPrefix)
            entry_addresses_byte := []byte(entry_messages[1])
            var entry_addresses AddressList
            err := json.Unmarshal(entry_addresses_byte, &entry_addresses)
            Error(err)

            for _, entry_addr_map := range entry_addresses {
                entered_flag := false
                for _, addr_map := range Addresses {
                    if addr_map.Address != entry_addr_map.Address {
                        continue
                    }

                    entered_flag = true
                    fmt.Printf("Duplicated entry address in multi-unicast addresses: %s\n", entry_addr_map.Address)
                    break
                }

                if !entered_flag {
                    entry := AddressMaps{Address: entry_addr_map.Address, Neary: 0, Group: 0}
                    Addresses = append(Addresses, entry)
                    fmt.Printf("Appended entry address into multi-unicast addresses: %s\n", entry_addr_map.Address)
                }
            }
            // Receive nearcast message
        } else if 0 == strings.Index(inbound_message, "multi-unicast:") {
            fmt.Printf("Reveived multi-unicast message %s > %s as “%s”\n", inbound_from_addr, DefaultAddr, inbound_message)
            // Receive nearcast message
        } else if 0 == strings.Index(inbound_message, "nearcast:") {
            fmt.Printf("Reveived nearcast message %s > %s as “%s”\n", inbound_from_addr, DefaultAddr, inbound_message)
            // Receive groupcast message
        } else if 0 == strings.Index(inbound_message, "groupcast:") {
            fmt.Printf("Reveived group message %s > %s as “%s”\n", inbound_from_addr, DefaultAddr, inbound_message)
            // others
        } else {
            // snip...
        }
    }
}

// Error
func Error(_err error) {
    if _err != nil {
        panic(_err)
    }
}

// Apply
func Apply() {
    applying_to_addr, err := net.ResolveUDPAddr("udp", ApplyingToAddr)
    applying, err := net.DialUDP("udp", nil, applying_to_addr)
    Error(err)
    defer applying.Close()
    fmt.Printf("Connected applying > %s\n", ApplyingToAddr)

    // Outbound applying message
    applying.Write([]byte(ApplyingMessage))
    fmt.Printf("Outbound applying > %s as “%s”\n", ApplyingToAddr, ApplyingMessage)
}

// Multi-unicast
func MultiUnicast(_message string) {
    for _, addr := range Addresses {

        go func(_addr AddressMaps) {
            outbound_to_addr, err := net.ResolveUDPAddr("udp", _addr.Address)
            multi_unicast, err := net.DialUDP("udp", nil, outbound_to_addr)
            Error(err)

            defer multi_unicast.Close()
            fmt.Printf("Connected multi-unicast > %s\n", outbound_to_addr)

            // Multi-unicast message
            multi_unicast.Write([]byte(_message))
            fmt.Printf("Multi-unicast > %v as “%s”\n", outbound_to_addr, _message)
        }(addr)
    }
}

// Check nearcast
func CheckNearcast() {
    for {
        time.Sleep(time.Duration(NearcastCheckInterval) * time.Second)

        if 0 == len(Addresses) {
            continue
        }

        for i, addr := range Addresses {
            go func(_i int, _addr AddressMaps) {
                addr_parts := strings.Split(_addr.Address, ":")
                stdout, err := exec.Command("traceroute", "-n", "-T", "-w", "1", "-q", "1", "-p", DefaultPort, addr_parts[0]).Output()
                Error(err)

                lines := strings.Split(string(stdout), "\n")
                for _, line := range lines {
                    line = strings.TrimSpace(line)
                    line = strings.Replace(line, "  ", " ", -1)
                    if line == "" {
                        continue
                    }
                    fmt.Printf("Check nearcast: %v\n", line)

                    columns := strings.Split(line, " ")
                    if columns[1] != addr_parts[0] {
                        continue
                    }

                    Addresses[_i].Neary, err = strconv.ParseFloat(columns[2], 64)
                    Error(err)
                    break
                }
            }(i, addr)
        }
    }
}

// Nearcast
func Nearcast(_message string) {
    // Reslice time for nearcast
    times := make([]float64, len(Addresses))
    for i, address := range Addresses {
        times[i] = address.Neary
    }

    // Sort time for nearcast
    sort.Slice(times, func(i, j int) bool {
        return times[i] < times[j]
    })

    // Extract fastest time for nearcast
    nearcast_addr := ""
ForNearcastBreak:
    for _, time := range times {
        for _, addr_map := range Addresses {
            if 1 < len(times) && addr_map.Address == DefaultIp {
                continue
            }

            if time == addr_map.Neary {
                nearcast_addr = addr_map.Address
                break ForNearcastBreak
            }
        }
    }

    if nearcast_addr != "" {
        // Connect nearcastic address
        outbound_to_addr, err := net.ResolveUDPAddr("udp", nearcast_addr)
        nearcast, err := net.DialUDP("udp", nil, outbound_to_addr)
        Error(err)

        defer nearcast.Close()
        fmt.Printf("Connected nearcast address > %v\n", outbound_to_addr)

        // Nearcast message
        nearcast.Write([]byte(_message))
        fmt.Printf("Nearcast > %v as “%s”\n", outbound_to_addr, _message)
    } else {
        // error
    }
}

// Groupcast
func Groupcast(_message string, _group int) {
    for _, addr_map := range Addresses {
        // Select groupcast
        if _group != addr_map.Group {
            continue
        }

        go func(_addr string) {
            // Connect groupcast address
            outbound_to_addr, err := net.ResolveUDPAddr("udp", _addr)
            groupcast, err := net.DialUDP("udp", nil, outbound_to_addr)
            Error(err)

            defer groupcast.Close()
            fmt.Printf("Connected groupcast > %s\n", outbound_to_addr)

            // Groupcast message
            groupcast.Write([]byte(_message))
            fmt.Printf("Groupcast > %v as “%s”\n", outbound_to_addr, _message)
        }(addr_map.Address)
    }
}

// Stdin
func Stdin() {
    for {
        //
        stdin := bufio.NewScanner(os.Stdin)
        for stdin.Scan() {
            // Extract command
            text := stdin.Text()
            text = strings.TrimSpace(text)
            text = strings.Replace(text, "  ", " ", -1)
            args := strings.Split(text, " ")

            // Check group command
            if args[0] == "group" {
                for i, addrmap := range Addresses {
                    addr_parts := strings.Split(addrmap.Address, ":")
                    ip_addr := addr_parts[0]

                    // Check address
                    if args[2] != ip_addr {
                        continue
                    }

                    // Update groupcast
                    if args[1] == "update" {
                        var err error
                        Addresses[i].Group, err = strconv.Atoi(args[3])
                        Error(err)
                        fmt.Printf("Update group %s to %s\n", ip_addr, args[3])
                        break
                        // Empty groupcast
                    } else if args[1] == "empty" {
                        Addresses[i].Group = 0
                        fmt.Printf("Update empty %s to 0\n", ip_addr)
                        break
                    }
                }
                // Multi-unicast
            } else if args[0] == "multi-unicast" {
                // Multi-unicast message
                MultiUnicast("multi-unicast:" + args[1])
                // Nearcast
            } else if args[0] == "nearcast" {
                // Nearcast message
                Nearcast("nearcast:" + args[1])
                // Groupcast
            } else if args[0] == "groupcast" {
                group, err := strconv.Atoi(args[2])
                Error(err)
                // Groupcast message
                Groupcast("groupcast:"+args[1], group)
            }
        }
    }
}
