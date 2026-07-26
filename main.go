package main

import (
	"groupie-tracker/backend/logics"
	"log"
	"net/http"
)

func main() {

	http.HandleFunc("/", logics.MainHandler)

	log.Fatal(http.ListenAndServe(":1010", nil))

}
