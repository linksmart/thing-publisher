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

	c := make(chan os.Signal, 2)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-c
		fmt.Println("C-CTRL detected")
		os.Exit(1)
	}()


	am := newAgentManager()
	go am.startAgentManager()
	defer am.stopAgentManager()





	for {
			//fmt.Println("LSTP  alive...")
			time.Sleep(10 * time.Second)
	}

}


