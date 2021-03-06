package apiRequests

import (
	"encoding/json"
	"fmt"
	"main/core"
	"main/models"
	"regexp"
	"strings"
)

// Response model
type CargoResponse struct {
	Id int `json:"id"`
}

// Request model
type WaypointListSource struct {
	Address string `json:"address"`
	CountrySign string `json:"countrySign"`
	AreaId string `json:"areaId"`
	TownId string `json:"townId"`
}

type WaypointListTarget struct {
	Address string `json:"address"`
	CountrySign string `json:"countrySign"`
	AreaId string `json:"areaId"`
	TownId string `json:"townId"`
}

type CreatePostRequest struct {
	WaypointListSource []WaypointListSource `json:"waypointListSource"`
	WaypointListTarget []WaypointListTarget `json:"waypointListTarget"`
}

func PostCargo(waypointListSource WaypointListSource, waypointListTarget WaypointListTarget, post models.Post, requestQuery map[string]string) CargoResponse {
	// Settings
	var cargoResponse CargoResponse
	requestHeader := map[string]string{}
	baseUrl := core.Config.LardiApiUrl
	lardiSecretKey := core.Config.LardiSecretKey
	endpointUrl := fmt.Sprintf("%s/v2/proposals/my/add/cargo", baseUrl)
	// Prepare Query Parameters
	if post.PaymentTypeId != "" {
		requestQuery["paymentTypeId"] = post.PaymentTypeId
	}
	if core.Config.ContactId != "" {
		requestQuery["contactId"] = core.Config.ContactId
	}
	requestQuery["sizeMassFrom"] = post.SizeMassFrom
	requestQuery["sizeMassTo"] = post.SizeMassTo
	requestQuery["sizeVolumeFrom"] = post.SizeVolumeFrom
	requestQuery["sizeVolumeTo"] = post.SizeVolumeTo
	requestQuery["paymentPrice"] = post.ProductPrice
	requestQuery["dateFrom"] = post.DateFrom // "2021-11-27"
	requestQuery["dateTo"] = post.DateTo // "2021-11-27"

	for valuesField, valuesValue := range post.Values {
		if apiField, exists := models.PostValuesSizeOnApiFields[valuesField]; exists {
			valuesValueReg := regexp.MustCompile(`(\d+[,]{0,1}\d*)`)
			valuesValueRes := valuesValueReg.FindAllSubmatch([]byte(valuesValue), -1)
			if len(valuesValueRes)-1 >= 0 {
				requestQuery[apiField] = strings.ReplaceAll(string(valuesValueRes[0][1]), ",", ".")
			}
		}
	}

	// Prepare Body Parameters
	requestBody := CreatePostRequest{
		WaypointListSource: []WaypointListSource{
			waypointListSource,
		},
		WaypointListTarget: []WaypointListTarget{
			waypointListTarget,
		},
	}
	// Prepare Headers
	requestHeader["Authorization"] = lardiSecretKey
	requestHeader["Accept"] = "application/json"
	requestHeader["Content-Type"] = "application/json"
	// Request
	body := core.Post(endpointUrl, requestQuery, requestBody, requestHeader)
	// Decode
	_ = json.Unmarshal([]byte(body), &cargoResponse)
	fmt.Println(cargoResponse)
	return cargoResponse
}
