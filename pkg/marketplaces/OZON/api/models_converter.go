package api

import (
	"encoding/json"
	"fmt"
	"log"
)

func PostingFbo(ClientId, ApiKey, PostingNumber string) PostingFBO {
	var posting PostingFBO
	jsonString, _ := v2PostingFboGet(ClientId, ApiKey, PostingNumber)
	err := json.Unmarshal([]byte(jsonString), &posting)
	if err != nil {
		log.Fatalf("Error decoding JSON: %v", err)
	}
	return posting
}

func ReturnsList(ClientId, ApiKey string, LastID int, since, to string) (Returns, error) {
	var returns Returns
	jsonSrting := returnsList(ClientId, ApiKey, LastID, since, to)
	err := json.Unmarshal([]byte(jsonSrting), &returns)
	if err != nil {
		return returns, fmt.Errorf("error decoding JSON: %v", err)
	}
	return returns, nil
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

func PostingsListFbs(ClientId, ApiKey, since, to string, offset int, status string) (PostingslistFbs, error) {
	var postingList PostingslistFbs
	jsonString, _ := V3PostingFbsList(ClientId, ApiKey, since, to, offset, status)
	if jsonString == "" {
		return postingList, fmt.Errorf("json пустой")
	}
	err := json.Unmarshal([]byte(jsonString), &postingList)
	if err != nil {
		return postingList, fmt.Errorf("error decoding JSON: %v", err)
	}
	return postingList, nil
}
func PostingsListFbo(ClientId, ApiKey, since, to string, offset int) PostingslistFbo {
	var postingList PostingslistFbo
	jsonString := V2PostingFboList(ClientId, ApiKey, since, to, offset)
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

func StocksAnalytics(ClientId, ApiKey string, skus []string) StocksNew {
	var stocks StocksNew
	jsonString := v1AnalyticsStocks(ClientId, ApiKey, skus)
	err := json.Unmarshal([]byte(jsonString), &stocks)
	if err != nil {
		log.Fatalf("Error decoding JSON: %v", err)
	}
	return stocks
}

func Products(ClientId, ApiKey string) ProductList {
	var products ProductList
	jsonString := v3ProductList(ClientId, ApiKey)
	err := json.Unmarshal([]byte(jsonString), &products)
	if err != nil {
		log.Fatalf("Error decoding JSON: %v", err)
	}
	return products
}

func ProductsWithAttributes(ClientId, ApiKey string) ProductListWithAttributes {
	var products ProductListWithAttributes
	jsonString := v4ProductInfoAttributes(ClientId, ApiKey)
	err := json.Unmarshal([]byte(jsonString), &products)
	if err != nil {
		log.Fatalf("Error decoding JSON: %v", err)
	}
	return products
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

//func GetLabel(ClientId, ApiKey, PostingNumber string) PackageLabel {
//	var label PackageLabel
//	jsonString := V2PostingFbsPackageLabel(ClientId, ApiKey, PostingNumber)
//	err := json.Unmarshal([]byte(jsonString), &label)
//	if err != nil {
//		log.Fatalf("Error decoding JSON: %v", err)
//	}
//	return label
//}
