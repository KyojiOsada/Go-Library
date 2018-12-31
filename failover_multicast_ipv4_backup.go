package main

import (
    "fmt"
    "net"
    "os/exec"
    "time"
)

func main() {
    mcast_ip4 := "224.0.0.1"
    mcast_port := ":9999"
    mcast_addr := mcast_ip4 + mcast_port
    pulse_interval := 1
    pulses := []int64{1, 1}
    pulse_check := 0
    pulse_retry := 3
    ucast_vip4 := "192.0.2.3"
    ucast_cidr := "/24"
    ucast_mask := ucast_vip4 + ucast_cidr
    ucast_if := "eth0"

    fmt.Printf("Listen multicast address: %s\n", mcast_addr)
    mcast_byte, err := net.ResolveUDPAddr("udp", mcast_addr)
    _Error(err)

    listener, err := net.ListenMulticastUDP("udp", nil, mcast_byte)
    _Error(err)
    defer listener.Close()

    // For master pulse receiver
    buffer := make([]byte, 8)
    go func() {
        for {
            length, remote_mip4, err := listener.ReadFrom(buffer)
            _Error(err)

            if string(buffer[:length]) == "master" {
                now := time.Now()
                //now.UnixNano()
                pulses[1] = now.Unix()

                fmt.Printf("Reveived multicast from Master MIP: %v\n", remote_mip4)
                //fmt.Printf("Unix Timestamp: %v\n", uts)
            }
        }
    }()

    // For master pulse checker
    for {
        // Sleep pulse check interval seconds
        time.Sleep(time.Duration(pulse_interval) * time.Second)

        // Wait master boot up
        if pulses[1] == 1 {
            continue
        }

        // Detected normal pulse
        if pulses[0] != pulses[1] {
            pulses[0] = pulses[1]
            pulse_check = 0
            continue
        }

        // Detected abnormal pulse
        pulse_check++

        // Retry pulse check
        if pulse_check < pulse_retry {
            fmt.Printf("Retry pulse check: %v\n", pulse_check)
            continue
        }

        // Detected master cardiac arrest
        fmt.Println("Detected master cardiac arrest.")
        pulses[1] = 0

        // Asign VIP
        fmt.Printf("Reasign Unicast VIP: %s\n", ucast_mask)
        //os.Exit(0)
        err = exec.Command("ip", "-f", "inet", "addr", "add", ucast_mask, "dev", ucast_if).Run()
        _Error(err)

        // Execute arping
        fmt.Println("Replace ARP tables.")
        err = exec.Command("arping", "-q", "-U", "-c5", "-w1", ucast_vip4, "-I", ucast_if).Run()
        _Error(err)

        fmt.Println("Suceeded failover.")
        break
    }

}

func _Error(_err error) {
    if _err != nil {
        panic(_err)
    }
}
