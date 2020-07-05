package apiRequests

import (
	"encoding/json"
	"fmt"
	"main/core"
)

// Response model
type Cargo struct {
	Id int `json:"id"`
	Name string `json:"name"`
}

func GetCargos(query string) []Cargo {
	// Settings
	var cargos []Cargo
	language := "ru"
	requestQuery := map[string]string{}
	requestHeader := map[string]string{}
	baseUrl := core.Config.LardiApiUrl
	lardiSecretKey := core.Config.LardiSecretKey
	endpointUrl := fmt.Sprintf("%s/v2/references/cargo", baseUrl)
	// Prepare Query Parameters
	requestQuery["query"] = query
	requestQuery["language"] = language
	// Prepare Headers
	requestHeader["Authorization"] = lardiSecretKey
	// Request
	body := core.Get(endpointUrl, requestQuery, requestHeader)
	// Decode
	_ = json.Unmarshal([]byte(body), &cargos)
	fmt.Println(cargos)
	return cargos
}
