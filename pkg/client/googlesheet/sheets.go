package googlesheet

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/option"
	"google.golang.org/api/sheets/v4"
)

type SheetsService struct {
	tokenPath       string
	credentialsPath string
}

func NewSheetsService(tokenPath, credentialsPath string) SheetsService {
	return SheetsService{
		tokenPath:       tokenPath,
		credentialsPath: credentialsPath,
	}
}

// Retrieve a token, saves the token, then returns the generated client.
func (gs SheetsService) getClient(config *oauth2.Config) (*http.Client, error) {
	// The file token.json stores the user's access and refresh tokens, and is
	// created automatically when the authorization flow completes for the first
	// time.
	tok, err := gs.tokenFromFile(gs.tokenPath)
	if err != nil {
		tok, err = gs.getTokenFromWeb(config)
		if err != nil {
			return nil, err
		}

		err = gs.saveToken(gs.tokenPath, tok)
		if err != nil {
			return nil, err
		}
	}
	return config.Client(context.Background(), tok), nil
}

// Request a token from the web, then returns the retrieved token.
func (SheetsService) getTokenFromWeb(config *oauth2.Config) (*oauth2.Token, error) {
	authURL := config.AuthCodeURL("state-token", oauth2.AccessTypeOffline)
	log.Printf("Go to the following link in your browser then type the "+
		"authorization code: \n%v\n", authURL)

	var authCode string
	if _, err := fmt.Scan(&authCode); err != nil {
		return nil, fmt.Errorf("unable to read authorization code: %w", err)
	}

	tok, err := config.Exchange(context.TODO(), authCode)
	if err != nil {
		fmt.Errorf("unable to retrieve token from web: %w", err)
	}
	return tok, err
}

// Retrieves a token from a local file.
func (SheetsService) tokenFromFile(file string) (*oauth2.Token, error) {
	f, err := os.Open(file)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	tok := &oauth2.Token{}
	err = json.NewDecoder(f).Decode(tok)
	return tok, err
}

func (SheetsService) saveToken(path string, token *oauth2.Token) error {
	log.Printf("Saving credential file to: %s\n", path)
	f, err := os.OpenFile(path, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		return fmt.Errorf("unable to cache oauth token: %v", err)
	}
	defer f.Close()

	err = json.NewEncoder(f).Encode(token)
	if err != nil {
		return fmt.Errorf("encode token failed: %v", err)
	}

	return nil
}

// func (gs SheetsService) read(spreadsheetId, readRange string) [][]interface{} {
//	ctx := context.Background()
//	b, err := os.ReadFile(gs.credentialsPath)
//	if err != nil {
//		log.Fatalf("Unable to read client secret file: %v", err)
//	}
//
//	// If modifying these scopes, delete your previously saved token.json.
//	config, err := google.ConfigFromJSON(b, "https://www.googleapis.com/auth/spreadsheets.readonly")
//	if err != nil {
//		log.Fatalf("Unable to parse client secret file to config: %v", err)
//	}
//	client := gs.getClient(config)
//
//	srv, err := sheets.NewService(ctx, option.WithHTTPClient(client))
//	if err != nil {
//		log.Fatalf("Unable to retrieve Sheets client: %v", err)
//	}
//
//	// Prints the names and majors of students in a sample spreadsheet:
//
//	// https://docs.google.com/spreadsheets/d/1_vD7wEx4ZaRdYn5pjJNelKAtzH7JA61TO2Q5QlVs0kQ/edit?usp=sharing
//	resp, err := srv.Spreadsheets.Values.Get(spreadsheetId, readRange).Do()
//	if err != nil {
//		log.Fatalf("Unable to retrieve data from sheet: %v", err)
//	}
//
//	if len(resp.Values) == 0 {
//		fmt.Println("No data found.")
//		return nil
//	} else {
//		return resp.Values
//	}
// }

func (gs SheetsService) Write(spreadsheetID, writeRange string, values [][]interface{}) error {
	ctx := context.Background()

	// Чтение файла с учетными данными клиента
	b, err := os.ReadFile(gs.credentialsPath)
	if err != nil {
		return fmt.Errorf("unable to read client secret file: %w", err)
	}

	// Настройка OAuth 2.0 конфигурации
	config, err := google.ConfigFromJSON(b, "https://www.googleapis.com/auth/spreadsheets")
	if err != nil {
		return fmt.Errorf("unable to parse client secret file to config: %w", err)
	}

	// Получение клиента OAuth 2.0
	client, _ := gs.getClient(config)

	// Создание сервиса для работы с Google Sheets
	srv, err := sheets.NewService(ctx, option.WithHTTPClient(client))
	if err != nil {
		return fmt.Errorf("unable to retrieve Sheets client: %w", err)
	}

	// Создание объекта ValueRange, который содержит данные для записи
	body := &sheets.ValueRange{
		Values: values,
	}

	// Вызов метода Update для записи данных
	_, err = srv.Spreadsheets.Values.Update(spreadsheetID, writeRange, body).
		ValueInputOption("RAW").Do()
	if err != nil {
		return fmt.Errorf("unable to update data in sheet: %w", err)
	}

	return nil
}
