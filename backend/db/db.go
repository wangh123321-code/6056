package db

import (
	"database/sql"
	"log"

	_ "github.com/mattn/go-sqlite3"
)

var DB *sql.DB

func Init(dbPath string) {
	var err error
	DB, err = sql.Open("sqlite3", dbPath+"?_journal_mode=WAL&_busy_timeout=5000")
	if err != nil {
		log.Fatal("Failed to open database:", err)
	}

	DB.SetMaxOpenConns(1)

	migrate()
}

func migrate() {
	_, err := DB.Exec(`
	CREATE TABLE IF NOT EXISTS courses (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		title TEXT NOT NULL,
		date TEXT NOT NULL,
		time_slot TEXT NOT NULL,
		capacity INTEGER NOT NULL DEFAULT 15,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);

	CREATE TABLE IF NOT EXISTS bookings (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		course_id INTEGER NOT NULL,
		user_name TEXT NOT NULL,
		user_phone TEXT NOT NULL,
		status TEXT NOT NULL DEFAULT 'booked',
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		cancelled_at DATETIME,
		FOREIGN KEY (course_id) REFERENCES courses(id),
		UNIQUE(course_id, user_phone, status)
	);

	CREATE INDEX IF NOT EXISTS idx_bookings_course_id ON bookings(course_id);
	CREATE INDEX IF NOT EXISTS idx_bookings_user_phone ON bookings(user_phone);
	CREATE INDEX IF NOT EXISTS idx_bookings_status ON bookings(status);
	CREATE INDEX IF NOT EXISTS idx_bookings_course_status ON bookings(course_id, status);
	CREATE INDEX IF NOT EXISTS idx_courses_date ON courses(date);
	`)
	if err != nil {
		log.Fatal("Migration failed:", err)
	}
}
