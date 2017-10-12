package main


import (
"testing"
	"time"
	"log"
	MQTT "github.com/eclipse/paho.mqtt.golang"
	"strings"
	"os"
)

const API_TIMEOUT =30*1e9

func TestNoAPI(t *testing.T){


	time.Sleep(1)

}
type TestResultmatrix struct {
	listthingsPassed chan bool
	thingstatusPassed chan bool
	removethingPassed chan bool
	addthingarchivePassed chan bool
}
func (matrix *TestResultmatrix) OnListThings(client MQTT.Client, msg MQTT.Message) {

		payload := string(msg.Payload())
		if strings.Contains(payload,"TRNG Generator") && strings.Contains(payload,"42"){
			matrix.listthingsPassed<-true
			log.Println("[OnListThings] PAYLOAD: ", payload)
			log.Println("[OnListThings] LSTP/listthings API works. GOOD")
		}else {
			log.Panic("[OnListThings] PAYLOAD: ", payload)
			log.Panic("[OnListThings] LSTP/listthings API broken")
			os.Exit(1)

		}


}

func (matrix *TestResultmatrix) OnUploadThing(client MQTT.Client, msg MQTT.Message) {


		payload := string(msg.Payload())
		if strings.Contains(payload,"running"){
			matrix.addthingarchivePassed<-true
			log.Println("[OnThingStatus] PAYLOAD: ", payload)
			log.Println("[OnUploadThing] LSTP/addthingarchive API works. GOOD")
		}else {
			//log.Panic("[OnThingStatus] PAYLOAD: ", payload)
			log.Panic("[OnUploadThing] LSTP/addthingarchive API broken")
			os.Exit(1)

		}

}

//func (matrix *TestResultmatrix) OnThingStatus(client MQTT.Client, msg MQTT.Message) {
//
//
//	//topic:= msg.Topic()
//		payload := string(msg.Payload())
//		if strings.Contains(payload,"running"){
//			matrix.thingstatusPassed<-true
//			log.Println("[OnThingStatus] PAYLOAD: ", payload)
//			log.Println("[OnThingStatus] LSTP/thingstatus API works. GOOD")
//		}else {
//			log.Panic("[OnThingStatus] PAYLOAD: ", payload)
//			log.Panic("[OnThingStatus] LSTP/thingstatus API broken")
//			os.Exit(1)
//
//		}
//
//}
func (matrix *TestResultmatrix) OnRemoveThing(client MQTT.Client, msg MQTT.Message) {


		payload := string(msg.Payload())
		if strings.Contains(payload,"not available"){
			matrix.removethingPassed<-true
			log.Println("[OnRemoveThing] PAYLOAD: ", payload)
			log.Println("[OnRemoveThing] LSTP/removething API works. GOOD")
		}else {
			log.Panic("[OnRemoveThing] PAYLOAD: ", payload)
			log.Panic("[OnRemoveThing] LSTP/removething API broken")
			os.Exit(1)

		}

}


func TestAPI(t *testing.T){

	cleanup()
	prepare()
	defer cleanup()


	matrix := TestResultmatrix{
		listthingsPassed: make (chan bool),
		thingstatusPassed:make (chan bool),
		removethingPassed:make (chan bool),
		addthingarchivePassed:make (chan bool),
	}


	manager := newAgentManager()
	go manager.startAgentManager()
	defer manager.stopAgentManager()

	opts := MQTT.NewClientOptions().AddBroker(manager.mConfig.Broker)
	//
	client := MQTT.NewClient(opts)
	//
	if token_connect := client.Connect(); token_connect.Wait() && token_connect.Error() != nil {
				log.Fatal(token_connect.Error())
	}
	defer client.Disconnect(250)

	time.Sleep(time.Second * VALIDATE_TIMER)

	client.Subscribe("LSTP/things",0,matrix.OnListThings)
	time.Sleep(time.Second*1)
	_ = client.Publish("LSTP/listthings", 0, false, "")
	select {
	case <- matrix.listthingsPassed:
		log.Println("(1) listthing API test passed. GOOD")
	case <- time.After(API_TIMEOUT):
		log.Println("(1) listthing API timeout")
		os.Exit(1)
	}


	// addarchive API test
	client.Subscribe("LSTP/thing/Temperature",0,matrix.OnUploadThing)

	workdir,_ := os.Getwd()
	thingarchive, _ := os.Open(workdir+ARCHIVES_DIR+"temperature.tar.gz")
	buffer := make([]byte,472)
	_,err := thingarchive.Read(buffer)
	if err == nil {
		log.Println("[TestUploadThing] temperature.tar.gz archive loaded")
	}else{
		log.Panic("[TestUploadThing] temperature.tar.gz archive not loaded")
		os.Exit(1)
	}
	_ = client.Publish("LSTP/addthingarchive", 0, false, buffer)
	select {
	case <- matrix.addthingarchivePassed:
		log.Println("(2) addthingarchive API test passed. GOOD")
	case <- time.After(API_TIMEOUT):
		log.Println("(2) addthingarchive API timeout")
		os.Exit(1)
	}

	// removething API test
	client.Subscribe("LSTP/thing/TRNG Generator",0,matrix.OnRemoveThing)
	_ = client.Publish("LSTP/removething/TRNG Generator", 0, false, "")
	select {
	case <- matrix.removethingPassed:
		log.Println("(2) removething API test passed. GOOD")
	case <- time.After(API_TIMEOUT):
		log.Println("(2) removething API timeout")
		os.Exit(1)
	}




}


