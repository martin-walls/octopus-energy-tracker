package store

import (
	"database/sql"
	"fmt"
	"log"
	"martin-walls/octopus-energy-tracker/internal/octopus"
	"strings"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

const dbPath = "./db.sqlite"

type Store struct {
	db *sql.DB
}

type reading struct {
	timestamp        string
	totalConsumption int
	consumptionDelta int
	demand           int
}

func NewStore() *Store {
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		log.Fatal("Error connecting to DB: ", err)
	}

	err = setupDb(db)
	if err != nil {
		log.Fatal("Error setting up DB: ", err)
	}

	return &Store{
		db: db,
	}
}

func setupDb(db *sql.DB) error {
	schemaStmt := `
		CREATE TABLE IF NOT EXISTS readings (
			timestamp TEXT PRIMARY KEY,
			total_consumption INTEGER NOT NULL,
			consumption_delta INTEGER NOT NULL,
			demand INTEGER NOT NULL
		);
	`

	_, err := db.Exec(schemaStmt)
	if err != nil {
		return fmt.Errorf("%w: %s", err, schemaStmt)
	}

	return nil
}

func (s *Store) InsertReadings(rs []*octopus.ConsumptionReading) error {
	insertStmt := `
		INSERT INTO readings (timestamp, total_consumption, consumption_delta, demand)
		VALUES 
	`
	values := []any{}

	for _, reading := range rs {
		insertStmt += "(?, ?, ?, ?),"
		values = append(
			values,
			reading.Timestamp.Format(time.RFC3339),
			reading.TotalConsumption,
			reading.ConsumptionDelta,
			reading.Demand)
	}

	insertStmt = strings.TrimSuffix(insertStmt, ",")

	res, err := s.db.Exec(insertStmt, values...)
	if err != nil {
		return fmt.Errorf("InsertReadings: %v", err)
	}

	rowCount, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("InsertReadings: %v", err)
	}

	log.Printf("Inserted %v readings into DB", rowCount)
	return nil
}

// TODO: since timestamp
func (s *Store) Readings() ([]*octopus.ConsumptionReading, error) {
	var readings []*octopus.ConsumptionReading

	rows, err := s.db.Query("SELECT * FROM readings")
	if err != nil {
		return nil, fmt.Errorf("Readings: %v", err)
	}
	defer rows.Close()

	for rows.Next() {
		var r reading

		err = rows.Scan(&r.timestamp, &r.totalConsumption, &r.consumptionDelta, &r.demand)
		if err != nil {
			return nil, fmt.Errorf("Readings: %v", err)
		}

		timestamp, err := time.Parse(time.RFC3339, r.timestamp)
		if err != nil {
			return nil, fmt.Errorf("Readings: %v", err)
		}

		readings = append(readings, &octopus.ConsumptionReading{
			Timestamp:        timestamp,
			TotalConsumption: r.totalConsumption,
			ConsumptionDelta: r.consumptionDelta,
			Demand:           r.demand,
		})
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("Reading: %v", err)
	}

	return readings, nil
}

func (s *Store) Close() {
	s.db.Close()
}
