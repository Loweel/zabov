package main

import (
	"bufio"
	"errors"
	"fmt"
	"net"
	"net/http"

	"strings"
	"time"
)

func init() {
	go downloadDoubleThread()
}

//DoubleIndexFilter puts the domains inside file
func DoubleIndexFilter(durl string, configs stringarray) error {

	fmt.Println("DoubleIndexFilter: Retrieving HostFile from: ", durl)

	// resets malformed HostLines for url
	setstatsvalue("Malformed HostLines "+durl, 0)

	var err error

	// Get the data
	resp, err := http.Get(durl)
	if err != nil {
		fmt.Println("HTTP problem: ", err)
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode == 200 { // OK
		fmt.Println(durl + " Response: OK")
	} else {
		fmt.Println("Server <"+durl+"> returned status code: ", resp.StatusCode)
		return errors.New("Server <" + durl + "> returned status code: " + resp.Status)
	}

	scanner := bufio.NewScanner(resp.Body)
	splitter := func(c rune) bool {
		return c == ' ' || c == '\t'
	}

	var numLines int64

	for scanner.Scan() {

		line := scanner.Text()

		if len(line) == 0 || strings.TrimSpace(line)[0] == '#' {
			continue
		}
		h := strings.FieldsFunc(line, splitter)

		if h == nil {
			continue
		}

		if len(h) < 2 {
			continue
		}

		if net.ParseIP(h[0]) != nil {

			DomainKill(h[1], durl, configs)

			// fmt.Println("MATCH: ", h[1])
			numLines++
		} else {
			incrementStats("Malformed HostLines "+durl, 1)
			// fmt.Println("Malformed line: <" + line + ">")
		}

	}

	fmt.Println("Finished to parse: "+durl+" ,number of lines", numLines)

	return err

}

func getDoubleFilters(urls urlsMap) {

	fmt.Println("getDoubleFilters: downloading all urls:", len(urls))
	for url, configs := range urls {
		DoubleIndexFilter(url, configs)
	}
	fmt.Println("getDoubleFilters: DONE!")

}

func downloadDoubleThread() {
	fmt.Println("Starting updater of DOUBLE lists, each (hours):", ZabovKillTTL)

	_urls := urlsMap{}

	for {
		fmt.Println("downloadDoubleThread: collecting urls from all configs...")
		for config := range ZabovConfigs {
			ZabovDoubleBL := ZabovConfigs[config].ZabovDoubleBL
			if len(ZabovDoubleBL) == 0 {
				continue
			}
			s := fileByLines(ZabovDoubleBL)
			for _, v := range s {
				configs := _urls[v]
				if configs == nil {
					configs = stringarray{}
					_urls[v] = configs
				}
				configs = append(configs, config)
				_urls[v] = configs
			}
		}

		getDoubleFilters(_urls)
		time.Sleep(time.Duration(ZabovKillTTL) * time.Hour)
	}

}
