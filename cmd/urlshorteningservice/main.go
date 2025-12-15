package main

import (
	"context"
	"database/sql"
	"embed"
	"fmt"
	"os"
	"strings"
	"time"
	"url-shortening-service/internal/application/stats"
	"url-shortening-service/internal/application/urlcases"
	"url-shortening-service/internal/domain"
	"url-shortening-service/internal/infrastructure/database"
	"url-shortening-service/internal/infrastructure/http"
	"url-shortening-service/internal/infrastructure/kafka/statsbus"
	rediswrap "url-shortening-service/internal/infrastructure/redis"

	"github.com/jackc/pgx/v5/pgxpool"
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
	kafkaHost := "localhost"
	kafkaPort := "9092"

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
	if envKafkaHost, found := os.LookupEnv(domain.KafkaHostEnv); found {
		kafkaHost = envKafkaHost
	}
	if envKafkaPort, found := os.LookupEnv(domain.KafkaPortEnv); found {
		kafkaPort = envKafkaPort
	}

	logger := domain.StdoutLogger
	databaseUrl := databaseSettings.GetUrl()
	kafkaUrl := kafkaHost + ":" + kafkaPort

	err := migrateDatabase(databaseUrl)
	if err != nil {
		logger.Error(fmt.Sprintf("Database migration failed: %v", err))
		return
	}

	dbpool, err := pgxpool.New(context.Background(), databaseUrl)
	if err != nil {
		logger.Error(fmt.Sprintf("Unable to connect to database: %v", err))
		return
	}
	defer dbpool.Close()

	storage := database.NewPostgresStorage(dbpool, logger)
	statsStorage := database.NewPostgersStatsStorage(dbpool, logger)

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

	getUrlCase := urlcases.NewUrlGetter(localCache, storage, logger)
	shortenUrlCase := urlcases.NewUrlShortener(idGenerator, storage)
	updateUrlCase := urlcases.NewUrlUpdater(localCache, storage, logger)
	deleteUrlCase := urlcases.NewUrlDeleter(localCache, storage, logger)
	statsProcessor := stats.NewRedirectStatsProcessor(statsStorage, logger)

	topicId := "url_stats_events"
	groupId := "url_stats_group"
	eventProducer := statsbus.NewKafkaEventProducer(topicId, 10*time.Millisecond, kafkaUrl)
	eventConsumer := statsbus.NewKafkaEventConsumer(topicId, groupId, statsProcessor, logger, kafkaUrl)

	go eventConsumer.StartConsuming(context.Background())

	server := http.NewSimpleServer(*shortenUrlCase, *getUrlCase, *updateUrlCase, *deleteUrlCase, eventProducer, logger, serverPort)
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
