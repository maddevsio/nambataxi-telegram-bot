package api

import (
	"testing"
	"github.com/maddevsio/simple-config"
	"github.com/stretchr/testify/assert"
)

var config = simple_config.NewSimpleConfig("../config", "yml")

func getApi() NambaTaxiApi{
	return NewNambaTaxiApi(
		config.GetString("partner_id"),
		config.GetString("server_token"),
		config.GetString("url"),
		config.GetString("version"),
	)
}

func TestGetFares(t *testing.T) {
	api := getApi()
	fares, err := api.GetFares()
	assert.NoError(t, err)
	assert.Equal(t, fares.Fare[0].Id, 1)
	assert.Equal(t, fares.Fare[1].Id, 11)
	assert.Equal(t, fares.Fare[1].Flagfall, 100.0)
	assert.Equal(t, 5, len(fares.Fare))
}

func TestGetPaymentMethods(t *testing.T) {
	api := getApi()
	paymentMethods, err := api.GetPaymentMethods()
	assert.NoError(t, err)
	assert.Equal(t, 4, len(paymentMethods.PaymentMethod))
}