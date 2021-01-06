package inlineHandlers

import (
	"fmt"
	"log"
	"main/apiRequests"
	"main/core"
	"main/models"
	"regexp"
	"strconv"
	"strings"
)

// Make request to lardi-trans.com to create new post
func CreatePostInlineHandler(requestId string) interface{} {
	log.Println(requestId)
	country := "UA"
	postData := models.GetPostByRequestId(requestId)
	core.DisconnectMongo()
	log.Println(postData)

	sourceTownName := prepareTownName(postData.SourceCity)
	sourceAutocompleteTowns := apiRequests.GetAutocompleteTowns(sourceTownName)

	if len(sourceAutocompleteTowns) <= 0 {
		log.Println(fmt.Sprintf("Bad naming of the city (%s). Request was skipped", sourceTownName))
		return false
	}

	sourceAutocompleteTown := sourceAutocompleteTowns[0]

	sourceTowns := apiRequests.GetTowns(sourceAutocompleteTown)
	sourceTown := sourceTowns[0]

	waypointListSource := apiRequests.WaypointListSource{
		CountrySign: country,
		TownId: strconv.Itoa(sourceTown.Id),
		AreaId: strconv.Itoa(sourceTown.AreaId),
	}

	targetTownName := prepareTownName(postData.DestinationCity)
	targetAutocompleteTowns := apiRequests.GetAutocompleteTowns(targetTownName)

	if len(targetAutocompleteTowns) <= 0 {
		log.Println(fmt.Sprintf("Bad naming of the city (%s). Request was skipped", targetTownName))
		return false
	}

	targetAutocompleteTown := targetAutocompleteTowns[0]

	targetTowns := apiRequests.GetTowns(targetAutocompleteTown)
	targetTown := targetTowns[0]

	waypointListTarget := apiRequests.WaypointListTarget{
		CountrySign: country,
		TownId: strconv.Itoa(targetTown.Id),
		AreaId: strconv.Itoa(targetTown.AreaId),
	}

	queryData := map[string]string{}

	// Search body type
	for _, bodyType := range apiRequests.GetBodyTypes() {
		if strings.ToLower(bodyType.Name) == strings.ToLower(postData.Truck) {
			queryData["bodyTypeId"] = strconv.Itoa(bodyType.Id)
			fmt.Println(bodyType)
			break
		}
	}

	// Search group type
	queryData["bodyGroupId"] = "1" // By default крытая (1)
	for _, bodyGroup := range apiRequests.GetBodyGroups() {
		if strings.ToLower(bodyGroup.Name) == strings.ToLower(postData.Truck) {
			queryData["bodyGroupId"] = strconv.Itoa(bodyGroup.Id)
			break
		}
	}

	queryData["contentName"] = postData.ProductType

	body := apiRequests.PostCargo(waypointListSource, waypointListTarget, postData, queryData)

	return body
}

func prepareTownName(townName string) string {
	townNameMapping := map[string] string {
		"Днипро": "Днепр",
	}

	for needle, replace := range townNameMapping{
		reg := regexp.MustCompile(needle)
		townName = reg.ReplaceAllString(townName, replace)
	}

	return townName
}

