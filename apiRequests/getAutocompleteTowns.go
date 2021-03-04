package apiRequests

import (
	"encoding/json"
	"fmt"
	"main/core"
	"strings"
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

func GetAutocompleteTowns(query string, district string) []AutocompleteTown {
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

	var filteredAutocompleteTowns []AutocompleteTown
	for _, autocompleteTown := range autocompleteTowns {
		if strings.Contains(autocompleteTown.AreaName, district) && autocompleteTown.Name == query {
			filteredAutocompleteTowns = append(filteredAutocompleteTowns, autocompleteTown)
		}
	}

	// If we do not have any matches for the name of the area
	if len(filteredAutocompleteTowns) <= 0 {
		fmt.Println(autocompleteTowns)
		return autocompleteTowns
	}

	fmt.Println(filteredAutocompleteTowns)
	return filteredAutocompleteTowns
}
