package repository

import (
	"database/sql"
	"errors"
	"github.com/fatemehkarimi/chronos_bot/entities"
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
	Init() error
	CreateTableFeatureFlag() error
	CreateTableSchedule() error
	AddFeatureFlag(ownerId int, featureFlag string) error
	AddSchedule(schedule entities.Schedule) (int, error)
	RemoveSchedule(scheduleId int) error
	GetFeatureFlagByName(name string) (*entities.FeatureFlag, error)
	GetFeatureFlagsByOwnerId(ownerId int) ([]entities.FeatureFlag, error)
}

type PostgresRepository struct {
	DB *sql.DB
}

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
		unix_time BIGINT
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
	    users_list TEXT,
		year INT,
		month INT,
		day INT,
		hour INT,
		minute INT,
		unix_time BIGINT
);`
	_, err := repo.DB.Exec(query)
	return err
}

func (repo *PostgresRepository) AddFeatureFlag(ownerId int, featureFlag string) error {
	query := `INSERT INTO feature_flag(owner_id, feature_flag, unix_time) VALUES ($1, $2, $3);`
	_, err := repo.DB.Exec(query, ownerId, featureFlag, time.Now().Unix())
	return err
}

func (repo *PostgresRepository) AddSchedule(schedule entities.Schedule) (int, error) {
	query := `
	INSERT INTO schedule(
	 	feature_flag,
	 	value,
	 	users_list,
	 	calendar_type,
	 	year,
	 	month,
	 	day,
	 	hour,
	 	minute,
	 	unix_time
	) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10) RETURNING schedule_id`
	var scheduleId int

	err := repo.DB.QueryRow(
		query,
		schedule.FeatureFlagName,
		schedule.Value,
		schedule.UsersList,
		schedule.CalendarType,
		schedule.Year,
		schedule.Month,
		schedule.Day,
		schedule.Hour,
		schedule.Minute,
		time.Now().Unix(),
	).Scan(&scheduleId)
	return scheduleId, err
}

func (repo *PostgresRepository) RemoveSchedule(scheduleId int) error {
	query := `
	DELETE FROM schedule where schedule_id=$1
	`
	_, err := repo.DB.Exec(query, scheduleId)
	return err
}

func (repo *PostgresRepository) GetFeatureFlagByName(name string) (*entities.FeatureFlag, error) {
	query := `
	SELECT feature_flag, owner_id, unix_time FROM feature_flag WHERE feature_flag=$1;
	`

	var featureFlag entities.FeatureFlag
	err := repo.DB.QueryRow(query, name).Scan(&featureFlag.Name, &featureFlag.OwnerId, &featureFlag.UnixTime)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, err
		}
	}

	return &featureFlag, nil
}

func (repo *PostgresRepository) GetFeatureFlagsByOwnerId(ownerId int) ([]entities.FeatureFlag, error) {
	query := `
	SELECT feature_flag, owner_id, unix_time FROM feature_flag WHERE owner_id=$1;
	`
	var featureFlags []entities.FeatureFlag
	rows, err := repo.DB.Query(query, ownerId)
	defer rows.Close()

	if err != nil {
		return featureFlags, err
	}

	var featureFlag entities.FeatureFlag
	for rows.Next() {
		err := rows.Scan(&featureFlag.Name, &featureFlag.OwnerId, &featureFlag.UnixTime)
		if err != nil {
			return featureFlags, err
		}
		featureFlags = append(featureFlags, featureFlag)
	}
	return featureFlags, nil
}

func (repo *PostgresRepository) Init() error {
	err := repo.CreateTableFeatureFlag()

	if err != nil {
		return err
	}

	err = repo.CreateTableSchedule()
	if err != nil {
		return err
	}
	return nil
}
