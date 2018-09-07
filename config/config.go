package config

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
)

// Config scheme struct for unmarshall
type Config struct {
	AccessToken string `json:"access_token"`
}

// Instance of unmarshalled config file
var Instance Config

func init() {
	configBytes, err := ioutil.ReadFile("config/config.json")
	if err != nil {
		fmt.Println(err)
		return
	}

	if err := json.Unmarshal(configBytes, &Instance); err != nil {
		fmt.Println(err)
		return
	}
}
