package ozon

import (
	"encoding/json"
	"log"
)

func PostingFbo(ClientId, ApiKey, PostingNumber string) PostingFBO {
	var posting PostingFBO
	jsonString := V2PostingFboGet(ClientId, ApiKey, PostingNumber) // assuming this function returns the JSON string
	err := json.Unmarshal([]byte(jsonString), &posting)
	if err != nil {
		log.Fatalf("Error decoding JSON: %v", err)
	}
	return posting
}

func ReturnsList(ClientId, ApiKey string, LastID int, since, to string) Returns {
	var returns Returns
	jsonSrting := returnsList(ClientId, ApiKey, LastID, since, to)
	err := json.Unmarshal([]byte(jsonSrting), &returns)
	if err != nil {
		log.Fatalf("Error decoding JSON: %v", err)
	}
	return returns
}

func PostingFbs(ClientId, ApiKey, PostingNumber string) PostingFBS {
	var posting PostingFBS
	jsonString := V3PostingFbsGet(ClientId, ApiKey, PostingNumber) // assuming this function returns the JSON string
	err := json.Unmarshal([]byte(jsonString), &posting)
	if err != nil {
		log.Fatalf("Error decoding JSON: %v", err)
	}
	return posting
}

func PostingsListFbs(ClientId, ApiKey, since, to string) PostingsList_FBS {
	var postingList PostingsList_FBS
	jsonString := V3PostingFbsList(ClientId, ApiKey, since, to)
	err := json.Unmarshal([]byte(jsonString), &postingList)
	if err != nil {
		log.Fatalf("Error decoding JSON: %v", err)
	}
	return postingList
}
func PostingsListFbo(ClientId, ApiKey, since, to string) PostingsList_FBO {
	var postingList PostingsList_FBO
	jsonString := V2PostingFboList(ClientId, ApiKey, since, to)
	err := json.Unmarshal([]byte(jsonString), &postingList)
	if err != nil {
		log.Fatalf("Error decoding JSON: %v", err)
	}
	return postingList
}

func Stocks(ClientId, ApiKey string) StocksList {
	var stocks StocksList
	jsonString := v2StockOnWarehouses(ClientId, ApiKey)
	err := json.Unmarshal([]byte(jsonString), &stocks)
	if err != nil {
		log.Fatalf("Error decoding JSON: %v", err)
	}
	return stocks
}

func Clusters(ClientId, ApiKey string) ClustersList {
	var clusters ClustersList
	jsonString := v1Clusters(ClientId, ApiKey)
	err := json.Unmarshal([]byte(jsonString), &clusters)
	if err != nil {
		log.Fatalf("Error decoding JSON: %v", err)
	}
	return clusters
}
