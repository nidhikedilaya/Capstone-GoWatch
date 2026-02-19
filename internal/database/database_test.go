package database

import (
	"testing"
	"time"
)

func TestDatabase(t *testing.T) {

	// Load env vars for testing (if needed, depending on your setup)
	// You can uncomment this if using godotenv:
	// godotenv.Load("../.env")

	// Create DB service using env variables
	db, err := NewMySQLService()
	if err != nil {
		t.Fatalf("failed to connect mysql: %v", err)
	}
	defer db.Close()

	alert := Alert{
		AgentID:   "agent-123",
		RuleName:  "HighCPU",
		Metric:    "cpu_usage",
		Value:     92.5,
		Threshold: 80.0,
		Timestamp: time.Now().Unix(),
	}

	t.Run("InsertAlert", func(t *testing.T) {
		if err := db.InsertAlert(alert); err != nil {
			t.Fatalf("insert alert failed: %v", err)
		}
	})

	t.Run("GetAlertHistory", func(t *testing.T) {
		alerts, err := db.GetAlertHistory()
		if err != nil {
			t.Fatalf("get alert history failed: %v", err)
		}
		if len(alerts) == 0 {
			t.Fatalf("expected alerts but got 0")
		}
	})
}
