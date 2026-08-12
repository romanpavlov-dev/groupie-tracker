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
	http.HandleFunc("/artist", logics.ArtistHandler)

	log.Fatal(http.ListenAndServe(":1010", nil))

}
