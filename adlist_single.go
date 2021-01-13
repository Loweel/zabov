package main

import (
	"bufio"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"
)

func init() {
	go downloadThread()
}

//SingleIndexFilter puts the domains inside file
func SingleIndexFilter(durl string, configs stringarray) error {

	fmt.Println("Retrieving DomainFile from: ", durl)

	// resets malformed HostLines for url
	setstatsvalue("Malformed DomainLines "+durl, 0)

	var err error

	// Get the data
	resp, err := http.Get(durl)
	if err != nil {
		fmt.Println("HTTP Problem: ", err)
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

		if len(h) < 1 {
			continue
		}

		if !strings.Contains(h[0], "#") {

			DomainKill(h[0], durl, configs)

			// fmt.Println("MATCH: ", h[1])
			numLines++
		} else {
			incrementStats("Malformed DomainLines "+durl, 1)
			// fmt.Println("Malformed line: <" + line + ">")
		}

	}

	fmt.Println("Finished to parse: "+durl+" ,number of lines", numLines)

	return err

}

func getSingleFilters(urls urlsMap) {

	fmt.Println("getSingleFilters: downloading all urls:", len(urls))
	for url, configs := range urls {
		SingleIndexFilter(url, configs)
	}
	fmt.Println("getSingleFilters: DONE!")

}

func downloadThread() {
	fmt.Println("Starting updater of SINGLE lists, each (hours): ", ZabovKillTTL)
	_urls := urlsMap{}

	for {
		fmt.Println("downloadThread: collecting urls from all configs...")
		for config := range ZabovConfigs {
			ZabovSingleBL := ZabovConfigs[config].ZabovSingleBL

			s := fileByLines(ZabovSingleBL)
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

		getSingleFilters(_urls)
		time.Sleep(time.Duration(ZabovKillTTL) * time.Hour)
	}

}
