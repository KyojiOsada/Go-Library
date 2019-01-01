package main

import (
    "fmt"
    "net"
    "os/exec"
    "strings"
    "sync"
    "time"
)

func main() {
    mcast_ip4 := "224.0.0.1"
    mcast_port := ":9999"
    mcast_addr := mcast_ip4 + mcast_port
    pulse_interval := 1
    pulse_retry := 5
    mip := ""
    mip_cidr := ""
    mip_mask := ""
    mip_port := ""
    mip_addr := ""
    vip4 := "192.0.2.9"
    vip4_cidr := "/24"
    vip4_mask := vip4 + vip4_cidr
    vip4_if := "eth0"
    master_ip := ""
    master_port := ""
    master_addr := ""
    state := ""
    nodes := make(map[string]map[string]string)
    pulses := make(map[string]map[string]int64)
    standings := make(map[string]bool)
    votings := make(map[string]int)
    sender_len := 0
    standing_message := ""
    voting_message := ""
    nominated_message := ""

    // Get node state
    own_ip_masks, err := net.InterfaceAddrs()
    _Error(err)

    state = "backup"
    for _, own_ip_mask := range own_ip_masks {
        if own_ip_mask.String() != vip4_mask {
            continue
        }
        state = "master"
        break
    }
    fmt.Printf("State: %s\n", state)

    // Start pulse sender
    fmt.Printf("Start pulse sender to: %s\n", mcast_addr)
    connector, err := net.Dial("udp", mcast_addr)
    _Error(err)
    defer connector.Close()

    go func() {
        for {
            time.Sleep(time.Duration(pulse_interval) * time.Second)

            // Send pulse
            connector.Write([]byte(state))
            fmt.Printf("Sent pulse: %s\n", state)
        }
    }()

    // Start pulse receiver
    fmt.Printf("Start pulse receiver from: %s\n", mcast_addr)
    mcast_byte, err := net.ResolveUDPAddr("udp", mcast_addr)
    _Error(err)

    listener, err := net.ListenMulticastUDP("udp", nil, mcast_byte)
    _Error(err)
    defer listener.Close()

    buffer := make([]byte, 64)
    go func() {
        for {
            // Receive pulse
            length, sender_addr_byte, err := listener.ReadFrom(buffer)
            _Error(err)

            receive_message := string(buffer[:length])

            // Split sender address to ip and port
            sender_addr := sender_addr_byte.(*net.UDPAddr).String()
            sender_addr_parts := strings.Split(sender_addr, ":")
            sender_ip := sender_addr_parts[0]
            sender_port := sender_addr_parts[1]

            // Receive pulse message
            if receive_message == "master" || receive_message == "backup" {
                // Build pulse information
                now := time.Now()
                pulses[sender_ip] = make(map[string]int64)
                pulses[sender_ip]["current"] = now.UnixNano()

                // Build nodes information
                nodes[sender_ip] = make(map[string]string)
                nodes[sender_ip]["state"] = receive_message

                fmt.Printf("Received pulse: %s from: %s\n", receive_message, sender_ip)
            }

            // Build master information
            if master_ip == "" {
                if receive_message == "master" {
                    master_ip = sender_ip
                    master_port = sender_port
                    master_addr = master_ip + ":" + master_port
                }
            }

            // Build own MIP information
            if mip == "" {
                for _, own_ip_mask := range own_ip_masks {
                    own_addr_parts := strings.Split(own_ip_mask.String(), "/")
                    if own_addr_parts[0] != sender_ip {
                        continue
                    }
                    mip = own_addr_parts[0]
                    mip_cidr = "/" + own_addr_parts[1]
                    mip_mask = mip + mip_cidr
                    mip_port = ":" + sender_port
                    mip_addr = mip + mip_port
                    break
                }
            }

            // Receive master standing message
            _, flag := standings[receive_message]
            if 0 == strings.Index(receive_message, "standing,") && !flag {
                standings[receive_message] = true
                fmt.Printf("Reveived standing message: %s\n", receive_message)

                // Send master voting message
                voting_message = "voting," + sender_ip + "," + receive_message
                connector.Write([]byte(voting_message))
                fmt.Printf("Sent voting message: %s\n", voting_message)
            }

            // Receive voting message
            if 0 == strings.Index(receive_message, "voting,"+mip+","+standing_message) {
                if _, flag := votings[receive_message]; !flag {
                    votings[receive_message] = 0
                }
                votings[receive_message]++
                nominated_message = receive_message
                fmt.Printf("Receive voting message: %s\n", receive_message)
            }

        }
    }()

    // Start pulse checker
    for {
        // Init waiting group
        wg := new(sync.WaitGroup)

        // Check nodes count
        sender_len = len(pulses)

        // To serialize processing
        for sender, _ := range pulses {
            // For parallel processing
            _sender := sender

            // Add waiting group
            wg.Add(1)

            // To parallel processing
            go func() {
                // Defer waiting group
                defer wg.Done()

                for {
                    // Check nodes count changing
                    if sender_len != len(pulses) {
                        return
                    }

                    // Sleep interval seconds
                    time.Sleep(time.Duration(pulse_interval) * time.Second)

                    // Detected normal pulse
                    if pulses[_sender]["prev"] != pulses[_sender]["current"] {
                        pulses[_sender]["prev"] = pulses[_sender]["current"]
                        pulses[_sender]["check"] = 0
                        continue
                    }

                    // Detected abnormal pulse
                    pulses[_sender]["check"]++

                    // Send master standing message
                    if pulses[_sender]["check"] == 1 &&
                        nodes[_sender]["state"] == "master" &&
                        nodes[mip]["state"] == "backup" {
                        standing_message = "standing," + _sender
                        connector.Write([]byte(standing_message))
                        fmt.Printf("Sent standing message: %v\n", standing_message)
                    }

                    // Retry pulse check
                    if pulses[_sender]["check"] <= int64(pulse_retry) {
                        fmt.Printf("Retry pulse check: %v\n", pulses[_sender]["check"])
                        continue
                    }

                    // Decided master node down
                    if nodes[_sender]["state"] == "master" {

                        nodes[_sender]["state"] = "down"
                        pulses[_sender]["current"] = 0
                        fmt.Println("Detected master node down.")

                        // Case self is backup
                        if nodes[mip]["state"] == "backup" {

                            // Promote to master
                            if votings[nominated_message] >= (len(pulses) / 2) {

                                // Reasign VIP
                                fmt.Printf("Reasign Unicast VIP: %s\n", vip4_mask)
                                err = exec.Command("ip", "-f", "inet", "addr", "add", vip4_mask, "dev", vip4_if).Run()
                                _Error(err)

                                // Execute arping
                                fmt.Println("Replace ARP tables.")
                                err = exec.Command("arping", "-q", "-U", "-c5", "-w1", vip4, "-I", vip4_if).Run()
                                _Error(err)

                                state = "master"
                                master_ip = ""

                                fmt.Println("Suceeded failover.")
                            }

                            // Case self is master
                        } else if nodes[mip]["state"] == "master" {

                            // Unasign VIP
                            fmt.Printf("unasign Unicast VIP: %s\n", vip4_mask)
                            err = exec.Command("ip", "-f", "inet", "addr", "delete", vip4_mask, "dev", vip4_if).Run()
                            _Error(err)

                            // Execute arping
                            fmt.Println("Replace ARP tables.")
                            err = exec.Command("arping", "-q", "-U", "-c5", "-w1", vip4, "-I", vip4_if).Run()
                            _Error(err)

                            state = "backup"
                            master_ip = ""

                            fmt.Println("Suceeded failover.")
                        }
                        // Decided backup node down
                    } else {
                        nodes[_sender]["state"] = "down"
                        pulses[_sender]["current"] = 0

                        fmt.Println("Detected backup node down.")
                    }

                    delete(pulses, _sender)
                    return
                } // End for child process loop
            }() // End child processes

        } // End for range pulses

        // Wait waiting group
        wg.Wait()

    } // End for pulse checker
}

func _Error(_err error) {
    if _err != nil {
        panic(_err)
    }
}
