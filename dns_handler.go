package main

import (
	"fmt"
	"net"
	"strings"
	"time"

	"github.com/miekg/dns"
)

var weekdays []string

func init() {
	weekdays = []string{"Su", "Mo", "Tu", "We", "Th", "Fr", "Sa"}
}

func getCurTime() (time.Time, error) {
	return time.Parse("15:04", time.Now().Format("15:04"))
}

func confFromTimeTable(timetable string) string {
	tt := ZabovTimetables[timetable]
	if tt == nil {
		fmt.Println("confFromTimeTable: return default")
		return "default"
	}
	for _, ttentry := range tt.table {
		now := time.Now()
		nowHour := now.Hour()
		nowMinute := now.Minute()
		weekday := weekdays[now.Weekday()]
		if ttentry.days == nil || len(ttentry.days) == 0 || ttentry.days[weekday] || ttentry.days[strings.ToLower(weekday)] {
			for _, t := range ttentry.times {

				if (nowHour > t.start.hour || (nowHour == t.start.hour && nowMinute >= t.start.minute)) &&
					(nowHour < t.stop.hour || (nowHour == t.stop.hour && nowMinute <= t.stop.minute)) {
					incrementStats("TIMETABLE IN: "+timetable, 1)
					fmt.Println("confFromTimeTable: return IN", tt.cfgin)
					return tt.cfgin
				}
			}
		}
	}
	incrementStats("TIMETABLE OUT: "+timetable, 1)
	fmt.Println("confFromTimeTable: return OUT", tt.cfgout)
	return tt.cfgout
}

func confFromIP(clientIP net.IP) string {

	for _, ipgroup := range ZabovIPGroups {
		for _, ip := range ipgroup.ips {
			if clientIP.Equal(ip) {
				if len(ipgroup.timetable) > 0 {
					return confFromTimeTable(ipgroup.timetable)
				}
				fmt.Println("confFromIP: ipgroup.cfg")
				return ipgroup.cfg
			}
		}
	}
	fmt.Println("confFromIP: return default")
	return "default"
}
func (mydns *handler) ServeDNS(w dns.ResponseWriter, r *dns.Msg) {
	go incrementStats("TotalQueries", 1)

	remIP, _, e := net.SplitHostPort(w.RemoteAddr().String())
	if e != nil {
		go incrementStats("CLIENT ERROR: "+remIP, 1)
	} else {
		go incrementStats("CLIENT: "+remIP, 1)
	}

	msg := dns.Msg{}
	msg.SetReply(r)

	config := confFromIP(net.ParseIP(remIP))
	incrementStats("CONFIG: "+config, 1)

	ZabovConfig := ZabovConfigs[config]
	switch r.Question[0].Qtype {
	case dns.TypeA:
		msg.Authoritative = true
		domain := msg.Question[0].Name
		fqdn := strings.TrimRight(domain, ".")

		if domainInKillfile(fqdn, config) {
			go incrementStats("Killed", 1)

			msg.Answer = append(msg.Answer, &dns.A{
				Hdr: dns.RR_Header{Name: domain, Rrtype: dns.TypeA, Class: dns.ClassINET, Ttl: 60},
				A:   net.ParseIP(ZabovConfig.ZabovAddBL),
			})
		} else {
			ret := ForwardQuery(r, config)
			w.WriteMsg(ret)
		}
	default:
		ret := ForwardQuery(r, config)
		w.WriteMsg(ret)
	}
	w.WriteMsg(&msg)

}
