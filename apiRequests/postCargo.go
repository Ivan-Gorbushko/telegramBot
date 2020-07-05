package apiRequests

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"main/core"
	"main/models"
	"net/http"
	"net/url"
)

// Response model

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

func PostCargo(waypointListSource WaypointListSource, waypointListTarget WaypointListTarget, post models.Post) interface{} {
	// Settings
	baseUrl := core.Config.LardiApiUrl
	lardiSecretKey := core.Config.LardiSecretKey
	endpointUrl := fmt.Sprintf("%s/v2/proposals/my/add/cargo", baseUrl)

	////params.Add("bodyTypeId", "2")

	// try to find id by name (name = post.truck) if it was found we stop

	// try to find id by name (name = post.truck) if it was found we stop
	//params.Add("bodyGroupId", "2")

	// If bodyTypeId == nill and bodyGroupId == nill by default bodyGroupId = крытая (1) params.Add("bodyGroupId", "1")

	// GET /v2/references/cargo?query=aprico&language=en - contentId get all products by name, get first  -  if it was found we stop
	//params.Add("contentId", contentId)

	//params.Add("contentName", post.productType)

	// sizeMassFrom - sizeMassTo
	// params.Add("sizeMassFrom", post.fromWeightTn)
	// params.Add("sizeMassTo", post.ToWeightTn)

	// sizeVolumeFrom - sizeVolumeTo
	// params.Add("sizeVolumeFrom", post.fromCube)
	// params.Add("sizeVolumeTo", post.toCube)

	//paymentPrice need to parse from post.productComment (get number by pattern (/\d{3,}/)), delete all letters
	// params.Add("paymentPrice", post.totalPrice)

	// Prepare Query Parameters
	params := url.Values{}
	params.Add("sizeMassFrom", post.WeightTn)
	params.Add("dateFrom", post.DateFrom) // "2020-11-27"
	params.Add("dateTo", post.DateTo)     // "2020-11-27"
	queryValue := params.Encode()

	// Prepare Body Parameters
	requestBody := CreatePostRequest{
		WaypointListSource: []WaypointListSource{
			waypointListSource,
		},
		WaypointListTarget: []WaypointListTarget{
			waypointListTarget,
		},
	}
	jsonValue, _ := json.MarshalIndent(requestBody, "", " ")
	readyUrl := fmt.Sprintf("%s/?%s", endpointUrl, queryValue)

	log.Println(readyUrl)

	// Prepare request
	req, _ := http.NewRequest("POST", readyUrl, bytes.NewBuffer(jsonValue))

	// Prepare Headers
	req.Header.Set("Authorization", lardiSecretKey)
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Content-Type", "application/json")

	// Do request
	clientRestApi := &http.Client{}
	resp, err := clientRestApi.Do(req)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()
	body, _ := ioutil.ReadAll(resp.Body)

	log.Println(string(body))

	// Logger result
	//fmt.Println("response Status:", resp.Status)
	//fmt.Println("response Headers:", resp.Header)
	//fmt.Println("response Body:", string(body))
	return body
}
