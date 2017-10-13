// Copyright 2017 Fraunhofer Institute for Applied Information Technology FIT

package main


import (
	"encoding/json"
	"io/ioutil"
	"os/exec"
	"time"
	"os"
	"reflect"
	"syscall"
	"io"
	"bufio"
	"strings"
	"log"
	"github.com/satori/go.uuid"
	"math/rand"
)

type Thing struct {
	Name                   string                 `json:"name,omitempty"`
	Description            string                 `json:"description,omitempty"`
	Properties             map[string]interface{} `json:"properties,omitempty"`
	Datastreams []*Datastream `json:"Datastreams,omitempty"`
}

type ObservedProperty struct {
	Name           string        `json:"name,omitempty"`
	Description    string        `json:"description,omitempty"`
	Definition     string        `json:"definition,omitempty"`
}

type Sensor struct {
	Name           string        `json:"name,omitempty"`
	Description    string        `json:"description,omitempty"`
	EncodingType   string        `json:"encodingType,omitempty"`
	Metadata       string        `json:"metadata,omitempty"`
}

type Datastream struct {
	Name                string                 `json:"name,omitempty"`
	Description         string                 `json:"description,omitempty"`
	UnitOfMeasurement   map[string]interface{} `json:"unitOfMeasurement,omitempty"`
	ObservationType     string                 `json:"observationType,omitempty"`
	Sensor              *Sensor                `json:"Sensor,omitempty"`
	ObservedProperty    *ObservedProperty      `json:"ObservedProperty,omitempty"`
}


type DataRequestType string

//
// An envelope data structure for requests of data from services
//
type DataRequest struct {
	ResourceId string
	Type       DataRequestType
	Arguments  []byte
	Reply      chan AgentResponse
}

//
// An envelope data structure for agent's data
//
type AgentResponse struct {
	AgentId string
	Payload    []byte
	IsError    bool
	Cached     time.Time
}


type AgentManager struct {

	things map[string]Thing
	thingFiles map[string]string
	agents map[string]*exec.Cmd
	uuids map[string]uuid.UUID
	agentOutputCache map[string]AgentResponse
	outputFromAgent chan AgentResponse
	quarantine      *Quarantine
	publisher       *Publisher
	subscriber 		*Subscriber
	mConfig         LSTPConfig

}
func newAgentManager() *AgentManager {

	c := loadConfig(CONFIG_DIR)

	manager := &AgentManager{
		agents: make(map[string]*exec.Cmd),
		uuids: make(map[string]uuid.UUID),
		things: make(map[string]Thing),
		thingFiles: make(map[string]string),
		outputFromAgent: make (chan AgentResponse),
		agentOutputCache: make(map[string]AgentResponse),
		quarantine:  newQuarantine(),
		mConfig: c,
		publisher: newPublisher(c),
	}

	manager.subscriber = newSubscriber(manager)

	manager.things,manager.thingFiles,manager.uuids = loadThings()

	return manager
}


func (am *AgentManager) startAgentManager() {

	go am.publisher.startPublisher(am)
	go am.subscriber.startSubscriber()

	for key,value := range am.things{
		log.Println("[AgentManager:startAgentManager] executing existing agents : ",value.Name)
		cmdService,_ := am.executeAgent(value,am.uuids[value.Name])
		am.agents[key] = cmdService
	}
	go am.quarantineListener()
	// start script quarantine in background
	go am.quarantine.startQuarantine()



	for {
		select {
		case resp := <-am.outputFromAgent:
			// forward agent output to publisher
			go func(){
				if t, ok:= am.things[resp.AgentId]; ok{
					if !t.IsEmpty() {
						am.publisher.toPublish<-resp
					}else{
						log.Println("[AgentManager:eventloop] Thing with name ",resp.AgentId, "is empty. Skipping publish.")
					}
				}else{
					log.Println("[AgentManager:eventloop] Thing with name ",resp.AgentId, "doesn't exist. Skipping publish.")
				}
			}()

		}

	}


}
func (am *AgentManager) dropzoneListener() {

	fromDropzone := AgentCandidate{}

	for {
		select {
		case fromDropzone = <-am.quarantine.dropzone.removedAgent:
			log.Println("[AgentManager:dropzoneListener] agent removed : ", fromDropzone.scriptFile)
		}
	}
}
// listens to quarantine for validated agents
func (am *AgentManager) quarantineListener(){
	fromQuarantine := AgentCandidate{}

	for {
		select {
		case fromQuarantine = <-am.quarantine.validatedAgent:
			log.Println("[AgentManager:quarantineListener] agent validated : ",fromQuarantine.thingFile)
			aThing :=newThing(fromQuarantine)
			if !aThing.IsEmpty(){
				_,ok := am.things[aThing.Name]
				if ok {
					log.Println("[AgentManager:quarantineListener] ignoring thing with already existing ID: ",fromQuarantine.thingFile)
					am.removeAgentFiles(fromQuarantine.uuid)
				}else {
					log.Println("[AgentManager:quarantineListener] thing loaded : ",fromQuarantine.thingFile)
					am.things[aThing.Name] = *aThing
					am.thingFiles[aThing.Name] = fromQuarantine.thingFile
					am.uuids[aThing.Name] = fromQuarantine.uuid
					log.Println("[AgentManager:quarantineListener] executing thing : ", fromQuarantine.thingFile)
					cmdService, _ := am.executeAgent(*aThing,am.uuids[aThing.Name])
					am.agents[aThing.Name] = cmdService
				}
			}
		}
	}
}

