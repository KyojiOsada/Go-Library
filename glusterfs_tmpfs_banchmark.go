package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"time"
)

func main() {
	// Configure
	file_paths := []string {"/cache_client/test.txt", "/file_client/test.txt"}
	systems := []string {"Cache System", "File System"}
	results := []float64 {0, 0}

	var content_string string
	for i := 0; i < 100000; i++ {
		content_string += "a"
	}
	content := []byte(content_string)

	for i := 0; i < 1; i++ {
		for j, _ := range file_paths {
			// Get processing start datetime
			start_datetime := time.Now()
			for k := 0; k < 1000; k++ {
				// Write file
				err := ioutil.WriteFile(file_paths[j], content, 0644)
				if err != nil {
					fmt.Println("File Writing Error: %s\n", err)
					os.Exit(1)
				}

				// Read file
				read, err := ioutil.ReadFile(file_paths[j])
				if err != nil {
					fmt.Println("File Reading Error: %s%s\n", err, read)
					os.Exit(1)
				}

				// Remove file
				os.Remove(file_paths[j])
				if err != nil {
					fmt.Println("File Removing Error: %s\n", err)
					fmt.Println(err)
					os.Exit(1)
				}
				//fmt.Printf("%v, ", k)
			}
			// Get processing end datetime
			end_datetime := time.Now()

			// Get processing total time
			total_time := end_datetime.Sub(start_datetime)
			results[j] += total_time.Seconds()
			fmt.Printf("[%v] %v: %v\n", i, systems[j], total_time)
		}
	}

	for i, v := range results {
		ave := v / 10
		fmt.Printf("%v Total: %vs\n", systems[i], v)
		fmt.Printf("%v Average: %vs\n", systems[i], ave)
	}

	os.Exit(0)
}
