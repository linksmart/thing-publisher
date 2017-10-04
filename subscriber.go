package main

import (
	MQTT "github.com/eclipse/paho.mqtt.golang"
	"log"
	"strings"
	"os"
	"github.com/satori/go.uuid"
	"io/ioutil"
)

type Subscriber struct {
	topicmap map[string]byte
	//config LSTPConfig
	subclient MQTT.Client
	stop chan bool
	agentManager *AgentManager
}
func (c *Subscriber) ThingPublisherAPI(client MQTT.Client, msg MQTT.Message) {

	log.Println("[ThingPublisherAPI] Command recieved : ",msg.Topic())
	topic_splited := strings.Split(msg.Topic(),"/")

	switch {
	case (msg.Topic()  == c.agentManager.mConfig.Prefix+c.agentManager.mConfig.AddThingArchiveTOPIC):
		log.Println("[ThingPublisherAPI] Incoming thing archive")
		destination,_ := os.Getwd()
		uuid := uuid.NewV4()
		destination = destination+DROPZONE+uuid.String()
		_ = ioutil.WriteFile(destination, msg.Payload(), os.FileMode(0700))
		log.Println("[ThingPublisherAPI] Thing archive written to :",destination)
	//case msg.Topic() == c.agentManager.mConfig.Prefix+c.agentManager.mConfig.RemoveThingTOPIC:
	case strings.Contains(msg.Topic(),c.agentManager.mConfig.RemoveThingTOPIC):
		name := topic_splited[len(topic_splited)-1]
		log.Println("[ThingPublisherAPI] Trying to remove thing :",name)
		if(!c.agentManager.removeAgent(c.agentManager.things[name])){
			log.Println("[ThingPublisherAPI] No thing with ID : >>",name, "<< exists")
		}
	case msg.Topic() == c.agentManager.mConfig.Prefix+c.agentManager.mConfig.ListThingsTOPIC:
		log.Println("[ThingPublisherAPI] Listing things")
		thingnames := "{\n"
		for _,value := range c.agentManager.things{
			thingnames = thingnames+"\"name\": \""+value.Name+"\",\n"
		}
		thingnames = thingnames+"}"
		topic := c.agentManager.mConfig.Prefix+"things"
		token := client.Publish(topic, 1, false, []byte(thingnames))
		token.Wait()
	case strings.Contains(msg.Topic(),c.agentManager.mConfig.ThingStatusTOPIC):
		name := topic_splited[len(topic_splited)-1]
		log.Println("[ThingPublisherAPI] Reporting status of thing "+name)
		if _,ok := c.agentManager.things[name];ok{
			c.agentManager.publisher.status2Publish<-AgentStatus{true,name}
		}else{
			log.Println("[ThingPublisherAPI] Thing with name >>",name,"<< doesn't exist")
			c.agentManager.publisher.status2Publish<-AgentStatus{false,name}
		}

	default:
		log.Println("[ThingPublisherAPI] unknown topic: ",msg.Topic())

	}
	//fmt.Printf("Topic: %s, Message: %s\n", msg.Topic(), msg.Payload())
}


func newSubscriber(am *AgentManager) *Subscriber {


	opts := MQTT.NewClientOptions().AddBroker(am.mConfig.Broker)


	subscriber := &Subscriber{
		topicmap: make(map[string]byte),
		subclient: MQTT.NewClient(opts),
		stop: make(chan bool),
		agentManager: am,
	}
	subscriber.topicmap[am.mConfig.Prefix+am.mConfig.AddThingArchiveTOPIC] = byte(0)
	subscriber.topicmap[am.mConfig.Prefix+am.mConfig.ListThingsTOPIC] = byte(0)
	subscriber.topicmap[am.mConfig.Prefix+am.mConfig.RemoveThingTOPIC+"/#"] = byte(0)
	subscriber.topicmap[am.mConfig.Prefix+am.mConfig.ThingStatusTOPIC+"/#"] = byte(0)


	return subscriber
}
func (s *Subscriber) startSubscriber(){

	if token := s.subclient.Connect(); token.Wait() && token.Error() != nil {
		log.Fatal(token.Error())
	}else{
		log.Println("[startSubscriber] Subscriber connected to : ",s.agentManager.mConfig.Broker)
	}

	s.subclient.SubscribeMultiple(s.topicmap,s.ThingPublisherAPI)
	log.Println("[startSubscriber] Subscriber started")

}
func (s *Subscriber) stopSubscriber(){
	s.subclient.Disconnect(250)
	log.Println("[startSubscriber] Subscriber stopped")

}
