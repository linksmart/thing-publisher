// Copyright 2017 Fraunhofer Institute for Applied Information Technology FIT

package main

import (
	"io/ioutil"
	"encoding/json"
	"log"
	"os"
)


const SCRIPT_DIR = "/sensors/"
const THING_DIR = "/things/"
const CONFIG_DIR = "/conf/"
const DROPZONE = "/dropzone/"
const AGENT_DIR = "/agents/"
const EXAMPLES_DIR = "/TEST/agents/"
const TEST_DIR ="/TEST/"
const ARCHIVES_DIR ="/TEST/agentarchives/"
const VALIDATE_TIMER = 10
const CONFIG_FILE ="thing-publisher.json"


//
// Main configuration container
//
type LSTPConfig struct {
	Id             string                       `json:"id"`
	Description    string                       `json:"description"`
	Broker         string                       `json:"broker"`
	Prefix         string 						`json:"prefix"`
	ValidateTimer  int 							`json:"validatetimer"`
	UUIDGeneration bool							`json:"uuidgeneration"`
	UploadArchive bool 							`json:"uploadarchive"`
	AddThingArchiveTOPIC string 				`json:"addthingarchive-topic"`
	ListThingsTOPIC string 						`json:"listthings-topic"`
	RemoveThingTOPIC string 					`json:"removething-topic"`
	ThingStatusTOPIC string 					`json:"thingstatus-topic"`
}

func loadConfig(confPath string) LSTPConfig {

	s, _ := os.Getwd()
	log.Println("[loadConfig] Using config file: ",s+confPath+CONFIG_FILE)
	content, err := ioutil.ReadFile(s+confPath+CONFIG_FILE)
	if err != nil {
		return LSTPConfig{}
	}
	var aConfig LSTPConfig
	err = json.Unmarshal(content,&aConfig)
	if err != nil{
		println("Cannot unmarshal json")
		return LSTPConfig{}
	}
	log.Println("[loadConfig] ID                           : ",aConfig.Id)
	log.Println("[loadConfig] Description                  : ",aConfig.Description)
	log.Println("[loadConfig] MQTT Broker URL              : ",aConfig.Broker)
	log.Println("[loadConfig] Prefix                       : ",aConfig.Prefix)
	log.Println("[loadConfig] Validate timer               : ",aConfig.ValidateTimer)
	log.Println("[loadConfig] UUID generation              : ",aConfig.UUIDGeneration)
	log.Println("[loadConfig] Archive upload allowed       : ",aConfig.UploadArchive)
	log.Println("[loadConfig] Add Thing Archive (MQTT-API) : ",aConfig.AddThingArchiveTOPIC)
	log.Println("[loadConfig] List Things       (MQTT-API) : ",aConfig.ListThingsTOPIC)
	log.Println("[loadConfig] Remove Thing      (MQTT-API) : ",aConfig.RemoveThingTOPIC)
	log.Println("[loadConfig] Thing Status      (MQTT-API) : ",aConfig.ThingStatusTOPIC)

	return aConfig

}
