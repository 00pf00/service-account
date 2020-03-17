package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"time"
)

func main() {
	logfile := flag.String("log","/home/serviceaccount.log","")
	flag.Parse()
	file,err := os.OpenFile(*logfile,os.O_APPEND|os.O_CREATE,666);
	if err != nil {
		os.Exit(1)
	}
	logger := log.New(file,time.Now().String(),log.Ldate|log.Ltime|log.Lshortfile);
	fmt.Printf(logger)

}
