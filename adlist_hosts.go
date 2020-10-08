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
func DoubleIndexFilter(durl string) error {

	fmt.Println("Retrieving HostFile from: ", durl)

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

		h := strings.FieldsFunc(line, splitter)

		if h == nil {
			continue
		}

		if len(h) < 2 {
			continue
		}

		if net.ParseIP(h[0]) != nil {
			DomainKill(h[1], durl)

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

func getDoubleFilters() {

	s := fileByLines(ZabovDoubleBL)

	for _, a := range s {
		DoubleIndexFilter(a)
	}

}

func downloadDoubleThread() {
	fmt.Println("Starting updater of DOUBLE lists, each (hours):", ZabovKillTTL)
	for {
		getDoubleFilters()
		time.Sleep(time.Duration(ZabovKillTTL) * time.Hour)
	}

}
