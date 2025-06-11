package repository

import (
	"database/sql"
	"time"
)

type DatabaseConfig struct {
	User     string
	Password string
	Host     string
	Port     int
	Name     string
	DBName   string
}

type Repository interface {
	CreateTableFeatureFlag() error
	CreateTableSchedule() error
	AddFeatureFlag(ownerId int, featureFlag string) error
	AddSchedule(featureFlag, value string, calendarType CalendarType, year, month, day, hour, minute int) (int, error)
	RemoveSchedule(scheduleId int) error
}

type PostgresRepository struct {
	DB *sql.DB
}

type CalendarType int

const (
	_ CalendarType = iota
	Shamsi
	Qama
)

func CreateNewRepository(db *sql.DB) Repository {
	return &PostgresRepository{
		DB: db,
	}
}

func (repo *PostgresRepository) CreateTableFeatureFlag() error {
	query := `
	CREATE TABLE IF NOT EXISTS feature_flag(
		feature_flag VARCHAR PRIMARY KEY,
		owner_id INT,
		time TIMESTAMP
	);`

	_, err := repo.DB.Exec(query)
	return err
}

func (repo *PostgresRepository) CreateTableSchedule() error {
	query := `
	CREATE TABLE IF NOT EXISTS schedule(
		schedule_id SERIAL PRIMARY KEY,
		feature_flag VARCHAR REFERENCES feature_flag(feature_flag),
		value TEXT,
		calendar_type SMALLINT,
		year INT,
		month INT,
		day INT,
		hour INT,
		minute INT,
		time TIMESTAMP,
);`
	_, err := repo.DB.Exec(query)
	return err
}

func (repo *PostgresRepository) AddFeatureFlag(ownerId int, featureFlag string) error {
	query := `INSERT INTO feature_flag(owner_id, feature_flag, time) VALUES ($1, $2, $3);`
	_, err := repo.DB.Exec(query, ownerId, featureFlag, time.Now())
	return err
}

func (repo *PostgresRepository) AddSchedule(featureFlag, value string, calendarType CalendarType, year, month, day, hour, minute int) (int, error) {
	query := `
	INSERT INTO schedule(feature_flag, value, calendar_type, year, month, day, hour, minute, time) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)`
	var scheduleId int
	err := repo.DB.QueryRow(query, featureFlag, value, calendarType, year, month, day, hour, minute, time.Now()).Scan(&scheduleId)
	return scheduleId, err
}

func (repo *PostgresRepository) RemoveSchedule(scheduleId int) error {
	query := `
	DELETE FROM schedule where schedule_id=$1
	`
	_, err := repo.DB.Exec(query, scheduleId)
	return err
}
