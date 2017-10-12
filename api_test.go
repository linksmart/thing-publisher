package main


import (
"testing"
	"time"
	"log"
	MQTT "github.com/eclipse/paho.mqtt.golang"
	"strings"
	"os"
)

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

	topic:= msg.Topic()
	if topic == "LSTP/things"{
		payload := string(msg.Payload())
		if strings.Contains(payload,"TRNG Generator") && strings.Contains(payload,"42"){
			matrix.listthingsPassed<-true
			log.Println("[OnListThings] LSTP/listthings API works. GOOD")
		}else {
			log.Panic("[OnListThings] PAYLOAD: ", payload)
			log.Panic("[OnListThings] LSTP/listthings API broken")
			os.Exit(1)

		}
	}else if topic == "LSTP/listthings"{

	}


}

func (matrix *TestResultmatrix) OnUploadThing(client MQTT.Client, msg MQTT.Message) {


	topic:= msg.Topic()
	if topic == "LSTP/thing/Temperature"{
		payload := string(msg.Payload())
		if strings.Contains(payload,"running"){
			matrix.addthingarchivePassed<-true
			log.Println("[OnUploadThing] LSTP/addthingarchive API works. GOOD")
		}else {
			//log.Panic("[OnThingStatus] PAYLOAD: ", payload)
			log.Panic("[OnUploadThing] LSTP/addthingarchive API broken")
			os.Exit(1)

		}
	}
}

func (matrix *TestResultmatrix) OnThingStatus(client MQTT.Client, msg MQTT.Message) {


	topic:= msg.Topic()
	if topic == "LSTP/thing/42"{
		payload := string(msg.Payload())
		if strings.Contains(payload,"running"){
			matrix.thingstatusPassed<-true
			log.Println("[OnThingStatus] LSTP/thingstatus API works. GOOD")
		}else {
			log.Panic("[OnThingStatus] PAYLOAD: ", payload)
			log.Panic("[OnThingStatus] LSTP/thingstatus API broken")
			os.Exit(1)

		}
	}

}
func (matrix *TestResultmatrix) OnRemoveThing(client MQTT.Client, msg MQTT.Message) {


	topic:= msg.Topic()
	if topic == "LSTP/thing/TRNG Generator"{
		payload := string(msg.Payload())
		if strings.Contains(payload,"not available"){
			matrix.removethingPassed<-true
			log.Println("[OnRemoveThing] LSTP/removething API works. GOOD")
		}else {
			log.Panic("[OnRemoveThing] PAYLOAD: ", payload)
			log.Panic("[OnRemoveThing] LSTP/removething API broken")
			os.Exit(1)

		}
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
	time.Sleep(time.Second * 5)

	opts := MQTT.NewClientOptions().AddBroker(manager.mConfig.Broker)
	//
	client := MQTT.NewClient(opts)
	//
	if token_connect := client.Connect(); token_connect.Wait() && token_connect.Error() != nil {
				log.Fatal(token_connect.Error())
	}
	defer client.Disconnect(250)

	// listthings API test
	client.Subscribe("LSTP/things",0,matrix.OnListThings)
	_ = client.Publish("LSTP/listthings", 1, false, "")
	select {
			case <- matrix.listthingsPassed:
				log.Println("(1) listthing API test passed. GOOD")
			case <- time.After(10*1e9):
				log.Println("(1) listthing API timeout")
				os.Exit(1)
	}

	// thingstatus API test
	client.Subscribe("LSTP/thing/42",0,matrix.OnThingStatus)
	_ = client.Publish("LSTP/thingstatus/42", 1, false, "")
	select {
		case <- matrix.thingstatusPassed:
			log.Println("(2) thingstatus API test passed. GOOD")
		case <- time.After(10*1e9):
			log.Println("(2) thingstatus API timeout")
			os.Exit(1)
	}

	// removething API test
	client.Subscribe("LSTP/thing/TRNG Generator",0,matrix.OnRemoveThing)
	_ = client.Publish("LSTP/removething/TRNG Generator", 0, false, "")
	select {
	case <- matrix.removethingPassed:
		log.Println("(3) removething API test passed. GOOD")
	case <- time.After(10*1e9):
		log.Println("(3) removething API timeout")
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
	_ = client.Publish("LSTP/addthingarchive", 1, false, buffer)
	select {
		case <- matrix.addthingarchivePassed:
			log.Println("(4) addthingarchive API test passed. GOOD")
		case <- time.After(30*1e9):
			log.Println("(4) addthingarchive API timeout")
			os.Exit(1)
	}

}


