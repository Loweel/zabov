package main

import (
	"fmt"
	"time"

	"math/rand"
	"strings"

	"github.com/miekg/dns"
)

//ForwardQuery forwards the query to the upstream server
//first server to answer wins
func ForwardQuery(query *dns.Msg) *dns.Msg {

	go incrementStats("ForwardQueries", 1)

	r := new(dns.Msg)
	r.SetReply(query)
	r.Authoritative = true

	fqdn := strings.TrimRight(query.Question[0].Name, ".")

	lfqdn := fmt.Sprintf("%d", query.Question[0].Qtype) + "." + fqdn
	if cached := GetDomainFromCache(lfqdn); cached != nil {
		go incrementStats("CacheHit", 1)
		cached.SetReply(query)
		cached.Authoritative = true
		return cached

	}

	c := new(dns.Client)

	c.ReadTimeout = 500 * time.Millisecond
	c.WriteTimeout = 500 * time.Millisecond

	for {
		// round robin with retry

		if !NetworkUp {
			time.Sleep(10 * time.Second)
			go incrementStats("Network Problems ", 1)
			continue
		}

		d := oneTimeDNS()

		in, _, err := c.Exchange(query, d)
		if err != nil {
			fmt.Printf("Problem with DNS %s : %s\n", d, err.Error())
			go incrementStats("DNS Problems "+d, 1)
			continue
		} else {
			go incrementStats(d, 1)
			in.SetReply(query)
			in.Authoritative = true
			go DomainCache(lfqdn, in)
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

func oneTimeDNS() (dns string) {

	rand.Seed(time.Now().Unix())

	upl := ZabovDNSArray

	if len(upl) < 1 {
		fmt.Println("No DNS defined, using default 127.0.0.53:53. Hope it works!")
		return "127.0.0.53:53"
	}

	n := rand.Intn(128*len(upl)) % len(upl)

	dns = upl[n]

	return

}
