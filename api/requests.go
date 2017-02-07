package api

import (
	"io/ioutil"
	"net/http"
	"net/url"
	"fmt"
	"errors"
	"log"
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

type Fare struct {

}

func NewNambaTaxiApi(partnerID string, serverToken string, url string, version string) NambaTaxiApi {
	return NambaTaxiApi{partnerID, serverToken, url, version}
}

func (api *NambaTaxiApi) GetFares() error {
	_, err := api.makePostRequest("fares")
	if err != nil {
		return err
	}
	return nil
}

func (api *NambaTaxiApi) makePostRequest(uri string) (string, error) {
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
		return "", errors.New(resp.Status)
	}

	return string(body), nil
}

func (api *NambaTaxiApi) getApiURL(uri string) string {
	urlString := fmt.Sprintf("%s/%s/%s/", api.url, api.version, uri)
	log.Printf("API URL is: %v", urlString)
	return urlString
}
