package testutils

import "time"

type TestMetrics struct {
	Name            string        `json:"name"`
	Status          string        `json:"status"`
	Duration        time.Duration `json:"duration_ns"`
	DurationMs      float64       `json:"duration_ms"`
	HeapAllocBefore uint64        `json:"heap_alloc_before"`
	HeapAllocAfter  uint64        `json:"heap_alloc_after"`
	HeapAllocDelta  int64         `json:"heap_alloc_delta"`
	TotalAlloc      uint64        `json:"total_alloc"`
	NumGCBefore     uint32        `json:"num_gc_before"`
	NumGCAfter      uint32        `json:"num_gc_after"`
	Goroutines      int           `json:"goroutines"`
}

type SuiteReport struct {
	SuiteName       string        `json:"suite_name"`
	Timestamp       time.Time     `json:"timestamp"`
	TotalTests      int           `json:"total_tests"`
	Passed          int           `json:"passed"`
	Failed          int           `json:"failed"`
	Skipped         int           `json:"skipped"`
	TotalDuration   time.Duration `json:"total_duration_ns"`
	TotalDurationMs float64       `json:"total_duration_ms"`
	PeakHeapAlloc   uint64        `json:"peak_heap_alloc"`
	SlowestTest     string        `json:"slowest_test"`
	SlowestTimeMs   float64       `json:"slowest_time_ms"`
	Tests           []TestMetrics `json:"tests"`
}

type ConsolidatedReport struct {
	Timestamp       time.Time     `json:"timestamp"`
	TotalTests      int           `json:"total_tests"`
	Passed          int           `json:"passed"`
	Failed          int           `json:"failed"`
	Skipped         int           `json:"skipped"`
	TotalDurationMs float64       `json:"total_duration_ms"`
	PeakHeapAlloc   uint64        `json:"peak_heap_alloc"`
	SlowestTest     string        `json:"slowest_test"`
	SlowestTimeMs   float64       `json:"slowest_time_ms"`
	Suites          []SuiteReport `json:"suites"`
}

type Thresholds struct {
	MaxTestDurationMs   float64 `json:"max_test_duration_ms"`
	MaxMemoryDeltaBytes int64   `json:"max_memory_delta_bytes"`
	MaxTotalAllocs      uint64  `json:"max_total_allocs"`
	MaxSuiteDurationMs  float64 `json:"max_suite_duration_ms"`
}
