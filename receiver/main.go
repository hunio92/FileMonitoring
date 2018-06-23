package main

import (
	"log"
	"FileMonitoring/receiver/receiver"
	"flag"
	"net/http"
	"strconv"
)

func main() { 
	var configFile string
	flag.StringVar(&configFile, "cf", "", "Choose Config File")
	flag.Parse()

	receiver.SaveFilesInfo()
	receiver.ReadConfig(configFile)
	
	http.HandleFunc("/checkfile", receiver.HandlerCheckFile)
	http.HandleFunc("/filetransfer", receiver.HandlerFileTransfer)
	port := strconv.Itoa(receiver.Config.Port)
	log.Println(http.ListenAndServe(":"+port, nil))
}
