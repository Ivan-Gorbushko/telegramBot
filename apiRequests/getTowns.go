package apiRequests

import (
	"encoding/json"
	"fmt"
	"main/core"
	"strconv"
)

type City struct {
	Id int `json:"id"`
	Name string `json:"name"`
	CountrySign string `json:"countrySign"`
	AreaId int `json:"areaId"`
	//Lat float32 `json:"lat"`
	//Lon float32 `json:"lon"`
}

func GetTowns(autocompleteTown AutocompleteTown) []City {
	// Settings
	var towns []City
	requestQuery := map[string]string{}
	requestHeader := map[string]string{}
	baseUrl := core.Config.LardiApiUrl
	lardiSecretKey := core.Config.LardiSecretKey
	endpointUrl := fmt.Sprintf("%s/v2/references/towns", baseUrl)
	// Prepare Query Parameters
	requestQuery["ids"] = strconv.Itoa(autocompleteTown.Id)
	// Prepare Headers
	requestHeader["Authorization"] = lardiSecretKey
	// Request
	body := core.Get(endpointUrl, requestQuery, requestHeader)
	// Decode
	_ = json.Unmarshal([]byte(body), &towns)
	fmt.Println(towns)
	return towns
}