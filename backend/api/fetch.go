package api

import (
	"encoding/json"
	"groupie-tracker/backend/models"
	"io"
	"log"
	"net/http"
	"sync"
)

const baseURL = "https://groupietrackers.herokuapp.com/api"

func FetchJSON(link string, target interface{}) error {
	resp, err := http.Get(link)

	if err != nil {
		return err
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return err
	}

	body, err := io.ReadAll(resp.Body)

	if err != nil {
		return err
	}

	if err := json.Unmarshal(body, target); err != nil {
		return err
	}

	return nil

}

var Views []models.ArtistView
var ViewsByID = make(map[int]models.ArtistView)

func JsonToMap() {
	// 	{"artists":"https://groupietrackers.herokuapp.com/api/artists",
	// 	"locations":"https://groupietrackers.herokuapp.com/api/locations",
	// 	"dates":"https://groupietrackers.herokuapp.com/api/dates",
	// 	"relation":"https://groupietrackers.herokuapp.com/api/relation"}

	var index models.Index
	var wg sync.WaitGroup

	if err := FetchJSON(baseURL, &index); err != nil {
		log.Println(err)
		return

	}
	wg.Add(3)

	var artists []models.Artist

	go func() {
		defer wg.Done()
		if err := FetchJSON(index.Artists, &artists); err != nil {
			log.Println(err)

			return
		}
	}()

	var location models.LocationsResponse

	go func() {
		defer wg.Done()
		if err := FetchJSON(index.Locations, &location); err != nil {
			log.Println(err)

			return
		}
	}()

	var dates models.DatesResponse

	go func() {
		defer wg.Done()
		if err := FetchJSON(index.Dates, &dates); err != nil {
			log.Println(err)

			return
		}
	}()

	wg.Wait()

	datesmap := make(map[int][]string)
	locationmap := make(map[int][]string)
	wg.Add(2)
	go func() {
		for _, d := range dates.Index { //вот тут можно применить горутины будто бы
			datesmap[d.ID] = d.Dates
		}
		wg.Done()
	}()
	go func() {
		for _, x := range location.Index {
			locationmap[x.ID] = x.Locations
		}
		wg.Done()
	}()

	wg.Wait()

	for _, a := range artists {
		view := models.ArtistView{
			Artist:    a,
			Locations: locationmap[a.ID],
			Dates:     datesmap[a.ID],
		}

		Views = append(Views, view)

		ViewsByID[a.ID] = view
	}

}
