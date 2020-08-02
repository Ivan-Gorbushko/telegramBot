package apiRequests

import (
	"encoding/json"
	"fmt"
	"main/core"
)

// Response model
type AutocompleteTown struct {
	Id int `json:"id"`
	Name string `json:"name"`
	CountrySign string `json:"countrySign"`
	CountryName string `json:"countryName"`
	AreaName string `json:"areaName"`
	//Lon float32 `json:"lon"`
	//Lat float32 `json:"lat"`
}

func GetAutocompleteTowns(query string) []AutocompleteTown {
	// Settings
	var autocompleteTowns []AutocompleteTown
	country := "UA"
	requestQuery := map[string]string{}
	requestHeader := map[string]string{}
	endpointUrl := "https://lardi-trans.com/webapi/geo/towns/autocomplete"
	// Prepare Query Parameters
	requestQuery["query"] = query
	requestQuery["countrySign"] = country
	// Request
	body := core.Get(endpointUrl, requestQuery, requestHeader)
	// Decode
	_ = json.Unmarshal([]byte(body), &autocompleteTowns)
	fmt.Println(autocompleteTowns)
	return autocompleteTowns
}
