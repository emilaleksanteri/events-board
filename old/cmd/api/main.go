package main

import (
	"context"

	"database/sql"
	"flag"
	"log/slog"
	"os"
	"strings"
	"sync"
	"time"

	"fmt"

	"github.com/emilaleksanteri/pubsub/internal/data"
	"github.com/gorilla/sessions"
	_ "github.com/lib/pq"
	redis "github.com/redis/go-redis/v9"
)

const VERSION = "1.0.0"

type config struct {
	port int
	env  string
	db   struct {
		dsn          string
		maxOpenConns int
		maxIdleConns int
		maxIdleTime  time.Duration
	}
	limiter struct {
		rps     float64
		burst   int
		enabled bool
	}
	cors struct {
		trustedOrigins []string
	}
	redis struct {
		redisAddr string
	}
	webhook struct {
		webhookAddr string
	}
}

type application struct {
	config       config
	logger       *slog.Logger
	wg           sync.WaitGroup
	models       data.Models
	redis        *redis.Client
	eventChan    chan *redis.Message
	sessionStore sessions.Store
}

func main() {
	var cfg config

	flag.IntVar(&cfg.port, "port", 4000, "API server port")
	flag.StringVar(&cfg.env, "env", "development", "Environment (development|production|staging)")
	flag.StringVar(&cfg.db.dsn, "db-dsn", "postgres://postgres:postgres@localhost:5432/postgres?sslmode=disable", "Postgres connection string")
	flag.IntVar(&cfg.db.maxOpenConns, "db-max-open-conns", 25, "Postgres max open connections")
	flag.IntVar(&cfg.db.maxIdleConns, "db-max-idle-conns", 25, "Postgres max idle connections")
	flag.DurationVar(&cfg.db.maxIdleTime, "db-max-idle-time", 15*time.Minute, "Postgres max connection idle time")
	flag.Float64Var(&cfg.limiter.rps, "limiter-rps", 2, "Rate limiter maximum requests per second")
	flag.IntVar(&cfg.limiter.burst, "limiter-burst", 4, "Rate limiter maximum burst size")
	flag.BoolVar(&cfg.limiter.enabled, "limiter-enabled", false, "Enable rate limiter")
	flag.Func("cors-trusted-origins", "Trusted CORS origins (space separated)", func(val string) error {
		cfg.cors.trustedOrigins = strings.Fields(val)
		return nil
	})
	flag.StringVar(&cfg.redis.redisAddr, "redis-dsn", "localhost:6379", "Redis connection string")
	flag.StringVar(&cfg.webhook.webhookAddr, "webhook-dsn", "localhost:9000", "Webhook connection string")

	flag.Parse()

	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))

	db, err := openDB(cfg)
	if err != nil {
		logger.Error(err.Error())
		os.Exit(1)
	}
	defer db.Close()

	logger.Info("database connection pool established")

	redisClient, err := createRedisClient(cfg)
	if err != nil {
		logger.Error(err.Error())
		os.Exit(1)
	}
	defer redisClient.Close()

	logger.Info("redis connection pool established")

	app := application{
		config:    cfg,
		logger:    logger,
		models:    data.NewModels(db),
		redis:     redisClient,
		eventChan: make(chan *redis.Message),
	}
	app.sessionStore = app.initAuth()

	subscribeTo := []string{POST_ADDED, COMMENT_ADDED, SUB_COMMENT_ADDED}
	sub := app.redis.Subscribe(context.Background(), subscribeTo...)
	iface, err := sub.Receive(context.Background())
	if err != nil {
		os.Exit(1)
		return
	}

	// some redis related misc stuff to help keep track
	for _, channel := range subscribeTo {
		logger.Info(fmt.Sprintf("subscribed to %s", channel))
	}

	switch iface.(type) {
	case *redis.Subscription:
		logger.Info("subscription received")
	}

	go func() {
		channel := sub.Channel()
		for msg := range channel {
			app.eventChan <- msg
		}
	}()

	err = app.serve()
	if err != nil {
		logger.Error(err.Error())
		os.Exit(1)
	}
}

func createRedisClient(cfg config) (*redis.Client, error) {
	client := redis.NewClient(&redis.Options{
		Addr:     cfg.redis.redisAddr,
		Password: "",
		DB:       0,
	})

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err := client.Ping(ctx).Result()
	if err != nil {
		return nil, err
	}

	return client, nil
}

func openDB(cfg config) (*sql.DB, error) {
	db, err := sql.Open("postgres", cfg.db.dsn)
	if err != nil {
		return nil, err
	}

	db.SetMaxOpenConns(cfg.db.maxOpenConns)
	db.SetMaxIdleConns(cfg.db.maxIdleConns)
	db.SetConnMaxIdleTime(cfg.db.maxIdleTime)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err = db.PingContext(ctx)
	if err != nil {
		return nil, err
	}

	return db, nil
}
