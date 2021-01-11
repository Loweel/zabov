package main

import (
	"fmt"
	"strings"
)

type killfileItem struct {
	Kdomain  string
	Ksource  string
	Kconfigs stringarray
}

var bChannel chan killfileItem

func init() {

	bChannel = make(chan killfileItem, 1024)
	fmt.Println("Initializing kill channel engine.")

	go bWriteThread()

}

func bWriteThread() {

	for item := range bChannel {

		for _, config := range item.Kconfigs {
			writeInKillfile(item.Kdomain, item.Ksource, config)
		}
		incrementStats("BL domains from "+item.Ksource, 1)
		incrementStats("TOTAL", 1)

	}

}

//DomainKill stores a domain name inside the killfile
func DomainKill(s, durl string, configs stringarray) {

	if len(s) > 2 {

		s = strings.ToLower(s)

		var k killfileItem

		k.Kdomain = s
		k.Ksource = durl
		k.Kconfigs = configs

		bChannel <- k

	}

}

func writeInKillfile(key, value string, config string) {

	stK := []byte(key)
	stV := []byte(value)

	MyZabovKDB := MyZabovKDBs[config]
	err := MyZabovKDB.Put(stK, stV, nil)
	if err != nil {
		fmt.Println("Cannot write to Killfile DB: ", err.Error())
	}

}

func domainInKillfile(domain string, config string) bool {

	s := strings.ToLower(domain)

	MyZabovKDB := MyZabovKDBs[config]
	has, err := MyZabovKDB.Has([]byte(s), nil)
	if err != nil {
		fmt.Println("Cannot read from Killfile DB: ", err.Error())
	}

	return has

}
