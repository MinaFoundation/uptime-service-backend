package delegation_backend

import (
	logging "github.com/ipfs/go-log/v2"
	sheets "google.golang.org/api/sheets/v4"
)

// Process rows retrieved from Google spreadsheet
// and extract public keys from the first column.
func processRows(rows [][](interface{})) Whitelist {
	wl := make(Whitelist)
	for _, row := range rows {
		if len(row) > 0 {
			switch v := row[0].(type) {
			case string:
				var pk Pk
				err := StringToPk(&pk, v)
				if err == nil {
					wl[pk] = true // we need something to be provided as value
				}
			}
		}
	}
	return wl
}

// Retrieve data from delegation program spreadsheet
// and extract public keys out of the column containing
// public keys of program participants.
func RetrieveWhitelist(service *sheets.Service, log *logging.ZapEventLogger, appCfg AppConfig, retries int) (Whitelist, error) {
	var resp *sheets.ValueRange
	var err error

	operation := func() error {
		col := appCfg.DelegationWhitelistColumn
		readRange := appCfg.DelegationWhitelistList + "!" + col + ":" + col
		spId := appCfg.GsheetId
		resp, err = service.Spreadsheets.Values.Get(spId, readRange).Do()
		if err != nil {
			return err
		}
		return nil
	}
	err = ExponentialBackoff(operation, retries, initialBackoff)
	if err != nil {
		log.Errorf("Unable to retrieve data from sheet after %v retries: %v", retries, err)
		return nil, err
	}

	return processRows(resp.Values), nil
}
