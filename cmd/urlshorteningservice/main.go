package main

import (
	"context"
	"database/sql"
	"embed"
	"fmt"
	"os"
	"strings"
	"url-shortening-service/internal/application"
	"url-shortening-service/internal/domain"
	"url-shortening-service/internal/infrastructure/database"
	"url-shortening-service/internal/infrastructure/http"
	rediswrap "url-shortening-service/internal/infrastructure/redis"

	"github.com/pressly/goose/v3"
	"github.com/redis/go-redis/v9"

	_ "github.com/jackc/pgx/v5/stdlib"
)

//go:embed migrations/*.sql
var embedMigrations embed.FS

func main() {
	redisUrl := "localhost"
	redisPort := "6379"
	serverPort := "8080"

	databaseSettings := domain.PostgresSettings{
		User:       "admin",
		Password:   "password",
		Host:       "localhost",
		Port:       "5432",
		DBName:     "url_shortener_db",
		SSlEnabled: false,
	}

	if envRedisUrl, found := os.LookupEnv(domain.RedisUrlEnv); found {
		redisUrl = envRedisUrl
	}
	if envRedisPort, found := os.LookupEnv(domain.RedisPortEnv); found {
		redisPort = envRedisPort
	}
	if envServerPort, found := os.LookupEnv(domain.ServerPortEnv); found {
		serverPort = envServerPort
	}
	if endDatabaseUser, found := os.LookupEnv(domain.DatabaseUserEnv); found {
		databaseSettings.User = endDatabaseUser
	}
	if endDatabasePassword, found := os.LookupEnv(domain.DatabasePasswordEnv); found {
		databaseSettings.Password = endDatabasePassword
	}
	if endDatabaseHost, found := os.LookupEnv(domain.DatabaseHostEnv); found {
		databaseSettings.Host = endDatabaseHost
	}
	if endDatabasePort, found := os.LookupEnv(domain.DatabasePortEnv); found {
		databaseSettings.Port = endDatabasePort
	}
	if endDatabaseName, found := os.LookupEnv(domain.DatabaseNameEnv); found {
		databaseSettings.DBName = endDatabaseName
	}
	if endDatabaseSsl, found := os.LookupEnv(domain.DatabaseSslEnv); found {
		if strings.EqualFold("false", endDatabaseSsl) {
			databaseSettings.SSlEnabled = false
		} else {
			databaseSettings.SSlEnabled = true
		}
	}

	logger := domain.StdoutLogger
	databaseUrl := databaseSettings.GetUrl()

	err := migrateDatabase(databaseUrl)
	if err != nil {
		logger.Error(fmt.Sprintf("Database migration failed: %v", err))
		return
	}

	storage, err := database.NewPostgresStorage(context.Background(), databaseSettings, logger)
	if err != nil {
		logger.Error(fmt.Sprintf("Error while creating Postgres storage: %v", err))
		return
	}
	defer storage.Close()

	redisClient := redis.NewClient(&redis.Options{
		Addr: redisUrl + ":" + redisPort,
	})
	defer redisClient.Close()

	localCache := rediswrap.NewRedisStorage(redisClient, logger)
	idGenerator, err := rediswrap.NewRedisIdGenerator(context.Background(), redisClient, storage)
	if err != nil {
		logger.Error("Failed to create Redis ID generator")
		return
	}

	getUrlCase := application.NewUrlGetter(localCache, storage, logger)
	shortenUrlCase := application.NewUrlShortener(idGenerator, storage)
	updateUrlCase := application.NewUrlUpdater(localCache, storage, logger)

	server := http.NewSimpleServer(*shortenUrlCase, *getUrlCase, *updateUrlCase, logger, serverPort)
	logger.Info("Starting server")
	server.Start()
	logger.Info("Server closed")
}

func migrateDatabase(databaseUrl string) error {
	goose.SetBaseFS(embedMigrations)

	db, err := sql.Open("pgx", databaseUrl)
	if err != nil {
		return err
	}
	defer db.Close()

	if err := goose.Up(db, "migrations"); err != nil {
		return err
	}

	return nil
}
