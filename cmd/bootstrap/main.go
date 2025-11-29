package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/DrC0ns0le/bind-api/config"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Database connection flags
var (
	dbPort = flag.String("db.port", "5432", "database port, env: DB_PORT")
	dbAddr = flag.String("db.addr", "127.0.0.1", "database address, env: DB_ADDR")
	dbUser = flag.String("db.user", "postgres", "database user, env: DB_USER")
	dbPass = flag.String("db.pass", "", "database password, env: DB_PASS")
	dbName = flag.String("db.name", "bind_dns", "database name, env: DB_NAME")
)

// Table name flags (customizable via env/flags)
var (
	schemaName      = flag.String("schema", "bind_dns", "schema name, env: DB_SCHEMA")
	configsTable    = flag.String("table.configs", "configs", "configs table name, env: TABLE_CONFIGS")
	zonesTable      = flag.String("table.zones", "zones", "zones table name, env: TABLE_ZONES")
	recordsTable    = flag.String("table.records", "records", "records table name, env: TABLE_RECORDS")
	tagsTable       = flag.String("table.tags", "tags", "tags table name, env: TABLE_TAGS")
	usersTable      = flag.String("table.users", "users", "users table name, env: TABLE_USERS")
	rolesTable      = flag.String("table.roles", "roles", "roles table name, env: TABLE_ROLES")
	userRolesTable  = flag.String("table.user_roles", "user_roles", "user_roles table name, env: TABLE_USER_ROLES")
)

// Operation flags
var (
	dropTables = flag.Bool("drop", false, "drop existing tables before creating")
)

// TableNames holds resolved table names with schema prefix
type TableNames struct {
	Schema    string
	Configs   string
	Zones     string
	Records   string
	Tags      string
	Users     string
	Roles     string
	UserRoles string
}

func (t TableNames) Full(table string) string {
	return t.Schema + "." + table
}

var tables TableNames

func main() {
	flag.Parse()

	// Resolve table names from env/flags
	tables = TableNames{
		Schema:    config.GetEnv("DB_SCHEMA", *schemaName),
		Configs:   config.GetEnv("TABLE_CONFIGS", *configsTable),
		Zones:     config.GetEnv("TABLE_ZONES", *zonesTable),
		Records:   config.GetEnv("TABLE_RECORDS", *recordsTable),
		Tags:      config.GetEnv("TABLE_TAGS", *tagsTable),
		Users:     config.GetEnv("TABLE_USERS", *usersTable),
		Roles:     config.GetEnv("TABLE_ROLES", *rolesTable),
		UserRoles: config.GetEnv("TABLE_USER_ROLES", *userRolesTable),
	}

	log.Printf("Using schema: %s", tables.Schema)
	log.Printf("Table names: configs=%s, zones=%s, records=%s, tags=%s, users=%s, roles=%s, user_roles=%s",
		tables.Configs, tables.Zones, tables.Records, tables.Tags, tables.Users, tables.Roles, tables.UserRoles)

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	// Connect to database
	connString := fmt.Sprintf("postgres://%s:%s@%s:%s/%s",
		config.GetEnv("DB_USER", *dbUser),
		config.GetEnv("DB_PASS", *dbPass),
		config.GetEnv("DB_ADDR", *dbAddr),
		config.GetEnv("DB_PORT", *dbPort),
		config.GetEnv("DB_NAME", *dbName),
	)

	pool, err := pgxpool.New(ctx, connString)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer pool.Close()

	if err := pool.Ping(ctx); err != nil {
		log.Fatalf("Failed to ping database: %v", err)
	}
	log.Println("Connected to database")

	// Create schema
	if err := createSchema(ctx, pool); err != nil {
		log.Fatalf("Failed to create schema: %v", err)
	}

	// Drop tables if requested
	if *dropTables {
		log.Println("Dropping existing tables...")
		if err := dropAllTables(ctx, pool); err != nil {
			log.Fatalf("Failed to drop tables: %v", err)
		}
		log.Println("Tables dropped successfully")
	}

	// Create/migrate tables
	log.Println("Creating/migrating tables...")
	if err := createOrMigrateTables(ctx, pool); err != nil {
		log.Fatalf("Failed to create/migrate tables: %v", err)
	}

	// Create indexes
	log.Println("Creating indexes...")
	if err := createIndexes(ctx, pool); err != nil {
		log.Fatalf("Failed to create indexes: %v", err)
	}

	// Seed default data
	log.Println("Seeding default data...")
	if err := seedDefaults(ctx, pool); err != nil {
		log.Fatalf("Failed to seed defaults: %v", err)
	}

	log.Println("Bootstrap completed successfully")
}

