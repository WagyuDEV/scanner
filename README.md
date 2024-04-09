# :construction:THIS PROJECT IS NOT DONE:construction:

# Scanner
 Scan the Internet

# Roadmap
- [ ] [Rewrite](#Rewrite)
- [ ] Turn into a CLI tool
- [ ] Add optional arguments for what data is collected

# Rewrite
The program in its current form suffers from 
memory leaks that build up after a few hours to either lock up your system or deadlock the program.
This is largely cause by a combination of the large number of hanging request and the spawning of a goroutine per request.

### Proposal
Instead of having a goroutine spawned for each request, have the user decide how many goroutines they want to spawn and then
split the request evenly between these routines. The manner in which actual requests are handled needs to be workshopped a bit.

#### Ideas for Handling Requests Inside Goroutines
- Use a threadpool to handle a set number of goroutines, each goroutine representing one request
- Use channels to make quasi-async/await
- Idea #1 with the addition of persistent routines taking in new data through channels
- [ChatGPT had an idea](#chatgpts-solution)

# Why
There are two reasons for this project to exist. The first reason is my curiosity of how the internet is laid out and works; I want go get a large dataset of functional IP addresses to do more research on.
The second reason is to practice Go, as this is my first Go project

# ChatGPT's Solution
>ChatGPT: To handle a large range of IP addresses efficiently in Go, while ensuring that each goroutine does a similar amount of work, requests are made concurrently with a specified timeout, and no request is made twice, you can use a combination of channels, goroutines, and a sync mechanism. Here's a general outline of how you could structure your program:
```go
package main

import (
    "fmt"
    "net/http"
    "sync"
    "time"
)

const (
    numWorkers = 10 // Number of goroutines to spawn
    timeout    = 5 * time.Second // Timeout for each HTTP request
)

func worker(id int, jobs <-chan string, results chan<- string, wg *sync.WaitGroup) {
    defer wg.Done()
    client := http.Client{
        Timeout: timeout,
    }
    for ip := range jobs {
        // Perform HTTP request
        resp, err := client.Get(fmt.Sprintf("http://%s", ip))
        if err != nil {
            results <- fmt.Sprintf("Worker %d: Error for %s: %v", id, ip, err)
            continue
        }
        defer resp.Body.Close()

        // Process response
        // For example, you can read the body or check the status code
        // For simplicity, we're just printing the response status code
        results <- fmt.Sprintf("Worker %d: %s - Status Code: %d", id, ip, resp.StatusCode)
    }
}

func main() {
    // Define a range of IP addresses
    ipRange := []string{
        "192.168.1.1",
        "192.168.1.2",
        // Add more IP addresses as needed...
    }

    // Create channels for sending jobs and receiving results
    jobs := make(chan string, len(ipRange))
    results := make(chan string, len(ipRange))

    // Populate jobs channel with IP addresses
    for _, ip := range ipRange {
        jobs <- ip
    }
    close(jobs)

    // Use a WaitGroup to wait for all goroutines to finish
    var wg sync.WaitGroup
    wg.Add(numWorkers)

    // Spawn worker goroutines
    for i := 1; i <= numWorkers; i++ {
        go worker(i, jobs, results, &wg)
    }

    // Close results channel when all workers are done
    go func() {
        wg.Wait()
        close(results)
    }()

    // Collect results from workers
    for result := range results {
        fmt.Println(result)
    }
}
```
