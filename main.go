package main

import (
	"groupie-tracker/backend/api"
	"groupie-tracker/backend/handlers"
	"log"
	"net/http"
)

func main() {
	store, err := api.LoadStore()
	if err != nil {
		log.Fatalf("failed to load artist data: %v", err)
	}

	h := &handlers.HandlerStore{Store: store}

	http.HandleFunc("/", h.MainHandler)
	http.HandleFunc("/artist", h.ArtistHandler)
	http.HandleFunc("/api/filter", h.HandleFilter)
	http.HandleFunc("/api/filters/meta", h.HandleFilterMeta)
	http.HandleFunc("/api/search", h.HandleSearch)
	http.HandleFunc("/api/artist/locations", h.ArtistLocationsHandler)

	log.Fatal(http.ListenAndServe(":1010", nil))

}
