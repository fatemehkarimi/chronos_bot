package main

import (
	"database/sql"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"time"

	"github.com/fatemehkarimi/chronos_bot/api"
	"github.com/fatemehkarimi/chronos_bot/repository"
	_ "github.com/lib/pq"
	"github.com/spf13/viper"
)

type Config struct {
	Database repository.DatabaseConfig
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
	defer db.Close()

	err = db.Ping()
	if err != nil {
		slog.Error("failed to ping database. err = ", slog.Any("err", err))
		os.Exit(1)
	}

	postgresRepo := repository.PostgresRepository{DB: db}
	// err = postgresRepo.Init()

	if err != nil {
		slog.Error("failed to init service. error = ", slog.Any("err", err))
		os.Exit(1)
	}

	httpHandler := api.NewHttpHandler(&postgresRepo)
	mux := http.NewServeMux()
	mux.HandleFunc("/getUpdates", httpHandler.GetUpdates)

	server := &http.Server{
		Addr:         ":8080",
		Handler:      mux,
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  120 * time.Second,
	}

	server.ListenAndServe()
}
