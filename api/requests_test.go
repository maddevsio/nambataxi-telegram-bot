package api

import (
	"testing"
	"github.com/maddevsio/simple-config"
)

var config = simple_config.NewSimpleConfig("../config", "yml")

func TestGetFares(t *testing.T) {
	api := NewNambaTaxiApi(config.GetString("partner_id"), config.GetString("server_token"))
	api.GetFares()
}