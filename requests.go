package main

import (
	"io/ioutil"
	"net/http"
	"net/url"
	"encoding/json"
	"errors"
)

// Shoutouts @mholt on GitHub for json-to-go
// that shit saved my life fr

// Representation of a Deezer search result
type deezerAlbumSearchResult struct {
	Data []struct {
		ID int `json:"id"`
	} `json:"data"`
}

// Representation of a Deezer album
type deezerAlbum struct {
	Title    string `json:"title"`
	Link     string `json:"link"`
	Cover    string `json:"cover"`
	CoverBig string `json:"cover_big"`
	CoverXl  string `json:"cover_xl"`
	Genres   struct {
		Data []struct {
			Name    string `json:"name"`
			Picture string `json:"picture"`
		} `json:"data"`
	} `json:"genres"`
	Released string `json:"release_date"`
	Artist struct {
		Name string `json:"name"`
	} `json:"artist"`
	Tracks struct {
		Data []struct {
			Title    string `json:"title"`
			Duration int `json:"duration"`
		} `json:"data"`
	} `json:"tracks"`
}

// Sends request to the Deezer API and returns the raw JSON in bytes for marshalling
func requestDeezerAPI(client *http.Client, path string) ([]byte, error)  {
	request, err := http.NewRequest("GET", "http://api.deezer.com" + path, nil)
	if err != nil {
		
		return nil, err
	}
	request.Header.Set("Content-Type", "application/json")

	response, err := client.Do(request)
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()

	body, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return nil, err
	}

	return body, nil
}

// Searches Deezer for an album and returns it's ID
func searchAlbum(client *http.Client, name string) (int, error) {
	escapedName := url.QueryEscape(name)

	body, err := requestDeezerAPI(client, "/search/album?q=" + escapedName)
	if err != nil {
		return -1, err
	}

	var search deezerAlbumSearchResult
	err = json.Unmarshal(body, &search)
	if err != nil {
		return -1, err
	}

	if len(search.Data) < 1 {
		err = errors.New("No Deezer results for " + name)
		return -1, err
	}

	return search.Data[0].ID, nil
}

/*
	Get a specific album, a search doesn't return all necessary
	meta-data, so you have to request the actual album
*/
func getAlbum(client *http.Client, ID string) (*deezerAlbum, error) {
	album := &deezerAlbum{}

	body, err := requestDeezerAPI(client, "/album/" + ID)
	if err != nil {
		return album, err
	}

	err = json.Unmarshal(body, &album)
	if err != nil {
		return album, err
	}

	return album, nil
}
