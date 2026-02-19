package grpc

import (
	pb "gowatch/gopherwatch/pkg/generated"
	"log/slog"
	"sync"
)

// -------------------- RULE STRUCTS --------------------

type AlertRule struct {
	Name       string
	Metric     string
	Threshold  float64
	Comparison string // ">", "<", ">=", "<="
}

// Evaluator interface
type Evaluator interface {
	Evaluate(metric *pb.MetricReport, rule AlertRule) bool
}

type SimpleEvaluator struct{}

func (e SimpleEvaluator) Evaluate(metric *pb.MetricReport, rule AlertRule) bool {
	switch rule.Metric {
	case "cpu":
		return compare(metric.CpuUsage, rule.Threshold, rule.Comparison)
	case "memory":
		return compare(metric.MemoryUsage, rule.Threshold, rule.Comparison)
	case "disk":
		return compare(metric.DiskUsage, rule.Threshold, rule.Comparison)
	default:
		return false
	}
}

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

var (
	metricChan = make(chan *pb.MetricReport, 200)
	state      sync.Map // map[string]CurrentState
)

type CurrentState struct {
	CpuUsage    float64
	MemoryUsage float64
	DiskUsage   float64
	Timestamp   int64
}

// -------------------- RULES --------------------

var rules = []AlertRule{
	{"High CPU", "cpu", 90.0, ">"},
	{"High Memory", "memory", 85.0, ">"},
	{"High Disk", "disk", 95.0, ">"},
}

// -------------------- WORKER POOL --------------------

func StartWorkers(n int) {
	evaluator := SimpleEvaluator{}

	for i := 0; i < n; i++ {
		go func(id int) {
			for metric := range metricChan {

				// Update state
				state.Store(metric.AgentId, CurrentState{
					CpuUsage:    metric.CpuUsage,
					MemoryUsage: metric.MemoryUsage,
					DiskUsage:   metric.DiskUsage,
					Timestamp:   metric.Timestamp,
				})

				// Evaluate rules
				for _, r := range rules {
					if evaluator.Evaluate(metric, r) {
						slog.Error("CRITICAL ALERT",
							"agent", metric.AgentId,
							"rule", r.Name,
							"value", getValue(metric, r.Metric),
							"threshold", r.Threshold,
						)
					}
				}
			}
		}(i)
	}
}

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
