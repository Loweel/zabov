package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

func init() {

	fmt.Println("Ingesting local hosts file")
	ingestLocalBlacklists()

}

func ingestLocalBlacklists() {

	fmt.Println("ingestLocalBlacklist: collecting urls from all configs...")
	_files := urlsMap{}
	for config := range ZabovConfigs {
		ZabovHostsFile := ZabovConfigs[config].ZabovHostsFile
		if len(ZabovHostsFile) == 0 {
			continue
		}
		configs := _files[ZabovHostsFile]
		if configs == nil {
			configs = stringarray{}
			_files[ZabovHostsFile] = configs
		}
		configs = append(configs, config)
		_files[ZabovHostsFile] = configs
	}

	for ZabovHostsFile, configs := range _files {
		file, err := os.Open(ZabovHostsFile)
		if err != nil {
			fmt.Println(err.Error())
		}
		defer file.Close()

		scanner := bufio.NewScanner(file)
		for scanner.Scan() {
			d := scanner.Text()
			if len(d) == 0 || strings.TrimSpace(d)[0] == '#' {
				continue
			}
			DomainKill(d, ZabovHostsFile, configs)
			incrementStats("Blacklist", 1)

		}

		if err := scanner.Err(); err != nil {
			fmt.Println(err.Error())
		}
	}

}

func fileByLines(filename string) (blurls []string) {

	file, err := os.Open(filename)
	if err != nil {
		fmt.Println(err.Error())
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		d := scanner.Text()
		blurls = append(blurls, d)

	}

	if err := scanner.Err(); err != nil {
		fmt.Println(err.Error())
	}

	return

}
