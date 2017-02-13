package api

import (
	"io/ioutil"
	"net/http"
	"net/url"
	"fmt"
	"errors"
	"log"
	"encoding/json"
	"strconv"
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

type PaymentMethods struct {
	PaymentMethod []struct {
		PaymentMethodId int `json:"payment_method_id"`
		Description string `json:"description"`
	} `json:"payment_methods"`
}

type RequestOptions struct {
	RequestOption []struct {
		Id int `json:"id"`
		Title string `json:"title"`
	} `json:"request_options"`
}

type Order struct {
	OrderId int `json:"order_id"`
	Message string `json:"message"`
	Status string `json:"status"`
	Driver struct {
		Name string `json:"name"`
		PhoneNumber string `json:"phone_number"`
		CabNumber string `json:"cab_number"`
		LicensePlate string `json:"license_plate"`
		Make string `json:"make"`
	} `json:"driver"`
}

func NewNambaTaxiApi(partnerID string, serverToken string, url string, version string) NambaTaxiApi {
	return NambaTaxiApi{partnerID, serverToken, url, version}
}

func (api *NambaTaxiApi) GetFares() (Fares, error) {
	structure := Fares{}
	err := api.makePostRequestAndMapStructure(&structure, "fares", make(map[string][]string))
	if err != nil {
		return Fares{}, err
	}
	return structure, nil
}

func (api *NambaTaxiApi) GetPaymentMethods() (PaymentMethods, error) {
	structure := PaymentMethods{}
	err := api.makePostRequestAndMapStructure(&structure, "payment-methods", make(map[string][]string))
	if err != nil {
		return PaymentMethods{}, err
	}
	return structure, nil
}

func (api *NambaTaxiApi) GetRequestOptions() (RequestOptions, error) {
	structure := RequestOptions{}
	err := api.makePostRequestAndMapStructure(&structure, "request-options", make(map[string][]string))
	if err != nil {
		return RequestOptions{}, err
	}
	return structure, nil
}

func (api *NambaTaxiApi) MakeOrder(orderOptions map[string][]string) (Order, error) {
	structure := Order{}
	err := api.makePostRequestAndMapStructure(&structure, "requests", orderOptions)
	if err != nil {
		return Order{}, err
	}
	return structure, nil
}

func (api *NambaTaxiApi) GetOrder(id int) (Order, error) {
	structure := Order{}
	err := api.makePostRequestAndMapStructure(&structure, "requests/"+strconv.Itoa(id), make(map[string][]string))
	if err != nil {
		return Order{}, err
	}
	return structure, nil
}

func (api *NambaTaxiApi) makePostRequestAndMapStructure(structure interface{}, uri string, postParams map[string][]string) (error) {
	jsonData, err := api.makePostRequest(uri, postParams)
	if err != nil {
		return err
	}
	err = json.Unmarshal(jsonData, structure)
	if err != nil {
		return err
	}
	return nil
}

func (api *NambaTaxiApi) makePostRequest(uri string, postParams map[string][]string) ([]byte, error) {
	var values url.Values = map[string][]string{
		PARTNER_ID:   {api.partnerID},
		SERVER_TOKEN: {api.serverToken},
	}

	for key, value := range postParams {
		values[key] = value
	}

	resp, err := http.PostForm(api.getApiURL(uri), values)

	if err != nil {
		panic(err)
	}

	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)

	if resp.StatusCode != 200 {
		return nil, errors.New(resp.Status)
	}

	//log.Print("%v", string(body))
	return body, nil
}

func (api *NambaTaxiApi) getApiURL(uri string) string {
	urlString := fmt.Sprintf("%s/%s/%s/", api.url, api.version, uri)
	log.Printf("API URL is: %v", urlString)
	return urlString
}
