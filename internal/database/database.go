package database

import (
	"database/sql"
	"fmt"
	"os"

	_ "github.com/go-sql-driver/mysql" // MySQL driver import (side effect only)
)

// -------------------- DATA MODEL --------------------

// Alert represents a single alert record stored in MySQL.
type Alert struct {
	ID        int64   // Auto-incremented primary key
	AgentID   string  // Identifier of the agent that sent the metric
	RuleName  string  // Name of the alert rule triggered
	Metric    string  // Metric type: cpu | memory | disk
	Value     float64 // Actual observed value
	Threshold float64 // Threshold that triggered the alert
	Timestamp int64   // Unix timestamp
}

// -------------------- DATABASE INTERFACE --------------------

// Service is an abstraction over the underlying DB.
// This allows swapping MySQL for SQLite/Postgres/mock DBs easily.
type Service interface {
	InsertAlert(alert Alert) error
	GetAlertHistory() ([]Alert, error)
	Health() map[string]string
	Close() error
}

// -------------------- MYSQL IMPLEMENTATION --------------------

// MySQLService wraps the sql.DB connection.
type MySQLService struct {
	DB *sql.DB
}

// NewMySQLService creates a new MySQL connection using environment variables.
func NewMySQLService() (*MySQLService, error) {

	// Load DSN components from environment variables.
	user := os.Getenv("DB_USER")
	pass := os.Getenv("DB_PASSWORD")
	host := os.Getenv("DB_HOST")
	port := os.Getenv("DB_PORT")
	dbname := os.Getenv("DB_NAME")

	// Construct the MySQL DSN
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?parseTime=true",
		user, pass, host, port, dbname,
	)

	// Open does NOT establish a real connection yet.
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return nil, fmt.Errorf("mysql open error: %w", err)
	}

	// Ping forces a real connection attempt.
	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("mysql ping error: %w", err)
	}

	return &MySQLService{DB: db}, nil
}

// -------------------- INSERT ALERT --------------------

// InsertAlert writes a new alert event into the database.
func (s *MySQLService) InsertAlert(alert Alert) error {
	query := `
        INSERT INTO alerts (agent_id, rule_name, metric, value, threshold, timestamp)
        VALUES (?, ?, ?, ?, ?, ?)
    `
	_, err := s.DB.Exec(query,
		alert.AgentID,
		alert.RuleName,
		alert.Metric,
		alert.Value,
		alert.Threshold,
		alert.Timestamp,
	)

	if err != nil {
		return fmt.Errorf("insert alert error: %w", err)
	}

	return nil
}

// -------------------- FETCH ALERT HISTORY --------------------

// GetAlertHistory loads all alerts ordered from newest → oldest.
func (s *MySQLService) GetAlertHistory() ([]Alert, error) {
	query := `
        SELECT id, agent_id, rule_name, metric, value, threshold, timestamp
        FROM alerts
        ORDER BY id DESC
    `

	rows, err := s.DB.Query(query)
	if err != nil {
		return nil, fmt.Errorf("select alerts error: %w", err)
	}
	defer rows.Close()

	var alerts []Alert

	// Scan all rows into Alert structs.
	for rows.Next() {
		var a Alert
		if err := rows.Scan(
			&a.ID,
			&a.AgentID,
			&a.RuleName,
			&a.Metric,
			&a.Value,
			&a.Threshold,
			&a.Timestamp,
		); err != nil {
			return nil, fmt.Errorf("scan alert error: %w", err)
		}
		alerts = append(alerts, a)
	}

	return alerts, nil
}

// -------------------- HEALTH CHECK --------------------

// Health returns a simple UP/DOWN response based on DB availability.
func (s *MySQLService) Health() map[string]string {
	if err := s.DB.Ping(); err != nil {
		return map[string]string{"database": "DOWN"}
	}
	return map[string]string{"database": "UP"}
}

// -------------------- CLOSE CONNECTION --------------------

// Close gracefully closes the underlying DB connection.
func (s *MySQLService) Close() error {
	return s.DB.Close()
}
