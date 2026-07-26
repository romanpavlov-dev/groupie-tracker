package logics

import (
	"groupie-tracker/backend/api"
	"groupie-tracker/backend/models"
	"log"
	"net/http"
	"text/template"
)

const baseURL = "https://groupietrackers.herokuapp.com/api"

func MainHandler(w http.ResponseWriter, r *http.Request) {
	var index models.Index

	if err := api.FetchJSON(baseURL, &index); err != nil {
		log.Println(err)
		http.Error(w, "Ошибка получения данных API", http.StatusBadRequest)
		return

	}

	var artists []models.Artist

	if err := api.FetchJSON(index.Artists, &artists); err != nil {
		log.Println(err)
		http.Error(w, "Ошибка получения данных API", http.StatusInternalServerError)

		return
	}

	var location models.LocationsResponse

	if err := api.FetchJSON(index.Locations, &location); err != nil {
		log.Println(err)
		http.Error(w, "Ошибка получения данных API", http.StatusInternalServerError)

		return
	}

	var dates models.DatesResponse

	if err := api.FetchJSON(index.Dates, &dates); err != nil {
		log.Println(err)
		http.Error(w, "Ошибка получения данных API", http.StatusInternalServerError)

		return
	}

	datesmap := make(map[int][]string)
	locationmap := make(map[int][]string)
	for _, d := range dates.Index {
		datesmap[d.ID] = d.Dates
	}
	for _, x := range location.Index {
		locationmap[x.ID] = x.Locations
	}

	views := make([]models.ArtistView, 0, len(artists))

	for _, a := range artists {
		views = append(views,
			models.ArtistView{
				Artist:    a,
				Locations: locationmap[a.ID],
				Dates:     datesmap[a.ID],
			})
	}

	tmp, err := template.ParseFiles("frontend/templates/artists.html")

	if err != nil {
		http.Error(w, "Ошибка парсинга", http.StatusInternalServerError)
		return
	}

	// if err := tmp.Execute(w, artists); err != nil {
	// 	log.Println(err)

	// 	return
	// }
	if err := tmp.Execute(w, views); err != nil {
		log.Println(err)

		return
	}

}
