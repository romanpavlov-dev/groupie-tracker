package api

import (
	"encoding/json"
	"io"
	"net/http"
)

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
