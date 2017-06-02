package api

import (
	"testing"
	"github.com/maddevsio/simple-config"
	"github.com/stretchr/testify/assert"
)

var config = simple_config.NewSimpleConfig("../config", "yml")

func getApi() NambaTaxiAPI {
	return NewNambaTaxiApi(
		config.GetString("partner_id"),
		config.GetString("server_token"),
		config.GetString("url"),
		config.GetString("version"),
	)
}

func getFakeApi() NambaTaxiAPI {
	return NewNambaTaxiApi(
		config.GetString("partner_id"),
		config.GetString("server_token"),
		config.GetString("url"),
		"xxx",
	)
}

func TestGetFares(t *testing.T) {
	nambaTaxiAPI := getApi()
	fares, err := nambaTaxiAPI.GetFares()
	assert.NoError(t, err)
	assert.Equal(t, 1, fares.Fare[0].ID)
	assert.Equal(t, 21, fares.Fare[1].ID)
	assert.Equal(t, 70.0, fares.Fare[1].Flagfall)
	assert.Equal(t, 5, len(fares.Fare))
}

func TestGetPaymentMethods(t *testing.T) {
	nambaTaxiAPI := getApi()
	paymentMethods, err := nambaTaxiAPI.GetPaymentMethods()
	assert.NoError(t, err)
	assert.Equal(t, 1, paymentMethods.PaymentMethod[0].PaymentMethodID)
	assert.Equal(t, "Наличными", paymentMethods.PaymentMethod[0].Description)
	assert.Equal(t, 4, len(paymentMethods.PaymentMethod))
}

func TestGetNearestDrivers(t *testing.T) {
	nambaTaxiAPI := getApi()
	nearestDrivers, err := nambaTaxiAPI.GetNearestDrivers("Московская Советская")
	assert.NoError(t, err)
	assert.Equal(t, "200", nearestDrivers.Status)
	assert.Equal(t, "Drivers found", nearestDrivers.Message)
	assert.Equal(t, 0, nearestDrivers.Drivers)
}

func TestGetNearestDriversError(t *testing.T) {
	nambaTaxiAPI := getFakeApi()
	_, err := nambaTaxiAPI.GetNearestDrivers("Московская Советская")
	assert.Error(t, err)
}

func TestGetRequestOptions(t *testing.T) {
	api := getApi()
	requestOptions, err := api.GetRequestOptions()
	assert.NoError(t, err)
	assert.Equal(t, 2, len(requestOptions.RequestOption))
	assert.Equal(t, "Курящий", requestOptions.RequestOption[0].Title)
}

func TestMakeOrder_GetOrder_DeleteOrder(t *testing.T) {
	api := getApi()

	orderOptions := map[string][]string{
		"phone_number": {"0555121314"},
		"address":      {"ул Советская, дом 1, палата 6"},
		"fare":         {"1"},
	}

	order1, err := api.MakeOrder(orderOptions)
	assert.NoError(t, err)
	assert.Equal(t, "success", order1.Message)

	order2, err := api.GetOrder(order1.OrderID)
	assert.NoError(t, err)
	assert.Equal(t, order1.OrderID, order2.OrderID)
	assert.Equal(t, "Новый заказ", order2.Status)

	cancel1, err := api.CancelOrder(order1.OrderID)
	assert.Equal(t, "200", cancel1.Status)

	cancel2, err := api.CancelOrder(order2.OrderID)
	assert.Equal(t, "400", cancel2.Status)
}

