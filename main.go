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

	var filepath, period, failureThreshold string
	var userTimeout time.Duration

	flag.StringVar(&filepath, "f", "", "file")
	flag.StringVar(&period, "p", "", "period")
	flag.DurationVar(&userTimeout, "t", 2, "timeout")
	flag.StringVar(&failureThreshold, "ft", "", "failure_threshold")
	flag.Parse()

	//httpClient := userTimeout * time.Second
	//fmt.Println(httpClient)

	file, err := os.Open(filepath)
	if err != nil{
		log.Fatal(err)
	}

	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.Split(scanner.Text(), " ")

		// To get the port number from the urlList, len - 1 is used as position.
		if len(line) > 1 {
			port := line[len(line)-1]
			getRequest(line[0], port, userTimeout)
			//fmt.Println(port)
		} else {
			getRequest(line[0], "", userTimeout)
		}
	}
}

func getRequest(url, port string, userTimeout time.Duration) {
	timeNow := time.Now().UTC()
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
		// This was a timeout
		//fmt.Printf("Website Down! [%v] TIMEOUT %v:%v \n", timeNow, url, port)
		timeOut(timeNow, url, port)
		return
	}

	if err != nil {
		// This was all errors but timeout
		//fmt.Printf("Website Down! [%v] 500 %v:%v \n", timeNow, url, port)
		down(timeNow, url, port)
		return
	}

	defer resp.Body.Close()
	fmt.Printf("Website Up! [%v] %v %v:%v \n", timeNow, resp.StatusCode, url, port)
}


func timeOut(timeNow time.Time, url, port string) {
	fmt.Printf("Website Down! [%v] TIMEOUT %v:%v \n", timeNow, url, port)
}

func down(timeNow time.Time, url, port string) {
	fmt.Printf("Website Down! [%v] 500 %v:%v \n", timeNow, url, port)
}
