package main

import (
	"fmt"
	"math/big"
	"net"
	"net/http"
	"runtime"
	"sync"
	"time"

	"github.com/pbnjay/memory"
)

var (
	// 1 Goroutine costs 2 kB
	numWorkers = memory.TotalMemory() / (8589934 / 2)
	timeout    = time.Second * 2 // request timeout
	wg         sync.WaitGroup
	bufferSize = numWorkers + 100
	//client     *http.Client
)

func get(id int, jobs <-chan string, clients []*http.Client) {
	defer wg.Done()
	for job := range jobs {
		client := clients[id%len(clients)] // Assign each worker to a client
		res, err := client.Get("https://" + job)
		if err != nil {
			//fmt.Printf("Error on job %d: %s\n", id, err)
		} else {
			fmt.Println(job, res.StatusCode)
		}
		//fmt.Println(id, runtime.NumGoroutine(), job)
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

	b, e := big.Int{}, big.Int{}
	b = *b.SetBytes(start.To4())
	e = *e.SetBytes(stop.To4())

	barr := make([]byte, 4)
	b.FillBytes(barr)

	go func() {
		for b.Cmp(&e) <= 0 { // x < y -> -1 | x == y -> 0 | x > y -> 1
			select {
			case jobs <- net.IP(barr).String():
			default:
				fmt.Println("FULLLLLLLLLL")
			}

			b.Add(&b, big.NewInt(1))
			b.FillBytes(barr)
		}
		close(jobs)
	}()
	for id := 1; id <= int(numWorkers); id++ {
		go get(id, jobs, clients)
		wg.Add(1)
	}
}

func main() {
	iter(net.IPv4(1, 0, 0, 0), net.IPv4(255, 255, 255, 255))
	wg.Wait()
}
