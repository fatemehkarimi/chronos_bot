package main

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"time"

	"github.com/fatemehkarimi/chronos_bot/handler"
	"github.com/fatemehkarimi/chronos_bot/scheduler"

	"github.com/fatemehkarimi/chronos_bot/api"
	"github.com/fatemehkarimi/chronos_bot/entities"
	"github.com/fatemehkarimi/chronos_bot/repository"
	_ "github.com/lib/pq"
	"github.com/spf13/viper"
)

type Config struct {
	Database   repository.DatabaseConfig
	BotToken   string
	LogChannel string
}

func LoadConfig() (Config, error) {
	var cfg Config
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(".")

	if err := viper.ReadInConfig(); err != nil {
		return cfg, err
	}

	err := viper.Unmarshal(&cfg)
	return cfg, err
}

func main() {
	slog.SetLogLoggerLevel(slog.LevelDebug)
	config, err := LoadConfig()
	if err != nil {
		os.Exit(-1)
	}

	connectionCredentials := fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
		config.Database.Host,
		config.Database.Port,
		config.Database.User,
		config.Database.Password,
		config.Database.DBName,
	)

	db, err := sql.Open("postgres", connectionCredentials)
	if err != nil {
		slog.Error("failed to open db", slog.Any("err", err))
		os.Exit(1)
	}
	defer func(db *sql.DB) {
		err := db.Close()
		if err != nil {
			slog.Error("failed to close db", slog.Any("err", err))
		}
	}(db)

	err = db.Ping()
	if err != nil {
		slog.Error("failed to ping database. err = ", slog.Any("err", err))
		os.Exit(1)
	}

	postgresRepo := repository.PostgresRepository{DB: db}
	err = postgresRepo.Init()

	if err != nil {
		slog.Error("failed to init service. error = ", slog.Any("err", err))
		os.Exit(1)
	}

	baleApi := api.NewBaleApi(config.BotToken)
	awxScheduler := scheduler.NewScheduler(
		&postgresRepo,
		baleApi,
		config.LogChannel,
	)
	go RunDailyJob(awxScheduler)

	httpHandler := handler.NewHttpHandler(&postgresRepo, baleApi, awxScheduler)

	mux := http.NewServeMux()
	mux.HandleFunc("/getUpdates", httpHandler.GetUpdates)

	server := &http.Server{
		Addr:         ":8080",
		Handler:      mux,
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  120 * time.Second,
	}

	//go checkForUpdates(config.BotToken, httpHandler)
	err = server.ListenAndServe()
	if err != nil {
		os.Exit(1)
	}
}

func checkForUpdates(botToken string, handler handler.Handler) {
	client := &http.Client{}
	for {
		time.Sleep(5 * time.Second)
		requestStruct := entities.RequestGetUpdates{}
		requestBytes, err := json.MarshalIndent(requestStruct, "", "  ")

		if err != nil {
			slog.Error("error marshaling request", slog.Any("error", err))
		}
		lastUpdateId := handler.GetLastProcessedUpdateId() + 1
		slog.Info("fetching updates from ", slog.Int("updateId", lastUpdateId))
		limit := 1
		timeout := 1
		endpoint := fmt.Sprintf(
			"https://tapi.bale.ai/bot%s/getUpdates?offset=%d&limit=%d&timeout=%d",
			botToken,
			lastUpdateId,
			limit,
			timeout,
		)

		go func() {
			req, err := http.NewRequest(
				"POST",
				endpoint,
				bytes.NewBuffer(requestBytes),
			)

			if err != nil {
				slog.Error("error creating new request", slog.Any("error", err))
				return
			}

			res, err := client.Do(req)

			if err != nil {
				slog.Error(
					"error creating sending request",
					slog.Any("error", err),
				)
				return
			}
			defer res.Body.Close()

			slog.Info(
				"getUpdates response from tapi",
				slog.Int("status", res.StatusCode),
			)

			updateResponse := entities.ResponseGetUpdates{}
			err = json.NewDecoder(res.Body).Decode(&updateResponse)
			if err != nil {
				slog.Error("error parsing response", slog.Any("error", err))
				return
			}

			// sending updates
			for _, update := range updateResponse.Result {
				requestBytes, err := json.MarshalIndent(update, "", "  ")

				if err != nil {
					slog.Error(
						"error marshaling request",
						slog.Any("error", err),
					)
				}

				req, err := http.NewRequest(
					"POST",
					"http://localhost:8080/getUpdates",
					bytes.NewBuffer(requestBytes),
				)

				if err != nil {
					slog.Error(
						"error creating new request",
						slog.Any("error", err),
					)
					continue
				}

				res, err := client.Do(req)

				if err != nil {
					slog.Error(
						"error creating sending request",
						slog.Any("error", err),
					)
					continue
				}

				slog.Info(
					"getUpdates response",
					slog.Int("status", res.StatusCode),
				)

			}
		}()

	}
}

func RunDailyJob(scheduler scheduler.Scheduler) {
	now := time.Now()
	startTime := entities.CalendarTime{Hour: now.Hour(), Minute: now.Minute()}
	endTime := entities.CalendarTime{Hour: 23, Minute: 59}
	for {
		err := LunchDailyScheduler(scheduler, startTime, endTime)
		if err != nil {
			slog.Error("error running daily job", slog.Any("error", err))
		}

		todayMidnight := time.Date(
			now.Year(),
			now.Month(),
			now.Day(),
			0,
			0,
			0,
			0,
			now.Location(),
		)
		tomorrowMidnight := todayMidnight.AddDate(0, 0, 1)
		duration := tomorrowMidnight.Sub(now)

		time.Sleep(duration)
		now = time.Now()
		startTime.Hour = 0
		startTime.Minute = 0
	}

}

func LunchDailyScheduler(
	scheduler scheduler.Scheduler,
	startTime, endTime entities.CalendarTime,
) error {
	go scheduler.LaunchSchedulesInRange(
		entities.KhorshidiCalendar{},
		startTime,
		endTime,
	)

	go scheduler.LaunchSchedulesInRange(
		entities.GeorgianCalendar{},
		startTime,
		endTime,
	)

	go scheduler.LaunchSchedulesInRange(
		entities.QamariCalendar{},
		startTime,
		endTime,
	)

	return nil
}
