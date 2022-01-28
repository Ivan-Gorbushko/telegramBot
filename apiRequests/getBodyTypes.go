package apiRequests

import (
	"encoding/json"
	"fmt"
	"main/core"
)

// Response model
type BodyType struct {
	Id int `json:"id"`
	Name string `json:"name"`
}

func GetBodyTypes() []BodyType {
	// Settings
	var bodyTypes []BodyType
	requestQuery := map[string]string{}
	requestHeader := map[string]string{}
	baseUrl := core.Config.LardiApiUrl
	lardiSecretKey := core.Config.LardiSecretKey
	endpointUrl := fmt.Sprintf("%s/v1/references/body/types", baseUrl)
	// Prepare Headers
	requestHeader["Authorization"] = lardiSecretKey
	// Request
	body := core.Get(endpointUrl, requestQuery, requestHeader)
	// Decode
	_ = json.Unmarshal([]byte(body), &bodyTypes)
	fmt.Println(bodyTypes)
	return bodyTypes
}
