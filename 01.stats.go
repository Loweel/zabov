package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"time"
)

type send struct {
	Payload   string
	Number    int64
	Operation string
}

//ZabovStats is used to keep statistics to print
var ZabovStats map[string]int64

var stats chan send

func init() {

	stats = make(chan send, 1024)

	ZabovStats = make(map[string]int64)

	fmt.Println("Initializing stats engine.")
	go reportPrintThread()
	go statsThread()
}

func statsPrint() {
	fmt.Println()
	stat, _ := json.Marshal(ZabovStats)
	fmt.Println(jsonPrettyPrint(string(stat)))
	fmt.Println()
}

func incrementStats(key string, value int64) {

	var s send

	s.Payload = key
	s.Number = value
	s.Operation = "INC"

	stats <- s

}

func setstatsvalue(key string, value int64) {

	var s send

	s.Payload = key
	s.Number = value
	s.Operation = "SET"

	stats <- s

}

func reportPrintThread() {
	for {
		var s send
		s.Operation = "PRI"
		s.Payload = "-"
		s.Number = 0
		stats <- s
		time.Sleep(2 * time.Minute)
	}
}

func statsThread() {

	fmt.Println("Starting Statistical Collection Thread")

	for item := range stats {

		switch item.Operation {
		case "INC":
			ZabovStats[item.Payload] += item.Number
		case "SET":
			if item.Number == 0 {

				delete(ZabovStats, item.Payload)
			} else {
				ZabovStats[item.Payload] = item.Number
			}
		case "PRI":
			statsPrint()
		}

	}

}

func jsonPrettyPrint(in string) string {
	var out bytes.Buffer
	err := json.Indent(&out, []byte(in), "", "\t")
	if err != nil {
		return in
	}
	return out.String()
}
