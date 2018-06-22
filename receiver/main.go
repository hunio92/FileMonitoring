package main

import (
	"net/http"
	"receiver/receiver"
)

func main() {
	receiver.SaveFilesInfo()
	receiver.ReadConfig()
	http.HandleFunc("/checkfile", receiver.HandlerCheckFile)
	http.HandleFunc("/filetransfer", receiver.HandlerFileTransfer)
	http.ListenAndServe(":8888", nil)
}
