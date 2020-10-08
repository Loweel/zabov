package main

import (
	"fmt"
	"runtime"
	"time"
)

func init() {

	fmt.Println("Garbage Collector Thread Starting")

	go memoryCleanerThread()

}

func memoryCleanerThread() {

	for {
		time.Sleep(10 * time.Minute)
		fmt.Println("Time to clean memory...")
		runtime.GC()
		fmt.Println("Garbage Collection done.")
	}

}
