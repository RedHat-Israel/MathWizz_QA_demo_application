package testutils

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"sync"
	"time"

	. "github.com/onsi/ginkgo/v2"
	"github.com/onsi/ginkgo/v2/types"
)

type MemSnapshot struct {
	HeapAlloc  uint64
	TotalAlloc uint64
	NumGC      uint32
	Goroutines int
}

func SnapshotMemStats() MemSnapshot {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	return MemSnapshot{
		HeapAlloc:  m.HeapAlloc,
		TotalAlloc: m.TotalAlloc,
		NumGC:      m.NumGC,
		Goroutines: runtime.NumGoroutine(),
	}
}

var (
	mu               sync.Mutex
	beforeSnapshots  map[string]MemSnapshot
	collectedMetrics []TestMetrics
	peakHeap         uint64
)

func init() {
	beforeSnapshots = make(map[string]MemSnapshot)
}

func AttachResourceReporter(reportDir string) bool {
	collectedMetrics = nil
	peakHeap = 0

	BeforeEach(func() {
		specName := CurrentSpecReport().FullText()
		mu.Lock()
		beforeSnapshots[specName] = SnapshotMemStats()
		mu.Unlock()
	})

	ReportAfterEach(func(report SpecReport) {
		after := SnapshotMemStats()
		specName := report.FullText()

		mu.Lock()
		before, exists := beforeSnapshots[specName]
		if !exists {
			before = MemSnapshot{}
		}
		delete(beforeSnapshots, specName)
		mu.Unlock()

		status := "passed"
		if report.State.Is(types.SpecStateFailed) {
			status = "failed"
		} else if report.State.Is(types.SpecStateSkipped) || report.State.Is(types.SpecStatePending) {
			status = "skipped"
		}

		delta := int64(after.HeapAlloc) - int64(before.HeapAlloc)

		metrics := TestMetrics{
			Name:            specName,
			Status:          status,
			Duration:        report.RunTime,
			DurationMs:      float64(report.RunTime.Milliseconds()),
			HeapAllocBefore: before.HeapAlloc,
			HeapAllocAfter:  after.HeapAlloc,
			HeapAllocDelta:  delta,
			TotalAlloc:      after.TotalAlloc - before.TotalAlloc,
			NumGCBefore:     before.NumGC,
			NumGCAfter:      after.NumGC,
			Goroutines:      after.Goroutines,
		}

		mu.Lock()
		collectedMetrics = append(collectedMetrics, metrics)
		if after.HeapAlloc > peakHeap {
			peakHeap = after.HeapAlloc
		}
		mu.Unlock()
	})

	ReportAfterSuite("Resource Report", func(suiteReport Report) {
		mu.Lock()
		metrics := make([]TestMetrics, len(collectedMetrics))
		copy(metrics, collectedMetrics)
		peak := peakHeap
		mu.Unlock()

		passed, failed, skipped := 0, 0, 0
		var slowestName string
		var slowestMs float64

		for _, m := range metrics {
			switch m.Status {
			case "passed":
				passed++
			case "failed":
				failed++
			case "skipped":
				skipped++
			}
			if m.DurationMs > slowestMs {
				slowestMs = m.DurationMs
				slowestName = m.Name
			}
		}

		report := SuiteReport{
			SuiteName:       suiteReport.SuiteDescription,
			Timestamp:       time.Now(),
			TotalTests:      len(metrics),
			Passed:          passed,
			Failed:          failed,
			Skipped:         skipped,
			TotalDuration:   suiteReport.RunTime,
			TotalDurationMs: float64(suiteReport.RunTime.Milliseconds()),
			PeakHeapAlloc:   peak,
			SlowestTest:     slowestName,
			SlowestTimeMs:   slowestMs,
			Tests:           metrics,
		}

		if err := os.MkdirAll(reportDir, 0o755); err != nil {
			fmt.Fprintf(os.Stderr, "failed to create report dir: %v\n", err)
			return
		}

		timestamp := time.Now().Format("2006-01-02T150405")
		baseName := fmt.Sprintf("%s-%s", suiteReport.SuiteDescription, timestamp)

		jsonPath := filepath.Join(reportDir, baseName+".json")
		if err := WriteSuiteReport(report, jsonPath); err != nil {
			fmt.Fprintf(os.Stderr, "failed to write JSON report: %v\n", err)
			return
		}
		fmt.Fprintf(os.Stdout, "Resource report (JSON): %s\n", jsonPath)

		consolidated, err := MergeReports(reportDir)
		if err != nil {
			fmt.Fprintf(os.Stderr, "failed to merge reports: %v\n", err)
			return
		}
		htmlPath := filepath.Join(reportDir, "resource-report.html")
		if err := WriteConsolidatedHTMLReport(consolidated, htmlPath); err != nil {
			fmt.Fprintf(os.Stderr, "failed to write HTML report: %v\n", err)
			return
		}
		fmt.Fprintf(os.Stdout, "Resource report (HTML): %s\n", htmlPath)
	})

	return true
}

