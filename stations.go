package main

import (
	"database/sql"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
)

type StationApi struct {
	Name string
	Code string
}

type StationDb struct {
	Id   string
	Code string
	Name string
}

func getRailStationsApi() []StationApi {
	// calling rail stations API
	log.Printf("Requesting data from API")
	req, err := http.Get(os.Getenv("INDRAIL_STATIONS_API"))
	if err != nil {
		log.Fatal(err)
	}

	// response status code check 200
	log.Printf("Response status code: %s", req.Status)
	if req.StatusCode != http.StatusOK {
		log.Fatalf("API failed: %s", req.Status)
	}

	// read doc from response
	body, err := io.ReadAll(req.Body)
	if err != nil {
		log.Fatal(err)
	}
	docContentString := string(body)

	// format string to create a list of stations obj
	docContentString = strings.ReplaceAll(docContentString, "[", "")
	docContentString = strings.ReplaceAll(docContentString, "]", "")
	docContentString = strings.ReplaceAll(docContentString, "\"", "")
	railStations := strings.Split(docContentString, ",")
	log.Printf("Total no. of rail stations (API): %d", len(railStations))
	railStationList := make([]StationApi, len(railStations))
	for i, station := range railStations {
		railStationList[i].Name = strings.Split(station, " - ")[0]
		railStationList[i].Code = strings.Split(station, " - ")[1]
	}
	return railStationList
}

func getRailStationsDb(db *sql.DB) []StationDb {
	table := os.Getenv("TABLE")
	// execute query
	log.Println("EXECUTING QUERY to fetch id, station code, station name")
	rows, err := db.Query(fmt.Sprintf("SELECT id, code, name FROM %s", table))
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()

	stationsDb := []StationDb{}
	for rows.Next() {
		rowData := StationDb{}
		err := rows.Scan(&rowData.Id, &rowData.Code, &rowData.Name)
		if err != nil {
			log.Fatal(err)
		}
		stationsDb = append(stationsDb, rowData)
	}
	log.Printf("Total no. of rail stations (DB): %d", len(stationsDb))
	return stationsDb
}

func railStations(db *sql.DB) {
	log.Println("START -> GET ALL rail stations API functionality <- START")
	stationsApi := getRailStationsApi()
	log.Println("END -> GET ALL rail stations API functionality <- END")

	log.Println("START -> GET ALL rail stations DB functionality <- START")
	stationsDb := getRailStationsDb(db)
	log.Println("END -> GET ALL rail stations DB functionality <- END")
}