func createSchema(ctx context.Context, pool *pgxpool.Pool) error {
	_, err := pool.Exec(ctx, fmt.Sprintf(`CREATE SCHEMA IF NOT EXISTS %s`, tables.Schema))
	return err
}

func dropAllTables(ctx context.Context, pool *pgxpool.Pool) error {
	// Drop in reverse order of dependencies
	tablesToDrop := []string{
		tables.UserRoles,
		tables.Users,
		tables.Roles,
		tables.Tags,
		tables.Records,
		tables.Zones,
		tables.Configs,
	}

	for _, t := range tablesToDrop {
		query := fmt.Sprintf(`DROP TABLE IF EXISTS %s.%s CASCADE`, tables.Schema, t)
		if _, err := pool.Exec(ctx, query); err != nil {
			return fmt.Errorf("failed to drop table %s: %w", t, err)
		}
	}
	return nil
}

// tableExists checks if a table exists in the schema
func tableExists(ctx context.Context, pool *pgxpool.Pool, tableName string) (bool, error) {
	var exists bool
	err := pool.QueryRow(ctx, `
		SELECT EXISTS (
			SELECT FROM information_schema.tables 
			WHERE table_schema = $1 AND table_name = $2
		)`, tables.Schema, tableName).Scan(&exists)
	return exists, err
}

// columnExists checks if a column exists in a table
func columnExists(ctx context.Context, pool *pgxpool.Pool, tableName, columnName string) (bool, error) {
	var exists bool
	err := pool.QueryRow(ctx, `
		SELECT EXISTS (
			SELECT FROM information_schema.columns 
			WHERE table_schema = $1 AND table_name = $2 AND column_name = $3
		)`, tables.Schema, tableName, columnName).Scan(&exists)
	return exists, err
}

// addColumnIfNotExists adds a column to a table if it doesn't exist
func addColumnIfNotExists(ctx context.Context, pool *pgxpool.Pool, tableName, columnName, columnDef string) error {
	exists, err := columnExists(ctx, pool, tableName, columnName)
	if err != nil {
		return err
	}
	if !exists {
		query := fmt.Sprintf(`ALTER TABLE %s.%s ADD COLUMN %s %s`, tables.Schema, tableName, columnName, columnDef)
		log.Printf("Adding column %s to %s", columnName, tableName)
		_, err = pool.Exec(ctx, query)
		return err
	}
	return nil
}

