package googleService

import (
	"context"
	"encoding/json"
	"fmt"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/option"
	"google.golang.org/api/sheets/v4"
	"log"
	"net/http"
	"os"
)

type GoogleService struct {
	tokenPath       string
	credentialsPath string
}

func NewGoogleService(tokenPath, credentialsPath string) GoogleService {
	return GoogleService{
		tokenPath:       tokenPath,
		credentialsPath: credentialsPath,
	}
}

// Retrieve a token, saves the token, then returns the generated client.
func (gs GoogleService) getClient(config *oauth2.Config) *http.Client {
	// The file token.json stores the user's access and refresh tokens, and is
	// created automatically when the authorization flow completes for the first
	// time.
	tok, err := gs.tokenFromFile(gs.tokenPath)
	if err != nil {
		tok = gs.getTokenFromWeb(config)
		gs.saveToken(gs.tokenPath, tok)
	}
	return config.Client(context.Background(), tok)
}

// Request a token from the web, then returns the retrieved token.
func (GoogleService) getTokenFromWeb(config *oauth2.Config) *oauth2.Token {
	authURL := config.AuthCodeURL("state-token", oauth2.AccessTypeOffline)
	fmt.Printf("Go to the following link in your browser then type the "+
		"authorization code: \n%v\n", authURL)

	var authCode string
	if _, err := fmt.Scan(&authCode); err != nil {
		log.Fatalf("Unable to read authorization code: %v", err)
	}

	tok, err := config.Exchange(context.TODO(), authCode)
	if err != nil {
		log.Fatalf("Unable to retrieve token from web: %v", err)
	}
	return tok
}

// Retrieves a token from a local file.
func (GoogleService) tokenFromFile(file string) (*oauth2.Token, error) {
	f, err := os.Open(file)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	tok := &oauth2.Token{}
	err = json.NewDecoder(f).Decode(tok)
	return tok, err
}

func (GoogleService) saveToken(path string, token *oauth2.Token) {
	fmt.Printf("Saving credential file to: %s\n", path)
	f, err := os.OpenFile(path, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		log.Fatalf("Unable to cache oauth token: %v", err)
	}
	defer f.Close()
	json.NewEncoder(f).Encode(token)
}

func (gs GoogleService) read(spreadsheetId, readRange string) [][]interface{} {
	ctx := context.Background()
	b, err := os.ReadFile(gs.credentialsPath)
	if err != nil {
		log.Fatalf("Unable to read client secret file: %v", err)
	}

	// If modifying these scopes, delete your previously saved token.json.
	config, err := google.ConfigFromJSON(b, "https://www.googleapis.com/auth/spreadsheets.readonly")
	if err != nil {
		log.Fatalf("Unable to parse client secret file to config: %v", err)
	}
	client := gs.getClient(config)

	srv, err := sheets.NewService(ctx, option.WithHTTPClient(client))
	if err != nil {
		log.Fatalf("Unable to retrieve Sheets client: %v", err)
	}

	// Prints the names and majors of students in a sample spreadsheet:

	// https://docs.google.com/spreadsheets/d/1_vD7wEx4ZaRdYn5pjJNelKAtzH7JA61TO2Q5QlVs0kQ/edit?usp=sharing
	resp, err := srv.Spreadsheets.Values.Get(spreadsheetId, readRange).Do()
	if err != nil {
		log.Fatalf("Unable to retrieve data from sheet: %v", err)
	}

	if len(resp.Values) == 0 {
		fmt.Println("No data found.")
		return nil
	} else {
		return resp.Values
	}
}

func (gs GoogleService) Write(spreadsheetId, writeRange string, values [][]interface{}) error {
	ctx := context.Background()

	// Чтение файла с учетными данными клиента
	b, err := os.ReadFile(gs.credentialsPath)
	if err != nil {
		return fmt.Errorf("Unable to read client secret file: %v", err)
	}

	// Настройка OAuth 2.0 конфигурации
	config, err := google.ConfigFromJSON(b, "https://www.googleapis.com/auth/spreadsheets")
	if err != nil {
		return fmt.Errorf("Unable to parse client secret file to config: %v", err)
	}

	// Получение клиента OAuth 2.0
	client := gs.getClient(config)

	// Создание сервиса для работы с Google Sheets
	srv, err := sheets.NewService(ctx, option.WithHTTPClient(client))
	if err != nil {
		return fmt.Errorf("Unable to retrieve Sheets client: %v", err)
	}

	// Создание объекта ValueRange, который содержит данные для записи
	body := &sheets.ValueRange{
		Values: values,
	}

	// Вызов метода Update для записи данных
	_, err = srv.Spreadsheets.Values.Update(spreadsheetId, writeRange, body).
		ValueInputOption("RAW").Do()
	if err != nil {
		return fmt.Errorf("Unable to update data in sheet: %v", err)
	}

	fmt.Println("Data written successfully")
	return nil
}
