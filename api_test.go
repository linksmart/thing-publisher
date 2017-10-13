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
			log.Println("[TestAPI:OnListThings] PAYLOAD: ", payload)
			log.Println("[TestAPI:OnListThings] LSTP/listthings API works. GOOD")
		}else {
			log.Panic("[TestAPI:OnListThings] PAYLOAD: ", payload)
			log.Panic("[TestAPI:OnListThings] LSTP/listthings API broken")
			os.Exit(1)

		}


}

func (matrix *TestResultmatrix) OnUploadThing(client MQTT.Client, msg MQTT.Message) {


		payload := string(msg.Payload())
		if strings.Contains(payload,"running"){
			matrix.addthingarchivePassed<-true
			log.Println("[TestAPI:OnThingStatus] PAYLOAD: ", payload)
			log.Println("[TestAPI:OnUploadThing] LSTP/addthingarchive API works. GOOD")
		}else {
			//log.Panic("[OnThingStatus] PAYLOAD: ", payload)
			log.Panic("[TestAPI:OnUploadThing] LSTP/addthingarchive API broken")
			os.Exit(1)

		}

}

func (matrix *TestResultmatrix) OnRemoveThing(client MQTT.Client, msg MQTT.Message) {


		payload := string(msg.Payload())
		if strings.Contains(payload,"not available"){
			matrix.removethingPassed<-true
			log.Println("[TestAPI:OnRemoveThing] PAYLOAD: ", payload)
			log.Println("[TestAPI:OnRemoveThing] LSTP/removething API works. GOOD")
		}else {
			log.Panic("[TestAPI:OnRemoveThing] PAYLOAD: ", payload)
			log.Panic("[TestAPI:OnRemoveThing] LSTP/removething API broken")
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
				log.Fatal("[TestAPI] ",token_connect.Error())
				os.Exit(1)
	}
	defer client.Disconnect(250)
	log.Println("[TestAPI] Connected to : ",manager.mConfig.Broker)

	time.Sleep(time.Second * (VALIDATE_TIMER+5))


	// *********************
	// listthings API test
	client.Subscribe("LSTP/things",0,matrix.OnListThings)
	time.Sleep(time.Second*3)
	_ = client.Publish("LSTP/listthings", 0, false, "")
	select {
	case <- matrix.listthingsPassed:
		log.Println("[TestAPI] (1) listthing API test passed. GOOD")
	case <- time.After(API_TIMEOUT):
		log.Println("[TestAPI] (1) listthing API timeout")
		os.Exit(1)
	}
	client.Unsubscribe("LSTP/things")


	// *********************
	// addarchive API test
	client.Subscribe("LSTP/thing/Temperature",0,matrix.OnUploadThing)

	workdir,_ := os.Getwd()
	thingarchive, _ := os.Open(workdir+ARCHIVES_DIR+"temperature.tar.gz")
	buffer := make([]byte,472)
	_,err := thingarchive.Read(buffer)
	if err == nil {
		log.Println("[TestAPI] temperature.tar.gz archive loaded")
	}else{
		log.Panic("[TestAPI] temperature.tar.gz archive not loaded")
		os.Exit(1)
	}
	_ = client.Publish("LSTP/addthingarchive", 0, false, buffer)
	select {
	case <- matrix.addthingarchivePassed:
		log.Println("[TestAPI] (2) addthingarchive API test passed. GOOD")
	case <- time.After(API_TIMEOUT):
		log.Println("[TestAPI] (2) addthingarchive API timeout")
		os.Exit(1)
	}
	client.Unsubscribe("LSTP/thing/Temperature")

	// *********************
	// removething API test
	client.Subscribe("LSTP/thing/Temperature",0,matrix.OnRemoveThing)
	_ = client.Publish("LSTP/removething/Temperature", 0, false, "")
	select {
	case <- matrix.removethingPassed:
		log.Println("[TestAPI] (3) removething API test passed. GOOD")
	case <- time.After(API_TIMEOUT):
		log.Println("[TestAPI] (3) removething API timeout")
		os.Exit(1)
	}
	client.Unsubscribe("LSTP/thing/TRNG Generator")



}


