package repository

import (
	"database/sql"
	"errors"
	"log"
	"time"

	"github.com/fatemehkarimi/chronos_bot/entities"
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
	GetScheduleByTime(
		calendarType entities.CalendarType,
		year int,
		month int,
		day int,
		startTime entities.CalendarTime,
		endTime entities.CalendarTime,
	) ([]entities.Schedule, error)
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
		feature_flag VARCHAR REFERENCES feature_flag(feature_flag) ON DELETE CASCADE,
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

func (repo *PostgresRepository) AddFeatureFlag(
	ownerId int,
	featureFlag string,
) error {
	query := `INSERT INTO feature_flag(owner_id, feature_flag, unix_time) VALUES ($1, $2, $3);`
	_, err := repo.DB.Exec(query, ownerId, featureFlag, time.Now().Unix())
	return err
}

func (repo *PostgresRepository) AddSchedule(schedule entities.Schedule) (
	int,
	error,
) {
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
		schedule.Calendar.Type,
		schedule.Calendar.Year,
		schedule.Calendar.Month,
		schedule.Calendar.Day,
		schedule.Calendar.Hour,
		schedule.Calendar.Minute,
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

func (repo *PostgresRepository) GetFeatureFlagByName(name string) (
	*entities.FeatureFlag,
	error,
) {
	query := `
	SELECT feature_flag, owner_id, unix_time FROM feature_flag WHERE feature_flag=$1;
	`

	var featureFlag entities.FeatureFlag
	err := repo.DB.QueryRow(query, name).Scan(
		&featureFlag.Name,
		&featureFlag.OwnerId,
		&featureFlag.UnixTime,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, err
		}
	}

	return &featureFlag, nil
}

func (repo *PostgresRepository) GetFeatureFlagsByOwnerId(ownerId int) (
	[]entities.FeatureFlag,
	error,
) {
	query := `
	SELECT feature_flag, owner_id, unix_time FROM feature_flag WHERE owner_id=$1;
	`
	var featureFlags []entities.FeatureFlag
	rows, err := repo.DB.Query(query, ownerId)
	defer func(rows *sql.Rows) {
		err := rows.Close()
		if err != nil {
			log.Fatal(err)
		}
	}(rows)

	if err != nil {
		return featureFlags, err
	}

	var featureFlag entities.FeatureFlag
	for rows.Next() {
		err := rows.Scan(
			&featureFlag.Name,
			&featureFlag.OwnerId,
			&featureFlag.UnixTime,
		)
		if err != nil {
			// todo: this is not really correct
			return featureFlags, err
		}
		featureFlags = append(featureFlags, featureFlag)
	}
	return featureFlags, nil
}

func (repo *PostgresRepository) GetScheduleByTime(
	calendarType entities.CalendarType,
	year int,
	month int,
	day int,
	startTime entities.CalendarTime,
	endTime entities.CalendarTime,
) ([]entities.Schedule, error) {
	query := `
	SELECT schedule_id, feature_flag, value, calendar_type, users_list, year, month, day, hour, minute, unix_time
	FROM schedule
	WHERE calendar_type = $1
	AND day = $2
	AND (year = 0 OR year = $3)
	AND (month = 0 OR month = $4)
	AND (
		(hour = $5 AND hour = $7 AND minute >= $6 AND minute <= $8)
		OR (hour = $5 AND minute >= $6)
		OR (hour > $5 AND hour < $7)
		OR (hour = $7 AND minute <= $8)
	)
	`

	var schedules []entities.Schedule
	rows, err := repo.DB.Query(
		query,
		calendarType,
		day,
		year,
		month,
		startTime.Hour,
		startTime.Minute,
		endTime.Hour,
		endTime.Minute,
	)
	defer func(rows *sql.Rows) {
		err := rows.Close()
		if err != nil {
			log.Fatal(err)
		}
	}(rows)

	if err != nil {
		return schedules, err
	}

	var schedule entities.Schedule
	for rows.Next() {
		err := rows.Scan(
			&schedule.ScheduleId,
			&schedule.FeatureFlagName,
			&schedule.Value,
			&schedule.Calendar.Type,
			&schedule.UsersList,
			&schedule.Calendar.Year,
			&schedule.Calendar.Month,
			&schedule.Calendar.Day,
			&schedule.Calendar.Hour,
			&schedule.Calendar.Minute,
			&schedule.UnixTime,
		)

		if err != nil {
			// todo: this is not really correct
			return schedules, err
		}

		schedules = append(schedules, schedule)
	}

	return schedules, nil
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
