package main

import (
	"testing"
	"os"
	"log"
)

func TestConfig(t *testing.T) {

	aConfig := loadConfig(CONFIG_DIR)

	if (aConfig.Id != "LSTP" || aConfig.Description !="Linksmart ThingPublisher"){
		log.Panic("[TestConfig] Loading config file failed")
		os.Exit(1)
	}

}
//func TestNoConfig(t *testing.T){
//
//
//	time.Sleep(1)
//	t.Fail()
//	os.Exit(1)
//}