package db

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	_ "github.com/go-sql-driver/mysql"
)

type DB struct {
	conn *sql.DB
}

type Event struct {
	ID     uint      `json:"id"`
	BotID  string    `json:"bot_id"`
	Ts     time.Time `json:"ts"`
	Symbol string    `json:"symbol"`
	Source string    `json:"source"`
	USDVal float64   `json:"usd_val"`
	Text   string    `json:"text"`
}

type EventVec struct {
	ID    uint            `json:"id"`
	BotID string          `json:"bot_id"`
	Ts    time.Time       `json:"ts"`
	Sym   string          `json:"sym"`
	Vec   json.RawMessage `json:"vec"`
	Text  string          `json:"text"`
}

type Prediction struct {
	ID     uint      `json:"id"`
	BotID  string    `json:"bot_id"`
	Ts     time.Time `json:"ts"`
	Symbol string    `json:"symbol"`
	Dir    string    `json:"dir"`
	Conv   int       `json:"conv"`
	Logic  string    `json:"logic"`
	FwdRet *float64  `json:"fwd_ret"`
}

type Trade struct {
	ID     uint      `json:"id"`
	BotID  string    `json:"bot_id"`
	Ts     time.Time `json:"ts"`
	Symbol string    `json:"symbol"`
	Side   string    `json:"side"`
	Qty    float64   `json:"qty"`
	Price  float64   `json:"price"`
	Status string    `json:"status"`
}

func Open(dsn string) (*DB, error) {
	conn, err := sql.Open("mysql", dsn)
	if err != nil {
		return nil, err
	}

	db := &DB{conn: conn}
	return db, nil
}

func (db *DB) GetConn() *sql.DB {
	return db.conn
}

func AutoMigrate(db *DB) error {
	if db == nil || db.conn == nil {
		return fmt.Errorf("database connection is nil")
	}

	queries := []string{
		// Create database if not exists
		`CREATE DATABASE IF NOT EXISTS sigforge`,
		`USE sigforge`,

		// Events table
		`CREATE TABLE IF NOT EXISTS events (
			id BIGINT AUTO_INCREMENT,
			bot_id VARCHAR(32) NOT NULL,
			ts DATETIME NOT NULL,
			symbol VARCHAR(16) NOT NULL,
			source ENUM('news','chain') NOT NULL,
			usd_val DOUBLE,
			text TEXT,
			PRIMARY KEY (bot_id, id),
			KEY idx_sym_ts (symbol, ts)
		)`,

		// Event vectors table with TTL
		`CREATE TABLE IF NOT EXISTS event_vecs (
			id BIGINT,
			bot_id VARCHAR(32) NOT NULL,
			ts DATETIME NOT NULL,
			sym VARCHAR(16) NOT NULL,
			vec JSON,
			text TEXT,
			PRIMARY KEY (bot_id, id)
		) TTL = ts + INTERVAL 30 DAY`,

		// Predictions table
		`CREATE TABLE IF NOT EXISTS predictions (
			id BIGINT AUTO_INCREMENT,
			bot_id VARCHAR(32) NOT NULL,
			ts DATETIME NOT NULL,
			symbol VARCHAR(16) NOT NULL,
			dir ENUM('LONG','SHORT','FLAT') NOT NULL,
			conv TINYINT NOT NULL,
			logic TEXT,
			fwd_ret DOUBLE,
			PRIMARY KEY (bot_id, id)
		)`,

		// Trades table
		`CREATE TABLE IF NOT EXISTS trades (
			id BIGINT AUTO_INCREMENT,
			bot_id VARCHAR(32) NOT NULL,
			ts DATETIME NOT NULL,
			symbol VARCHAR(16) NOT NULL,
			side VARCHAR(8) NOT NULL,
			qty DOUBLE NOT NULL,
			price DOUBLE NOT NULL,
			status VARCHAR(20) NOT NULL,
			PRIMARY KEY (bot_id, id)
		)`,
	}

	for _, query := range queries {
		if _, err := db.conn.Exec(query); err != nil {
			return err
		}
	}

	return nil
}
