package main

import (
	"context"
	"fmt"

	"database/sql"
	"flag"
	"time"

	//"github.com/aws/aws-lambda-go/events"
	//"github.com/aws/aws-lambda-go/lambda"
	"log/slog"
	"os"
	"strings"
	"sync"

	//awsConf "github.com/aws/aws-sdk-go-v2/config"
	//"github.com/aws/aws-sdk-go-v2/feature/rds/auth"
	_ "github.com/lib/pq"
)

type config struct {
	port int
	env  string
	db   struct {
		dbUsername string
		dbPassword string
		dbHost     string
		dbPort     int
		dbName     string
		region     string
	}
	limiter struct {
		rps     float64
		burst   int
		enabled bool
	}
	cors struct {
		trustedOrigins []string
	}
}

type application struct {
	config config
	logger *slog.Logger
	wg     sync.WaitGroup
	models Models
}

func main() {
	var cfg config

	flag.IntVar(&cfg.port, "port", 4001, "API server port")
	flag.StringVar(&cfg.env, "env", "development", "Environment (development|production|staging)")
	flag.StringVar(&cfg.db.dbUsername, "db-username", "postgres", "Database username")
	flag.StringVar(&cfg.db.dbPassword, "db-password", "postgres", "Database password")
	flag.StringVar(&cfg.db.dbHost, "db-host", "localhost", "Database host")
	flag.IntVar(&cfg.db.dbPort, "db-port", 5432, "Database port")
	flag.StringVar(&cfg.db.dbName, "db-name", "postgres", "Database name")
	flag.StringVar(&cfg.db.region, "db-region", "eu-west", "AWS region")
	flag.IntVar(&cfg.limiter.burst, "limiter-burst", 4, "Rate limiter maximum burst size")
	flag.BoolVar(&cfg.limiter.enabled, "limiter-enabled", false, "Enable rate limiter")
	flag.Func("cors-trusted-origins", "Trusted CORS origins (space separated)", func(val string) error {
		cfg.cors.trustedOrigins = strings.Fields(val)
		return nil
	})

	flag.Parse()

	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))

	db, err := openDB(cfg)
	if err != nil {
		logger.Error(err.Error())
		os.Exit(1)
	}
	defer db.Close()

	logger.Info("database connection pool established")

	app := application{
		config: cfg,
		logger: logger,
		models: NewModels(db),
	}

	err = app.serve()
	if err != nil {
		logger.Error(err.Error())
		os.Exit(1)
	}
}

func openDB(cfg config) (*sql.DB, error) {
	uri := fmt.Sprintf("postgres://%s:%s@%s:%d/%s",
		cfg.db.dbUsername,
		cfg.db.dbPassword,
		cfg.db.dbHost,
		cfg.db.dbPort,
		cfg.db.dbName,
	)

	db, err := sql.Open("postgres", uri)
	if err != nil {
		return nil, err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err = db.PingContext(ctx)
	if err != nil {
		return nil, err
	}

	return db, nil
}
