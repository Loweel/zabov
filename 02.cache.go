package main

import (
	"bytes"
	"encoding/gob"
	"fmt"
	"time"

	"github.com/miekg/dns"
)



type cacheItem struct {
	Query []byte
	Date  time.Time
}

//DomainCache stores a domain name inside the cache
func DomainCache(s string, resp *dns.Msg) {

	var domain2cache cacheItem
	var err error
	var dom2 bytes.Buffer
	enc := gob.NewEncoder(&dom2)

	domain2cache.Query, err = resp.Pack()
	if err != nil {
		fmt.Println("Problems packing the response: ", err.Error())
	}
	domain2cache.Date = time.Now()

	err = enc.Encode(domain2cache)

	if err != nil {
		fmt.Println("Cannot GOB the domain to cache: ", err.Error())
	}

	cacheDomain(s, dom2.Bytes())

}

func cacheDomain(key string, domain []byte) {

	err := MyZabovCDB.Put([]byte(key), domain, nil)
	if err != nil {
		fmt.Println("Cannot write to Cache DB: ", err.Error())
	}

}

//GetDomainFromCache stores a domain name inside the cache
func GetDomainFromCache(s string) *dns.Msg {

	ret := new(dns.Msg)
	var cache bytes.Buffer
	dec := gob.NewDecoder(&cache)
	var record cacheItem
	var conf []byte
	var errDB error

	if domainInCache(s) == false {
		return nil
	}

	conf, errDB = MyZabovCDB.Get([]byte(s), nil)
	if errDB != nil {
		fmt.Println("Cant READ DB :" , errDB.Error() )
		return nil
	}

	cache.Write(conf)

	err := dec.Decode(&record)
	if err != nil {
		fmt.Println("Decode error :", err.Error())
		return nil
	}

	if time.Since(record.Date) > (time.Duration(ZabovCacheTTL) * time.Hour) {
		return nil
	}

	err = ret.Unpack(record.Query)
	if err != nil {
		fmt.Println("Problem unpacking response: ", err.Error())
		return nil
	}

	return ret

}

func domainInCache(domain string) bool {

	has, err := MyZabovCDB.Has([]byte(domain), nil)
	if err != nil {
		fmt.Println("Cannot search Cache DB: ", err.Error())
		return false
	}

	return has

}
