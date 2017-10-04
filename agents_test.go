package main


import (
	"testing"
	"time"
	"os/exec"
	"os"
	"log"
	"path/filepath"
	//"strconv"
	"strconv"
)
// a preparation routine copy example agents for a test
func prepare(){
	s,_ := os.Getwd()

	srcFolder := s+EXAMPLES_DIR
	destFolder := s+AGENT_DIR
	log.Println("[prepare] copying src:",srcFolder, "to dest:",destFolder)
	cpCmd := exec.Command("cp", "-TRf", srcFolder, destFolder)
	err := cpCmd.Run()
	if err!=nil{
		println(err.Error())
		return
	}

}
func prepareBrokenThings(){
		s,_ := os.Getwd()

		srcFolder := s+TEST_DIR+"broken_agents/"
		destFolder := s+AGENT_DIR
		log.Println("[prepare] copying src:",srcFolder, "to dest:",destFolder)
		cpCmd := exec.Command("cp", "-TRf", srcFolder, destFolder)
		err := cpCmd.Run()
		if err!=nil{
			println(err.Error())
			return
		}
}
// a cleanup routine deletes example agents from script and thing directories
func cleanup(){

	s,_ := os.Getwd()
	destFolder := s+AGENT_DIR

	d, _ := os.Open(destFolder)
	defer d.Close()
	names, _ := d.Readdirnames(-1)
	for _, name := range names {
		os.RemoveAll(filepath.Join(destFolder, name))
		log.Println("[cleanup] cleaning up ",destFolder+name)
	}
}
//// a newly created agent manager tries to load and validate things
func TestThingLoadingAndValidation(t *testing.T) {


	prepare()
	defer cleanup()

	expecting := 2

	manager := newAgentManager()

	if len(manager.things) != expecting{
		log.Panic("[TestThingLoadingAndValidation] Propper number for loaded and validated things should be ",expecting,". Found:",len(manager.things))
	}

}
func TestAgentManagerShutdown(t *testing.T) {

	prepare()
	defer cleanup()

	manager := newAgentManager()
	go manager.startAgentManager()
	time.Sleep(time.Second * 3)
	if (!manager.stopAgentManager()) {
		log.Panic("[TestAgentManagerShutdown] Agents did't stopped properly")
	}
}


// 2 agent archives are used in the test. 1 is broken_agents. Expecting 1 to run.
func TestAgentManagerWithDropzone(t *testing.T) {

	cleanup()

	expecting := 1

	manager := newAgentManager()
	go manager.startAgentManager()
	defer manager.stopAgentManager()
	time.Sleep(time.Second * 2)

	// adds proper agent archives
	prepareDropzone()
	// add broken_agents agent archive
	prepareBrokenDropzone()
	defer cleanDropzone()
	defer cleanup()

	time.Sleep(time.Second * 15)
	if(len(manager.agents)!=expecting){
		log.Panic("Expecting ",expecting," running agent. Found ",len(manager.agents))
	}else {
		log.Println("Found ",len(manager.agents)," running agents. Good")
	}

}
//// 2 agent archives are used in the test. 1 is a doppelganger of the first one. Expecting 1 to run.
func TestAgentManagerWithDoppelgaengerAgent(t *testing.T) {

	cleanup()

	expecting := 1

	manager := newAgentManager()
	go manager.startAgentManager()
	defer manager.stopAgentManager()

	time.Sleep(time.Second * 2)

	// adds proper agent archives
	prepareDropzone()
	time.Sleep(time.Second * 1)
	// add doppelgaenger agent archive
	prepareDropzoneWithDoppelgaenger()
	defer cleanDropzone()
	defer cleanup()

	time.Sleep(time.Second * 15)
	if(len(manager.agents)!=expecting){
		log.Panic("Expecting ",expecting," running agent. Found ",len(manager.agents))
	}else {
		log.Println("Found ",len(manager.agents)," running agents. Good")
	}

}

