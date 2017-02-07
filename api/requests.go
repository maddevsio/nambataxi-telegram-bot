package api

import (
	"log"
	"io/ioutil"
	"net/http"
	"net/url"
)

type NambaTaxiApi struct {
	partnerID string
	serverToken string
}

func NewNambaTaxiApi(partnerID string, serverToken string) NambaTaxiApi {
	return NambaTaxiApi{partnerID, serverToken}
}

func (api *NambaTaxiApi) GetFares() {
	log.Print(api.partnerID)
	log.Print(api.serverToken)
	api.makePostRequest("https://partners.staging.swift.kg/api/v1/fares/")
}

func (api *NambaTaxiApi) makePostRequest(apiURL string) string {
	resp, err := http.PostForm(apiURL,
		url.Values{"partner_id":   {api.partnerID},
			   "server_token": {api.serverToken},
		})
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	log.Printf("status %v", resp.Status)
	log.Printf("body %v", string(body))
	return ""
}
