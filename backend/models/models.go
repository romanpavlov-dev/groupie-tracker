package models

import (
	"groupie-tracker/backend/geo"
	"time"
)

type Index struct {
	Artists   string `json:"artists"`
	Locations string `json:"locations"`
	Dates     string `json:"dates"`
	Relation  string `json:"relation"`
}

type ArtistView struct {
	Artist
	Locations    []string
	Dates        []string
	DateLocation map[string][]string
	Markers      []geo.Marker
}

type Artist struct {
	ID           int      `json:"id"`
	Image        string   `json:"image"`
	Name         string   `json:"name"`
	Members      []string `json:"members"`
	CreationDate int      `json:"creationDate"`
	FirstAlbum   string   `json:"firstAlbum"`
	Locations    string   `json:"locations"`
	ConcertDates string   `json:"concertDates"`
	Relations    string   `json:"relations"`
}

type LocationsResponse struct {
	Index []struct {
		ID        int      `json:"id"`
		Locations []string `json:"locations"`
		Dates     string   `json:"dates"`
	} `json:"index"`
}

type DatesResponse struct {
	Index []struct {
		ID    int      `json:"id"`
		Dates []string `json:"dates"`
	} `json:"index"`
}

type RelationsResponse struct {
	Index []ArtistRelation `json:"index"`
}

type ArtistRelation struct {
	ID             int                 `json:"id"`
	DatesLocations map[string][]string `json:"datesLocations"`
}

type FilterMeta struct {
	CreationMin int      `json:"creation_min"`
	CreationMax int      `json:"creation_max"`
	AlbumMin    string   `json:"album_min"`
	AlbumMax    string   `json:"album_max"`
	Locations   []string `json:"locations"`
}

type FilterCriteria struct {
	CreationMin    int
	HasCreationMin bool
	CreationMax    int
	HasCreationMax bool

	AlbumMin    time.Time
	HasAlbumMin bool
	AlbumMax    time.Time
	HasAlbumMax bool

	MembersMin    int
	HasMembersMin bool
	MembersMax    int
	HasMembersMax bool

	Locations []string
}

type FieldError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}



func (c *FilterCriteria) IsEmpty() bool {
	return !c.HasCreationMin && !c.HasCreationMax &&
		!c.HasAlbumMin && !c.HasAlbumMax &&
		!c.HasMembersMax && !c.HasMembersMin &&
		len(c.Locations) == 0
}
