package core

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
)

func Post(endpointUrl string, requestQuery map[string]string, requestBody interface{}, requestHeader map[string]string) []byte {
	// Prepare Query Parameters
	params := url.Values{}
	for key, value := range requestQuery {
		params.Add(key, value)
	}
	queryValue := params.Encode()
	// Prepare Body Parameters
	jsonValue, _ := json.MarshalIndent(requestBody, "", " ")
	readyUrl := fmt.Sprintf("%s/?%s", endpointUrl, queryValue)
	log.Println(readyUrl)
	// Prepare request
	req, _ := http.NewRequest("POST", readyUrl, bytes.NewBuffer(jsonValue))
	// Prepare Headers
	for key, value := range requestHeader {
		req.Header.Set(key, value)
	}
	// Do request
	clientRestApi := &http.Client{}
	resp, err := clientRestApi.Do(req)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()
	body, _ := ioutil.ReadAll(resp.Body)
	log.Println(string(body))
	return body
}


func Get(endpointUrl string, requestQuery map[string]string, requestHeader map[string]string) []byte {
	// Prepare Query Parameters
	params := url.Values{}
	for key, value := range requestQuery {
		params.Add(key, value)
	}
	queryValue := params.Encode()
	readyUrl := fmt.Sprintf("%s/?%s", endpointUrl, queryValue)
	log.Println(readyUrl)
	// Prepare request
	req, _ := http.NewRequest("GET", readyUrl, bytes.NewBuffer([]byte{}))
	// Prepare Headers
	for key, value := range requestHeader {
		req.Header.Set(key, value)
	}
	// Do request
	clientRestApi := &http.Client{}
	resp, err := clientRestApi.Do(req)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()
	body, _ := ioutil.ReadAll(resp.Body)
	log.Println(string(body))
	return body
}