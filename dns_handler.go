package main

import (
	"fmt"
	"log"
	"net"
	"os"
	"path"
	"strings"
	"time"

	"github.com/miekg/dns"
)

var reqTypes map[uint16]string

var weekdays []string

type logItem struct {
	clientIP  string
	name      string
	reqType   uint16
	config    string
	timetable string
	killed    string
}

// logChannel used by logging thread
var logChannel chan logItem

func init() {

	weekdays = []string{"Su", "Mo", "Tu", "We", "Th", "Fr", "Sa"}

	if len(ZabovDebugDBPath) > 0 {
		os.MkdirAll(ZabovDebugDBPath, 0755)
	}

	reqTypes = map[uint16]string{
		dns.TypeNone:       "TypeNone",
		dns.TypeA:          "TypeA",
		dns.TypeNS:         "TypeNS",
		dns.TypeMD:         "TypeMD",
		dns.TypeMF:         "TypeMF",
		dns.TypeCNAME:      "TypeCNAME",
		dns.TypeSOA:        "TypeSOA",
		dns.TypeMB:         "TypeMB",
		dns.TypeMG:         "TypeMG",
		dns.TypeMR:         "TypeMR",
		dns.TypeNULL:       "TypeNULL",
		dns.TypePTR:        "TypePTR",
		dns.TypeHINFO:      "TypeHINFO",
		dns.TypeMINFO:      "TypeMINFO",
		dns.TypeMX:         "TypeMX",
		dns.TypeTXT:        "TypeTXT",
		dns.TypeRP:         "TypeRP",
		dns.TypeAFSDB:      "TypeAFSDB",
		dns.TypeX25:        "TypeX25",
		dns.TypeISDN:       "TypeISDN",
		dns.TypeRT:         "TypeRT",
		dns.TypeNSAPPTR:    "TypeNSAPPTR",
		dns.TypeSIG:        "TypeSIG",
		dns.TypeKEY:        "TypeKEY",
		dns.TypePX:         "TypePX",
		dns.TypeGPOS:       "TypeGPOS",
		dns.TypeAAAA:       "TypeAAAA",
		dns.TypeLOC:        "TypeLOC",
		dns.TypeNXT:        "TypeNXT",
		dns.TypeEID:        "TypeEID",
		dns.TypeNIMLOC:     "TypeNIMLOC",
		dns.TypeSRV:        "TypeSRV",
		dns.TypeATMA:       "TypeATMA",
		dns.TypeNAPTR:      "TypeNAPTR",
		dns.TypeKX:         "TypeKX",
		dns.TypeCERT:       "TypeCERT",
		dns.TypeDNAME:      "TypeDNAME",
		dns.TypeOPT:        "TypeOPT",
		dns.TypeAPL:        "TypeAPL",
		dns.TypeDS:         "TypeDS",
		dns.TypeSSHFP:      "TypeSSHFP",
		dns.TypeRRSIG:      "TypeRRSIG",
		dns.TypeNSEC:       "TypeNSEC",
		dns.TypeDNSKEY:     "TypeDNSKEY",
		dns.TypeDHCID:      "TypeDHCID",
		dns.TypeNSEC3:      "TypeNSEC3",
		dns.TypeNSEC3PARAM: "TypeNSEC3PARAM",
		dns.TypeTLSA:       "TypeTLSA",
		dns.TypeSMIMEA:     "TypeSMIMEA",
		dns.TypeHIP:        "TypeHIP",
		dns.TypeNINFO:      "TypeNINFO",
		dns.TypeRKEY:       "TypeRKEY",
		dns.TypeTALINK:     "TypeTALINK",
		dns.TypeCDS:        "TypeCDS",
		dns.TypeCDNSKEY:    "TypeCDNSKEY",
		dns.TypeOPENPGPKEY: "TypeOPENPGPKEY",
		dns.TypeCSYNC:      "TypeCSYNC",
		dns.TypeSPF:        "TypeSPF",
		dns.TypeUINFO:      "TypeUINFO",
		dns.TypeUID:        "TypeUID",
		dns.TypeGID:        "TypeGID",
		dns.TypeUNSPEC:     "TypeUNSPEC",
		dns.TypeNID:        "TypeNID",
		dns.TypeL32:        "TypeL32",
		dns.TypeL64:        "TypeL64",
		dns.TypeLP:         "TypeLP",
		dns.TypeEUI48:      "TypeEUI48",
		dns.TypeEUI64:      "TypeEUI64",
		dns.TypeURI:        "TypeURI",
		dns.TypeCAA:        "TypeCAA",
		dns.TypeAVC:        "TypeAVC",
		dns.TypeTKEY:       "TypeTKEY",
		dns.TypeTSIG:       "TypeTSIG",
		dns.TypeIXFR:       "TypeIXFR",
		dns.TypeAXFR:       "TypeAXFR",
		dns.TypeMAILB:      "TypeMAILB",
		dns.TypeMAILA:      "TypeMAILA",
		dns.TypeANY:        "TypeANY",
		dns.TypeTA:         "TypeTA",
		dns.TypeDLV:        "TypeDLV",
		dns.TypeReserved:   "TypeReserved"}

	fmt.Println("Local Time:", getLocalTime().Format(time.ANSIC))

	if len(ZabovDebugDBPath) > 0 {
		logChannel = make(chan logItem, 1024)
		go logWriteThread()
	}
}

