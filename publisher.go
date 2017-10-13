// Copyright 2017 Fraunhofer Institute for Applied Information Technology FIT

package main

import (
	MQTT "github.com/eclipse/paho.mqtt.golang"
	"log"
	"time"
	"encoding/json"
)

type SensorThingPayload struct {
	Result     string                 `json:"result,omitempty"`
	Time       string                 `json:"phenomenonTime,omitempty"`
}
type SensorThingTopic struct {
	SensorID string
	AreaID string
	SensorName string
}

type Publisher struct {
	toPublish chan AgentResponse
	status2Publish chan AgentStatus
	brokerUrl string
	id string
	stop chan bool
	manager *AgentManager
}
type AgentStatus struct {
	status bool
	agentName string
}


func newPublisher(aConfig LSTPConfig) *Publisher {


	publisher := &Publisher{
		brokerUrl: aConfig.Broker,
		toPublish: make (chan AgentResponse),
		status2Publish: make(chan AgentStatus),
		id: aConfig.Id,
		stop: make (chan bool),
	}
	return publisher
}
func (p* Publisher) stopPublisher(){
	p.stop<-true
}
func (p *Publisher) startPublisher(am *AgentManager){

	p.manager = am

	var f MQTT.MessageHandler = func(client MQTT.Client, msg MQTT.Message) {
		log.Print("[Publisher:MessageHandler] TOPIC: %s\n", msg.Topic())
		log.Print("[Publisher:MessageHandler] MSG: %s\n", msg.Payload())
	}



	opts := MQTT.NewClientOptions()
	log.Println("[Publisher:startPublisher] Using MQTT broker: ",p.brokerUrl)
	opts.AddBroker(p.brokerUrl)
	opts.SetClientID(p.id)
	opts.SetDefaultPublishHandler(f)

	client := MQTT.NewClient(opts)
	if token := client.Connect(); token.Wait() && token.Error() != nil {
		log.Panic(token.Error())
	}
	defer client.Disconnect(250)
	log.Println("[Publisher:startPublisher] Connected to : ",p.brokerUrl)



	go func() {
		log.Println("[Publisher:startPublisher] Payload Publisher started.")
		for {
			data:=<-p.toPublish
			// create payload/topic and publish
			go func(){
				payload := SensorThingPayload{string(data.Payload),time.Now().UTC().Format(time.RFC3339)}
				payloadJSON, err := json.Marshal(payload)
				if err != nil {
					log.Println("[Publisher:payloadloop] Error: %s", err)
					return;
				}
				topic := p.manager.mConfig.Prefix+"Datastreams(" + data.AgentId + ")/" + p.manager.things[data.AgentId].Datastreams[0].Sensor.Description
				token := client.Publish(topic, 1, false, payloadJSON)
				token.Wait()
			}()
		}
	}()
	go func() {
		log.Println("[Publisher:startPublisher] Status Publisher started.")
		for {
			status:=<-p.status2Publish
			// create payload/topic and publish
			go func(){
				payload := ""
				if status.status{
					payload = "{ \"status\" : \"running\" }"
				}else{
					payload ="{ \"status\" : \"not available\" }"
				}
				topic := p.manager.mConfig.Prefix+"thing/"+status.agentName
				token := client.Publish(topic, 1, false, payload)
				token.Wait()
			}()
		}
	}()
	<-p.stop
	log.Println("[Publisher:startPublisher] Publisher stopped.")


}
