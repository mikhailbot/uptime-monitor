package state

import (
	"database/sql"
	"time"

	_ "modernc.org/sqlite"
)

type DB struct {
	conn *sql.DB
}

type CheckResult struct {
	Name      string
	Type      string
	Status    string // "ok" or "fail"
	Message   string
	Timestamp time.Time
}

func Init(path string) (*DB, error) {
	connStr := path + "?_busy_timeout=5000&_foreign_keys=on&_journal_mode=WAL"

	conn, err := sql.Open("sqlite", connStr)
	if err != nil {
		return nil, err
	}

	if _, err := conn.Exec(`PRAGMA journal_mode = WAL;`); err != nil {
		return nil, err
	}

	if _, err := conn.Exec(`PRAGMA busy_timeout = 5000;`); err != nil {
		return nil, err
	}

	schema := `
	CREATE TABLE IF NOT EXISTS results (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		name TEXT,
		type TEXT,
		status TEXT,
		message TEXT,
		timestamp DATETIME
	);`
	if _, err := conn.Exec(schema); err != nil {
		return nil, err
	}

	return &DB{conn: conn}, nil
}

func (db *DB) SaveResult(result CheckResult) error {
	_, err := db.conn.Exec(
		`INSERT INTO results (name, type, status, message, timestamp) VALUES (?, ?, ?, ?, ?)`,
		result.Name, result.Type, result.Status, result.Message, result.Timestamp.Format(time.RFC3339Nano),
	)
	return err
}

func (db *DB) LatestStatuses() ([]CheckResult, error) {
	query := `
		SELECT name, type, status, message, MAX(timestamp)
		FROM results
		GROUP BY name
		ORDER BY name;
	`

	rows, err := db.conn.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var results []CheckResult
	for rows.Next() {
		var r CheckResult
		var ts string
		if err := rows.Scan(&r.Name, &r.Type, &r.Status, &r.Message, &ts); err != nil {
			return nil, err
		}
		r.Timestamp, _ = time.Parse(time.RFC3339Nano, ts)
		results = append(results, r)
	}

	return results, nil
}

func (db *DB) Close() error {
	return db.conn.Close()
}
