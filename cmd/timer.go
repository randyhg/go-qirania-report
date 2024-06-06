package cmd

import (
	"context"
	"fmt"
	"github.com/robfig/cron"
	"go-qirania/config"
	"go-qirania/utils/milog"
	"google.golang.org/api/sheets/v4"
	"log"
	"net/http"
	"time"
)

func InitTimer(client *http.Client) {
	c := cron.New()

	err := c.AddFunc("0 0 * * *", func() { // 1 day
		createNextMonthSheet(client)
	})

	//err := c.AddFunc("* * * * *", func() { // 1 minute
	//	createNextMonthSheet(client)
	//})

	if err != nil {
		milog.Fatal("Error adding cron job:", err)
	}

	milog.Info("cron job created successfully!")
	c.Start()
	select {}
}

func createNextMonthSheet(client *http.Client) {
	spreadSheetId := config.Conf.SpreadSheetId
	tempSheetId := config.Conf.TemplateSheetId

	nextMonthSheetName := (time.Now().Month() + 1).String()

	srv, err := sheets.New(client)
	if err != nil {
		milog.Fatalf("Unable to retrieve Sheets client: %s", err.Error())
	}

	exists, _, err := sheetExists(srv, spreadSheetId, nextMonthSheetName)
	if err != nil {
		log.Fatalf("Error checking if sheet exists: %v", err)
	}

	if !exists {
		// Duplicate the sheet
		duplicateSheetRequest := &sheets.Request{
			DuplicateSheet: &sheets.DuplicateSheetRequest{
				SourceSheetId: tempSheetId,
				NewSheetName:  nextMonthSheetName,
			},
		}

		batchUpdateRequest := &sheets.BatchUpdateSpreadsheetRequest{
			Requests: []*sheets.Request{duplicateSheetRequest},
		}

		_, err = srv.Spreadsheets.BatchUpdate(spreadSheetId, batchUpdateRequest).Context(context.Background()).Do()
		if err != nil {
			log.Fatalf("Unable to duplicate sheet: %v", err)
		}

		milog.Infof("Duplicated sheet with name: %s\n", nextMonthSheetName)
	}
}

func sheetExists(srv *sheets.Service, spreadsheetId string, sheetName string) (bool, int64, error) {
	// Get spreadsheet metadata
	spreadsheet, err := srv.Spreadsheets.Get(spreadsheetId).Fields("sheets(properties(sheetId,title))").Do()
	if err != nil {
		return false, 0, fmt.Errorf("unable to retrieve spreadsheet: %v", err)
	}

	// Check if the sheet exists
	for _, sheet := range spreadsheet.Sheets {
		if sheet.Properties.Title == sheetName {
			return true, sheet.Properties.SheetId, nil
		}
	}
	return false, 0, nil
}
