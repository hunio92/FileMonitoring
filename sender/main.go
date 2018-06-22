package main

import (
	"FileMonitoring/sender/sender"
	"net/http"
)

func main() {
	http.HandleFunc("/register", sender.HandlerRegister)
	http.ListenAndServe(":9999", nil)
}
