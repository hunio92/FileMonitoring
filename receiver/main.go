package main

import (
	"FileMonitoring/receiver/receiver"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"
)

// ToDo: 1. BAD FILE TRANSFER: last mofidied not match ! //****// 2. send file by extension

func main() {
	var configFile string
	flag.StringVar(&configFile, "cf", "", "Enter Config File")
	flag.Parse()

	if configFile != "" {
		recName := receiver.ReadConfig(configFile)

		c := waitForSignal()
		go func() {
			<-c // If get anything
			receiver.SaveFilesInfo(recName)
			os.Exit(3)
		}()

		_, err := os.Stat("FilesInfo/" + recName + ".json")
		if err != nil {
			receiver.SaveFilesInfo(recName)
		}

		go receiver.PostData()

		http.HandleFunc("/checkfile", receiver.HandlerCheckFile)
		http.HandleFunc("/filetransfer", receiver.HandlerFileTransfer)
		port := strconv.Itoa(receiver.GetConfig().Port)
		log.Println(http.ListenAndServe(":"+port, nil))

	} else {
		fmt.Println("Please select config file !")
	}
}

func waitForSignal() chan os.Signal {
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, os.Kill)
	return c
}
