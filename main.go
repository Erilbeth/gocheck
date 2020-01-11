package main

import (
	"bufio"
	"crypto/tls"
	"flag"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"strings"
	"time"
)

func main() {

	var filepath  string
	var userTimeout time.Duration
	var failureThreshold, period int

	flag.StringVar(&filepath, "f", "", "file")
	flag.IntVar(&period, "p", 1, "period")
	flag.DurationVar(&userTimeout, "t", 2, "timeout")
	flag.IntVar(&failureThreshold, "ft", 10, "failure_threshold")
	flag.Parse()

	file, err := os.Open(filepath)
	if err != nil{
		log.Fatal(err)
	}

	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.Split(scanner.Text(), " ")

		if len(line) > 1 {

			port := line[len(line)-1]
			logic(failureThreshold, line[0], port, userTimeout)

		} else {

			logic(failureThreshold, line[0], "", userTimeout)
		}
	}
}

func logic(failureThreshold int, url, port  string, userTimeout time.Duration) {

	timeNow := time.Now().UTC()
	count := 0
	for i:=1; i <= failureThreshold; i++ {
		response := getRequest(url, port, userTimeout)
		if response == "error" || response == "timeout" {
			count++
		}
		fmt.Println(count)

		if count == failureThreshold && response == "timeout" {
			timeOut(timeNow, url, port)
		}

		if count == failureThreshold && response == "error" {
			down(timeNow, url, port)
		}

		time.Sleep(1 * time.Second)
	}
}

func getRequest(url, port string, userTimeout time.Duration) string{
	var (
		httpClientTimeout = userTimeout * time.Second
		httpClient        = &http.Client{
			Transport: &http.Transport{
				DialContext: (&net.Dialer{
					Timeout:   httpClientTimeout,
					KeepAlive: httpClientTimeout / 2,
				}).DialContext,
				DisableKeepAlives:     true,
				TLSClientConfig:       &tls.Config{InsecureSkipVerify: true},
				TLSHandshakeTimeout:   2 * time.Second,
				ExpectContinueTimeout: 1 * time.Second,
			},
		}
	)

	resp, err := httpClient.Get(url+":"+port)

	if e, ok := err.(net.Error); ok && e.Timeout() {
		return "timeout"
	}

	if err != nil {
		return "error"
	}

	defer resp.Body.Close()
	return "200"
}


func timeOut(timeNow time.Time, url, port string) {
	fmt.Printf("Website Down! [%v] TIMEOUT %v:%v \n", timeNow, url, port)
}

func down(timeNow time.Time, url, port string) {
	fmt.Printf("Website Down! [%v] 500 %v:%v \n", timeNow, url, port)
}