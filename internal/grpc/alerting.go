package grpc

import (
	pb "gowatch/gopherwatch/pkg/generated"
	"gowatch/internal/database"
	"log/slog"
	"sync"
)

// -------------------- RULE STRUCTS --------------------

// AlertRule defines a single alert condition.
// Example: "CPU > 90%"
type AlertRule struct {
	Name       string  // Human-readable rule name
	Metric     string  // Metric to evaluate: cpu | memory | disk
	Threshold  float64 // Threshold value to compare against
	Comparison string  // Operator: ">", "<", ">=", "<="
}

// Evaluator is an interface allowing custom evaluation engines.
// Future: AI-based evaluator, weighted scoring, anomaly detection, etc.
type Evaluator interface {
	Evaluate(metric *pb.MetricReport, rule AlertRule) bool
}

// SimpleEvaluator implements basic threshold comparison.
type SimpleEvaluator struct{}

// Evaluate checks whether a metric triggers a rule.
func (e SimpleEvaluator) Evaluate(metric *pb.MetricReport, rule AlertRule) bool {
	switch rule.Metric {
	case "cpu":
		return compare(metric.CpuUsage, rule.Threshold, rule.Comparison)
	case "memory":
		return compare(metric.MemoryUsage, rule.Threshold, rule.Comparison)
	case "disk":
		return compare(metric.DiskUsage, rule.Threshold, rule.Comparison)
	default:
		// Unknown metric type → never trigger
		return false
	}
}

// compare performs the actual numeric operation.
func compare(value, threshold float64, cmp string) bool {
	switch cmp {
	case ">":
		return value > threshold
	case "<":
		return value < threshold
	case ">=":
		return value >= threshold
	case "<=":
		return value <= threshold
	}
	return false
}

// -------------------- GLOBAL STATE --------------------

// MetricChan is a buffered channel where gRPC pushes incoming metric reports.
// Workers consume from this channel.
var (
	MetricChan = make(chan *pb.MetricReport, 200)

	// state stores the *latest metrics per agent*.
	// Key: AgentID (string), Value: CurrentState struct
	state sync.Map
)

// CurrentState holds the latest known status of an agent.
type CurrentState struct {
	CpuUsage    float64
	MemoryUsage float64
	DiskUsage   float64
	Timestamp   int64
}

// -------------------- PREDEFINED ALERT RULES --------------------

// These are default rules the system evaluates for every incoming metric.
var rules = []AlertRule{
	{"High CPU", "cpu", 90.0, ">"},
	{"High Memory", "memory", 85.0, ">"},
	{"High Disk", "disk", 95.0, ">"},
}

// -------------------- WORKER POOL --------------------

// StartWorkers creates N worker goroutines that:
//  1. Consume metrics from MetricChan
//  2. Update in-memory state
//  3. Evaluate alert rules
//  4. Insert alerts into the database
//
// This design ensures concurrency, scalability, and smooth load distribution.
func StartWorkers(n int, db database.Service) {
	evaluator := SimpleEvaluator{}

	for i := 0; i < n; i++ {
		go func(id int) {

			// Each goroutine continuously processes metrics
			for metric := range MetricChan {

				// -------------------- UPDATE LATEST STATE --------------------
				state.Store(metric.AgentId, CurrentState{
					CpuUsage:    metric.CpuUsage,
					MemoryUsage: metric.MemoryUsage,
					DiskUsage:   metric.DiskUsage,
					Timestamp:   metric.Timestamp,
				})

				// -------------------- EVALUATE ALERT RULES --------------------
				for _, r := range rules {
					if evaluator.Evaluate(metric, r) {

						// Log critical alert
						slog.Error("CRITICAL ALERT",
							"agent", metric.AgentId,
							"rule", r.Name,
							"value", getValue(metric, r.Metric),
							"threshold", r.Threshold,
						)

						// -------------------- STORE ALERT IN DB --------------------
						if db != nil { // database layer is optional for testing
							_ = db.InsertAlert(database.Alert{
								AgentID:   metric.AgentId,
								RuleName:  r.Name,
								Metric:    r.Metric,
								Value:     getValue(metric, r.Metric),
								Threshold: r.Threshold,
								Timestamp: metric.Timestamp,
							})
						}
					}
				}
			}
		}(i)
	}
}

// getValue maps a metric name string to the corresponding MetricReport value.
func getValue(metric *pb.MetricReport, metricName string) float64 {
	switch metricName {
	case "cpu":
		return metric.CpuUsage
	case "memory":
		return metric.MemoryUsage
	case "disk":
		return metric.DiskUsage
	default:
		return 0
	}
}
