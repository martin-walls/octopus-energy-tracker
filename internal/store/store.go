package store

import (
	"database/sql"
	"errors"
	"fmt"
	"log"
	"martin-walls/octopus-energy-tracker/internal/octopus"
	"strings"
	"time"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/sqlite3"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	_ "github.com/mattn/go-sqlite3"
)

const dbPath = "./db.sqlite"
const migrationsPath = "./migrations"

type Store struct {
	db *sql.DB
}

type reading struct {
	timestamp        string
	totalConsumption int
	demand           int
}

func NewStore() *Store {
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		log.Fatal("NewStore: Error connecting to DB: ", err)
	}

	err = runMigrations(db)
	if err != nil {
		log.Fatal("NewStore: ", err)
	}

	return &Store{
		db: db,
	}
}

func runMigrations(db *sql.DB) error {
	driver, err := sqlite3.WithInstance(db, &sqlite3.Config{})
	if err != nil {
		return fmt.Errorf("migrate: %v", err)
	}

	m, err := migrate.NewWithDatabaseInstance(
		"file://"+migrationsPath,
		"sqlite3",
		driver)
	if err != nil {
		return fmt.Errorf("migrate: %v", err)
	}

	err = m.Up()
	if err != nil {
		if !errors.Is(err, migrate.ErrNoChange) {
			return fmt.Errorf("migrate: %v", err)
		}
	}

	return nil
}

func (s *Store) InsertReadings(rs []*octopus.ConsumptionReading) error {
	insertStmt := `
		INSERT INTO readings (timestamp, total_consumption, demand)
		VALUES 
	`
	values := []any{}

	for _, reading := range rs {
		insertStmt += "(?, ?, ?),"
		values = append(
			values,
			reading.Timestamp.Format(time.RFC3339),
			reading.TotalConsumption,
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

		err = rows.Scan(&r.timestamp, &r.totalConsumption, &r.demand)
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
