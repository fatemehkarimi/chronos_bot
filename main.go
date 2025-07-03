package main

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"fmt"
	"github.com/fatemehkarimi/chronos_bot/handler"
	"github.com/fatemehkarimi/chronos_bot/scheduler"
	"log/slog"
	"net/http"
	"os"
	"time"

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

	go awxScheduler.LaunchSchedulesInRange(
		entities.GeorgianCalendar,
		entities.CalendarTime{
			Hour:   time.Now().Hour(),
			Minute: time.Now().Minute(),
		},
		entities.CalendarTime{Hour: 23, Minute: 0},
	)

	go checkForUpdates(config.BotToken, httpHandler)
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

		req, err := http.NewRequest(
			"POST",
			endpoint,
			bytes.NewBuffer(requestBytes),
		)

		if err != nil {
			slog.Error("error creating new request", slog.Any("error", err))
			continue
		}

		res, err := client.Do(req)

		if err != nil {
			slog.Error("error creating sending request", slog.Any("error", err))
			continue
		}
		slog.Info(
			"getUpdates response from tapi",
			slog.Int("status", res.StatusCode),
		)

		updateResponse := entities.ResponseGetUpdates{}
		err = json.NewDecoder(res.Body).Decode(&updateResponse)
		if err != nil {
			slog.Error("error parsing response", slog.Any("error", err))
			continue
		}

		// sending updates
		for _, update := range updateResponse.Result {
			requestBytes, err := json.MarshalIndent(update, "", "  ")

			if err != nil {
				slog.Error("error marshaling request", slog.Any("error", err))
			}

			req, err := http.NewRequest(
				"POST",
				"http://localhost:8080/getUpdates",
				bytes.NewBuffer(requestBytes),
			)

			if err != nil {
				slog.Error("error creating new request", slog.Any("error", err))
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

			slog.Info("getUpdates response", slog.Int("status", res.StatusCode))

		}
	}
}
