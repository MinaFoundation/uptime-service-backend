package itn_uptime_analyzer

import (
	"fmt"
	"strings"
	"time"

	logging "github.com/ipfs/go-log/v2"
	sheets "google.golang.org/api/sheets/v4"
)

// This function returns true if the identity is present
// It also returns the row index
// If the identity is not present it returns false and searches for the closest relative of the identity (same pubkey or same ip)
// If nothing was found it returns false and 0 as the row index
func (identity Identity) GetCell(config AppConfig, client *sheets.Service, log *logging.ZapEventLogger, sheetTitle string) (exactMatch bool, rowIndex int, firstEmptyRow int) {
	exactMatch = false
	rowIndex = 0
	firstEmptyRow = 1
	col := IDENTITY_COLUMN
	readRange := sheetTitle + "!" + col + ":" + col
	spId := config.AnalyzerOutputGsheetId
	var identityString string
	resp, err := client.Spreadsheets.Values.Get(spId, readRange).Do()
	if err != nil {
		log.Fatalf("Unable to retrieve data from sheet: %v", err)
	}

	if identity["graphql-port"] != "" {
		identityString = strings.Join([]string{identity["public-key"], identity["public-ip"], identity["graphql-port"]}, "-")
	} else {
		identityString = strings.Join([]string{identity["public-key"], identity["public-ip"]}, "-")
	}
	for index, row := range resp.Values {
		if row[0] == identityString {
			rowIndex = index + 1
			exactMatch = true
			break
		}
		firstEmptyRow = firstEmptyRow + 1
	}

	if !exactMatch {
		for index, row := range resp.Values {
			str := fmt.Sprintf("%v\n", row[0])
			if strings.Split(str, "-")[0] == identity["public-key"] {
				rowIndex = index + 1
				exactMatch = false
			}
		}
	}

	return exactMatch, rowIndex, firstEmptyRow
}

// Appends the identity string of the node to the first column
func (identity Identity) AppendNext(config AppConfig, client *sheets.Service, log *logging.ZapEventLogger, sheetTitle string) {
	col := IDENTITY_COLUMN
	readRange := sheetTitle + "!" + col + ":" + col
	spId := config.AnalyzerOutputGsheetId
	var identityString string

	if identity["graphql-port"] != "" {
		identityString = strings.Join([]string{identity["public-key"], identity["public-ip"], identity["graphql-port"]}, "-")
	} else {
		identityString = strings.Join([]string{identity["public-key"], identity["public-ip"]}, "-")
	}

	cellValue := []interface{}{identityString}

	valueRange := sheets.ValueRange{
		Values: [][]interface{}{cellValue},
	}

	_, err := client.Spreadsheets.Values.Append(spId, readRange, &valueRange).ValueInputOption("USER_ENTERED").Do()
	if err != nil {
		log.Fatalf("Unable to append data to sheet: %v\n", err)
	}
}

// Inserts the identity string of the node in the first column under rowIndex
func (identity Identity) InsertBelow(config AppConfig, client *sheets.Service, log *logging.ZapEventLogger, sheetTitle string, rowIndex int) {
	col := IDENTITY_COLUMN
	readRange := fmt.Sprintf("%s!%s%d:%s%d", sheetTitle, col, rowIndex+1, col, rowIndex+1)
	spId := config.AnalyzerOutputGsheetId
	var identityString string

	if identity["graphql-port"] != "" {
		identityString = strings.Join([]string{identity["public-key"], identity["public-ip"], identity["graphql-port"]}, "-")
	} else {
		identityString = strings.Join([]string{identity["public-key"], identity["public-ip"]}, "-")
	}

	cellValue := []interface{}{identityString}

	valueRange := sheets.ValueRange{
		Values: [][]interface{}{cellValue},
	}

	_, err := client.Spreadsheets.Values.Append(spId, readRange, &valueRange).ValueInputOption("USER_ENTERED").InsertDataOption("INSERT_ROWS").Do()
	if err != nil {
		log.Fatalf("Unable to insert data in sheet: %v\n", err)
	}
}

