package database

import (
    "context"
    "fmt"
    "log"
    "os"

    "github.com/jackc/pgx/v5/pgxpool"
)

var DB *pgxpool.Pool

func ConnectDB() error {
    dsn := os.Getenv("DATABASE_URL")
    if dsn == "" {
        host := os.Getenv("DB_HOST")
        port := os.Getenv("DB_PORT")
        user := os.Getenv("DB_USER")
        password := os.Getenv("DB_PASSWORD")
        dbname := os.Getenv("DB_NAME")

        dsn = fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable", user, password, host, port, dbname)
    }

    var err error
    DB, err = pgxpool.New(context.Background(), dsn)
    if err != nil {
        return fmt.Errorf("gagal membuat connection pool: %w", err)
    }

    if err := DB.Ping(context.Background()); err != nil {
        return fmt.Errorf("gagal melakukan ping ke database: %w", err)
    }

    log.Println("Berhasil terhubung ke database PostgreSQL!")
    return nil
}

func CloseDB() {
    if DB != nil {
        DB.Close()
    }
}
