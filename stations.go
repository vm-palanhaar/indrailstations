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

func (stationDb StationDb) updateRailStationDb(db *sql.DB, stationApi StationApi) {
	table := os.Getenv("TABLE")
	if stationApi.Code == stationDb.Code && stationApi.Name == stationDb.Name {
		// DO NOTHING
	} else if stationApi.Code == stationDb.Code && stationApi.Name != stationDb.Name {
		stmt, err := db.Prepare(
			fmt.Sprintf("UPDATE %s SET %s = $1 WHERE %s = $2 AND %s = $3",
				table, "name", "id", "code"))
		if err != nil {
			log.Fatal(err)
		}
		defer stmt.Close()
		_, err = stmt.Exec(stationApi.Name, stationDb.Id, stationDb.Code)
		if err != nil {
			log.Fatal(err)
		}
		log.Printf("---UPDATE SUCCESS [%s - %s]---", stationApi.Name, stationApi.Code)
	}
}

func railStations(db *sql.DB) {
	log.Println("START -> GET ALL rail stations API functionality <- START")
	stationsApi := getRailStationsApi()
	log.Println("END -> GET ALL rail stations API functionality <- END")

	log.Println("START -> GET ALL rail stations DB functionality <- START")
	stationsDb := getRailStationsDb(db)
	log.Println("END -> GET ALL rail stations DB functionality <- END")

	/*
		Iterate through each station to check for any changes in station name.
		1. If unchanged, do nothing
		2. If changed [station name], update station name w.r.t to station code
	*/
	if len(stationsApi) == len(stationsDb) {
		for _, sapi := range stationsApi {
			for _, sdb := range stationsDb {
				sdb.updateRailStationDb(db, sapi)
			}
		}
	} else if len(stationsApi) < len(stationsDb) {
		/*
			Send mail to admin user(s)
		*/
		stations := []string{}
		for _, sdb := range stationsDb {
			checkStation := true
			for _, sapi := range stationsApi {
				sdb.updateRailStationDb(db, sapi)
				if sapi.Code == sdb.Code {
					checkStation = false
					break
				}
			}
			if checkStation {
				stations = append(stations, fmt.Sprintf("%s - %s", sdb.Name, sdb.Name))
			}
		}
		log.Print("<---Rail Stations in DB--->")
		for i, station := range stations {
			log.Printf("%d. %s", i+1, station)
		}
	} else if len(stationsApi) > len(stationsDb) {
		table := os.Getenv("TABLE")
		log.Print("<---Rail Stations in API--->")
		for i, sapi := range stationsApi {
			checkStation := true
			for _, sdb := range stationsDb {
				sdb.updateRailStationDb(db, sapi)
				if sapi.Code == sdb.Code {
					checkStation = false
					break
				}
			}
			if checkStation {
				stmt, err := db.Prepare(
					fmt.Sprintf("INSERT INTO %s (%s, %s) VALUES ($1, $2)",
						table, "code", "name"))
				if err != nil {
					log.Fatal(err)
				}
				defer stmt.Close()
				_, err = stmt.Exec(sapi.Code, sapi.Name)
				if err != nil {
					log.Fatal(err)
				}
				log.Printf("%d. INSERT SUCCESS [%s - %s]", i+1, sapi.Name, sapi.Code)
			}
		}
	}
}
