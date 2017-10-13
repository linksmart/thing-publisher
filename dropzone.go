// Copyright 2017 Fraunhofer Institute for Applied Information Technology FIT

package main

import (
	"github.com/fsnotify/fsnotify"
	"os"
	"log"
	"archive/tar"
	"compress/gzip"
	"io"
	"strings"
	"os/exec"
    "github.com/satori/go.uuid"
)


type AgentCandidate struct {
	thingFile string
	scriptFile string
	archiveFile string
	validated bool
	uuid uuid.UUID
}

type Dropzone struct{
	archivesByScript map[string]AgentCandidate
	newAgent chan AgentCandidate
	removedAgent chan AgentCandidate
	stop chan bool

}

func newDropzone() *Dropzone {
	dropzone := &Dropzone{
		archivesByScript : make(map[string]AgentCandidate),
		newAgent : make (chan AgentCandidate),
		removedAgent : make (chan AgentCandidate),
		stop: make (chan bool),
	}
	return dropzone
}

//TODO Refactoring needed
func (d* Dropzone) untarArchive(mpath string) {
	log.Println("[Dropzone:untarArchive] qualified path to unpack : ",mpath)
	agent := AgentCandidate{"","","",false,uuid.NewV4()}
	unpackHere ,_ := os.Getwd()

	unpackHere = unpackHere+AGENT_DIR+agent.uuid.String()
	log.Println("[Dropzone:untraArchive] will unpack here: ",unpackHere)
	os.Mkdir(unpackHere, os.FileMode(0700))


	archive := strings.Split(mpath,"/")
	log.Println("[Dropzone:untarArchive] will unpack archive : ",archive[len(archive)-1])
	agent.archiveFile = archive[len(archive)-1]
	fr, err := read(mpath)
	defer fr.Close()
	if err != nil {
		panic(err)
	}
	gr, err := gzip.NewReader(fr)
	defer gr.Close()
	if err != nil {
		panic(err)
	}
	tr := tar.NewReader(gr)

	currentdir := ""


	for {
		header, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			panic(err)
		}
		path := header.Name
		log.Println("[Dropzone:untarArchive] header name:",path)
		a := strings.Split(path,"/")
		if(len(a)==3) {
			if (a[1] == "sensors" && a[2] != "") {
				agent.scriptFile=a[2]
			} else if (a[1] == "things" && a[2] != "") {
				agent.thingFile=a[2]
			}
		}
		switch header.Typeflag {
		case tar.TypeDir:

			newdir := strings.Replace(header.Name,"./","",-1)
			newdir = strings.Replace(newdir,"/","",-1)
			currentdir = newdir
			if err := os.MkdirAll(unpackHere+"/"+newdir, os.FileMode(header.Mode)); err != nil {
				log.Panic(err)
			}
		case tar.TypeReg:
			log.Println("[Dropzone:untarArchive] unpacking file : ",header.Name)
			newfile := strings.Split(header.Name,"/")
			outFile, err := os.Create(unpackHere+"/"+currentdir+"/"+newfile[len(newfile)-1])
			if err != nil{
				log.Fatalf("Dropzone:untarArchive] extraction failed: %s",err.Error())
			}
			defer outFile.Close()

			if _, err := io.Copy(outFile, tr); err != nil {
				log.Panic(err)
			}
		default:
			log.Printf("[Dropzone:untarArchive] Unknown header: %c, %s\n", header.Typeflag, path)
		}
	}

	// fill dropzone's internal inventory
	d.archivesByScript[agent.scriptFile] = agent
	setExecutable(unpackHere+SCRIPT_DIR+agent.scriptFile)
	// notify quarantine about a new agent
	go func() {
		d.newAgent <- agent
	}()
	// agent archive consumed, removing the archive file
	d.removeArchive(agent)


}


func read(mpath string) (*os.File, error) {
	f, err := os.OpenFile(mpath, os.O_RDONLY, 0444)
	if err != nil {
		return f, err
	}
	return f, nil
}
func setExecutable(mpath string)(error){
	cmd := exec.Command("chmod","0777",mpath)
	_,err := cmd.Output()
	log.Println("[setExecutable] executable flag set for ",mpath)
	return err
}
func (d* Dropzone) stopDropzone(){
	d.stop<-true
}

func (d *Dropzone) startDropzone(){


	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		log.Println("[Dropzone:startDropzone] Couldn't start dropzone")
		log.Panic(err)
	}
	defer watcher.Close()
	log.Println("[Dropzone:startDropzone] Dropzone started.")

	s,_ := os.Getwd()
	go func() {
		for{
			select{
				case event:= <- watcher.Events:
					if event.Op&fsnotify.Create == fsnotify.Create {
						log.Println("[Dropzone:eventloop] new file in the dropzone :", event.Name)
						go d.untarArchive(event.Name)
					}else if event.Op&fsnotify.Remove == fsnotify.Remove {
						log.Println("[Dropzone:eventloop] removed file:", event.Name)
					}
				case err := <- watcher.Errors:
					if err != nil {
						log.Println("[Dropzone:eventloop] error from watcher : ",err)
					}
			}
		}

	}()
	toWatch := s+"/dropzone"
	log.Println("[Dropzone:startDropzone] watching directory : ",toWatch)
	err = watcher.Add(toWatch)
	if err != nil{
		log.Panic("[Dropzone:startDropzone] Clouldn't add dir to watcher: ",err)
	}
	<-d.stop
	log.Println("[Dropzone:startDropzone] Dropzone stopped.")

}
func (d* Dropzone) removeArchive(ac AgentCandidate) bool{

	s,_ := os.Getwd()

	log.Println("[Dropzone:removeAgentCandidate] deleting :",s+DROPZONE+ac.archiveFile)
	os.RemoveAll(s+DROPZONE+ac.archiveFile)

	delete(d.archivesByScript,ac.scriptFile)

	return true

}