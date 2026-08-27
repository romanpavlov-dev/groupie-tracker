package search

import (
	"groupie-tracker/backend/models"
	"strconv"
	"strings"
)



const (
	TypeArtistBand     = "artist/band"
	TypeMember         = "member"
	TypeLocation       = "location"
	TypeFirstAlbumDate = "first album date"
	TypeCreationDate   = "creation date"
)

type Suggestion struct {
	Text      string `json:"text"`
	Type      string `json:"type"`
	ArtistIDs []int  `json:"artistIds"`
}

type suggestionKey struct {
	text string
	typ  string
}

func Suggestions(views []models.ArtistView, query string) []Suggestion {
	query = strings.TrimSpace(query)
	if query == "" {
		return []Suggestion{}
	}

	needle := normalize(query)
	grouped := make(map[suggestionKey][]int)
	order := make([]suggestionKey, 0)

	add := func(text, typ string, artistID int) {
		key := suggestionKey{text: text, typ: typ}
		ids, exists := grouped[key]
		if !exists {
			order = append(order, key)
		}
		for _, id := range ids {
			if id == artistID {
				return
			}
		}
		grouped[key] = append(ids, artistID)
	}

	for _, v := range views {
		if matches(v.Name, needle) {
			add(v.Name, TypeArtistBand, v.ID)
		}
		for _, member := range v.Members {
			if matches(member, needle) {
				add(member, TypeMember, v.ID)
			}
		}
		for _, loc := range v.Locations {
			if matches(loc, needle) {
				add(loc, TypeLocation, v.ID)
			}
		}
		if matches(v.FirstAlbum, needle) {
			add(v.FirstAlbum, TypeFirstAlbumDate, v.ID)
		}
		creation := strconv.Itoa(v.CreationDate)
		if matches(creation, needle) {
			add(creation, TypeCreationDate, v.ID)
		}
	}

	out := make([]Suggestion, 0, len(order))
	for _, key := range order {
		out = append(out, Suggestion{
			Text:      key.text,
			Type:      key.typ,
			ArtistIDs: grouped[key],
		})
	}
	return out
}

func Apply(views []models.ArtistView, query, matchType, matchValue string) []models.ArtistView {
	if strings.TrimSpace(matchType) != "" && strings.TrimSpace(matchValue) != "" {
		return filterExact(views, matchType, matchValue)
	}

	query = strings.TrimSpace(query)
	if query == "" {
		return views
	}

	needle := normalize(query)
	result := make([]models.ArtistView, 0, len(views))
	for _, v := range views {
		if artistMatchesQuery(v, needle) {
			result = append(result, v)
		}
	}
	return result
}

func artistMatchesQuery(v models.ArtistView, needle string) bool {
	if matches(v.Name, needle) {
		return true
	}
	for _, member := range v.Members {
		if matches(member, needle) {
			return true
		}
	}
	for _, loc := range v.Locations {
		if matches(loc, needle) {
			return true
		}
	}
	if matches(v.FirstAlbum, needle) {
		return true
	}
	return matches(strconv.Itoa(v.CreationDate), needle)
}

func filterExact(views []models.ArtistView, matchType, matchValue string) []models.ArtistView {
	result := make([]models.ArtistView, 0, len(views))
	for _, v := range views {
		if artistMatchesExact(v, matchType, matchValue) {
			result = append(result, v)
		}
	}
	return result
}

func artistMatchesExact(v models.ArtistView, matchType, matchValue string) bool {
	switch matchType {
	case TypeArtistBand:
		return strings.EqualFold(v.Name, matchValue)
	case TypeMember:
		for _, member := range v.Members {
			if strings.EqualFold(member, matchValue) {
				return true
			}
		}
	case TypeLocation:
		for _, loc := range v.Locations {
			if strings.EqualFold(loc, matchValue) {
				return true
			}
		}
	case TypeFirstAlbumDate:
		clean := func(s string) string {
			return strings.ReplaceAll(s, "*", "")
		}
		return strings.EqualFold(clean(v.FirstAlbum), clean(matchValue))
	case TypeCreationDate:
		return strconv.Itoa(v.CreationDate) == matchValue
	}
	return false
}

func matches(value, needle string) bool {
	return strings.Contains(normalize(value), needle)
}

func normalize(s string) string {
	s = strings.ToLower(strings.TrimSpace(s))
	var b strings.Builder
	prevSpace := false
	for _, r := range s {
		if r == '_' || r == '-' || r == ',' || r == '*' {
			if !prevSpace {
				b.WriteByte(' ')
				prevSpace = true
			}
			continue
		}
		b.WriteRune(r)
		prevSpace = false
	}
	return strings.TrimSpace(b.String())
}
