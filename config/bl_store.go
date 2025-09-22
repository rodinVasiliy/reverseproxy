package config

import (
	"database/sql"
	"log"

	_ "github.com/mattn/go-sqlite3"
)

type BL struct {
	db *sql.DB
}

// Инициализация: подключаемся к базе и создаём таблицу
func NewBlacklistStore(path string) (*BL, error) {
	db, err := sql.Open("sqlite3", path)
	if err != nil {
		return nil, err
	}

	sqlStmt := `
	CREATE TABLE IF NOT EXISTS blacklist (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		ip TEXT NOT NULL UNIQUE,
		added_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
	);`
	_, err = db.Exec(sqlStmt)
	if err != nil {
		return nil, err
	}

	return &BL{db: db}, nil
}

// Добавить IP в blacklist
func (b *BL) Add(ip string) error {
	_, err := b.db.Exec("INSERT OR IGNORE INTO blacklist(ip) VALUES (?)", ip)
	return err
}

// Удалить IP из blacklist
func (b *BL) Remove(ip string) error {
	_, err := b.db.Exec("DELETE FROM blacklist WHERE ip = ?", ip)
	return err
}

// Проверить наличие IP в blacklist
func (b *BL) Exists(ip string) (bool, error) {
	var exists int
	err := b.db.QueryRow("SELECT 1 FROM blacklist WHERE ip = ?", ip).Scan(&exists)
	if err == sql.ErrNoRows {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	return true, nil
}

// Закрыть соединение с базой
func (b *BL) Close() {
	if b.db != nil {
		if err := b.db.Close(); err != nil {
			log.Printf("Error closing DB: %v", err)
		}
	}
}
