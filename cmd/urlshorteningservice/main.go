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
	"url-shortening-service/internal/infrastructure/location"
	rediswrap "url-shortening-service/internal/infrastructure/redis"

	"github.com/ClickHouse/clickhouse-go/v2"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/pressly/goose/v3"
	"github.com/redis/go-redis/v9"
	"github.com/segmentio/kafka-go"

	_ "github.com/ClickHouse/clickhouse-go/v2"
	_ "github.com/jackc/pgx/v5/stdlib"
)

//go:embed migrations/*.sql
var embedMigrations embed.FS

//go:embed clickhouse-migrations/*.sql
var embedStatsMigrations embed.FS

func main() {
	redisHost := "localhost"
	redisPort := "6379"

	serverPort := "8080"

	kafkaHost := "localhost"
	kafkaPort := "9094"

	clickhouseHost := "localhost"
	clickhousePort := "9000"
	clickhouseDB := "url_shortener_analytics"
	clickhouseUser := "default"
	clickhousePassword := "default"

	databaseSettings := domain.PostgresSettings{
		User:       "admin",
		Password:   "password",
		Host:       "localhost",
		Port:       "5432",
		DBName:     "url_shortener_db",
		SSlEnabled: false,
	}

	trySetEnvVariable(domain.RedisUrlEnv, &redisHost)
	trySetEnvVariable(domain.RedisPortEnv, &redisPort)
	trySetEnvVariable(domain.ServerPortEnv, &serverPort)
	trySetEnvVariable(domain.DatabaseUserEnv, &databaseSettings.User)
	trySetEnvVariable(domain.DatabasePasswordEnv, &databaseSettings.Password)
	trySetEnvVariable(domain.DatabaseHostEnv, &databaseSettings.Host)
	trySetEnvVariable(domain.DatabasePortEnv, &databaseSettings.Port)
	trySetEnvVariable(domain.DatabaseNameEnv, &databaseSettings.DBName)
	trySetEnvVariable(domain.KafkaHostEnv, &kafkaHost)
	trySetEnvVariable(domain.KafkaPortEnv, &kafkaPort)
	trySetEnvVariable(domain.ClickhouseHostEnv, &clickhouseHost)
	trySetEnvVariable(domain.ClickhousePortEnv, &clickhousePort)
	trySetEnvVariable(domain.ClickhouseDBEnv, &clickhouseDB)
	trySetEnvVariable(domain.ClickhouseUserEnv, &clickhouseUser)
	trySetEnvVariable(domain.ClickhousePasswordEnv, &clickhousePassword)
	if endDatabaseSsl, found := os.LookupEnv(domain.DatabaseSslEnv); found {
		if strings.EqualFold("false", endDatabaseSsl) {
			databaseSettings.SSlEnabled = false
		} else {
			databaseSettings.SSlEnabled = true
		}
	}

	clickhouseSettings := &clickhouse.Options{
		Addr: []string{fmt.Sprintf("%s:%s", clickhouseHost, clickhousePort)},
		Auth: clickhouse.Auth{
			Database: clickhouseDB,
			Username: clickhouseUser,
			Password: clickhousePassword,
		},
		DialTimeout:     5 * time.Second,
		MaxOpenConns:    100,
		MaxIdleConns:    20,
		ConnMaxLifetime: 1 * time.Hour,
	}
	clickhouseDsn := fmt.Sprintf("clickhouse://%s:%s@%s/%s?secure=false&async_insert=1&wait_for_async_insert=0",
		clickhouseSettings.Auth.Username,
		clickhouseSettings.Auth.Password,
		clickhouseSettings.Addr[0],
		clickhouseSettings.Auth.Database,
	)

	logger := domain.StdoutLogger
	databaseUrl := databaseSettings.GetUrl()
	kafkaUrl := kafkaHost + ":" + kafkaPort

	err := migrateDatabase(databaseUrl, &embedMigrations, "migrations", "pgx", "postgres")
	if err != nil {
		logger.Error(fmt.Sprintf("Database migration failed: %v", err))
		return
	}

	err = migrateDatabase(clickhouseDsn, &embedStatsMigrations, "clickhouse-migrations", "clickhouse", "clickhouse")
	if err != nil {
		logger.Error(fmt.Sprintf("Clickhouse migration failed: %v", err))
		return
	}

	dbpool, err := pgxpool.New(context.Background(), databaseUrl)
	if err != nil {
		logger.Error(fmt.Sprintf("Unable to connect to database: %v", err))
		return
	}
	defer dbpool.Close()

	clickhouseConn, err := clickhouse.Open(clickhouseSettings)
	if err != nil {
		logger.Error(fmt.Sprintf("Unable to connect to Clickhouse: %v", err))
		return
	}
	defer clickhouseConn.Close()

	redisClient := redis.NewClient(&redis.Options{
		Addr: redisHost + ":" + redisPort,
	})
	defer redisClient.Close()

	geo2ipDb, err := location.OpenGeoIPDatabase(domain.LocationDbPath)
	if err != nil {
		logger.Error("Failed to open GeoIP database: " + err.Error())
		return
	}
	defer geo2ipDb.Close()

	storage := database.NewPostgresStorage(dbpool, logger)
	statsStorage := database.NewClickhouseStatsStorage(clickhouseConn)
	cache := rediswrap.NewRedisStorage(redisClient, logger)

	idGenerator, err := rediswrap.NewRedisIdGenerator(context.Background(), redisClient, storage)
	if err != nil {
		logger.Error("Failed to create Redis ID generator")
		return
	}

	ipLocator := location.NewGeoIpLocator(geo2ipDb)

	getUrlCase := urlcases.NewUrlGetter(cache, storage, logger)
	shortenUrlCase := urlcases.NewUrlShortener(idGenerator, storage)
	updateUrlCase := urlcases.NewUrlUpdater(cache, storage, logger)
	deleteUrlCase := urlcases.NewUrlDeleter(cache, storage, logger)

	statsProcessor := stats.NewRedirectStatsProcessor(statsStorage, ipLocator, logger)
	statsCalculator := database.NewClickhouseStatsCalculator(clickhouseConn)

	topicId := "url_stats_events"
	groupId := "url_stats_group"
	kafkaWriter := &kafka.Writer{
		Addr:         kafka.TCP(kafkaUrl),
		Topic:        topicId,
		Balancer:     &kafka.LeastBytes{},
		BatchTimeout: 10 * time.Millisecond,
		Async:        true,
	}
	kafkaReader := kafka.NewReader(kafka.ReaderConfig{
		Brokers: []string{kafkaUrl},
		Topic:   topicId,
		GroupID: groupId,
	})
	eventProducer := statsbus.NewKafkaEventProducer(kafkaWriter)
	eventConsumer := statsbus.NewStatsEventConsumer(kafkaReader, statsProcessor, logger)

	go eventConsumer.StartConsuming(context.Background())

	server := http.NewSimpleServer(shortenUrlCase, getUrlCase, updateUrlCase, deleteUrlCase,
		eventProducer, statsCalculator, logger, serverPort)

	logger.Info("Starting server")
	server.Start()
	logger.Info("Server closed")
}

func trySetEnvVariable(envName string, val *string) {
	if envVal, found := os.LookupEnv(envName); found {
		*val = envVal
	}
}

func migrateDatabase(databaseUrl string, migrations *embed.FS, dir, driverName, dialect string) error {
	db, err := sql.Open(driverName, databaseUrl)
	if err != nil {
		return err
	}
	defer db.Close()

	goose.SetBaseFS(migrations)

	if err := goose.SetDialect(dialect); err != nil {
		return err
	}

	if err := goose.Up(db, dir); err != nil {
		return err
	}

	return nil
}