func (am *AgentManager) stopAgentManager() bool{

	go am.subscriber.startSubscriber()
	go am.publisher.stopPublisher()
	go am.quarantine.stopQuarantine()

	run_counter := len(am.things)
	log.Println("[AgentManager:stopAgentManager] agent counter :",run_counter)
	for _,value := range am.things{
		if (am.stopAgent(value)){
			run_counter--
		}
		log.Println("[AgentManager:stopAgentManager] agent counter :",run_counter)

	}

	if run_counter == 0{
		log.Println("[AgentManager:stopAgentManager] all agents stopped")
		return true
	}
	return false
}
func (am *AgentManager) stopAgent(stopme Thing) bool{

	if am.agents[stopme.Name] != nil {
		log.Println("[AgentManager:stopAgent] stopping process with pid: ", am.agents[stopme.Name].Process.Pid)
		pid := am.agents[stopme.Name].Process.Pid
		err := syscall.Kill(-pid, 15)
		if err == nil {
			log.Println("[AgentManager:stopAgent] ",err.Error())
			// some clean up
		}
		state,err := am.agents[stopme.Name].Process.Wait()
		if state != nil {
			log.Println("[AgentManager:stopAgent] process state      : -->", state.String(), "<--")
			log.Println("[AgentManager:stopAgent] process terminated : -->", strings.ContainsAny(state.String(), "terminated"), "<--")

			if !strings.ContainsAny(state.String(), "terminated") || err != nil {
				log.Fatal("AgentManager:stopAgent] process.Signal on pid %d returned: %v\n", pid, err)
				return false
			}
			return true
		}else{
			log.Println("[AgentManager:stopAgent] process state unknown")
			return true
		}

	}else{
		log.Println("[AgentManager:stopAgent] ignoring stop request for : ", stopme.Name)
		return false
	}







}
func (am* AgentManager) removeAgentFiles(removeme uuid.UUID) bool{
	s,_ := os.Getwd()

	workingdir := s+AGENT_DIR+removeme.String()

	err := os.RemoveAll(workingdir)
	if err!=nil {
		log.Fatal("[AgentManager:removeAgentFiles] ",err.Error())
		return false
	}
	log.Println("[AgentManager:removeAgentFiles] agent files deleted from ",workingdir)

	return true

}
func (am* AgentManager) removeAgent(removeme Thing) bool{

	log.Println("[AgentManager:removeAgent] removing agent: ",removeme.Name)
	// stop agent script
	if (!am.stopAgent(removeme)){
		return false
	}
	// physically remove agent's files
	am.removeAgentFiles(am.uuids[removeme.Name])
	// notify subscribers
	am.publisher.status2Publish<-AgentStatus{false,removeme.Name}

	// delete internal representation
	delete(am.thingFiles,removeme.Name)
	delete(am.uuids,removeme.Name)
	delete(am.agents,removeme.Name)
	delete(am.things,removeme.Name)

	return true
}
func randomInRange(minimum, maximum int) int {
	rand.Seed(time.Now().Unix())
	return rand.Intn(maximum - minimum) + minimum
}