// createOrMigrateTables creates tables if they don't exist, or migrates them if they do
func createOrMigrateTables(ctx context.Context, pool *pgxpool.Pool) error {
	// Define table schemas
	tableSchemas := []struct {
		name    string
		create  string
		columns []struct {
			name string
			def  string
		}
	}{
		{
			name: tables.Configs,
			create: fmt.Sprintf(`
				CREATE TABLE IF NOT EXISTS %s.%s (
					config_key VARCHAR(255) NOT NULL,
					config_value TEXT NOT NULL,
					created_at TIMESTAMP NOT NULL DEFAULT NOW(),
					modified_at TIMESTAMP NOT NULL DEFAULT NOW(),
					deleted_at TIMESTAMP,
					staging BOOLEAN NOT NULL DEFAULT FALSE,
					PRIMARY KEY (config_key, config_value)
				)`, tables.Schema, tables.Configs),
			columns: []struct {
				name string
				def  string
			}{
				{"staging", "BOOLEAN NOT NULL DEFAULT FALSE"},
			},
		},
		{
			name: tables.Zones,
			create: fmt.Sprintf(`
				CREATE TABLE IF NOT EXISTS %s.%s (
					uuid UUID PRIMARY KEY DEFAULT gen_random_uuid(),
					name VARCHAR(255) NOT NULL,
					created_at TIMESTAMP NOT NULL DEFAULT NOW(),
					modified_at TIMESTAMP NOT NULL DEFAULT NOW(),
					deleted_at TIMESTAMP,
					staging BOOLEAN NOT NULL DEFAULT FALSE,
					primary_ns VARCHAR(255) NOT NULL DEFAULT '',
					admin_email VARCHAR(255) NOT NULL DEFAULT '',
					refresh INTEGER NOT NULL DEFAULT 3600,
					retry INTEGER NOT NULL DEFAULT 600,
					expire INTEGER NOT NULL DEFAULT 604800,
					minimum INTEGER NOT NULL DEFAULT 86400,
					ttl INTEGER NOT NULL DEFAULT 3600
				)`, tables.Schema, tables.Zones),
			columns: []struct {
				name string
				def  string
			}{
				{"primary_ns", "VARCHAR(255) NOT NULL DEFAULT ''"},
				{"admin_email", "VARCHAR(255) NOT NULL DEFAULT ''"},
				{"refresh", "INTEGER NOT NULL DEFAULT 3600"},
				{"retry", "INTEGER NOT NULL DEFAULT 600"},
				{"expire", "INTEGER NOT NULL DEFAULT 604800"},
				{"minimum", "INTEGER NOT NULL DEFAULT 86400"},
				{"ttl", "INTEGER NOT NULL DEFAULT 3600"},
			},
		},
		{
			name: tables.Records,
			create: fmt.Sprintf(`
				CREATE TABLE IF NOT EXISTS %s.%s (
					uuid UUID PRIMARY KEY DEFAULT gen_random_uuid(),
					type VARCHAR(16) NOT NULL,
					host VARCHAR(255) NOT NULL,
					content TEXT NOT NULL,
					ttl INTEGER NOT NULL DEFAULT 3600,
					add_ptr BOOLEAN NOT NULL DEFAULT FALSE,
					created_at TIMESTAMP NOT NULL DEFAULT NOW(),
					modified_at TIMESTAMP NOT NULL DEFAULT NOW(),
					deleted_at TIMESTAMP,
					zone_uuid UUID NOT NULL REFERENCES %s.%s(uuid) ON DELETE CASCADE,
					staging BOOLEAN NOT NULL DEFAULT FALSE
				)`, tables.Schema, tables.Records, tables.Schema, tables.Zones),
			columns: []struct {
				name string
				def  string
			}{
				{"add_ptr", "BOOLEAN NOT NULL DEFAULT FALSE"},
			},
		},
		{
			name: tables.Tags,
			create: fmt.Sprintf(`
				CREATE TABLE IF NOT EXISTS %s.%s (
					id SERIAL PRIMARY KEY,
					tag VARCHAR(255) NOT NULL,
					zone_uuid UUID REFERENCES %s.%s(uuid) ON DELETE CASCADE,
					record_uuid UUID REFERENCES %s.%s(uuid) ON DELETE CASCADE,
					CONSTRAINT chk_tag_reference CHECK (
						(zone_uuid IS NOT NULL AND record_uuid IS NULL) OR
						(zone_uuid IS NULL AND record_uuid IS NOT NULL)
					),
					UNIQUE (tag, zone_uuid),
					UNIQUE (tag, record_uuid)
				)`, tables.Schema, tables.Tags, tables.Schema, tables.Zones, tables.Schema, tables.Records),
			columns: nil,
		},
		{
			name: tables.Users,
			create: fmt.Sprintf(`
				CREATE TABLE IF NOT EXISTS %s.%s (
					id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
					username VARCHAR(255) NOT NULL UNIQUE,
					password_hash VARCHAR(255) NOT NULL,
					email VARCHAR(255),
					created_at TIMESTAMP NOT NULL DEFAULT NOW(),
					modified_at TIMESTAMP NOT NULL DEFAULT NOW(),
					deleted_at TIMESTAMP
				)`, tables.Schema, tables.Users),
			columns: nil,
		},
		{
			name: tables.Roles,
			create: fmt.Sprintf(`
				CREATE TABLE IF NOT EXISTS %s.%s (
					id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
					name VARCHAR(255) NOT NULL UNIQUE,
					description TEXT,
					created_at TIMESTAMP NOT NULL DEFAULT NOW(),
					modified_at TIMESTAMP NOT NULL DEFAULT NOW(),
					deleted_at TIMESTAMP
				)`, tables.Schema, tables.Roles),
			columns: nil,
		},
		{
			name: tables.UserRoles,
			create: fmt.Sprintf(`
				CREATE TABLE IF NOT EXISTS %s.%s (
					user_id UUID NOT NULL REFERENCES %s.%s(id) ON DELETE CASCADE,
					role_id UUID NOT NULL REFERENCES %s.%s(id) ON DELETE CASCADE,
					created_at TIMESTAMP NOT NULL DEFAULT NOW(),
					PRIMARY KEY (user_id, role_id)
				)`, tables.Schema, tables.UserRoles, tables.Schema, tables.Users, tables.Schema, tables.Roles),
			columns: nil,
		},
	}

	for _, ts := range tableSchemas {
		exists, err := tableExists(ctx, pool, ts.name)
		if err != nil {
			return fmt.Errorf("failed to check if table %s exists: %w", ts.name, err)
		}

		if !exists {
			// Create table
			log.Printf("Creating table: %s", ts.name)
			if _, err := pool.Exec(ctx, ts.create); err != nil {
				return fmt.Errorf("failed to create table %s: %w", ts.name, err)
			}
		} else {
			// Migrate: add missing columns
			log.Printf("Table %s exists, checking for migrations...", ts.name)
			for _, col := range ts.columns {
				if err := addColumnIfNotExists(ctx, pool, ts.name, col.name, col.def); err != nil {
					return fmt.Errorf("failed to add column %s to %s: %w", col.name, ts.name, err)
				}
			}
		}
	}

	return nil
}

