package initializers

import (
	"database/sql"
	"log"
	"os"

	_ "github.com/lib/pq" // Import PostgreSQL driver
)

var DB *sql.DB

func ConnectToDb() {
	dsn := os.Getenv("DB_URL")
	var err error
	DB, err = sql.Open("postgres", dsn)
	if err != nil {
		log.Fatal("Failed to connect to DB:", err)
	}

	// Test connection
	err = DB.Ping()
	if err != nil {
		log.Fatal("Failed to ping DB:", err)
	}
}
