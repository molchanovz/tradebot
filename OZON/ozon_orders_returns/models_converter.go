package ozon_orders_returns

import (
	"WildberriesGo_bot/OZON/API"
	"encoding/json"
	"log"
)

func returnsFbo(ClientId, ApiKey string, LastID int) ([]ReturnFBO, int) {
	var returns ReturnsFBO
	jsonString := API.V3ReturnsCompanyFbo(ClientId, ApiKey, LastID) // assuming this function returns the JSON string
	err := json.Unmarshal([]byte(jsonString), &returns)
	if err != nil {
		log.Fatalf("Error decoding JSON: %v", err)
	}
	return returns.Returns, returns.LastID
}
func postingFbo(ClientId, ApiKey, PostingNumber string) PostingFBO {
	var posting PostingFBO
	jsonString := API.V2PostingFboGet(ClientId, ApiKey, PostingNumber) // assuming this function returns the JSON string
	err := json.Unmarshal([]byte(jsonString), &posting)
	if err != nil {
		log.Fatalf("Error decoding JSON: %v", err)
	}
	return posting
}

func returnsFbs(ClientId, ApiKey string, LastID int) ([]ReturnFBS, int) {
	var returns ReturnsFBS
	jsonString := API.V3ReturnsCompanyFbs(ClientId, ApiKey, LastID) // assuming this function returns the JSON string
	err := json.Unmarshal([]byte(jsonString), &returns)
	if err != nil {
		log.Fatalf("Error decoding JSON: %v", err)
	}
	return returns.Returns, returns.LastID
}
func postingFbs(ClientId, ApiKey, PostingNumber string) PostingFBS {
	var posting PostingFBS
	jsonString := API.V3PostingFbsGet(ClientId, ApiKey, PostingNumber) // assuming this function returns the JSON string
	err := json.Unmarshal([]byte(jsonString), &posting)
	if err != nil {
		log.Fatalf("Error decoding JSON: %v", err)
	}
	return posting
}

func postingsListFbs(ClientId, ApiKey string, daysAgo int) PostingsList_FBS {
	var postingList PostingsList_FBS
	jsonString := API.V3PostingFbsList(ClientId, ApiKey, daysAgo)
	err := json.Unmarshal([]byte(jsonString), &postingList)
	if err != nil {
		log.Fatalf("Error decoding JSON: %v", err)
	}
	return postingList
}
func postingsListFbo(ClientId, ApiKey string, daysAgo int) PostingsList_FBO {
	var postingList PostingsList_FBO
	jsonString := API.V2PostingFboList(ClientId, ApiKey, daysAgo)
	err := json.Unmarshal([]byte(jsonString), &postingList)
	if err != nil {
		log.Fatalf("Error decoding JSON: %v", err)
	}
	return postingList
}
