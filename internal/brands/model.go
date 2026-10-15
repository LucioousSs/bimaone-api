package brands

import "time"

// Brand adalah struct yang merepresentasikan tabel brands di database
type Brand struct {
    ID          int       `json:"id"`
    Name        string    `json:"name"`
    Description string    `json:"description"`
    CreatedAt   time.Time `json:"created_at"`
    UpdatedAt   time.Time `json:"updated_at"`
}