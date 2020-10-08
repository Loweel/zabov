package main

import (
	"fmt"
	"strings"
)

var zabovKbucket = []byte("killfile")

type killfileItem struct {
	Kdomain string
	Ksource string
}

var bChannel chan killfileItem

func init() {

	bChannel = make(chan killfileItem, 1024)
	fmt.Println("Initializing kill channel engine.")

	go bWriteThread()

}

func bWriteThread() {

	for item := range bChannel {

		writeInKillfile(item.Kdomain, item.Ksource)
		incrementStats("BL domains from "+item.Ksource, 1)
		incrementStats("TOTAL", 1)

	}

}

//DomainKill stores a domain name inside the killfile
func DomainKill(s, durl string) {

	if len(s) > 2 {

		s = strings.ToLower(s)

		var k killfileItem

		k.Kdomain = s
		k.Ksource = durl

		bChannel <- k

	}

}

func writeInKillfile(key, value string) {

	stK := []byte(key)
	stV := []byte(value)

	err := MyZabovKDB.Put(stK, stV, nil)
	if err != nil {
		fmt.Println("Cannot write to Killfile DB: ", err.Error())
	}

}

func domainInKillfile(domain string) bool {

	s := strings.ToLower(domain)

	has, err := MyZabovKDB.Has([]byte(s), nil)
	if err != nil {
		fmt.Println("Cannot read from Killfile DB: ", err.Error())
	}

	return has

}
