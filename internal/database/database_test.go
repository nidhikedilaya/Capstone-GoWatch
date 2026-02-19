package database

import (
	"testing"
	"time"
)

func TestDatabase(t *testing.T) {

	// -------------------------------------------------------
	// Initialize database (using environment variables)
	//
	// NOTE:
	// - Ensure DB_USER, DB_PASSWORD, DB_HOST, DB_PORT, DB_NAME
	//   are set before running tests.
	//
	// - If your test environment requires dotenv loading,
	//   uncomment the godotenv loader below.
	// -------------------------------------------------------

	// godotenv.Load("../.env")

	db, err := NewMySQLService()
	if err != nil {
		t.Fatalf("failed to connect mysql: %v", err)
	}
	defer db.Close()

	// -------------------------------------------------------
	// Create a sample alert record for testing Insert + Fetch
	// -------------------------------------------------------
	alert := Alert{
		AgentID:   "agent-123",
		RuleName:  "HighCPU",
		Metric:    "cpu_usage",
		Value:     92.5,
		Threshold: 80.0,
		Timestamp: time.Now().Unix(),
	}

	// -------------------------------------------------------
	// TEST: InsertAlert
	// -------------------------------------------------------
	t.Run("InsertAlert", func(t *testing.T) {
		if err := db.InsertAlert(alert); err != nil {
			t.Fatalf("insert alert failed: %v", err)
		}
	})

	// -------------------------------------------------------
	// TEST: GetAlertHistory
	//
	// Confirms:
	// - History loads without error
	// - At least one record exists (the one we inserted)
	// -------------------------------------------------------
	t.Run("GetAlertHistory", func(t *testing.T) {
		alerts, err := db.GetAlertHistory()
		if err != nil {
			t.Fatalf("get alert history failed: %v", err)
		}

		if len(alerts) == 0 {
			t.Fatalf("expected at least 1 alert but got 0")
		}
	})
}
