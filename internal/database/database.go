package database

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"time"

	_ "github.com/go-sql-driver/mysql"
)

//
// -------------------- DATA MODEL --------------------
//

type Alert struct {
	ID          int64   `json:"id"`
	AgentID     string  `json:"agent_id"`
	ServiceName string  `json:"service_name"`
	RuleName    string  `json:"rule_name"`
	Metric      string  `json:"metric"`
	Value       float64 `json:"value"`
	Threshold   float64 `json:"threshold"`
	Timestamp   int64   `json:"timestamp"` // Unix timestamp
}

//
// -------------------- SERVICE INTERFACE --------------------
//

type Service interface {
	InsertAlert(ctx context.Context, alert Alert) error
	GetAlertHistory(ctx context.Context) ([]Alert, error)
	Health() map[string]string
	Close() error
}

//
// -------------------- MYSQL SERVICE --------------------
//

type MySQLService struct {
	DB *sql.DB
}

//
// -------------------- CONNECT --------------------
//

func NewMySQLService() (*MySQLService, error) {

	user := os.Getenv("DB_USER")
	pass := os.Getenv("DB_PASSWORD")
	host := os.Getenv("DB_HOST")
	port := os.Getenv("DB_PORT")
	dbname := os.Getenv("DB_NAME")

	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?parseTime=true",
		user, pass, host, port, dbname,
	)

	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return nil, fmt.Errorf("mysql open error: %w", err)
	}

	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("mysql ping error: %w", err)
	}

	return &MySQLService{DB: db}, nil
}

//
// -------------------- INSERT ALERT --------------------
//

func (s *MySQLService) InsertAlert(ctx context.Context, alert Alert) error {

	// Prevent DB hanging forever
	ctx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()

	// Store UNIX time safely using FROM_UNIXTIME
	query := `
        INSERT INTO alerts 
            (agent_id, service_name, rule_name, metric, value, threshold, timestamp)
        VALUES 
            (?, ?, ?, ?, ?, ?, FROM_UNIXTIME(?))
    `

	_, err := s.DB.ExecContext(ctx, query,
		alert.AgentID,
		alert.ServiceName,
		alert.RuleName,
		alert.Metric,
		alert.Value,
		alert.Threshold,
		alert.Timestamp, // Raw Unix time → converted in SQL
	)

	if err != nil {
		return fmt.Errorf("insert alert error: %w", err)
	}

	return nil
}

//
// -------------------- GET ALERT HISTORY --------------------
//

func (s *MySQLService) GetAlertHistory(ctx context.Context) ([]Alert, error) {

	ctx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()

	// Return Unix timestamp using UNIX_TIMESTAMP()
	query := `
        SELECT 
            id, 
            agent_id, 
            service_name, 
            rule_name, 
            metric, 
            value, 
            threshold, 
            UNIX_TIMESTAMP(timestamp)
        FROM alerts
        ORDER BY id DESC
    `

	rows, err := s.DB.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("select alerts error: %w", err)
	}
	defer rows.Close()

	var alerts []Alert

	for rows.Next() {
		var a Alert
		if err := rows.Scan(
			&a.ID,
			&a.AgentID,
			&a.ServiceName,
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

//
// -------------------- HEALTH CHECK --------------------
//

func (s *MySQLService) Health() map[string]string {
	if err := s.DB.Ping(); err != nil {
		return map[string]string{"database": "DOWN"}
	}
	return map[string]string{"database": "UP"}
}

//
// -------------------- CLOSE CONNECTION --------------------
//

func (s *MySQLService) Close() error {
	return s.DB.Close()
}
