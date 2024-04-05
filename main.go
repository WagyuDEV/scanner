package main

import (
	"fmt"
	"math/big"
	"net"
	"net/http"
	"os"
	"runtime"
	"strconv"
	"sync"
	"time"
)

var (
	maxConcurrentRequest = 500
	sem                  = make(chan int, maxConcurrentRequest)
	wg                   sync.WaitGroup
	mutex                sync.Mutex
)

func get(url string, file string, client *http.Client) {
	sem <- runtime.NumGoroutine()
	defer func() {
		<-sem
		wg.Done()
	}()

	res, err := client.Get(url)

	if err != nil {
		fmt.Println(err)
		return
	}
	defer res.Body.Close()

	resStr := url + ";" + strconv.Itoa(res.StatusCode) + "\n"

	mutex.Lock()
	defer mutex.Unlock()

	f, err := os.OpenFile(file, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		fmt.Println("Error opening file", err)
		return
	}
	defer f.Close()

	_, err = f.WriteString(resStr)
	if err != nil {
		fmt.Println("Error writing to file:", err)
	}
}

func iter(start net.IP, stop net.IP) {
	client := &http.Client{
		Timeout: time.Second * 2,
	}

	b, e := big.Int{}, big.Int{}
	b = *b.SetBytes(start.To4())
	e = *e.SetBytes(stop.To4())

	barr := make([]byte, 4)
	b.FillBytes(barr)

	for b.Cmp(&e) <= 0 { // x < y -> -1 | x == y -> 0 | x > y -> 1
		for runtime.NumGoroutine() > 50000 {
			time.Sleep(time.Duration(time.Second))
		}
		wg.Add(1)
		go get("http://"+net.IP(barr).String(), "scan.csv", client)
		b.Add(&b, big.NewInt(1))
		b.FillBytes(barr)
	}
}

func main() {
	// TODO make it not error when an octet lower than 1 0 0 0 is presented
	iter(net.IPv4(0, 0, 0, 0), net.IPv4(255, 255, 255, 255))
	wg.Wait()
	// b := big.NewInt(1)
	// fmt.Println(b.Bytes())
	// barr := make([]byte, 4)
	// b.FillBytes(barr)
	// fmt.Println(barr)
}
