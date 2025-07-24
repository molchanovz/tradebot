package ozon

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
)

func PostingFbo(clientID, apiKey, postingNumber string) PostingFBO {
	var posting PostingFBO
	jsonString, _ := v2PostingFboGet(clientID, apiKey, postingNumber)
	err := json.Unmarshal([]byte(jsonString), &posting)
	if err != nil {
		log.Fatalf("Error decoding JSON: %v", err)
	}
	return posting
}

func ReturnsList(clientID, apiKey string, lastID int, since, to string) (Returns, error) {
	var returns Returns
	jsonSrting := returnsList(clientID, apiKey, lastID, since, to)
	err := json.Unmarshal([]byte(jsonSrting), &returns)
	if err != nil {
		return returns, fmt.Errorf("error decoding JSON: %w", err)
	}
	return returns, nil
}

func PostingFbs(clientID, apiKey, postingNumber string) PostingFBS {
	var posting PostingFBS
	jsonString := V3PostingFbsGet(clientID, apiKey, postingNumber) // assuming this function returns the JSON string
	err := json.Unmarshal([]byte(jsonString), &posting)
	if err != nil {
		log.Fatalf("Error decoding JSON: %v", err)
	}
	return posting
}

func PostingsListFbs(clientID, apiKey, since, to string, offset int, status string) (PostingslistFbs, error) {
	var postingList PostingslistFbs
	jsonString, _ := V3PostingFbsList(clientID, apiKey, since, to, offset, status)
	if jsonString == "" {
		return postingList, errors.New("json пустой")
	}
	err := json.Unmarshal([]byte(jsonString), &postingList)
	if err != nil {
		return postingList, fmt.Errorf("error decoding JSON: %w", err)
	}
	return postingList, nil
}
func PostingsListFbo(clientID, apiKey, since, to string, offset int) PostingslistFbo {
	var postingList PostingslistFbo
	jsonString := V2PostingFboList(clientID, apiKey, since, to, offset)
	err := json.Unmarshal([]byte(jsonString), &postingList)
	if err != nil {
		log.Fatalf("Error decoding JSON: %v", err)
	}
	return postingList
}

func Stocks(clientID, apiKey string) StocksList {
	var stocks StocksList
	jsonString := v2StockOnWarehouses(clientID, apiKey)
	err := json.Unmarshal([]byte(jsonString), &stocks)
	if err != nil {
		log.Fatalf("Error decoding JSON: %v", err)
	}
	return stocks
}

func StocksAnalytics(clientID, apiKey string, skus []string) StocksNew {
	var stocks StocksNew
	jsonString := v1AnalyticsStocks(clientID, apiKey, skus)
	err := json.Unmarshal([]byte(jsonString), &stocks)
	if err != nil {
		log.Fatalf("Error decoding JSON: %v", err)
	}
	return stocks
}

func Products(clientID, apiKey string) ProductList {
	var products ProductList
	jsonString := v3ProductList(clientID, apiKey)
	err := json.Unmarshal([]byte(jsonString), &products)
	if err != nil {
		log.Fatalf("Error decoding JSON: %v", err)
	}
	return products
}

func ProductsWithAttributes(clientID, apiKey string) ProductListWithAttributes {
	var products ProductListWithAttributes
	jsonString := v4ProductInfoAttributes(clientID, apiKey)
	err := json.Unmarshal([]byte(jsonString), &products)
	if err != nil {
		log.Fatalf("Error decoding JSON: %v", err)
	}
	return products
}

func Clusters(clientID, apiKey string) ClustersList {
	var clusters ClustersList
	jsonString := v1Clusters(clientID, apiKey)
	err := json.Unmarshal([]byte(jsonString), &clusters)
	if err != nil {
		log.Fatalf("Error decoding JSON: %v", err)
	}
	return clusters
}

//func GetLabel(clientID, apiKey, postingNumber string) PackageLabel {
//	var label PackageLabel
//	jsonString := V2PostingFbsPackageLabel(clientID, apiKey, postingNumber)
//	err := json.Unmarshal([]byte(jsonString), &label)
//	if err != nil {
//		log.Fatalf("Error decoding JSON: %v", err)
//	}
//	return label
//}
