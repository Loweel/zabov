package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"os"
	"strconv"
	"strings"

	"github.com/miekg/dns"
)

type stringarray []string
type urlsMap map[string]stringarray

func init() {
	var MyConfRaw interface{}

	file, err := ioutil.ReadFile("config.json")

	if err != nil {
		log.Println("Cannot open config file", err.Error())
		os.Exit(1)
	}

	err = json.Unmarshal([]byte(file), &MyConfRaw)

	if err != nil {
		log.Println("Cannot unmarshal json: ", err.Error())
		os.Exit(1)
	}

	// now we read configuration file
	fmt.Println("Reading configuration file...")

	MyConf := MyConfRaw.(map[string]interface{})

	zabov := MyConf["zabov"].(map[string]interface{})

	ZabovPort := zabov["port"].(string)
	ZabovType := zabov["proto"].(string)
	ZabovAddr := zabov["ipaddr"].(string)
	ZabovCacheTTL = int(zabov["cachettl"].(float64))
	ZabovKillTTL = int(zabov["killfilettl"].(float64))

	configs := MyConf["configs"].(map[string]interface{})

	if len(configs) == 0 {
		log.Println("you shall set at least default config")
		os.Exit(1)
	}

	if configs["default"] == nil {
		log.Println("default config is required")
		os.Exit(1)
	}

	zabovString := ZabovAddr + ":" + ZabovPort

	MyDNS = new(dns.Server)
	MyDNS.Addr = zabovString
	MyDNS.Net = ZabovType

	ZabovConfigs = map[string]ZabovConfig{}
	ZabovIPGroups = []ZabovIPGroup{}
	ZabovTimetables = map[string]*ZabovTimetable{}
	ZabovIPAliases = map[string]string{}

	IPAliasesRaw := MyConf["ipaliases"].(map[string]interface{})

	for alias, ip := range IPAliasesRaw {
		fmt.Println("IP Alias:", alias, ip)
		ZabovIPAliases[alias] = ip.(string)
	}

	for name, v := range configs {
		fmt.Println("evaluaing config name:", name)
		confRaw := v.(map[string]interface{})
		var conf ZabovConfig
		conf.ZabovUpDNS = confRaw["upstream"].(string)
		conf.ZabovSingleBL = confRaw["singlefilters"].(string)
		conf.ZabovDoubleBL = confRaw["doublefilters"].(string)
		conf.ZabovAddBL = confRaw["blackholeip"].(string)
		conf.ZabovHostsFile = confRaw["hostsfile"].(string)

		conf.ZabovDNSArray = fileByLines(conf.ZabovUpDNS)
		ZabovConfigs[name] = conf
		ZabovCreateKDB(name)
	}

	timetables := MyConf["timetables"].(map[string]interface{})

	for name, v := range timetables {
		fmt.Println("evaluaing timetable name:", name)
		timetableRaw := v.(map[string]interface{})
		var timetable ZabovTimetable

		timetable.cfgin = timetableRaw["cfgin"].(string)
		timetable.cfgout = timetableRaw["cfgout"].(string)

		if timetable.cfgin == "" {
			timetable.cfgin = "default"
		}
		if timetable.cfgout == "" {
			timetable.cfgout = "default"
		}

		_, ok := ZabovConfigs[timetable.cfgin]
		if !ok {
			log.Println("timetable: inexistent cfgin:", timetable.cfgin)
			os.Exit(1)
		}

		_, ok = ZabovConfigs[timetable.cfgout]
		if !ok {
			log.Println("timetable: inexistent cfgout:", timetable.cfgout)
			os.Exit(1)
		}

		tables := timetableRaw["tables"].([]interface{})

		for i := range tables {
			table := tables[i].(map[string]interface{})
			var ttEntry ZabovTimetableEntry
			ttEntry.times = []*ZabovTimeRange{}
			for _, tRaw := range strings.Split(table["times"].(string), ";") {
				tRawArr := strings.Split(tRaw, "-")
				if len(tRawArr) > 1 {
					startArr := strings.Split(tRawArr[0], ":")
					stopArr := strings.Split(tRawArr[1], ":")

					if len(startArr) > 1 && len(stopArr) > 1 {
						hourStart, _ := strconv.Atoi(startArr[0])
						minuteStart, _ := strconv.Atoi(startArr[1])
						start := ZabovTime{hour: hourStart, minute: minuteStart}

						hourStop, _ := strconv.Atoi(stopArr[0])
						minuteStop, _ := strconv.Atoi(stopArr[1])
						stop := ZabovTime{hour: hourStop, minute: minuteStop}
						t := ZabovTimeRange{start: start, stop: stop}
						ttEntry.times = append(ttEntry.times, &t)
					}
				}

			}

			ttEntry.days = map[string]bool{}
			for _, day := range strings.Split(table["days"].(string), ";") {
				ttEntry.days[day] = true
			}

			timetable.table = append(timetable.table, &ttEntry)
		}
		ZabovTimetables[name] = &timetable
	}

	IPGroups := MyConf["ipgroups"].([]interface{})

	fmt.Println("evaluating IP Groups: ", len(IPGroups))
	for i := range IPGroups {
		fmt.Println("evaluating IP Group n.", i)
		var groupStruct ZabovIPGroup
		groupMap := IPGroups[i].(map[string]interface{})
		IPsRaw := groupMap["ips"].([]interface{})
		groupStruct.ips = []net.IP{}
		for x := range IPsRaw {
			ipRaw := IPsRaw[x].(string)
			ip := net.ParseIP(ipRaw)
			fmt.Println("adding IP ", ipRaw)

			alias, ok := ZabovIPAliases[ipRaw]
			if ok {
				fmt.Println("IP alias: ", ipRaw, alias)
				ip = net.ParseIP(alias)
			}
			groupStruct.ips = append(groupStruct.ips, ip)
		}
		groupStruct.cfg = groupMap["cfg"].(string)
		groupStruct.timetable = groupMap["timetable"].(string)
		fmt.Println("cfg:", groupStruct.cfg)
		fmt.Println("timetable:", groupStruct.timetable)
		_, ok := ZabovTimetables[groupStruct.timetable]
		if !ok {
			log.Println("inexistent timetable:", groupStruct.timetable)
			os.Exit(1)
		}
		ZabovIPGroups = append(ZabovIPGroups, groupStruct)
	}
	//fmt.Println("ZabovConfigs:", ZabovConfigs)
	//fmt.Println("ZabovTimetables:", ZabovTimetables)
	//fmt.Println("ZabovIPAliases:", ZabovIPAliases)
	//fmt.Println("ZabovIPGroups:", ZabovIPGroups)

}
