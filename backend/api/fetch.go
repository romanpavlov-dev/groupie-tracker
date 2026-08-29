package api

import (
	"encoding/json"
	"fmt"
	"groupie-tracker/backend/filters"
	"groupie-tracker/backend/geo"
	"groupie-tracker/backend/models"
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
		return fmt.Errorf("unexpected status %d for %s", resp.StatusCode, link)
	}

	if err := json.NewDecoder(resp.Body).Decode(target); err != nil { //more idiomatic, cuz unmarshall/marshall designed to work directly with the slice of bytes (ex. looking for the error)
		return err
	}

	return nil

}

type Store struct {
	views     []models.ArtistView
	viewsByID map[int]models.ArtistView
	meta      models.FilterMeta
}

func NewStore(views []models.ArtistView) *Store { //why do we need this?
	viewsByID := make(map[int]models.ArtistView, len(views))

	for _, view := range views {
		viewsByID[view.ID] = view
	}

	return &Store{
		views:     views,
		viewsByID: viewsByID,
		meta:      filters.BuildFilterMeta(views),
	}
}

func (s *Store) All() []models.ArtistView {
	return s.views
}

func (s *Store) ByID(id int) (models.ArtistView, bool) {
	v, ok := s.viewsByID[id]
	return v, ok
}

func (s *Store) Meta() models.FilterMeta {
	return s.meta
}

func LoadStore() (*Store, error) {
	// 	{"artists":"https://groupietrackers.herokuapp.com/api/artists",
	// 	"locations":"https://groupietrackers.herokuapp.com/api/locations",
	// 	"dates":"https://groupietrackers.herokuapp.com/api/dates",
	// 	"relation":"https://groupietrackers.herokuapp.com/api/relation"}
	var wg sync.WaitGroup

	var index models.Index
	var artists []models.Artist
	var location models.LocationsResponse
	var dates models.DatesResponse
	var relations models.RelationsResponse

	if err := FetchJSON(baseURL, &index); err != nil {
		log.Println(err)
		return nil, err

	}
	wg.Add(4)

	errs := make(chan error, 4)

	go func() {
		defer wg.Done()
		if err := FetchJSON(index.Artists, &artists); err != nil {
			log.Println(err)

			errs <- fmt.Errorf("fetch artists: %w", err)
			return
		}
	}()

	go func() {
		defer wg.Done()
		if err := FetchJSON(index.Locations, &location); err != nil {
			log.Println(err)

			errs <- fmt.Errorf("fetch locations: %w", err)
			return
		}
	}()

	go func() {
		defer wg.Done()
		if err := FetchJSON(index.Dates, &dates); err != nil {
			log.Println(err)

			errs <- fmt.Errorf("fetch dates: %w", err)
			return
		}
	}()

	go func() {
		defer wg.Done()
		if err := FetchJSON(index.Relation, &relations); err != nil {
			log.Println(err)

			errs <- fmt.Errorf("fetch relations: %w", err)
			return
		}
	}()

	wg.Wait()
	close(errs)

	for err := range errs {
		return nil, err
	}

	datesmap := make(map[int][]string)
	locationmap := make(map[int][]string)
	relationsByID := make(map[int]map[string][]string)

	for _, d := range dates.Index { //вот тут можно применить горутины будто бы
		datesmap[d.ID] = d.Dates
	}

	for _, x := range location.Index {
		locationmap[x.ID] = x.Locations
	}

	for _, x := range relations.Index {
		relationsByID[x.ID] = x.DatesLocations
	}

	// Warm the geocoding cache with every unique raw location key
	// BEFORE building views, so BuildMarkers() below can attach
	// coordinates from cache alone (no network calls per artist).
	allRawLocations := collectUniqueLocations(relationsByID)
	geo.EnsureCache(allRawLocations)

	views := make([]models.ArtistView, 0, len(artists))
	viewsByID := make(map[int]models.ArtistView, len(artists))

	for _, a := range artists {
		dateLocation := relationsByID[a.ID]

		view := models.ArtistView{
			Artist:       a,
			Locations:    locationmap[a.ID],
			Dates:        datesmap[a.ID],
			DateLocation: dateLocation,
			Markers:      geo.BuildMarkers(dateLocation), 
		}

		views = append(views, view)

		viewsByID[a.ID] = view
	}

	meta := filters.BuildFilterMeta(views)

	return &Store{views: views, viewsByID: viewsByID, meta: meta}, nil

}

// collectUniqueLocations pulls every distinct raw location key out of
// every artist's DateLocation map, so we can warm the geo cache in
// one pass before building views.
func collectUniqueLocations(relationsByID map[int]map[string][]string) []string {
	seen := make(map[string]struct{})
	result := make([]string, 0)

	for _, dateLocation := range relationsByID {
		for raw := range dateLocation {
			if _, ok := seen[raw]; !ok {
				seen[raw] = struct{}{}
				result = append(result, raw)
			}
		}
	}
	return result
}
