package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
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

	defaultConf := configs["default"].(map[string]interface{})

	ZabovUpDNS = defaultConf["upstream"].(string)
	ZabovSingleBL = defaultConf["singlefilters"].(string)
	ZabovDoubleBL = defaultConf["doublefilters"].(string)
	ZabovAddBL = defaultConf["blackholeip"].(string)
	ZabovHostsFile = defaultConf["hostsfile"].(string)

	zabovString := ZabovAddr + ":" + ZabovPort

	MyDNS = new(dns.Server)
	MyDNS.Addr = zabovString
	MyDNS.Net = ZabovType

	ZabovDNSArray = fileByLines(ZabovUpDNS)

	ZabovConfigs = map[string]ZabovConfig{}
	ZabovIPGroups = []ZabovIPGroup{}
	ZabovTimetables = map[string]ZabovTimetable{}
	ZabovIPAliases = map[string]string{}
	ZabovDNSArrays = map[string][]string{}
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

		ZabovDNSArrays[name] = fileByLines(conf.ZabovUpDNS)
		ZabovConfigs[name] = conf
		if name == "default" {
			ZabovConfigDefault = conf
		}
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
			log.Println("inexistent cfgin:", timetable.cfgin)
			os.Exit(1)
		}

		_, ok = ZabovConfigs[timetable.cfgout]
		if !ok {
			log.Println("inexistent cfgout:", timetable.cfgout)
			os.Exit(1)
		}

		tables := timetableRaw["tables"].([]interface{})

		for i := range tables {
			table := tables[i].(map[string]interface{})
			var ttEntry ZabovTimetableEntry
			ttEntry.times = strings.Split(table["times"].(string), ";")
			ttEntry.days = strings.Split(table["days"].(string), ";")
			timetable.table = append(timetable.table, ttEntry)
		}
		ZabovTimetables[name] = timetable
	}

	IPGroups := MyConf["ipgroups"].([]interface{})

	fmt.Println("evaluating IP Groups: ", len(IPGroups))
	for i := range IPGroups {
		fmt.Println("evaluating IP Group n.", i)
		var groupStruct ZabovIPGroup
		groupMap := IPGroups[i].(map[string]interface{})
		IPsRaw := groupMap["ips"].([]interface{})
		groupStruct.ips = []string{}
		for x := range IPsRaw {
			ip := IPsRaw[x].(string)
			fmt.Println("adding IP ", ip)

			alias, ok := ZabovIPAliases[ip]
			if ok {
				fmt.Println("IP alias: ", ip, alias)
				ip = alias
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
	fmt.Println("ZabovConfigs:", ZabovConfigs)
	fmt.Println("ZabovTimetables:", ZabovTimetables)
	fmt.Println("ZabovIPAliases:", ZabovIPAliases)
	fmt.Println("ZabovIPGroups:", ZabovIPGroups)

}
