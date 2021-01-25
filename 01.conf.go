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

var localresponderConfigName string

type stringarray []string
type urlsMap map[string]stringarray

func init() {
	localresponderConfigName = "__localresponder__"
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

	//******************************
	// zabov section (global config)
	//******************************
	zabov := MyConf["zabov"].(map[string]interface{})

	ZabovPort := zabov["port"].(string)
	ZabovType := zabov["proto"].(string)
	ZabovAddr := zabov["ipaddr"].(string)

	ZabovCacheTTL = int(zabov["cachettl"].(float64))
	ZabovKillTTL = int(zabov["killfilettl"].(float64))

	if zabov["debug"] != nil {
		ZabovDebug = zabov["debug"].(string) == "true"
	}
	if zabov["debugdbpath"] != nil {
		ZabovDebugDBPath = (zabov["debugdbpath"].(string))
	}

	if MyConf["configs"] == nil {
		log.Println("configs not set: you shall set at least 'default' config")
		os.Exit(1)
	}

	configs := MyConf["configs"].(map[string]interface{})

	if len(configs) == 0 {
		log.Println("you shall set at least 'default' config")
		os.Exit(1)
	}

	if configs["default"] == nil {
		log.Println("'default' config is required")
		os.Exit(1)
	}

	zabovString := ZabovAddr + ":" + ZabovPort

	MyDNS = new(dns.Server)
	MyDNS.Addr = zabovString
	MyDNS.Net = ZabovType

	ZabovConfigs = map[string]*ZabovConfig{}
	ZabovIPGroups = []ZabovIPGroup{}
	ZabovTimetables = map[string]*ZabovTimetable{}
	ZabovIPAliases = map[string]string{}

	//*******************
	// IP aliases section
	//*******************
	if MyConf["ipaliases"] != nil {
		IPAliasesRaw := MyConf["ipaliases"].(map[string]interface{})

		for alias, ip := range IPAliasesRaw {
			fmt.Println("IP Alias:", alias, ip)
			ZabovIPAliases[alias] = ip.(string)
		}
	}

	//****************
	// configs section
	//****************
	for name, v := range configs {
		fmt.Println("evaluaing config name:", name)
		confRaw := v.(map[string]interface{})
		var conf ZabovConfig
		conf.ZabovUpDNS = confRaw["upstream"].(string)
		conf.ZabovSingleBL = confRaw["singlefilters"].(string)
		conf.ZabovDoubleBL = confRaw["doublefilters"].(string)
		conf.ZabovAddBL = net.ParseIP(confRaw["blackholeip"].(string))
		conf.ZabovHostsFile = confRaw["hostsfile"].(string)

		conf.ZabovDNSArray = fileByLines(conf.ZabovUpDNS)
		ZabovConfigs[name] = &conf

	}

	// default config is mandatory
	ZabovConfigs["default"].references++

	//*******************
	// timetables section
	//*******************
	if MyConf["timetables"] != nil {
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

			refConfig, ok := ZabovConfigs[timetable.cfgin]
			if !ok {
				log.Println("timetable: inexistent cfgin:", timetable.cfgin)
				os.Exit(1)
			}

			refConfig.references++
			refConfig, ok = ZabovConfigs[timetable.cfgout]
			if !ok {
				log.Println("timetable: inexistent cfgout:", timetable.cfgout)
				os.Exit(1)
			}
			refConfig.references++

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
	}

	//******************
	// IP groups section
	//******************
	if MyConf["ipgroups"] != nil {
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
			if len(groupStruct.cfg) > 0 {
				refConfig, ok := ZabovConfigs[groupStruct.cfg]
				if !ok {
					log.Println("ipgroups: inexistent cfg:", groupStruct.cfg)
					os.Exit(1)
				} else {
					refConfig.references++
				}
			}
			fmt.Println("cfg:", groupStruct.cfg)
			fmt.Println("timetable:", groupStruct.timetable)
			_, ok := ZabovTimetables[groupStruct.timetable]
			if !ok {
				log.Println("inexistent timetable:", groupStruct.timetable)
				os.Exit(1)
			}
			ZabovIPGroups = append(ZabovIPGroups, groupStruct)
		}
	}

	if zabov["timetable"] != nil {
		ZabovDefaultTimetable = zabov["timetable"].(string)
		_, ok := ZabovTimetables[ZabovDefaultTimetable]
		if !ok {
			log.Println("inexistent timetable:", ZabovDefaultTimetable)
			os.Exit(1)
		}
	}

	//************************
	// Local responder section
	//************************
	if MyConf["localresponder"] != nil {
		localresponder := MyConf["localresponder"].(map[string]interface{})

		if localresponder != nil {
			if localresponder["responder"] != nil {
				ZabovLocalResponder = localresponder["responder"].(string)
				if len(ZabovLocalResponder) > 0 {
					local := ZabovConfig{ZabovDNSArray: []string{ZabovLocalResponder}, references: 1}
					ZabovConfigs[localresponderConfigName] = &local
					fmt.Println("ZabovLocalResponder:", ZabovLocalResponder)
				}
			}
			if localresponder["localdomain"] != nil {
				ZabovLocalDomain = localresponder["localdomain"].(string)
			}
		}
	}
	//******************************************
	// clearing unused configs to save resources
	//******************************************
	for name, conf := range ZabovConfigs {
		if conf.references == 0 {
			log.Println("WARNING: disabling unused configuration:", name)
			delete(ZabovConfigs, name)
		} else {
			ZabovCreateKDB(name)
		}
	}

}
