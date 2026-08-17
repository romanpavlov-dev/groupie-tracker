package filters

import (
	"groupie-tracker/backend/models"
	"log"
	"net/http"
	"sort"
	"strconv"
	"strings"
	"time"
)

const dateLayout = "02-01-2006"

func ParseFilterCriteria(r *http.Request) (models.FilterCriteria, []models.FieldError) {

	q := r.URL.Query()	//это тот самый длинный квери? какой тип он получает?

	var criteria models.FilterCriteria
	var errs []models.FieldError

	if q.Has("creation_min") { //а что делает вообще q.Has(), точнее как работает этот метод?
		v, err := strconv.Atoi(q.Get("creation_min"))
		if err != nil {
			errs = append(errs, models.FieldError{Field: "creation_min", Message: "must be a valid integer"})
		} else {
			criteria.CreationMin = v
			criteria.HasCreationMin = true
		}
	}
	if q.Has("creation_max") {
		v, err := strconv.Atoi(q.Get("creation_max"))
		if err != nil {
			errs = append(errs, models.FieldError{Field: "creation_max", Message: "must be a valid integer"})
		} else {
			criteria.CreationMax = v
			criteria.HasCreationMax = true
		}
	}
	if q.Has("album_min") {
		t, err := ParseAlbumDate(q.Get("album_min"))
		if err != nil {
			errs = append(errs, models.FieldError{Field: "album_min", Message: "must be in format DD-MM-YYYY"})
		} else {
			criteria.AlbumMin = t
			criteria.HasAlbumMin = true
		}
	}
	if q.Has("album_max") {
		t, err := ParseAlbumDate(q.Get("album_max"))
		if err != nil {
			errs = append(errs, models.FieldError{Field: "album_max", Message: "must be in format DD-MM-YYYY"})
		} else {
			criteria.AlbumMax = t
			criteria.HasAlbumMax = true
		}
	}
	if q.Has("members_min") {
		v, err := strconv.Atoi(q.Get("members_min"))
		if err != nil {
			errs = append(errs, models.FieldError{Field: "members_min", Message: "must be a valid integer"})
		} else {
			criteria.MembersMin = v
			criteria.HasMembersMin = true
		}
	}
	if q.Has("members_max") {
		v, err := strconv.Atoi(q.Get("members_max"))
		if err != nil {
			errs = append(errs, models.FieldError{Field: "members_max", Message: "must be a valid integer"})
		} else {
			criteria.MembersMax = v
			criteria.HasMembersMax = true
		}
	}
	if q.Has("locations") {
		raw := q.Get("locations")
		criteria.Locations = strings.Split(raw, ",")
	}

	return criteria, errs

}

func ApplyFilters(views []models.ArtistView, c models.FilterCriteria) []models.ArtistView {
	if c.IsEmpty() {
		return views
	}

	result := make([]models.ArtistView, 0, len(views)) // Создаем не nil слайс!!! Если создали бы по var result []models.... то получили бы nil слайс который в json обработке превартился бы в null а не [] - пустой массив информации

	for _, v := range views {
		if !matchesCreationDate(v, c) {
			continue
		}
		if !matchesFirstAlbum(v, c) {
			continue
		}
		if !matchesMembers(v, c) {
			continue
		}
		if !matchesLocations(v, c) {
			continue
		}
		result = append(result, v)

	}

	return result
}

func matchesCreationDate(v models.ArtistView, c models.FilterCriteria) bool {
	if c.HasCreationMin && v.CreationDate < c.CreationMin {
		return false
	}

	if c.HasCreationMax && v.CreationDate > c.CreationMax {
		return false
	}

	return true
}

func ParseAlbumDate(raw string) (time.Time, error) {
	clean := strings.ReplaceAll(raw, "*", "")
	return time.Parse(dateLayout, clean)
}

func matchesFirstAlbum(v models.ArtistView, c models.FilterCriteria) bool {
	if !c.HasAlbumMin && !c.HasAlbumMax {
		return true
	}

	albumDate, err := ParseAlbumDate(v.FirstAlbum)
	if err != nil {
		return false
	}

	if c.HasAlbumMin && albumDate.Before(c.AlbumMin) {
		return false
	}
	if c.HasAlbumMax && albumDate.After(c.AlbumMax) {
		return false
	}

	return true
}

func matchesLocations(v models.ArtistView, c models.FilterCriteria) bool {
	if len(c.Locations) == 0 {
		return true
	}

	for _, wanted := range c.Locations {
		for _, have := range v.Locations {
			if strings.EqualFold(have, wanted) {
				return true
			}
		}
	}

	return false
}

func matchesMembers(v models.ArtistView, c models.FilterCriteria) bool {
	if c.HasMembersMin && len(v.Members) < c.MembersMin {
		return false
	}

	if c.HasMembersMax && len(v.Members) > c.MembersMax {
		return false
	}

	return true
}

func BuildFilterMeta(views []models.ArtistView) models.FilterMeta {

	if len(views) == 0 {
		return models.FilterMeta{}
	}

	locationSet := make(map[string]struct{})

	meta := models.FilterMeta{
		CreationMin: views[0].CreationDate,
		CreationMax: views[0].CreationDate,
	}

	var minAlbum, maxAlbum time.Time

	for i, v := range views {
		if v.CreationDate < meta.CreationMin {
			meta.CreationMin = v.CreationDate
		}

		if v.CreationDate > meta.CreationMax {
			meta.CreationMax = v.CreationDate
		}

		albumDate, err := ParseAlbumDate(v.FirstAlbum)
		if err != nil {
			log.Printf("skipping bad firstAlbum date for artist id=%d: %q", v.ID, v.FirstAlbum)
		} else {
			if i == 0 || albumDate.Before(minAlbum) {
				minAlbum = albumDate
			}
			if i == 0 || albumDate.After(maxAlbum) {
				maxAlbum = albumDate
			}

		}

		for _, loc := range v.Locations {
			locationSet[loc] = struct{}{}
		}
	}

	meta.AlbumMin = minAlbum.Format(dateLayout)
	meta.AlbumMax = maxAlbum.Format(dateLayout)

	meta.Locations = make([]string, 0, len(locationSet))
	for loc := range locationSet {
		meta.Locations = append(meta.Locations, loc)
	}

	sort.Strings(meta.Locations)

	return meta
}
