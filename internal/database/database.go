package database

import (
	"database/sql"
	"fmt"

	_ "github.com/go-sql-driver/mysql"
)

type Alert struct {
	ID        int64
	AgentID   string
	RuleName  string
	Metric    string
	Value     float64
	Threshold float64
	Timestamp int64
}

type Service interface {
	InsertAlert(alert Alert) error
	GetAlertHistory() ([]Alert, error)
	Health() map[string]string
	Close() error
}

type MySQLService struct {
	DB *sql.DB
}

func NewMySQLService(user, pass, host, dbname string) (*MySQLService, error) {
	dsn := fmt.Sprintf("%s:%s@tcp(%s)/%s?parseTime=true",
		user, pass, host, dbname,
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

func (s *MySQLService) Health() map[string]string {
	if err := s.DB.Ping(); err != nil {
		return map[string]string{"database": "DOWN"}
	}
	return map[string]string{"database": "UP"}
}

func (s *MySQLService) Close() error {
	return s.DB.Close()
}