func WriteSuiteReport(report SuiteReport, path string) error {
	data, err := json.MarshalIndent(report, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal report: %w", err)
	}
	return os.WriteFile(path, data, 0o644)
}

func MergeReports(reportDir string) (ConsolidatedReport, error) {
	entries, err := os.ReadDir(reportDir)
	if err != nil {
		return ConsolidatedReport{}, fmt.Errorf("failed to read report dir: %w", err)
	}

	var consolidated ConsolidatedReport
	consolidated.Timestamp = time.Now()

	for _, entry := range entries {
		if entry.IsDir() || filepath.Ext(entry.Name()) != ".json" {
			continue
		}

		data, err := os.ReadFile(filepath.Join(reportDir, entry.Name()))
		if err != nil {
			continue
		}

		var suite SuiteReport
		if err := json.Unmarshal(data, &suite); err != nil {
			continue
		}

		consolidated.Suites = append(consolidated.Suites, suite)
		consolidated.TotalTests += suite.TotalTests
		consolidated.Passed += suite.Passed
		consolidated.Failed += suite.Failed
		consolidated.Skipped += suite.Skipped
		consolidated.TotalDurationMs += suite.TotalDurationMs

		if suite.PeakHeapAlloc > consolidated.PeakHeapAlloc {
			consolidated.PeakHeapAlloc = suite.PeakHeapAlloc
		}
		if suite.SlowestTimeMs > consolidated.SlowestTimeMs {
			consolidated.SlowestTimeMs = suite.SlowestTimeMs
			consolidated.SlowestTest = suite.SlowestTest
		}
	}

	return consolidated, nil
}

func CheckThresholds(report SuiteReport, thresholds Thresholds) []string {
	var violations []string

	if thresholds.MaxSuiteDurationMs > 0 && report.TotalDurationMs > thresholds.MaxSuiteDurationMs {
		violations = append(violations, fmt.Sprintf(
			"suite duration %.0fms exceeds threshold %.0fms",
			report.TotalDurationMs, thresholds.MaxSuiteDurationMs))
	}

	for _, t := range report.Tests {
		if thresholds.MaxTestDurationMs > 0 && t.DurationMs > thresholds.MaxTestDurationMs {
			violations = append(violations, fmt.Sprintf(
				"test %q took %.0fms, exceeds threshold %.0fms",
				t.Name, t.DurationMs, thresholds.MaxTestDurationMs))
		}
		if thresholds.MaxMemoryDeltaBytes > 0 && t.HeapAllocDelta > thresholds.MaxMemoryDeltaBytes {
			violations = append(violations, fmt.Sprintf(
				"test %q used %d bytes, exceeds threshold %d bytes",
				t.Name, t.HeapAllocDelta, thresholds.MaxMemoryDeltaBytes))
		}
	}

	return violations
}
