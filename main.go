package main

import (
	"log"
	"net"

	"github.com/miekg/dns"
)

//MyDNS is my dns server
var MyDNS *dns.Server

//ZabovCacheTTL is the amount of hours we cache records of DNS (global)
var ZabovCacheTTL int

//ZabovKillTTL is the amount of hours we cache the killfile (global)
var ZabovKillTTL int

//ZabovLocalResponder is the default DNS server for loca domains
var ZabovLocalResponder string

//ZabovLocalDomain is the default local domain
var ZabovLocalDomain string

type handler struct{}

// ZabovConfig contains all Zabov configs
type ZabovConfig struct {
	ZabovSingleBL  string   // json:singlefilters -> ZabovSingleBL list of urls returning a file with just names of domains
	ZabovDoubleBL  string   // json:doublefilters -> ZabovDoubleBL list of urls returning a file with  IP<space>domain
	ZabovAddBL     string   // json:blackholeip  -> ZabovAddBL is the IP we want to send all the clients to. Usually is 127.0.0.1
	ZabovHostsFile string   // json:hostsfile -> ZabovHostsFile is the file we use to keep our hosts
	ZabovUpDNS     string   // json:upstream -> ZabovUpDNS keeps the name of upstream DNSs
	ZabovDNSArray  []string // contains all the DNS we mention, parsed from ZabovUpDNS file
	references     int      // contains references to this config; if zero, config shall be removed
}

// ZabovConfigs contains all Zabov configs
var ZabovConfigs map[string]*ZabovConfig

// ZabovIPGroup contains Zabov groups of IPs
type ZabovIPGroup struct {
	ips       []net.IP // IPs in this group
	cfg       string   // config name to be used if there is no timetable
	timetable string   // timetable name to be used for this group; timetable SHALL reference to config name to use
}

// ZabovIPGroups contains an array of all Zabov groups of IP rules
var ZabovIPGroups []ZabovIPGroup

// ZabovTime contains Zabov single time
type ZabovTime struct {
	hour   int
	minute int
}

// ZabovTimeRange contains Zabov single time range
type ZabovTimeRange struct {
	start ZabovTime
	stop  ZabovTime
}

// ZabovTimetableEntry contains Zabov single time table entry
type ZabovTimetableEntry struct {
	times []*ZabovTimeRange
	days  map[string]bool
}

// ZabovTimetable contains a Zabov time table
type ZabovTimetable struct {
	table  []*ZabovTimetableEntry
	cfgin  string // configuration name to be used if "inside" timetable
	cfgout string // configuration name to be used if "outiside" timetable
}

// ZabovTimetables contains all Zabov time tables, by name
var ZabovTimetables map[string]*ZabovTimetable

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
