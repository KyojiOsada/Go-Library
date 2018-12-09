package main

import (
	"fmt"
	"net"
	"os"
	"os/exec"
	"strings"
	"time"
	"unsafe"
)

func main() {

	cluster_vip := os.Args[1]
	cluster_port := os.Args[2]
	cluster_address := cluster_vip + ":" + cluster_port
	self_ip := os.Args[3]
	self_port := os.Args[4]
	greeting_interval := 1
	greeting_timeout := 5
	retry_interval := 3
	retry_count := 3
	retry_timeout := 5

	for {
		// Sleep healthcheck interval seconds
		time.Sleep(time.Duration(greeting_interval) * time.Second)

		// Greet to cluster
		conn, err := net.DialTimeout("tcp", cluster_address, time.Duration(greeting_timeout)*time.Second)

		// Succeeded greeting
		if err == nil {
			conn.Close()
			fmt.Println("Could greet to cluster.")
			continue
		}

		// Failed greeting
		// Log
		//...
		fmt.Println("Could not greet to cluster.")

		// Loop re-connection to cluster
		health_flag := false
		for j := 0; j < retry_count; j++ {

			// Sleep re-connection interval seconds
			time.Sleep(time.Duration(retry_interval) * time.Second)

			// Re-connect
			conn, err = net.DialTimeout("tcp", cluster_address, time.Duration(retry_timeout)*time.Second)

			// Failed re-connection
			if err != nil {
				// Log
				//...
				fmt.Println("Could not re-connect to cluster.")
				continue
			}

			// Succeeded re-connection
			fmt.Println("Succeeded re-connection to cluster.")
			health_flag = true
			break
		}

		// Alive
		if health_flag {
			continue
		}

		// Dead
		personality_byte, err := exec.Command("lizardfs-admin", "metadataserver-status", self_ip, self_port).Output()
		personality := *(*string)(unsafe.Pointer(&personality_byte))

		// Failed get personality
		if err != nil {
			// Log
			//...
			fmt.Println("Could not get personality.")
			os.Exit(1)
		}

		// Demote to master
		if -1 != strings.Index(personality, "personality: master") {
			err := exec.Command("lizardfs-uraft-helper", "demote").Run()

			// Failed master demotion
			if err != nil {
				// Log
				//...
				fmt.Println("Could not demote to master.")
				os.Exit(1)
			}
			fmt.Println("Succeeded master demotion.")
			// Promote to master
		} else if -1 != strings.Index(personality, "personality: shadow") {
			err := exec.Command("lizardfs-uraft-helper", "remote").Run()

			// Failed master promotion
			if err != nil {
				// Log
				//...
				fmt.Println("Could not promote to master.")
				os.Exit(1)
			}
			fmt.Println("Succeeded master promotion.")
		}

		// Restart LizardFS
		err = exec.Command("systemctl", "restart", "lizardfs-uraft").Run()

		// Failed restart
		if err != nil {
			// Log
			//...
			fmt.Println("Could not restart LizardFS.")
			os.Exit(1)
		}
		fmt.Println("Succeeded restart LizardFs.")
	} // End Loop
} // End main
