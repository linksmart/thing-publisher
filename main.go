// Copyright 2017 Fraunhofer Institute for Applied Information Technology FIT

package main


import (
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"
	"log"
)

var Version = "No Version Provided"

func main() {
	log.Println("****************************************")
	log.Println("Linksmart ThingPublisher v", Version)
	log.Println("****************************************")
	fmt.Println("Press ctrl-C to exit")
	
	am := newAgentManager()
	go am.startAgentManager()
	defer am.stopAgentManager()
	
	c := make(chan os.Signal, 2)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	<-c
	fmt.Println("Stopping")
}


