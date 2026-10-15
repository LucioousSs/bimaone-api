package database

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/jackc/pgx/v5/pgxpool"
)

// DB adalah instance global untuk connection pool PostgreSQL
var DB *pgxpool.Pool

// ConnectDB melakukan inisialisasi koneksi ke database menggunakan pgxpool
func ConnectDB() error {
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		// Fallback jika menggunakan variabel terpisah di .env
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

	// Test koneksi
	if err := DB.Ping(context.Background()); err != nil {
		return fmt.Errorf("gagal melakukan ping ke database: %w", err)
	}

	log.Println("Berhasil terhubung ke database PostgreSQL!")
	return nil
}

// CloseDB menutup koneksi database
func CloseDB() {
	if DB != nil {
		DB.Close()
	}
}
