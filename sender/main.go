package main

import (
	"net/http"
	"sender/sender"
)

func main() {
	http.HandleFunc("/register", sender.HandlerRegister)
	sender.CheckFiles()
	http.ListenAndServe(":9999", nil)
}
