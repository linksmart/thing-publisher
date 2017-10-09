package main

import (
	"testing"
	"time"
	"log"
	"os"
	"os/exec"
	"path/filepath"
)



//func TestNothingDropzone(t *testing.T){
//	time.Sleep(1)
//	t.Fail()
//	os.Exit(1)
//}

// test the extraction capability of the dropzone
// putting two agent archives into the dropzone.
// test extract data from one and compares it with proper values.
func TestDropzone(t *testing.T){

	cleanup()
	//prepare()
	//defer cleanup()


	dz := newDropzone()
	go dz.startDropzone()

	time.Sleep(time.Second * 2)


	prepareDropzone()
	defer cleanDropzone()
	defer cleanup()

	// receive information about newly detected archive
	// consume agent and parse the data
	log.Println("[TestDropzone] waiting for agent candidates...")
	c:=<-dz.newAgent


	log.Println("[TestDropzone] thing file : ",c.thingFile)
	log.Println("[TestDropzone] script file : ",c.scriptFile)
	log.Println("[TestDropzone] archive file : ",c.archiveFile)

	time.Sleep(time.Second*10)
	log.Println("[TestDropzone] woke up.")


	if c.thingFile!="42.json" || c.scriptFile !="42.sh" || c.archiveFile !="42.tar.gz" {
		log.Panic("[TestDropzone] Wrong content from archive")
		os.Exit(1)
	}
	log.Println("[TestDropzone] unpack test passed.")
	// extraction test passed

	dz.stopDropzone()
	time.Sleep(time.Second * 1)
	log.Println("[TestDropzone] dropzone stopped")
	workingdir,_ := os.Getwd()
	workingdir = workingdir+DROPZONE
	dropzonefiles := scanDirectory(workingdir)
	// dropzone code should delete all archives upon consumption
	// expects 0 archives in the dropzone.
	if (len(dropzonefiles)!=0){
		log.Panic("[TestDropzone] Expecting 0 archives in dropzone. Found ",len(dropzonefiles))
		os.Exit(1)
	}else{
		log.Println("[TestDropzone] Found 0 archives in the dropzone. Good")
	}

	return
}

func prepareDropzone(){
	s,_ := os.Getwd()

	srcFolder := s+ARCHIVES_DIR+"42.tar.gz"
	destFolder := s+DROPZONE
	log.Println("[prepareDropzone] copying archive :",srcFolder, "to dest:",destFolder)
	//cpCmd := exec.Command("cp", "-TRf", srcFolder, destFolder)
	cpCmd := exec.Command("cp", "-f", srcFolder, destFolder)
	err := cpCmd.Run()
	if err!=nil{
		log.Println(err.Error())
		return
	}

}
func prepareBrokenDropzone(){
	s,_ := os.Getwd()

	srcFolder := s+ARCHIVES_DIR+"42broken.tar.gz"
	destFolder := s+DROPZONE
	log.Println("[prepareBrokenDropzone] copying archive :",srcFolder, "to dest:",destFolder)
	//cpCmd := exec.Command("cp", "-TRf", srcFolder, destFolder)
	cpCmd := exec.Command("cp", "-f", srcFolder, destFolder)
	err := cpCmd.Run()
	if err!=nil{
		log.Println(err.Error())
		return
	}
}
func prepareDropzoneWithDoppelgaenger(){
	s,_ := os.Getwd()

	srcFolder := s+ARCHIVES_DIR+"42double.tar.gz"
	destFolder := s+DROPZONE
	log.Println("[prepareDropzoneWithDoppelgaenger] copying archive :",srcFolder, "to dest:",destFolder)
	//cpCmd := exec.Command("cp", "-TRf", srcFolder, destFolder)
	cpCmd := exec.Command("cp", "-f", srcFolder, destFolder)
	err := cpCmd.Run()
	if err!=nil{
		log.Println(err.Error())
		return
	}
}
func cleanDropzone(){

	s,_ := os.Getwd()

	destFolder := s+DROPZONE


	d, _ := os.Open(destFolder)
	defer d.Close()
	names, _ := d.Readdirnames(-1)
	for _, name := range names {
		os.RemoveAll(filepath.Join(destFolder, name))
	}

}
