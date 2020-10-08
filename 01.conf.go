package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"

	"github.com/miekg/dns"
)

func init() {

	//ZabovConf describes the Json we use for configuration
	type ZabovConf struct {
		Zabov struct {
			Port          string `json:"port"`
			Proto         string `json:"proto"`
			Ipaddr        string `json:"ipaddr"`
			Upstream      string `json:"upstream"`
			Cachettl      int    `json:"cachettl"`
			Killfilettl   int    `json:"killfilettl"`
			Singlefilters string `json:"singlefilters"`
			Doublefilters string `json:"doublefilters"`
			Blackholeip   string `json:"blackholeip"`
			Hostsfile     string `json:"hostsfile"`
		} `json:"zabov"`
	}

	var MyConf ZabovConf

	file, err := ioutil.ReadFile("config.json")

	if err != nil {
		log.Println("Cannot open config file", err.Error())
		os.Exit(1)
	}

	err = json.Unmarshal([]byte(file), &MyConf)

	if err != nil {
		log.Println("Cannot marshal json: ", err.Error())
		os.Exit(1)
	}

	// now we read configuration file
	fmt.Println("Reading configuration file...")

	ZabovPort := MyConf.Zabov.Port
	ZabovType := MyConf.Zabov.Proto
	ZabovAddr := MyConf.Zabov.Ipaddr
	ZabovUpDNS = MyConf.Zabov.Upstream
	ZabovSingleBL = MyConf.Zabov.Singlefilters
	ZabovDoubleBL = MyConf.Zabov.Doublefilters
	ZabovAddBL = MyConf.Zabov.Blackholeip
	ZabovCacheTTL = MyConf.Zabov.Cachettl
	ZabovKillTTL = MyConf.Zabov.Killfilettl
	ZabovHostsFile = MyConf.Zabov.Hostsfile

	zabovString := ZabovAddr + ":" + ZabovPort

	MyDNS = new(dns.Server)
	MyDNS.Addr = zabovString
	MyDNS.Net = ZabovType

	ZabovDNSArray = fileByLines(ZabovUpDNS)

}
