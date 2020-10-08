package main

import (
	"fmt"
	"net/http"
	"time"
)

//NetworkUp tells the system if the network is up or not
var NetworkUp bool

func checkNetworkUp() bool {
	// RFC2606 test domain, should always work, unless internet is down.
	_, err := http.Get("http://example.com")
	if err != nil {
		return false
	}
	return true
}

func checkNetworkUpThread() {

	ticker := time.NewTicker(2 * time.Minute)

	for range ticker.C {
		NetworkUp = checkNetworkUp()
	}

}

func init() {

	fmt.Println("Network Checker starting....")

	go checkNetworkUpThread()

}