func logWriteThread() {
	for item := range logChannel {
		var header string
		d := time.Now().Format("2006-01-02")
		logpath := path.Join(ZabovDebugDBPath, strings.Replace(item.clientIP, ":", "_", -1)+"-"+d+".log")

		_, err1 := os.Stat(logpath)
		if os.IsNotExist(err1) {
			header = strings.Join([]string{"time", "clientIP", "name", "reqType", "config", "timetable", "killed"}, "\t")
		}
		f, err := os.OpenFile(logpath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if err == nil {
			reqTypeName, err := reqTypes[item.reqType]
			if !err {
				reqTypeName = fmt.Sprintf("%d", item.reqType)
			}
			ct := time.Now().Format(time.RFC3339)
			log := strings.Join([]string{ct, item.clientIP, strings.TrimRight(item.name, "."), reqTypeName, item.config, item.timetable, item.killed}, "\t")
			if len(header) > 0 {
				f.Write([]byte(header))
				f.Write([]byte("\n"))
			}
			f.Write([]byte(log))
			f.Write([]byte("\n"))
			f.Close()
		}
	}
}

func logQuery(clientIP string, name string, reqType uint16, config string, timetable string, killed string) {
	if len(ZabovDebugDBPath) > 0 {
		k := logItem{clientIP: clientIP, name: name, reqType: reqType, config: config, timetable: timetable, killed: killed}

		logChannel <- k

	}
}

func getLocalTime() time.Time {
	return time.Now().Local()
}

func confFromTimeTable(timetable string) string {
	tt := ZabovTimetables[timetable]
	if tt == nil {
		if ZabovDebug {
			log.Println("confFromTimeTable: return default")
		}
		return "default"
	}
	for _, ttentry := range tt.table {
		now := getLocalTime()

		nowHour := now.Hour()
		nowMinute := now.Minute()
		weekday := weekdays[now.Weekday()]
		if ttentry.days == nil || len(ttentry.days) == 0 || ttentry.days[weekday] || ttentry.days[strings.ToLower(weekday)] {
			for _, t := range ttentry.times {

				if (nowHour > t.start.hour || (nowHour == t.start.hour && nowMinute >= t.start.minute)) &&
					(nowHour < t.stop.hour || (nowHour == t.stop.hour && nowMinute <= t.stop.minute)) {
					go incrementStats("TIMETABLE IN: "+timetable, 1)
					if ZabovDebug {
						log.Println("confFromTimeTable: return IN", tt.cfgin)
					}
					return tt.cfgin
				}
			}
		}
	}
	go incrementStats("TIMETABLE OUT: "+timetable, 1)
	if ZabovDebug {
		log.Println("confFromTimeTable: return OUT", tt.cfgout)
	}
	return tt.cfgout
}

func confFromIP(clientIP net.IP) (string, string) {
	for _, ipgroup := range ZabovIPGroups {
		for _, ip := range ipgroup.ips {
			if clientIP.Equal(ip) {
				if len(ipgroup.timetable) > 0 {
					return confFromTimeTable(ipgroup.timetable), ipgroup.timetable
				}
				if ZabovDebug {
					log.Println("confFromIP: ipgroup.cfg", ipgroup.cfg)
				}
				return ipgroup.cfg, ""
			}
		}
	}
	if len(ZabovDefaultTimetable) > 0 {
		return confFromTimeTable(ZabovDefaultTimetable), ZabovDefaultTimetable
	}

	if ZabovDebug {
		log.Println("confFromIP: return default")
	}
	return "default", ""
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

	config, timetable := confFromIP(net.ParseIP(remIP))

	if ZabovDebug {
		log.Println("REQUEST:", remIP, config)
	}
	ZabovConfig := ZabovConfigs[config]
	QType := r.Question[0].Qtype
	switch QType {
	case dns.TypeA:
		msg.Authoritative = true
		domain := msg.Question[0].Name
		fqdn := strings.TrimRight(domain, ".")

		if ZabovDebug {
			log.Println("TypeA: fqdn:", fqdn)
		}

		if len(ZabovIPAliases[fqdn]) > 0 {
			config = "__aliases__"
			msg.Answer = append(msg.Answer, &dns.A{
				Hdr: dns.RR_Header{Name: domain, Rrtype: dns.TypeA, Class: dns.ClassINET, Ttl: 60},
				A:   net.ParseIP(ZabovIPAliases[fqdn]),
			})
			go logQuery(remIP, fqdn, QType, config, timetable, "alias")
			break
		}
		if len(ZabovLocalResponder) > 0 {
			if !strings.Contains(fqdn, ".") ||
				(len(ZabovLocalDomain) > 0 && strings.HasSuffix(fqdn, ZabovLocalDomain)) {
				config = localresponderConfigName
				ret := ForwardQuery(r, config, true)
				w.WriteMsg(ret)
				go logQuery(remIP, fqdn, QType, config, timetable, "localresponder")
				break
			}

		}
		if domainInKillfile(fqdn, config) {
			go incrementStats("Killed", 1)

			msg.Answer = append(msg.Answer, &dns.A{
				Hdr: dns.RR_Header{Name: domain, Rrtype: dns.TypeA, Class: dns.ClassINET, Ttl: 60},
				A:   net.ParseIP(ZabovConfig.ZabovAddBL),
			})
			go logQuery(remIP, fqdn, QType, config, timetable, "killed")
		} else {
			go logQuery(remIP, fqdn, QType, config, timetable, "forwarded")
			ret := ForwardQuery(r, config, false)
			w.WriteMsg(ret)
		}
	case dns.TypePTR:
		if ZabovDebug {
			log.Println("TypePTR: Name:", msg.Question[0].Name)
		}

		if len(ZabovLocalResponder) > 0 {
			// if set use local responder for reverse lookup (suffix ".in-addr.arpa.")
			config = localresponderConfigName
		}
		ret := ForwardQuery(r, config, true)
		w.WriteMsg(ret)
		go logQuery(remIP, msg.Question[0].Name, QType, config, timetable, "localresponder")
	default:
		ret := ForwardQuery(r, config, false)
		w.WriteMsg(ret)
		if len(ZabovDebugDBPath) > 0 {
			go logQuery(remIP, msg.Question[0].Name, QType, config, timetable, "forwarded")
		}
	}
	go incrementStats("CONFIG: "+config, 1)
	w.WriteMsg(&msg)

}
