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

//var FilePath, Period, Failure_threshold string
//var userTimeout time.Duration
func main() {

	timeNow := time.Now().UTC()
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

		// To get the port number from the urlList, len - 1 is used as position.
		count := 0
		if len(line) > 1 {

			port := line[len(line)-1]

			for i:=1; i <= failureThreshold; i++ {
				response := getRequest(line[0], port, userTimeout)
				if response == "error" || response == "timeout" {
					count++
				}
				fmt.Println(count)

				if count == failureThreshold && response == "timeout" {
					timeOut(timeNow, line[0], port)
				}

				if count == failureThreshold && response == "error" {
					down(timeNow, line[0], port)
				}

				time.Sleep(1 * time.Second)
			}

		} else {

			for i := 1; i <= failureThreshold; i++ {
				response := getRequest(line[0], "", userTimeout)
				if response == "error" || response == "timeout" {
					count++
				}
				fmt.Println(count)

				if count == failureThreshold && response == "timeout" {
					timeOut(timeNow, line[0], "")
				}

				if count == failureThreshold && response == "error" {
					down(timeNow, line[0], "")
				}

				time.Sleep(1 * time.Second)

			}
		}
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