package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"strconv"
)

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


type AutocompleteTown struct {
	Id int `json:"id"`
	Name string `json:"name"`
	CountrySign string `json:"countrySign"`
	//Lon float32 `json:"lon"`
	//Lat float32 `json:"lat"`
	CountryName string `json:"countryName"`
	AreaName string `json:"areaName"`
}

type City struct {
	Id int `json:"id"`
	Name string `json:"name"`
	CountrySign string `json:"countrySign"`
	AreaId int `json:"areaId"`
	//Lat float32 `json:"lat"`
	//Lon float32 `json:"lon"`
}

var Cities []City

func getAutocompleteTowns(query string) []AutocompleteTown {
	// Settings
	var autocompleteTowns []AutocompleteTown
	country := "UA"
	autocompleteUrlTowns := "https://lardi-trans.com/webapi/geo/towns/autocomplete"

	// Prepare Query Parameters
	params := url.Values{}
	params.Add("query", query)
	params.Add("countrySign", country)
	queryValue := params.Encode()
	readyUrl := fmt.Sprintf("%s/?%s", autocompleteUrlTowns, queryValue)

	log.Println(readyUrl)

	// Prepare request
	getTowns, _ := http.NewRequest("GET", readyUrl, bytes.NewBuffer([]byte{}))

	// Do request
	clientRestApi := &http.Client{}
	resp, _ := clientRestApi.Do(getTowns)
	defer resp.Body.Close()
	body, _ := ioutil.ReadAll(resp.Body)
	_ = json.Unmarshal([]byte(body), &autocompleteTowns)

	log.Println(autocompleteTowns)

	return autocompleteTowns
}

func getTowns(autocompleteTown AutocompleteTown) []City {
	// Settings
	var towns []City
	baseUrl := getEnvData("lardi_api_url", "")
	getTownApiUrl := fmt.Sprintf("%s/v2/references/towns", baseUrl)
	lardiSecretKey := getEnvData("lardi_secret_key", "")

	// Prepare Query Parameters
	params := url.Values{}
	params.Add("ids", strconv.Itoa(autocompleteTown.Id))
	queryValue := params.Encode()
	readyUrl := fmt.Sprintf("%s/?%s", getTownApiUrl, queryValue)

	log.Println(readyUrl)

	// Prepare request
	req, _ := http.NewRequest("GET", readyUrl, bytes.NewBuffer([]byte{}))

	// Prepare Headers
	req.Header.Set("Authorization", lardiSecretKey)
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Content-Type", "application/json")

	// Do request
	clientRestApi := &http.Client{}
	resp, _ := clientRestApi.Do(req)
	defer resp.Body.Close()
	body, _ := ioutil.ReadAll(resp.Body)
	_ = json.Unmarshal([]byte(body), &towns)

	log.Println(towns)

	return towns
}

func postCargo(waypointListSource WaypointListSource, waypointListTarget WaypointListTarget, post Post) interface{} {
	// Settings
	baseUrl := getEnvData("lardi_api_url", "")
	endpointUrl := fmt.Sprintf("%s/v2/proposals/my/add/cargo", baseUrl)
	lardiSecretKey := getEnvData("lardi_secret_key", "")





	//GET /v2/references/body/types  - bodyTypeId get all types
//	[
//	{
//		"id": 34,
//		"name": "Тент"
//	},
//	{
//		"id": 27,
//		"name": "Контейнер"
//	},
//	{
//		"id": 25,
//		"name": "Изотерм"
//	}
//]
////params.Add("bodyTypeId", "2")

// try to find id by name (name = post.truck) if it was found we stop


	// GET /v2/references/body/groups  - bodyGroupId get all types
	//	[
	//	    {
	//        "id": 1,
	//        "name": "Крытая"
	//    },
	//    {
	//        "id": 2,
	//        "name": "Открытая"
	//    },
	//    {
	//        "id": 3,
	//        "name": "Цистерна"
	//    },
	//    {
	//        "id": 4,
	//        "name": "Специальная техника"
	//    },
	//    {
	//        "id": 5,
	//        "name": "Пассажирский"
	//    },
	//    {
	//        "id": 6,
	//        "name": "Типы кузова USA"
	//    }
	//]

	// try to find id by name (name = post.truck) if it was found we stop
	//params.Add("bodyGroupId", "2")

	// If bodyTypeId == nill and bodyGroupId == nill by default bodyGroupId = крытая (1) params.Add("bodyGroupId", "1")



	// GET /v2/references/cargo?query=aprico&language=en - contentId get all products by name, get first  -  if it was found we stop
	//params.Add("contentId", contentId)

	//[
	//    {
	//        "id": 12,
	//        "name": "apricot in boxes"
	//    },
	//    {
	//        "id": 6,
	//        "name": "apricots in boxes"
	//    },
	//    {
	//        "id": 7,
	//        "name": "apricots"
	//    }
	//]

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


/*
	Example request to create Cargo Post

	curl -X POST -H "Accept: application/json" -H "Content-Type: application/json" -H "Authorization: 3WQ1EQ465C4005000130" \
	"https://api.lardi-trans.com/v2/proposals/my/add/cargo?\
	dateFrom=2020-11-27\
	&dateTo=2020-11-30\
	&contentId=18\
	&bodyGroupId=2\
	&bodyTypeId=63\
	&loadTypes=24,25\
	&unloadTypes=26,27\
	&adr=3\
	&cmr=true\
	&cmrInsurance=true\
	&groupage=true\
	&t1=true\
	&tir=true\
	&lorryAmount=2\
	&note=some%20useful%20note\
	&paymentPrice=1000\
	&paymentCurrencyId=2\
	&paymentUnitId=2\
	&paymentTypeId=8\
	&paymentMomentId=4\
	&paymentPrepay=10\
	&paymentDelay=5\
	&paymentVat=true\
	&medicalRecords=true\
	&customsControl=true\
	&sizeMassFrom=24\
	&sizeMassTo=36\
	&sizeVolumeFrom=30\
	&sizeVolumeTo=40\
	&sizeLength=10.1\
	&sizeWidth=2.5\
	&sizeHeight=3\
	" -d '{
	    "waypointListSource": [
	        {
	            "address": "уточнение адреса",
	            "countrySign": "UA",
	            "areaId": 23,
	            "townId": 137
	        }
	    ],
	    "waypointListTarget": [
	        {
	            "address": "уточнение адреса",
	            "countrySign": "UA",
	            "areaId": 34,
	            "townId": 69
	        }
	    ]
	}'
*/