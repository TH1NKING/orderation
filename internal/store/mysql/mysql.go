package mysql

import (
    "context"
    "database/sql"
    _ "github.com/go-sql-driver/mysql"
)

func Open(dsn string) (*sql.DB, error) {
    db, err := sql.Open("mysql", dsn)
    if err != nil { return nil, err }
    if err := db.Ping(); err != nil { return nil, err }
    return db, nil
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

