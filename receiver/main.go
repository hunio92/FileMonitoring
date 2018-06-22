package main

import (
	"FileMonitoring/receiver/receiver"
	"flag"
	"net/http"
)

func main() {
	var configFile string
	flag.StringVar(&configFile, "cf", "", "Choose Config File")
	flag.Parse()

	receiver.SaveFilesInfo()
	receiver.ReadConfig(configFile)
	http.HandleFunc("/checkfile", receiver.HandlerCheckFile)
	http.HandleFunc("/filetransfer", receiver.HandlerFileTransfer)
	http.ListenAndServe(":8888", nil)
}
