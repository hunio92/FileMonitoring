package main

import (
	"FileMonitoring/sender/sender"
	"net/http"
)

func main() {
	sender.StoreFilesInfo()
	http.HandleFunc("/register", sender.HandlerRegister)
	http.ListenAndServe(":8080", nil)
}
