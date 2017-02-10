package api

import (
	"io/ioutil"
	"net/http"
	"net/url"
	"fmt"
	"errors"
	"log"
	"encoding/json"
)

const (
	PARTNER_ID   = "partner_id"
	SERVER_TOKEN = "server_token"
)

type NambaTaxiApi struct {
	partnerID string
	serverToken string
	url string
	version string
}

type Fares struct {
	Fare []struct {
		Flagfall float64 `json:"flagfall"`
		Free_waiting float64 `json:"free_waiting"`
		Full_description string `json:"full_description"`
		Include_kilometers int `json:"include_kilometers"`
		Id int `json:"id"`
		Cost_per_kilometer float64 `json:"cost_per_kilometer"`
		Name string `json:"name"`
	} `json:"fares"`
}

func NewNambaTaxiApi(partnerID string, serverToken string, url string, version string) NambaTaxiApi {
	return NambaTaxiApi{partnerID, serverToken, url, version}
}

func (api *NambaTaxiApi) GetFares() (Fares, error) {
	jsonData, err := api.makePostRequest("fares")
	if err != nil {
		return Fares{}, err
	}
	fares := Fares{}
	err = json.Unmarshal(jsonData, &fares)
	if err != nil {
		return Fares{}, err
	}
	return fares, nil
}

func (api *NambaTaxiApi) makePostRequest(uri string) ([]byte, error) {
	resp, err := http.PostForm(api.getApiURL(uri),
		url.Values{
			PARTNER_ID:   {api.partnerID},
			SERVER_TOKEN: {api.serverToken},
		})

	if err != nil {
		panic(err)
	}

	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)

	if resp.StatusCode != 200 {
		return nil, errors.New(resp.Status)
	}

	return body, nil
}

func (api *NambaTaxiApi) getApiURL(uri string) string {
	urlString := fmt.Sprintf("%s/%s/%s/", api.url, api.version, uri)
	log.Printf("API URL is: %v", urlString)
	return urlString
}
