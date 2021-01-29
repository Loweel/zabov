package main

import (
	"fmt"
	"log"
	"time"

	"math/rand"
	"strings"

	"github.com/miekg/dns"
)

//ForwardQuery forwards the query to the upstream server
//first server to answer wins
//accepts config name to select the UP DNS source list
func ForwardQuery(query *dns.Msg, config string, nocache bool) *dns.Msg {
	if ZabovDebug {
		log.Println("ForwardQuery: nocache", nocache)
	}
	go incrementStats("ForwardQueries", 1)

	r := new(dns.Msg)
	r.SetReply(query)
	r.Authoritative = true

	fqdn := strings.TrimRight(query.Question[0].Name, ".")

	lfqdn := fmt.Sprintf("%d", query.Question[0].Qtype) + "." + fqdn
	if !nocache {
		if cached := GetDomainFromCache(lfqdn); cached != nil {
			go incrementStats("CacheHit", 1)
			cached.SetReply(query)
			cached.Authoritative = true
			if ZabovDebug {
				log.Println("ForwardQuery: CacheHit")
			}
			cached.Compress = true
			return cached

		}
	}

	c := new(dns.Client)

	c.ReadTimeout = 500 * time.Millisecond
	c.WriteTimeout = 500 * time.Millisecond

	for {
		// round robin with retry

		// local responder should always be available also if no internet connection
		if !NetworkUp && localresponderConfigName != config {
			time.Sleep(10 * time.Second)
			go incrementStats("Network Problems ", 1)
			continue
		}

		d := oneTimeDNS(config)

		in, _, err := c.Exchange(query, d)
		if err != nil {
			fmt.Printf("Problem with DNS %s : %s\n", d, err.Error())
			go incrementStats("DNS Problems "+d, 1)
			continue
		} else {
			go incrementStats(d, 1)
			in.SetReply(query)
			in.Authoritative = true
			in.Compress = true
			go DomainCache(lfqdn, in)
			if ZabovDebug {
				log.Println("ForwardQuery: OK!")
			}

			return in

		}

	}

}

func init() {

	fmt.Println("DNS client engine starting")
	NetworkUp = checkNetworkUp()

	if NetworkUp {
		fmt.Println("[OK]: Network is UP")
	} else {
		fmt.Println("[KO] Network is DOWN: system will check again in 2 minutes")
	}

}

func oneTimeDNS(config string) (dns string) {

	rand.Seed(time.Now().Unix())

	upl := ZabovConfigs[config].ZabovDNSArray

	if len(upl) < 1 {

		if len(ZabovLocalResponder) > 0 {
			fmt.Println("No DNS defined, fallback to local responder:", ZabovLocalResponder)
			return ZabovLocalResponder
		}
		fmt.Println("No DNS defined, using default 127.0.0.53:53. Hope it works!")
		return "127.0.0.53:53"
	}

	n := rand.Intn(128*len(upl)) % len(upl)

	dns = upl[n]

	return

}
