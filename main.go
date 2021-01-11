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

//ZabovDNSArrays contains the arrays containing all the DNS we mention
var ZabovDNSArrays map[string][]string

// ZabovConfig contains all Zabov configs
type ZabovConfig struct {
	ZabovUpDNS     string // json:upstream -> ZabovDNSArray
	ZabovSingleBL  string // json:singlefilters
	ZabovDoubleBL  string // json:doublefilters
	ZabovAddBL     string // json:blackholeip
	ZabovHostsFile string // json:hostsfile
}

// ZabovConfigs contains all Zabov configs
var ZabovConfigs map[string]ZabovConfig

// ZabovConfigDefault contains only "default" config
var ZabovConfigDefault ZabovConfig

// ZabovIPGroup contains Zabov groups of IPs
type ZabovIPGroup struct {
	ips       []string // IPs in this group
	cfg       string   // config name to be used if there is no timetable
	timetable string   // timetable name to be used for this group; timetable SHALL reference to config name to use
}

// ZabovIPGroups contains an array of all Zabov groups of IP rules
var ZabovIPGroups []ZabovIPGroup

// ZabovTimetableEntry contains Zabov single time table entry
type ZabovTimetableEntry struct {
	times []string
	days  []string
}

// ZabovTimetable contains a Zabov time table
type ZabovTimetable struct {
	table  []ZabovTimetableEntry
	cfgin  string // configuration name to be used if "inside" timetable
	cfgout string // configuration name to be used if "outiside" timetable
}

// ZabovTimetables contains all Zabov time tables, by name
var ZabovTimetables map[string]ZabovTimetable

// ZabovIPAliases contains an array of all Zabov IP aliases
var ZabovIPAliases map[string]string

func main() {

	MyDNS.Handler = &handler{}
	if err := MyDNS.ListenAndServe(); err != nil {
		log.Printf("Failed to set udp listener %s\n", err.Error())
	} else {
		log.Printf("Listener running \n")
	}
}
