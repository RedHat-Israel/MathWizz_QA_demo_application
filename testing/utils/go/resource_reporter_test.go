package testutils_test

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	testutils "github.com/mathwizz/testing/utils"
)

func TestResourceReporter(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Resource Reporter Suite")
}

var _ = Describe("Resource Reporter", func() {
	var reportDir string

	BeforeEach(func() {
		var err error
		reportDir, err = os.MkdirTemp("", "testutils-reports-*")
		Expect(err).ShouldNot(HaveOccurred())
	})

	AfterEach(func() {
		os.RemoveAll(reportDir)
	})

	Describe("CollectMetrics", func() {
		It("captures heap allocation and goroutine count", func() {
			before := testutils.SnapshotMemStats()

			// Allocate some memory
			data := make([]byte, 1024*1024)
			_ = data

			after := testutils.SnapshotMemStats()

			Expect(after.HeapAlloc).Should(BeNumerically(">=", before.HeapAlloc))
			Expect(after.Goroutines).Should(BeNumerically(">", 0))
		})
	})

	Describe("WriteSuiteReport", func() {
		It("writes a valid JSON report file", func() {
			report := testutils.SuiteReport{
				SuiteName:  "test-suite",
				TotalTests: 2,
				Passed:     1,
				Failed:     1,
				Tests: []testutils.TestMetrics{
					{Name: "passes", Status: "passed", DurationMs: 10.5, HeapAllocDelta: 1024},
					{Name: "fails", Status: "failed", DurationMs: 20.3, HeapAllocDelta: 2048},
				},
			}

			outPath := filepath.Join(reportDir, "report.json")
			err := testutils.WriteSuiteReport(report, outPath)
			Expect(err).ShouldNot(HaveOccurred())

			data, err := os.ReadFile(outPath)
			Expect(err).ShouldNot(HaveOccurred())

			var loaded testutils.SuiteReport
			Expect(json.Unmarshal(data, &loaded)).Should(Succeed())
			Expect(loaded.SuiteName).Should(Equal("test-suite"))
			Expect(loaded.Tests).Should(HaveLen(2))
		})
	})

	Describe("CheckThresholds", func() {
		It("returns violations when metrics exceed thresholds", func() {
			report := testutils.SuiteReport{
				TotalDurationMs: 70000,
				Tests: []testutils.TestMetrics{
					{Name: "slow-test", DurationMs: 6000, HeapAllocDelta: 20_000_000},
				},
			}
			thresholds := testutils.Thresholds{
				MaxTestDurationMs:   5000,
				MaxMemoryDeltaBytes: 10_485_760,
				MaxSuiteDurationMs:  60000,
			}

			violations := testutils.CheckThresholds(report, thresholds)
			Expect(violations).Should(HaveLen(3))
		})

		It("returns no violations when all metrics are within thresholds", func() {
			report := testutils.SuiteReport{
				TotalDurationMs: 5000,
				Tests: []testutils.TestMetrics{
					{Name: "fast-test", DurationMs: 100, HeapAllocDelta: 1024},
				},
			}
			thresholds := testutils.Thresholds{
				MaxTestDurationMs:   5000,
				MaxMemoryDeltaBytes: 10_485_760,
				MaxSuiteDurationMs:  60000,
			}

			violations := testutils.CheckThresholds(report, thresholds)
			Expect(violations).Should(BeEmpty())
		})
	})

	Describe("MergeReports and WriteConsolidatedHTMLReport", func() {
		It("merges multiple JSON reports into one consolidated HTML", func() {
			suite1 := testutils.SuiteReport{
				SuiteName:       "Web-Server Suite",
				TotalTests:      2,
				Passed:          2,
				TotalDurationMs: 100.0,
				PeakHeapAlloc:   4096,
				SlowestTest:     "solver test",
				SlowestTimeMs:   80.0,
				Tests: []testutils.TestMetrics{
					{Name: "solver test", Status: "passed", DurationMs: 80.0, HeapAllocDelta: 1024},
					{Name: "handler test", Status: "passed", DurationMs: 20.0, HeapAllocDelta: 512},
				},
			}
			suite2 := testutils.SuiteReport{
				SuiteName:       "History-Worker Suite",
				TotalTests:      1,
				Passed:          1,
				TotalDurationMs: 50.0,
				PeakHeapAlloc:   8192,
				SlowestTest:     "worker test",
				SlowestTimeMs:   50.0,
				Tests: []testutils.TestMetrics{
					{Name: "worker test", Status: "passed", DurationMs: 50.0, HeapAllocDelta: 2048},
				},
			}

			Expect(testutils.WriteSuiteReport(suite1, filepath.Join(reportDir, "web-server.json"))).Should(Succeed())
			Expect(testutils.WriteSuiteReport(suite2, filepath.Join(reportDir, "history-worker.json"))).Should(Succeed())

			consolidated, err := testutils.MergeReports(reportDir)
			Expect(err).ShouldNot(HaveOccurred())
			Expect(consolidated.TotalTests).Should(Equal(3))
			Expect(consolidated.Passed).Should(Equal(3))
			Expect(consolidated.Suites).Should(HaveLen(2))
			Expect(consolidated.PeakHeapAlloc).Should(Equal(uint64(8192)))
			Expect(consolidated.SlowestTest).Should(Equal("solver test"))

			htmlPath := filepath.Join(reportDir, "resource-report.html")
			err = testutils.WriteConsolidatedHTMLReport(consolidated, htmlPath)
			Expect(err).ShouldNot(HaveOccurred())

			data, err := os.ReadFile(htmlPath)
			Expect(err).ShouldNot(HaveOccurred())

			html := string(data)
			Expect(html).Should(ContainSubstring("Web-Server Suite"))
			Expect(html).Should(ContainSubstring("History-Worker Suite"))
			Expect(html).Should(ContainSubstring("solver test"))
			Expect(html).Should(ContainSubstring("worker test"))
			Expect(html).Should(ContainSubstring("<!DOCTYPE html>"))
		})
	})
})