// Finds the first empty column on the row specified and puts the percentage of uptime of the identity
func (identity Identity) AppendUptime(config AppConfig, client *sheets.Service, log *logging.ZapEventLogger, sheetTitle string, rowIndex int) {
	readRange := fmt.Sprintf("%s!A%d:Z%d", sheetTitle, 1, 1)
	spId := config.AnalyzerOutputGsheetId

	resp, err := client.Spreadsheets.Values.Get(spId, readRange).Do()
	if err != nil {
		log.Fatalf("Unable to retrieve data from sheet: %v\n", err)
	}

	var nextEmptyColumn int = len(resp.Values[0])

	updateRange := fmt.Sprintf("%s!%s%d", sheetTitle, string(nextEmptyColumn+65), rowIndex)

	cellValue := []interface{}{identity["uptime"]}

	valueRange := sheets.ValueRange{
		Values: [][]interface{}{cellValue},
	}

	_, err = client.Spreadsheets.Values.Append(spId, updateRange, &valueRange).ValueInputOption("USER_ENTERED").Do()
	if err != nil {
		log.Fatalf("Unable to insert data in sheet: %v\n", err)
	}
}

// Creates a sheet with the title sheetTitle
func CreateSheet(config AppConfig, client *sheets.Service, log *logging.ZapEventLogger, sheetTitle string) error {
	spId := config.AnalyzerOutputGsheetId

	// Prepare the request to add a new sheet
	req := &sheets.BatchUpdateSpreadsheetRequest{
		Requests: []*sheets.Request{
			{
				AddSheet: &sheets.AddSheetRequest{
					Properties: &sheets.SheetProperties{
						Title: sheetTitle,
					},
				},
			},
		},
	}

	// Execute the request
	_, err := client.Spreadsheets.BatchUpdate(spId, req).Do()
	if err != nil {
		log.Fatalf("failed to create sheet: %v", err)
	}

	updateRange := fmt.Sprintf("%s!%s%d", sheetTitle, IDENTITY_COLUMN, 1)
	cellValue := []interface{}{"Node Identity ↓ Execution Time Window →"}

	valueRange := sheets.ValueRange{
		Values: [][]interface{}{cellValue},
	}

	_, err = client.Spreadsheets.Values.Append(spId, updateRange, &valueRange).ValueInputOption("USER_ENTERED").Do()
	if err != nil {
		log.Fatalf("Unable to insert data in sheet: %v\n", err)
	}

	return nil
}

// Returns the sheets of the specified spreadsheet
func GetSheets(config AppConfig, client *sheets.Service, log *logging.ZapEventLogger) ([]*sheets.Sheet, error) {
	// Retrieve the spreadsheet
	spreadsheet, err := client.Spreadsheets.Get(config.AnalyzerOutputGsheetId).Do()
	if err != nil {
		log.Fatalf("failed to retrieve spreadsheet: %v\n", err)
	}

	// Get the sheets from the spreadsheet
	sheets := spreadsheet.Sheets

	return sheets, nil
}

// Tracks the date of execution on the top row of the spreadsheet
func MarkExecution(config AppConfig, client *sheets.Service, log *logging.ZapEventLogger, sheetTitle string, currentTime time.Time, executionInterval int) {
	readRange := fmt.Sprintf("%s!A%d:Z%d", sheetTitle, 1, 1)
	spId := config.AnalyzerOutputGsheetId

	lastExecutionTime := GetLastExecutionTime(config, client, log, sheetTitle, currentTime, executionInterval)

	timeInterval := strings.Join([]string{lastExecutionTime.Format(time.RFC3339), currentTime.Format(time.RFC3339)}, " - ")

	resp, err := client.Spreadsheets.Values.Get(spId, readRange).Do()
	if err != nil {
		log.Fatalf("Unable to retrieve data from sheet: %v\n", err)
	}

	var nextEmptyColumn int = len(resp.Values[0])

	updateRange := fmt.Sprintf("%s!%s%d", sheetTitle, string(nextEmptyColumn+65), 1)

	cellValue := []interface{}{timeInterval}

	valueRange := sheets.ValueRange{
		Values: [][]interface{}{cellValue},
	}

	_, err = client.Spreadsheets.Values.Append(spId, updateRange, &valueRange).ValueInputOption("USER_ENTERED").Do()
	if err != nil {
		log.Fatalf("Unable to insert data in sheet: %v\n", err)
	}
}
