package apiRequests

import (
	"encoding/json"
	"fmt"
	"main/core"
)

// Response model
type BodyGroup struct {
	Id int `json:"id"`
	Name string `json:"name"`
}

func GetBodyGroups() []BodyGroup {
	// Settings
	var bodyGroups []BodyGroup
	requestQuery := map[string]string{}
	requestHeader := map[string]string{}
	baseUrl := core.Config.LardiApiUrl
	lardiSecretKey := core.Config.LardiSecretKey
	endpointUrl := fmt.Sprintf("%s/v1/references/body/groups", baseUrl)
	// Prepare Headers
	requestHeader["Authorization"] = lardiSecretKey
	// Request
	body := core.Get(endpointUrl, requestQuery, requestHeader)
	// Decode
	_ = json.Unmarshal([]byte(body), &bodyGroups)
	fmt.Println(bodyGroups)
	return bodyGroups
}
