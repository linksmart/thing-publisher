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
	subclient MQTT.Client
	stop chan bool
	agentManager *AgentManager
	dropzoneDir string
}
func (c *Subscriber) ThingPublisherAPI(client MQTT.Client, msg MQTT.Message) {

	log.Println("[Subscriber:ThingPublisherAPI] Command recieved : ",msg.Topic())


	switch {
	case (msg.Topic()  == c.agentManager.mConfig.Prefix+c.agentManager.mConfig.AddThingArchiveTOPIC):
		uuid := uuid.NewV4()
		_ = ioutil.WriteFile(c.dropzoneDir + uuid.String(), msg.Payload(), os.FileMode(0700))
		log.Println("[Subscriber:ThingPublisherAPI] Thing archive written to :", c.dropzoneDir + uuid.String())
	case strings.Contains(msg.Topic(),c.agentManager.mConfig.RemoveThingTOPIC):
		topic_splited := strings.Split(msg.Topic(),"/")
		name := topic_splited[len(topic_splited)-1]
		if(!c.agentManager.removeAgent(c.agentManager.things[name])){
			log.Println("[Subscriber:ThingPublisherAPI] No thing with ID : >>",name, "<< exists")
		}
	case msg.Topic() == c.agentManager.mConfig.Prefix+c.agentManager.mConfig.ListThingsTOPIC:
		thingnames := "{\n"
		for _,value := range c.agentManager.things{
			thingnames = thingnames+"\"name\": \""+value.Name+"\",\n"
		}
		thingnames = thingnames+"}"
		topic := c.agentManager.mConfig.Prefix+"things"
		token := client.Publish(topic, 1, false, []byte(thingnames))
		token.Wait()
	case strings.Contains(msg.Topic(),c.agentManager.mConfig.ThingStatusTOPIC):
		topic_splited := strings.Split(msg.Topic(),"/")
		name := topic_splited[len(topic_splited)-1]
		if _,ok := c.agentManager.things[name];ok{
			c.agentManager.publisher.status2Publish<-AgentStatus{true,name}
		}else{
			log.Println("[Subscriber:ThingPublisherAPI] Thing with name >>",name,"<< doesn't exist")
			c.agentManager.publisher.status2Publish<-AgentStatus{false,name}
		}

	default:
		log.Println("[Subscriber:ThingPublisherAPI] unknown topic: ",msg.Topic())
	}
}

func newSubscriber(am *AgentManager) *Subscriber {


	opts := MQTT.NewClientOptions().AddBroker(am.mConfig.Broker)

	s, _ := os.Getwd()

	subscriber := &Subscriber{
		topicmap: make(map[string]byte),
		subclient: MQTT.NewClient(opts),
		stop: make(chan bool),
		agentManager: am,
		dropzoneDir: s+DROPZONE,
	}
	if(am.mConfig.UploadArchive){
		subscriber.topicmap[am.mConfig.Prefix+am.mConfig.AddThingArchiveTOPIC] = byte(0)
		log.Println("[Subscriber:newSubscriber] remote upload of LSTP archives allowed. Modify thing-publisher.json to disable it")
	}
	subscriber.topicmap[am.mConfig.Prefix+am.mConfig.ListThingsTOPIC] = byte(0)
	subscriber.topicmap[am.mConfig.Prefix+am.mConfig.RemoveThingTOPIC+"/#"] = byte(0)
	subscriber.topicmap[am.mConfig.Prefix+am.mConfig.ThingStatusTOPIC+"/#"] = byte(0)

	return subscriber
}
func (s *Subscriber) startSubscriber(){

	if token := s.subclient.Connect(); token.Wait() && token.Error() != nil {
		log.Fatal(token.Error())
	}else{
		log.Println("[Subscriber:startSubscriber] Subscriber connected to : ",s.agentManager.mConfig.Broker)
	}

	s.subclient.SubscribeMultiple(s.topicmap,s.ThingPublisherAPI)
	log.Println("[Subscriber:startSubscriber] Subscriber started")

}
func (s *Subscriber) stopSubscriber(){
	s.subclient.Disconnect(250)
	log.Println("[Subscriber:startSubscriber] Subscriber stopped")

}



