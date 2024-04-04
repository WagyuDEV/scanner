package main

import (
	"context"
	"fmt"
	"math/big"
	"net"
	"net/http"
	"os"
	"strconv"
	"sync"
	"time"

	"golang.org/x/time/rate"
)

// ChatGPT helped with concurrent file writing and rate limiting

var mutex sync.Mutex
var wg sync.WaitGroup
var limiter = rate.NewLimiter(rate.Every(time.Second), 100) // Limit to 100 requests per second

func get(url string, file string) {
	defer wg.Done()
	if err := limiter.Wait(context.Background()); err != nil {
		fmt.Println("Rate limit exceeded. Waiting...")
		time.Sleep(limiter.Reserve().Delay())
	}
	res, err := http.Get(url)
	if err != nil {
		return
	}
	resStr := url + ";" + strconv.Itoa(res.StatusCode) + "\n"
	mutex.Lock()
	defer mutex.Unlock()
	f, err := os.OpenFile(file, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		fmt.Println("Error opening file:", err)
		return
	}
	defer f.Close()
	_, err = f.WriteString(resStr)
	if err != nil {
		fmt.Println("Error writing to file:", err)
	}
	//fmt.Println(url, res.StatusCode)
}

func iter(start net.IP, stop net.IP) {
	b, e := big.Int{}, big.Int{}
	b = *b.SetBytes(start.To4())
	e = *e.SetBytes(stop.To4())
	for b.Cmp(&e) <= 0 {
		wg.Add(1)
		go get("http://"+net.IP(b.Bytes()).String(), "scan.csv")
		b.Add(&b, big.NewInt(1))
	}
	wg.Wait()
}

func main() {
	// TODO: make this not crash when an ip less than 1.0.0.0 is presented
	iter(net.IPv4(1, 0, 0, 0), net.IPv4(255, 255, 255, 255))
}
