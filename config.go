package main

import (
	"flag"
	"os"
	"strconv"
)

var (
	dbPort = flag.Int("db.port", 5432, "database port")
	dbAddr = flag.String("db.addr", "127.0.0.1", "database address")
	dbUser = flag.String("db.user", "postgres", "database user")
	dbPass = flag.String("db.pass", "", "database password")
	dbName = flag.String("db.table", "bind_dns", "database name")

	listenAddr = flag.String("listen.addr", "0.0.0.0", "listen address")
	listenPort = flag.String("listen.port", "8080", "listen port")

	gitToken = flag.String("git.token", "", "git token")
)

func getEnv(key, fallback string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return fallback
}

func getEnvInt(key string, fallback int) int {
	if value, exists := os.LookupEnv(key); exists {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return fallback
}

func loadConfig() {
	flag.Parse()

	*dbPort = getEnvInt("DB_PORT", *dbPort)
	*dbAddr = getEnv("DB_ADDR", *dbAddr)
	*dbUser = getEnv("DB_USER", *dbUser)
	*dbPass = getEnv("DB_PASS", *dbPass)
	*dbName = getEnv("DB_NAME", *dbName)

	*listenAddr = getEnv("LISTEN_ADDR", *listenAddr)
	*listenPort = getEnv("LISTEN_PORT", *listenPort)

	*gitToken = getEnv("GIT_TOKEN", *gitToken)
}
