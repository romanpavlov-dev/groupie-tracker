package models

type Index struct {
	Artists   string `json:"artists"`
	Locations string `json:"locations"`
	Dates     string `json:"dates"`
	Relation  string `json:"relation"`
}

type Artist struct {
	ID            string   `json:"id"`
	Image         string   `json:"image"`
	Name          string   `json:"name"`
	Members       []string `json:"members"`
	Creation_date int      `json:"creationDate"`
	First_album   string   `json:"firstAlbum"`
	Locations     string   `json:"locations"`
	Concert_dates string   `json:"concertDates"`
	Relations     string   `json:"relations"`
}


type LocationsResponse struct {
	Index []struct {
		ID int `json:"id"`
		Locations []string `json:"locations"`
		Dates string `json:"dates"`
	} `json:"index"`
}

type DatesResponse struct {
	Index []struct {
		ID int `json:"id"`
		Dates []string `json:"dates"`
	} `json:"index"`
}

type RelationsResponse struct {
	Index []struct {
		ID int `json:"int"`
		DatesLocations map[string][]string `json:"datesLocations"`
	} `json:"index"`
}
