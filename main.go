package main

import (
	"log"

	"github.com/miekg/dns"
)

//MyDNS is my dns server
var MyDNS *dns.Server

//ZabovUpDNS keeps the name of upstream DNSs
var ZabovUpDNS string

//ZabovSingleBL list of urls returning a file with just names of domains
var ZabovSingleBL string

//ZabovDoubleBL list of urls returning a file with  IP<space>domain
var ZabovDoubleBL string

//ZabovAddBL is the IP we want to send all the clients to. Usually is 127.0.0.1
var ZabovAddBL string

//ZabovCacheTTL is the amount of hours we cache records of DNS
var ZabovCacheTTL int

//ZabovKillTTL is the amount of hours we cache the killfile
var ZabovKillTTL int

//ZabovHostsFile is the file we use to keep our hosts
var ZabovHostsFile string

//ZabovDNSArray is the array containing all the DNS we mention
var ZabovDNSArray []string

type handler struct{}

func main() {

	MyDNS.Handler = &handler{}
	if err := MyDNS.ListenAndServe(); err != nil {
		log.Printf("Failed to set udp listener %s\n", err.Error())
	} else {
		log.Printf("Listener running \n")
	}
}
