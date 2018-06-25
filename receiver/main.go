package main

import (
	"FileMonitoring/receiver/receiver"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
)

func main() {
	var configFile string
	flag.StringVar(&configFile, "cf", "", "Choose Config File")
	flag.Parse()

	if configFile != "" {
		defer receiver.SaveFilesInfo()

		_, err := os.Stat("filesinfo.json")
		if err != nil {
			receiver.SaveFilesInfo()
		}

		receiver.ReadConfig(configFile)
		go receiver.PostData()

		http.HandleFunc("/checkfile", receiver.HandlerCheckFile)
		http.HandleFunc("/filetransfer", receiver.HandlerFileTransfer)
		port := strconv.Itoa(receiver.Config.Port)
		log.Println(http.ListenAndServe(":"+port, nil))
	} else {
		fmt.Println("Please select config file !")
	}
}
