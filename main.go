package main

import (
	"groupie-tracker/backend/api"
	"groupie-tracker/backend/logics"
	"log"
	"net/http"
)

func main() {
	api.JsonToMap()
	http.HandleFunc("/", logics.MainHandler)

	log.Fatal(http.ListenAndServe(":1010", nil))

}
