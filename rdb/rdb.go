package rdb

import (
	"context"
	"flag"
	"fmt"
	"time"

	"github.com/DrC0ns0le/bind-api/config"
	"github.com/jackc/pgx/v5/pgxpool"
)

var (
	dbPort     = flag.String("db.port", "5432", "database port, env: DB_PORT")
	dbAddr     = flag.String("db.addr", "127.0.0.1", "database address, env: DB_ADDR")
	dbUser     = flag.String("db.user", "postgres", "database user, env: DB_USER")
	dbPass     = flag.String("db.pass", "", "database password, env: DB_PASS")
	dbName     = flag.String("db.table", "bind_dns", "database name, env: DB_NAME")
	dbMaxConns = flag.Int("db.max_conns", 10, "database max connections, env: DB_MAX_CONNS")
	dbMinConns = flag.Int("db.min_conns", 5, "database min connections, env: DB_MIN_CONNS")
)

var db *pgxpool.Pool

type DBConfig struct {
	Host        string
	Port        string
	User        string
	Password    string
	DBName      string
	MaxConns    int32
	MinConns    int32
	MaxConnIdle time.Duration
	MaxConnLife time.Duration
}

func Init() error {
	ctx := context.Background()
	// Connect to the database
	if err := connect(ctx, DBConfig{
		Host:     config.GetEnv("DB_ADDR", *dbAddr),
		Port:     config.GetEnv("DB_PORT", *dbPort),
		User:     config.GetEnv("DB_USER", *dbUser),
		Password: config.GetEnv("DB_PASS", *dbPass),
		DBName:   config.GetEnv("DB_NAME", *dbName),
		MaxConns: int32(config.GetEnvInt("DB_MAX_CONNS", *dbMaxConns)),
		MinConns: int32(config.GetEnvInt("DB_MIN_CONNS", *dbMinConns)),
	}); err != nil {
		return fmt.Errorf("error connecting to the database: %w", err)
	}

	// Test the connection
	if err := db.Ping(ctx); err != nil {
		return fmt.Errorf("error pinging the database: %w", err)
	}

	return nil
}

// Connect establishes a connection pool to the database
func connect(ctx context.Context, config DBConfig) error {
	connString := fmt.Sprintf("postgres://%s:%s@%s:%s/%s",
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
