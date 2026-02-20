package database

import (
	"context"
	"testing"
	"time"
)

func TestDatabase(t *testing.T) {

	db, err := NewMySQLService()
	if err != nil {
		t.Fatalf("failed to connect mysql: %v", err)
	}
	defer db.Close()

	alert := Alert{
		AgentID:     "agent-123",
		ServiceName: "service-A",
		RuleName:    "HighCPU",
		Metric:      "cpu_usage",
		Value:       92.5,
		Threshold:   80.0,
		Timestamp:   time.Now().Unix(),
	}

	ctx := context.Background()

	t.Run("InsertAlert", func(t *testing.T) {
		if err := db.InsertAlert(ctx, alert); err != nil {
			t.Fatalf("insert alert failed: %v", err)
		}
	})

	t.Run("GetAlertHistory", func(t *testing.T) {
		alerts, err := db.GetAlertHistory(ctx)
		if err != nil {
			t.Fatalf("get alert history failed: %v", err)
		}
		if len(alerts) == 0 {
			t.Fatalf("expected alerts but got 0")
		}
	})
}