//// tests if proper data from a certain agents arrives
func TestAgentOutput(t *testing.T) {

	cleanup()

	prepare()
	defer cleanup()

	manager := newAgentManager()
	go manager.startAgentManager()
	defer manager.stopAgentManager()

	time.Sleep(time.Second * 2)

	for i:=0 ; i < 5; i++{
		log.Println("[TestAgentExecution] waiting for agent data from stdout...")
		data:=<-manager.outputFromAgent
		log.Println("[TestAgentExecution] output from script ",data.AgentId," : ",string(data.Payload))
		if data.AgentId=="Temperature"{
			if string(data.Payload) != "20.0"{
				log.Panic("[TestAgentExecution] expecting 20.0 from Temperature agent")
			}
		}
		if data.AgentId=="TRNG Generator"{
			_, err := strconv.Atoi(string(data.Payload))
			if err != nil{
				log.Panic("[TestAgentExecution] expecting integer from TRNG Generator")
			}
		}

	}

}
func TestRemoveAgent(t *testing.T){

	cleanup()

	expecting := 0
	prepare()
	defer cleanup()

	workingdir,_ := os.Getwd()
	agentdir := workingdir+AGENT_DIR
	//	workingdir = workingdir+AGENT_DIR+"db554710-a2be-11e7-8506-c7a676b6fa69"

	manager := newAgentManager()
	go manager.startAgentManager()
	defer manager.stopAgentManager()
	time.Sleep(time.Second * 2)


	thingToRemove := manager.things["42"]
	manager.removeAgent(thingToRemove)

	thingToRemove = manager.things["TRNG Generator"]
	manager.removeAgent(thingToRemove)


	if len(manager.things) != expecting{
		log.Panic("[TestRemoveAgent] thing list should contain ",expecting," entries. Found ",len(manager.things))
	}

	if len(manager.uuids) != expecting{
		log.Panic("[TestRemoveAgent] uuid list should contain ",expecting," entries. Found ",len(manager.uuids))
	}

	if len(manager.thingFiles) != expecting{
		log.Panic("[TestRemoveAgent] thing file list should contain ",expecting," entries. Found ",len(manager.thingFiles))
	}

	if len(manager.agents) != expecting{
		log.Panic("[TestRemoveAgent] agent list should contain ",expecting," entries. Found ",len(manager.agents))
	}

	entries := scanDirectory(agentdir)
	if len(entries) != expecting{
		log.Panic("[TestRemoveAgent] directory should contain ",expecting," entries. Found ",len(entries))
	}
}
func TestLoadThing(t *testing.T){

	expecting := "TRNG Generator"

	prepare()
	defer cleanup()

	workingdir,_ := os.Getwd()

	workingdir = workingdir+AGENT_DIR+"de1bf1ba-a2be-11e7-8672-532f56551d51"+"/things/"

	thing := loadThing(workingdir+"thing2.json")
	if thing.Name != expecting{
		log.Panic("[TestLoadThing] thing2.json not loaded")
	}


}
// Don't load things with same ID (Name)
func TestLoadDoubleThings(t *testing.T){

	expecting := 1

	prepareBrokenThings()
	defer cleanup()

	things,thingfiles,_ := loadThings()
	log.Println("[TestLoadValidThings] things found: ",len(things))
	if (len(things) != expecting) || (len(thingfiles) != expecting) {
			log.Panic("[TestLoadValidThings] things not loaded. Expecting ",expecting," valid things")
	}



}
func TestLoadEmptyThing_1(t *testing.T){

	prepareBrokenThings()
	defer cleanup()

	workingdir,_ := os.Getwd()
	//
	workingdir = workingdir+AGENT_DIR+"4de88642-a2c5-11e7-82ee-c7c3818f111e"+"/things/"


	thing := loadThing(workingdir+"empty.json")
	if !thing.IsEmpty(){
		log.Panic("[TestLoadErrornousThing] Expected empty thing. ")

	}
}
func TestLoadEmptyThing_2(t *testing.T){

	prepareBrokenThings()
	defer cleanup()

	workingdir,_ := os.Getwd()
	//
	workingdir = workingdir+AGENT_DIR+"7bbe7cd2-a2c7-11e7-b00a-3b376cf8f6c6"+"/things/"


	thing := loadThing(workingdir+"empty2.json")
	if !thing.IsEmpty(){
		log.Panic("[TestLoadErrornousThing] Expected empty thing. ")

	}
}
func TestLoadWrongNameThing(t *testing.T){

	prepareBrokenThings()
	defer cleanup()

	workingdir,_ := os.Getwd()
	//
	workingdir = workingdir+AGENT_DIR+"4de88642-a2c5-11e7-82ee-c7c3818f111e"+"/things/"

	thing := loadThing(workingdir+"notexisting.json")
	if !thing.IsEmpty(){
		log.Panic("[TestLoadWrongNameThing] Expected empty thing. ")

	}
}
//////// test validation of thing json files
//////// four json files are valid things
func TestLoadValidThings(t *testing.T){

	expecting := 2

	prepare()
	defer cleanup()

	things,thingfiles,_ := loadThings()
	log.Println("[TestLoadValidThings] things found: ",len(things))
	if (len(things) != expecting) || (len(thingfiles) != expecting) {
		log.Panic("[TestLoadValidThings] things not loaded. Expecting ",expecting," valid things")
	}

}
func TestScanDirectory(t *testing.T){


	expecting := 2

	prepare()
	defer cleanup()

	workingdir,_ := os.Getwd()
	//
	workingdir = workingdir+AGENT_DIR

	names := scanDirectory(workingdir)
	if len(names) != expecting{
		log.Panic("[TestScanDirectory] directory should contain ",expecting," json files. Found ",len(names))
	}
}



