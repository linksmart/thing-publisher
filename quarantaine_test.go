package main

import (
	"testing"
	"log"
	"time"
	"os"
)

//func TestNothing(t *testing.T){
//	log.Println("[TestNothing]")
//	time.Sleep(1)
//	t.Fail()
//	os.Exit(1)
//}

func TestQuarantineWithDropzone(t *testing.T) {

	cleanup()
	expecting := 1

	quarantaine := newQuarantine()
	go quarantaine.startQuarantine()
	defer quarantaine.stopQuarantine()

	time.Sleep(time.Second * 2)

	prepareDropzone()
	defer cleanDropzone()
	defer cleanup()

	timeout := make(chan bool, 1)
	go func() {
		time.Sleep(time.Second * VALIDATE_TIMER*3)
		log.Println("[TestQuarantineWithDropzone] timeout after : ",time.Second * VALIDATE_TIMER*3," seconds")
		timeout <- true
	}()

	validatedAgentCounter := 0
	// wait until quarantine produces
	for {
		if validatedAgentCounter == expecting {
			log.Println("[TestQuarantineWithDropzone] Found total of 1 validated agent. Good")
			break
		}

		select {
		case <-quarantaine.validatedAgent:
			{
				log.Println("[TestQuarantineWithDropzone] validated agent received. Good")
				validatedAgentCounter++
			}
		case <-timeout:
			{
				log.Panic("[TestQuarantineWithDropzone] Timeout detected. Expected ",expecting," validated agents. Found : ",validatedAgentCounter)
				os.Exit(1)
			}

		}
	}

}
func TestQuarantineWithBrokenAgent(t *testing.T) {
	cleanup()

	expecting := 1

	quarantaine := newQuarantine()
	go quarantaine.startQuarantine()
	defer quarantaine.stopQuarantine()

	time.Sleep(time.Second * 2)


	timeout := make(chan bool, 1)
	go func() {
		time.Sleep(time.Second * VALIDATE_TIMER*3)
		log.Println("[TestQuarantineWithDropzone] timeout after : ",time.Second * VALIDATE_TIMER*3," seconds")
		timeout <- true
	}()

	validatedAgentCounter := 0

	prepareBrokenDropzone()
	prepareDropzone()
	defer cleanDropzone()
	defer cleanup()


	// wait until quarantine produces
	for {
		if validatedAgentCounter == expecting {
			log.Println("[TestQuarantineWithBrokenAgent] Found total of 1 validated agents. Good")
			break
		}

		select {
		case <-quarantaine.validatedAgent:
			{
				log.Println("[TestQuarantineWithBrokenAgent] validated agent received. Good")
				validatedAgentCounter++
			}
		case <-timeout:
			{
				log.Panic("[TestQuarantineWithBrokenAgent] Timeout detected. Expected ",expecting," validated agents. Found : ",validatedAgentCounter)
				os.Exit(1)
			}

		}
	}

}
