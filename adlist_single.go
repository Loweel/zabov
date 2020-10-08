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
func SingleIndexFilter(durl string) error {

	fmt.Println("Retrieving DomainFile from: ", durl)

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

		h := strings.FieldsFunc(line, splitter)

		if h == nil {
			continue
		}

		if len(h) < 1 {
			continue
		}

		if !strings.Contains(h[0], "#") {
			DomainKill(h[0], durl)
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

func getSingleFilters() {

	s := fileByLines(ZabovSingleBL)

	for _, a := range s {
		SingleIndexFilter(a)
	}

}

func downloadThread() {
	fmt.Println("Starting updater of SINGLE lists, each (hours): ", ZabovKillTTL)
	for {
		getSingleFilters()
		time.Sleep(time.Duration(ZabovKillTTL) * time.Hour)
	}

}
