package apiRequests

import (
	"encoding/json"
	"fmt"
	"main/core"
	"main/models"
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
	// GET /v2/references/cargo?query=aprico&language=en -
	// Settings
	var cargoResponse CargoResponse
	requestHeader := map[string]string{}
	baseUrl := core.Config.LardiApiUrl
	lardiSecretKey := core.Config.LardiSecretKey
	endpointUrl := fmt.Sprintf("%s/v2/proposals/my/add/cargo", baseUrl)
	// Prepare Query Parameters
	// sizeMassFrom - sizeMassTo
	// params.Add("sizeMassFrom", post.fromWeightTn)
	// params.Add("sizeMassTo", post.ToWeightTn)
	requestQuery["sizeMassFrom"] = "1"
	requestQuery["sizeMassTo"] = "2"
	// sizeVolumeFrom - sizeVolumeTo
	// params.Add("sizeVolumeFrom", post.fromCube)
	// params.Add("sizeVolumeTo", post.toCube)
	requestQuery["sizeVolumeFrom"] = "1"
	requestQuery["sizeVolumeTo"] = "2"
	//paymentPrice need to parse from post.productComment (get number by pattern (/\d{3,}/)), delete all letters
	// params.Add("paymentPrice", post.totalPrice)
	requestQuery["paymentPrice"] = "1000"
	requestQuery["sizeMassFrom"] = post.WeightTn
	requestQuery["dateFrom"] = post.DateFrom // "2020-11-27"
	requestQuery["dateTo"] = post.DateTo // "2020-11-27"
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
