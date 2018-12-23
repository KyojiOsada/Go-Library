package main

import (
    "fmt"
    "io/ioutil"
    "os"
    "strconv"
    "time"
)

func main() {
    // Configure
    file_paths := []string{"/mnt/lizardfs/test", "/mnt/glusterfs/test"}
    systems := []string{"LizardFS", "GlusterFS"}
    results := []float64{0, 0}
    contents_size := 1000000
    benchmark_times := 10
    processing_times := 1000

    // Create contents
    var contents_string string
    for i := 0; i < contents_size; i++ {
        contents_string += "a"
    }
    contents_byte := []byte(contents_string)

    // Loop benchmark times
    for i := 0; i < benchmark_times; i++ {

        // Loop GlusterFS and LizardFS
        for j, _ := range file_paths {
            // Get processing start datetime
            start_datetime := time.Now()

            // Write files
            for k := 0; k < processing_times; k++ {
                err := ioutil.WriteFile(file_paths[j] + strconv.Itoa(k), contents_byte, 0644)
                if err != nil {
                    fmt.Println("File Writing Error: ", err)
                    os.Exit(1)
                }
            }

            // Read files
            for k := 0; k < processing_times; k++ {
                content_read, err := ioutil.ReadFile(file_paths[j] + strconv.Itoa(k))
                if err != nil {
                    fmt.Println("File Reading Error: ", err, content_read)
                    os.Exit(1)
                }
            }

            // Remove files
            for k := 0; k < processing_times; k++ {
                err := os.Remove(file_paths[j] + strconv.Itoa(k))
                if err != nil {
                    fmt.Println("File Removing Error: ", err)
                    os.Exit(1)
                }
            }

            // Get processing end datetime
            end_datetime := time.Now()

            // Get processing total time
            total_time := end_datetime.Sub(start_datetime)
            results[j] += total_time.Seconds()
            fmt.Printf("[%v] %v: %v\n", i, systems[j], total_time)
        }
    }

    // Get average processing time
    for i, v := range results {
        average := v / float64(benchmark_times)
        fmt.Printf("%v Average: %vs\n", systems[i], average)
    }

    os.Exit(0)
}
