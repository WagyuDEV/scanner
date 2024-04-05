package main

import (
	"fmt"
	"math/big"
	"net"
	"net/http"
	"os"
	"runtime"
	"strconv"
	"strings"
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

	resStr := url + ";" + strconv.Itoa(res.StatusCode) + ";" + res.Request.URL.String() + "\n"

	http.Post("https://discord.com/api/webhooks/1220197342888726609/WWaC8NndSGSe5rOTwmVPccrefg_rJIKirdcAVgwCM3aOtsUH0jqQeeQvikB_HkekRSqv", "application/json", strings.NewReader(`{"content" : "`+url+" | "+strconv.Itoa(res.StatusCode)+" | "+res.Request.URL.String()+`", "username" : "Internet Scanner"}`))

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
		// if big.NewInt(0).Mod(&b, big.NewInt(1000)).Cmp(big.NewInt(0)) == 0 {
		// 	go fmt.Println(barr, b.String())
		// }
	}
}

func main() {
	iter(net.IPv4(0, 0, 0, 0), net.IPv4(255, 255, 255, 255))
	wg.Wait()
}
