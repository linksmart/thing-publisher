// Copyright 2017 Fraunhofer Institute for Applied Information Technology FIT

package main

import (
	"os/exec"
	"syscall"
	"io"
	"bufio"
	"time"
	"log"
	"os"
)

type Quarantine struct{
	agents map[string]AgentCandidate
	stop chan bool
	dropzone* Dropzone
	validatedAgent chan AgentCandidate
}

func newQuarantine() *Quarantine {
	quarantine := &Quarantine{
		agents : make(map[string]AgentCandidate),
	}
	quarantine.stop = make (chan bool)
	quarantine.validatedAgent = make (chan AgentCandidate)
	return quarantine
}
func (q *Quarantine) startQuarantine(){
	q.dropzone = newDropzone()
	go q.dropzone.startDropzone()
	go func() {
		for {
			log.Println("[startQuarantine] Quarantine started.")
			log.Println("[startQuarantine] waiting for new quarantine candidates...")
			agentCandidate:= <-q.dropzone.newAgent
			log.Println("[startQuarantine] agent candidate recieved : ",agentCandidate.scriptFile)
			q.agents[agentCandidate.scriptFile] = agentCandidate
			log.Println("[startQuarantine] Number of quarantained agents: ",len(q.agents))
			go q.validateAgent(agentCandidate)
		}
	}()
	<-q.stop
	log.Println("[startQuarantie] Quarantine stopped.")
}

func (q *Quarantine) stopQuarantine(){
	go q.dropzone.stopDropzone()
	q.stop<-true
}

func (q *Quarantine) validateAgent(qa AgentCandidate){


	script := qa.scriptFile
	workingdir,_ := os.Getwd()
	workingdir = workingdir+AGENT_DIR+qa.uuid.String()

	log.Println("[validateAgent] validating agent: ",workingdir+SCRIPT_DIR+script)
	command := []string{"/bin/bash", "-c",workingdir+SCRIPT_DIR+script}
	cmd := exec.Command(command[0],command[1:]...)

	cmd.SysProcAttr = &syscall.SysProcAttr{}
	cmd.SysProcAttr.Setsid = true

	serviceOutput, err := cmd.StdoutPipe()
	defer serviceOutput.Close()

	if err != nil {
		return
	}
	counter := 0
	log.Println("[validateAgent] entering output scanning routine...")
	go func(out io.ReadCloser) {
		log.Println("[validateAgent] executing script : ",workingdir+SCRIPT_DIR+script)
		scanner := bufio.NewScanner(out)
		for scanner.Scan() {
			log.Println("[validateAgent] stdout from script: ",string(scanner.Bytes()))
			counter++
			if counter > 1{
				log.Println("[validateAgent] script: ",script," validated. Exiting validation loop")
				return
			}
		}
		if err = scanner.Err(); err != nil {
			log.Println("[validateAgent] error from script: ",err.Error())
		}else{
			log.Println("[validateAgent] no output from script. ")
		}
		out.Close()
	}(serviceOutput)
	// high load leads to missing output from the scripts . sleep introduced as workaround
	time.Sleep(time.Millisecond*100)
	cmd.Start()
	log.Println("[validateAgent] quarantained script executed. Waiting...")

	time.Sleep(VALIDATE_TIMER*time.Second)
	log.Println("[validateAgent] Behaviour analysis of the script...")

	if counter > 1 {
		// inform reciever about a validated agent
		log.Println("[validateAgent]  counter > 1, agent validated, :", q.agents[script].scriptFile)
		// notify agent manager about a validated agent
		q.validatedAgent<-qa
	}else{
		log.Println("[validateAgent] counter != 1, agent in-valid, :", q.agents[script].scriptFile)
		if counter == 1{
			log.Println("[validateAgent] Found one output line. Possible task type script detected")
		}else if counter == 0{
			log.Println("[validateAgent]  no output from candidate script")
		}
	}
	log.Println("[validateAgent] sending SIGTERM to script: ",script)
	group, _:= os.FindProcess(-1 * cmd.Process.Pid)
	group.Signal(syscall.SIGTERM)
	if cmd.Process == nil {
		return
	}
	log.Println("[validateAgent] sending SIGKILL to script: ",script)
	group, _ = os.FindProcess(-1 * cmd.Process.Pid)
	group.Signal(syscall.SIGKILL)

}