// createIndexes creates all indexes
func createIndexes(ctx context.Context, pool *pgxpool.Pool) error {
	indexes := []string{
		fmt.Sprintf(`CREATE INDEX IF NOT EXISTS idx_%s_name ON %s.%s(name) WHERE deleted_at IS NULL`, tables.Zones, tables.Schema, tables.Zones),
		fmt.Sprintf(`CREATE INDEX IF NOT EXISTS idx_%s_zone_uuid ON %s.%s(zone_uuid) WHERE deleted_at IS NULL`, tables.Records, tables.Schema, tables.Records),
		fmt.Sprintf(`CREATE INDEX IF NOT EXISTS idx_%s_type ON %s.%s(type) WHERE deleted_at IS NULL`, tables.Records, tables.Schema, tables.Records),
		fmt.Sprintf(`CREATE INDEX IF NOT EXISTS idx_%s_host ON %s.%s(host) WHERE deleted_at IS NULL`, tables.Records, tables.Schema, tables.Records),
		fmt.Sprintf(`CREATE INDEX IF NOT EXISTS idx_%s_zone_uuid ON %s.%s(zone_uuid)`, tables.Tags, tables.Schema, tables.Tags),
		fmt.Sprintf(`CREATE INDEX IF NOT EXISTS idx_%s_record_uuid ON %s.%s(record_uuid)`, tables.Tags, tables.Schema, tables.Tags),
		fmt.Sprintf(`CREATE INDEX IF NOT EXISTS idx_%s_username ON %s.%s(username) WHERE deleted_at IS NULL`, tables.Users, tables.Schema, tables.Users),
		fmt.Sprintf(`CREATE INDEX IF NOT EXISTS idx_%s_email ON %s.%s(email) WHERE deleted_at IS NULL`, tables.Users, tables.Schema, tables.Users),
		fmt.Sprintf(`CREATE INDEX IF NOT EXISTS idx_%s_name ON %s.%s(name) WHERE deleted_at IS NULL`, tables.Roles, tables.Schema, tables.Roles),
		fmt.Sprintf(`CREATE INDEX IF NOT EXISTS idx_%s_key ON %s.%s(config_key) WHERE deleted_at IS NULL`, tables.Configs, tables.Schema, tables.Configs),
	}

	for _, idx := range indexes {
		if _, err := pool.Exec(ctx, idx); err != nil {
			// Extract index name for error message
			parts := strings.Split(idx, " ")
			idxName := ""
			for i, p := range parts {
				if p == "EXISTS" && i+1 < len(parts) {
					idxName = parts[i+1]
					break
				}
			}
			return fmt.Errorf("failed to create index %s: %w", idxName, err)
		}
	}
	return nil
}

