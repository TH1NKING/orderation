package mysql

import (
    "context"
    "database/sql"
    "fmt"
    "os"
    "strconv"
    "strings"
    _ "github.com/go-sql-driver/mysql"
)

type Config struct {
    Host     string
    Port     int
    Username string
    Password string
    Database string
    Params   string
}

func NewConfigFromEnv() *Config {
    if dsn := os.Getenv("MYSQL_DSN"); dsn != "" {
        return parseFromDSN(dsn)
    }

    host := getEnvOrDefault("MYSQL_HOST", "localhost")
    port, _ := strconv.Atoi(getEnvOrDefault("MYSQL_PORT", "3306"))
    username := getEnvOrDefault("MYSQL_USER", "root")
    password := getEnvOrDefault("MYSQL_PASSWORD", "")
    database := getEnvOrDefault("MYSQL_DATABASE", "orderation")
    params := getEnvOrDefault("MYSQL_PARAMS", "parseTime=true&charset=utf8mb4&collation=utf8mb4_unicode_ci")

    return &Config{
        Host:     host,
        Port:     port,
        Username: username,
        Password: password,
        Database: database,
        Params:   params,
    }
}

func parseFromDSN(dsn string) *Config {
    config := &Config{
        Host:     "localhost",
        Port:     3306,
        Username: "root",
        Database: "orderation",
        Params:   "parseTime=true&charset=utf8mb4&collation=utf8mb4_unicode_ci",
    }

    parts := strings.Split(dsn, "@tcp(")
    if len(parts) >= 2 {
        userPart := parts[0]
        if strings.Contains(userPart, ":") {
            userPassParts := strings.Split(userPart, ":")
            config.Username = userPassParts[0]
            if len(userPassParts) > 1 {
                config.Password = userPassParts[1]
            }
        } else {
            config.Username = userPart
        }

        remaining := parts[1]
        hostDbParts := strings.Split(remaining, ")/")
        if len(hostDbParts) >= 2 {
            hostPort := hostDbParts[0]
            if strings.Contains(hostPort, ":") {
                hostPortParts := strings.Split(hostPort, ":")
                config.Host = hostPortParts[0]
                if port, err := strconv.Atoi(hostPortParts[1]); err == nil {
                    config.Port = port
                }
            } else {
                config.Host = hostPort
            }

            dbParams := hostDbParts[1]
            if strings.Contains(dbParams, "?") {
                dbParamsParts := strings.Split(dbParams, "?")
                config.Database = dbParamsParts[0]
                if len(dbParamsParts) > 1 {
                    config.Params = dbParamsParts[1]
                }
            } else {
                config.Database = dbParams
            }
        }
    }

    return config
}

func (c *Config) ToDSN() string {
    return fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?%s",
        c.Username, c.Password, c.Host, c.Port, c.Database, c.Params)
}

func getEnvOrDefault(key, defaultValue string) string {
    if value := os.Getenv(key); value != "" {
        return value
    }
    return defaultValue
}

func Open(dsn string) (*sql.DB, error) {
    db, err := sql.Open("mysql", dsn)
    if err != nil { return nil, err }
    if err := db.Ping(); err != nil { return nil, err }
    return db, nil
}

func OpenWithConfig(config *Config) (*sql.DB, error) {
    return Open(config.ToDSN())
}

// EnsureSchema creates tables if not exist.
func EnsureSchema(ctx context.Context, db *sql.DB) error {
    stmts := []string{
        `CREATE TABLE IF NOT EXISTS users (
            id VARCHAR(32) PRIMARY KEY,
            name VARCHAR(255) NOT NULL,
            email VARCHAR(255) NOT NULL UNIQUE,
            pass_hash TEXT NOT NULL,
            role VARCHAR(32) NOT NULL,
            created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
        ) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;`,
        `CREATE TABLE IF NOT EXISTS restaurants (
            id VARCHAR(32) PRIMARY KEY,
            name VARCHAR(255) NOT NULL,
            address VARCHAR(512) NOT NULL,
            open_time VARCHAR(16) NOT NULL,
            close_time VARCHAR(16) NOT NULL,
            created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
        ) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;`,
        `CREATE TABLE IF NOT EXISTS tables (
            id VARCHAR(32) PRIMARY KEY,
            restaurant_id VARCHAR(32) NOT NULL,
            name VARCHAR(255) NOT NULL,
            capacity INT NOT NULL,
            created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
            INDEX idx_tables_restaurant (restaurant_id),
            CONSTRAINT fk_tables_restaurant FOREIGN KEY (restaurant_id) REFERENCES restaurants(id) ON DELETE CASCADE
        ) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;`,
        `CREATE TABLE IF NOT EXISTS reservations (
            id VARCHAR(32) PRIMARY KEY,
            restaurant_id VARCHAR(32) NOT NULL,
            table_id VARCHAR(32) NOT NULL,
            user_id VARCHAR(32) NOT NULL,
            start_time DATETIME NOT NULL,
            end_time DATETIME NOT NULL,
            guests INT NOT NULL,
            status VARCHAR(32) NOT NULL,
            created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
            INDEX idx_resv_user (user_id),
            INDEX idx_resv_table (table_id),
            INDEX idx_resv_rest (restaurant_id),
            INDEX idx_resv_time (start_time, end_time),
            CONSTRAINT fk_resv_rest FOREIGN KEY (restaurant_id) REFERENCES restaurants(id) ON DELETE CASCADE,
            CONSTRAINT fk_resv_table FOREIGN KEY (table_id) REFERENCES tables(id) ON DELETE CASCADE,
            CONSTRAINT fk_resv_user FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
        ) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;`,
    }
    for _, s := range stmts {
        if _, err := db.ExecContext(ctx, s); err != nil { return err }
    }
    return nil
}