func (am *AgentManager) executeAgent(thing Thing,uuid uuid.UUID)(*exec.Cmd,error) {
	s, _ := os.Getwd()
	filename:= thing.Datastreams[0].Sensor.Metadata
	workingdir := s+AGENT_DIR+uuid.String()+SCRIPT_DIR

	command := []string{"/bin/bash", "-c", workingdir + filename}
	cmd := exec.Command(command[0], command[1:]...)

	cmd.SysProcAttr = &syscall.SysProcAttr{}
	cmd.SysProcAttr.Setsid = true

	cmd.Stderr = os.Stderr

	serviceOutput, err := cmd.StdoutPipe()
	if err!=nil{
		return nil,err
	}
	//defer serviceOutput.Close()

	go func(out io.ReadCloser) {
		log.Println("[AgentManager:executeAgent] executing script : ", workingdir+filename)
		scanner := bufio.NewScanner(out)
		reply := AgentResponse{}
		reply.AgentId = thing.Name
		counter := 0
		for scanner.Scan() {
			reply.Cached = time.Now()
			reply.IsError = false
			reply.Payload = scanner.Bytes()
			am.outputFromAgent <- reply
			counter++
			// TODO a tick / alive functionality can be added
			if counter > (30+randomInRange(0,60)){
				log.Println("[AgentManager:executeAgent] ",reply.AgentId +" is alive")
				counter = 0
			}
		}
		if err = scanner.Err(); err != nil {

			reply.Cached = time.Now()
			reply.IsError = true
			reply.Payload = []byte(err.Error())
			am.outputFromAgent <- reply
			log.Println("[AgentManager:executeAgent] error from script: ", err.Error())
		}
		out.Close()
	}(serviceOutput)
	cmd.Start()
	am.agents[thing.Name] = cmd
	// publish current status of the agent
	am.publisher.status2Publish<-AgentStatus{true,thing.Name}
	return cmd, nil
}

func loadThings() (map[string]Thing,map[string]string,map[string]uuid.UUID){

	workingdir,_ := os.Getwd()
	workingdir = workingdir+AGENT_DIR

	log.Println("[loadThings] working dir: ",workingdir)

	things :=  make(map[string]Thing)
    thingfilemap := make(map[string]string)
	uuidmap := make(map[string]uuid.UUID)

	agentdirectories := scanDirectory(workingdir)

	for _,uuid_dir := range agentdirectories{
		log.Println("[loadThings] scanning agent files with uuid: ",uuid_dir)
		thingfiles := scanDirectory(workingdir+uuid_dir+THING_DIR)
		log.Println("[loadThings] ",thingfiles)
		for _,thingfile := range thingfiles{
			log.Println("[loadThings] trying to load: ",thingfile)
			aThing := loadThing(workingdir+uuid_dir+THING_DIR+thingfile)
			if !aThing.IsEmpty(){
				_,ok := things[aThing.Name]
				if ok {
					log.Println("[loadThings] omitting already existing thing")

				}else{
					things[aThing.Name] = *aThing
					thingfilemap[aThing.Name] = thingfile
					uuidmap[aThing.Name], _ = uuid.FromString(uuid_dir)
				}

			}else{
				log.Println("[loadThings] omitting empty thing")
			}


		}

	}
	return things,thingfilemap,uuidmap
}

// constructs a new thing representation
func newThing(candidate AgentCandidate) *Thing{
	workingdir,_ := os.Getwd()
	thing_file := workingdir+AGENT_DIR+candidate.uuid.String()+THING_DIR+candidate.thingFile
	log.Println("[newThing] new thing file : ",thing_file)
	return loadThing(thing_file)

}
// loads and parses a valid Thing object
// needs an absolute location
func loadThing(thingfile string) *Thing {
	log.Println("[loadThing] Trying to load thing file : ",thingfile)
	var aThing Thing
	content, err := ioutil.ReadFile(thingfile)

	if err !=nil || len(content) == 0 {
		log.Println("[loadThing] Error reading thing file :",err.Error())
		return &Thing{}
	}

	err = json.Unmarshal(content,&aThing)
	if err != nil{
		log.Println("[loadThing] Error unmarshaling data :",err)
		return &Thing{}
	}
	//add more sophisticated validation ?
	log.Println("[loadThing] Parsing thing candidate : ",aThing.Name)
	//log.Println("[loadThing] Datastreams : ",len(aThing.Datastreams))
	if(len(aThing.Datastreams) > 0){
		if(aThing.Datastreams[0].Sensor != nil) {
			if(len(aThing.Datastreams[0].Sensor.Metadata) > 0) {
				log.Println("[loadThing] Returning proper thing. GOOD")
				return &aThing
			}else{
				log.Println("[loadThing] No Metadata field inside JSON . Ignoring thing file : ",thingfile)
			}
		}else{
			log.Println("[loadThing] No Sensor field inside JSON . Ignoring thing file : ",thingfile)
		}
	}else {
		log.Println("[loadThing] No Datastream inside JSON . Ingonring thing file : ",thingfile)
	}
	return &Thing{}

}
func scanDirectory(directory string)([]string){
	var fileNames []string
	//s,_ := os.Getwd()
	files, _:= ioutil.ReadDir(directory)
	for _, f := range files {
		fileNames = append(fileNames,f.Name())
		log.Println("[scanDirectory] found : ",f.Name())
	}
	return fileNames
}

func (s *Thing) IsEmpty() bool {
	return reflect.DeepEqual(s,&Thing{})
}
func (s *Datastream) IsEmpty() bool {
	return reflect.DeepEqual(s,&Datastream{})
}