// seedDefaults inserts default data
func seedDefaults(ctx context.Context, pool *pgxpool.Pool) error {
	// Default roles
	defaultRoles := []struct {
		name        string
		description string
	}{
		{"admin", "Full administrative access"},
		{"editor", "Can create and modify DNS records"},
		{"viewer", "Read-only access to DNS records"},
	}

	for _, role := range defaultRoles {
		_, err := pool.Exec(ctx, fmt.Sprintf(`
			INSERT INTO %s.%s (name, description) 
			VALUES ($1, $2) 
			ON CONFLICT (name) DO NOTHING
		`, tables.Schema, tables.Roles), role.name, role.description)
		if err != nil {
			return fmt.Errorf("failed to insert role %s: %w", role.name, err)
		}
	}

	// Default SOA configs for reverse DNS zones
	defaultConfigs := []struct {
		key   string
		value string
	}{
		{"rdns_primary_ns", "ns.example.com"},
		{"rdns_admin_email", "admin.example.com"},
		{"rdns_refresh", "1800"},
		{"rdns_retry", "1800"},
		{"rdns_expire", "604800"},
		{"rdns_minimum", "1800"},
		{"rdns_ttl", "3600"},
	}

	for _, cfg := range defaultConfigs {
		_, err := pool.Exec(ctx, fmt.Sprintf(`
			INSERT INTO %s.%s (config_key, config_value) 
			VALUES ($1, $2) 
			ON CONFLICT (config_key, config_value) DO NOTHING
		`, tables.Schema, tables.Configs), cfg.key, cfg.value)
		if err != nil {
			return fmt.Errorf("failed to insert config %s: %w", cfg.key, err)
		}
	}

	return nil
}

func init() {
	// Load .env file if present
	if data, err := os.ReadFile(".env"); err == nil {
		for _, line := range splitLines(string(data)) {
			if len(line) == 0 || line[0] == '#' {
				continue
			}
			if idx := indexOf(line, '='); idx > 0 {
				key := line[:idx]
				value := line[idx+1:]
				if os.Getenv(key) == "" {
					os.Setenv(key, value)
				}
			}
		}
	}
}

func splitLines(s string) []string {
	var lines []string
	start := 0
	for i := 0; i < len(s); i++ {
		if s[i] == '\n' {
			line := s[start:i]
			if len(line) > 0 && line[len(line)-1] == '\r' {
				line = line[:len(line)-1]
			}
			lines = append(lines, line)
			start = i + 1
		}
	}
	if start < len(s) {
		lines = append(lines, s[start:])
	}
	return lines
}

func indexOf(s string, c byte) int {
	for i := 0; i < len(s); i++ {
		if s[i] == c {
			return i
		}
	}
	return -1
}
