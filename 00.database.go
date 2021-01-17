package main

import (
	"fmt"
	"os"

	"github.com/syndtr/goleveldb/leveldb"
)

//MyZabovCDB is the storage where we'll put domains to cache (global for all configs)
var MyZabovCDB *leveldb.DB

//MyZabovKDBs is the storage where we'll put domains to block (one for each config)
var MyZabovKDBs map[string]*leveldb.DB

func init() {

	var err error

	os.RemoveAll("./db")

	os.MkdirAll("./db", 0755)

	MyZabovCDB, err = leveldb.OpenFile("./db/cache", nil)
	if err != nil {
		fmt.Println("Cannot create Cache db: ", err.Error())
	} else {
		fmt.Println("Cache DB created")
	}

	MyZabovKDBs = map[string]*leveldb.DB{}
}

// ZabovCreateKDB creates Kill DBs
func ZabovCreateKDB(conf string) {
	var err error

	dbname := "./db/killfile_" + conf
	KDB, err := leveldb.OpenFile(dbname, nil)
	if err != nil {
		fmt.Println("Cannot create Killfile db: ", err.Error())
	} else {
		fmt.Println("Killfile DB created:", dbname)
	}

	MyZabovKDBs[conf] = KDB

}
