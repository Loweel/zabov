package main

import (
	"fmt"
	"os"

	"github.com/syndtr/goleveldb/leveldb"
)

//MyZabovKDB is the storage where we'll put domains to block
var MyZabovKDB *leveldb.DB

//MyZabovCDB is the storage where we'll put domains to cache
var MyZabovCDB *leveldb.DB

func init() {

	var err error

	os.RemoveAll("./db")

	os.MkdirAll("./db", 0755)

	MyZabovKDB, err = leveldb.OpenFile("./db/killfile", nil)
	if err != nil {
		fmt.Println("Cannot create Killfile db: ", err.Error())
	} else {
		fmt.Println("Killfile DB created")
	}

	MyZabovCDB, err = leveldb.OpenFile("./db/cache", nil)
	if err != nil {
		fmt.Println("Cannot create Cache db: ", err.Error())
	} else {
		fmt.Println("Cache DB created")
	}

}
