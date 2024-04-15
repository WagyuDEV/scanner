package main

import (
	"context"
	"fmt"
	"math/big"
	"net"
	"net/http"
	"os"
	"runtime"
	"strconv"
	"sync"
	"time"

	"github.com/pbnjay/memory"
	"golang.org/x/time/rate"
)

var (
	// 1 Goroutine costs 2 kB
	numWorkers = memory.TotalMemory() / (8589934 / 2)
	timeout    = time.Second * 2 // request timeout
	wg         sync.WaitGroup
	bufferSize = numWorkers + 100
	//client     *http.Client
	//mutex sync.Mutex
)

func get(id int, jobs <-chan string, data chan string, clients []*http.Client, rl *rate.Limiter) {
	defer wg.Done()
	for job := range jobs {
		rl.Wait(context.Background())

		client := clients[id%len(clients)] // Assign each worker to a client

		res, err := client.Get("https://" + job)

		if err != nil {
			fmt.Printf("Error on job %d: %s\n", id, err)
		} else {
			//fmt.Println(job, res.StatusCode)
			resStr := job + ";" + strconv.Itoa(res.StatusCode) + ";" + res.Request.URL.String() + "\n"
			data <- resStr
		}
		//fmt.Println(id, runtime.NumGoroutine(), job)
	}
}

func writeToFile(file string, data chan string) {
	defer close(data)
	f, err := os.OpenFile(file, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		fmt.Println("Error opening file", err)
		os.Exit(1) // exit program if unable to open file
	}
	defer f.Close()
	for item := range data {
		_, err = f.WriteString(item)
		if err != nil {
			fmt.Println("Error writing to file:", err)
		}
	}
}

func iter(start net.IP, stop net.IP) {
	tr := &http.Transport{
		MaxIdleConns:        100,                  // Maximum idle connections to keep open
		MaxIdleConnsPerHost: runtime.NumCPU() * 2, // Maximum idle connections per host
		IdleConnTimeout:     5 * time.Second,      // Timeout for idle connections
		// TLSHandshakeTimeout:   1 * time.Second,      // Timeout for TLS handshake
		// ExpectContinueTimeout: 1 * time.Second,      // Timeout for waiting for a '100 Continue' response
	}
	clients := make([]*http.Client, runtime.NumCPU())
	for i := range clients {
		clients[i] = &http.Client{Timeout: timeout, Transport: tr}
	}

	jobs := make(chan string, bufferSize)
	data := make(chan string, bufferSize)

	r := rate.NewLimiter(rate.Limit(800), 2000)

	b, e := big.Int{}, big.Int{}
	b = *b.SetBytes(start.To4())
	e = *e.SetBytes(stop.To4())

	barr := make([]byte, 4)
	b.FillBytes(barr)

	go func() {
		for b.Cmp(&e) <= 0 { // x < y -> -1 | x == y -> 0 | x > y -> 1
			// select {
			// case jobs <- net.IP(barr).String():
			// default:
			// 	time.Sleep(time.Millisecond)
			// 	//fmt.Println("FULLLLLLLLLL")
			// }
			jobs <- net.IP(barr).String()
			b.Add(&b, big.NewInt(1))
			b.FillBytes(barr)
		}
		close(jobs)
	}()

	go writeToFile("data.csv", data)

	for id := 1; id <= int(numWorkers); id++ {
		go get(id, jobs, data, clients, r)
		wg.Add(1)
	}
}

func main() {
	fmt.Println(numWorkers)
	iter(net.IPv4(1, 1, 0, 0), net.IPv4(255, 255, 255, 255))
	wg.Wait()
}
