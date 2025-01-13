package rdb

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

var db *pgxpool.Pool

type DBConfig struct {
	Host        string
	Port        int
	User        string
	Password    string
	DBName      string
	MaxConns    int32
	MinConns    int32
	MaxConnIdle time.Duration
	MaxConnLife time.Duration
}

func Init(config DBConfig) {
	ctx := context.Background()
	// Connect to the database
	if err := connect(ctx, config); err != nil {
		log.Fatal(err)
	}

	// Test the connection
	if err := db.Ping(ctx); err != nil {
		log.Fatal(err)
	}
	log.Printf("Connected to the database successfully.\n")
}

// Connect establishes a connection pool to the database
func connect(ctx context.Context, config DBConfig) error {
	connString := fmt.Sprintf("postgres://%s:%s@%s:%d/%s",
		config.User, config.Password, config.Host, config.Port, config.DBName)

	poolConfig, err := pgxpool.ParseConfig(connString)
	if err != nil {
		return fmt.Errorf("unable to parse connection string: %v", err)
	}

	// Set connection pool settings
	if config.MaxConns == 0 {
		poolConfig.MaxConns = 50
	} else {
		poolConfig.MaxConns = config.MaxConns
	}

	if config.MinConns == 0 {
		poolConfig.MinConns = 5
	} else {
		poolConfig.MinConns = config.MinConns
	}

	if config.MaxConnIdle == 0 {
		poolConfig.MaxConnIdleTime = 10 * time.Minute
	} else {
		poolConfig.MaxConnIdleTime = config.MaxConnIdle
	}

	if config.MaxConnLife == 0 {
		poolConfig.MaxConnLifetime = 10 * time.Minute
	} else {

		poolConfig.MaxConnLifetime = config.MaxConnLife
	}

	// Create the connection pool
	db, err = pgxpool.NewWithConfig(ctx, poolConfig)
	if err != nil {
		return fmt.Errorf("unable to create connection pool: %v", err)
	}

	return nil
}
