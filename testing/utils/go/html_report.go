package testutils

import (
	"fmt"
	"html/template"
	"os"
)

const consolidatedHTMLTemplate = `<!DOCTYPE html>
<html lang="en">
<head>
<meta charset="UTF-8">
<title>MathWizz Resource Report</title>
<style>
  body { font-family: monospace; background: #1a1a2e; color: #e0e0e0; padding: 20px; }
  h1 { color: #4da6ff; }
  h2 { color: #4da6ff; margin-top: 30px; border-bottom: 1px solid #0f3460; padding-bottom: 5px; }
  .summary { background: #16213e; padding: 15px; border-radius: 8px; margin-bottom: 20px; }
  .summary span { margin-right: 30px; }
  .suite-summary { background: #16213e; padding: 10px 15px; border-radius: 8px; margin-bottom: 10px; font-size: 14px; }
  .suite-summary span { margin-right: 20px; }
  table { border-collapse: collapse; width: 100%; margin-bottom: 30px; }
  th { background: #0f3460; color: #4da6ff; text-align: left; padding: 8px 12px; }
  td { padding: 8px 12px; border-bottom: 1px solid #16213e; }
  tr:hover { background: #16213e; }
  .passed { color: #3bda48; }
  .failed { color: #ff4757; }
  .skipped { color: #ffa502; }
</style>
</head>
<body>
<h1>MathWizz Resource Report</h1>
<div class="summary">
  <span>Total: {{.TotalTests}}</span>
  <span class="passed">Passed: {{.Passed}}</span>
  <span class="failed">Failed: {{.Failed}}</span>
  <span class="skipped">Skipped: {{.Skipped}}</span>
  <span>Duration: {{printf "%.1f" .TotalDurationMs}}ms</span>
  <span>Peak Heap: {{.PeakHeapAlloc}} bytes</span>
  <span>Slowest: {{.SlowestTest}} ({{printf "%.1f" .SlowestTimeMs}}ms)</span>
  <span>Suites: {{len .Suites}}</span>
</div>
{{range .Suites}}
<h2>{{.SuiteName}}</h2>
<div class="suite-summary">
  <span>Tests: {{.TotalTests}}</span>
  <span class="passed">Passed: {{.Passed}}</span>
  <span class="failed">Failed: {{.Failed}}</span>
  <span>Duration: {{printf "%.1f" .TotalDurationMs}}ms</span>
  <span>Peak Heap: {{.PeakHeapAlloc}} bytes</span>
</div>
<table>
<thead>
  <tr>
    <th>Test</th>
    <th>Status</th>
    <th>Duration (ms)</th>
    <th>Heap Delta (bytes)</th>
    <th>Total Alloc (bytes)</th>
    <th>Goroutines</th>
  </tr>
</thead>
<tbody>
{{range .Tests}}
  <tr>
    <td>{{.Name}}</td>
    <td class="{{.Status}}">{{.Status}}</td>
    <td>{{printf "%.1f" .DurationMs}}</td>
    <td>{{.HeapAllocDelta}}</td>
    <td>{{.TotalAlloc}}</td>
    <td>{{.Goroutines}}</td>
  </tr>
{{end}}
</tbody>
</table>
{{end}}
</body>
</html>`

const perSuiteHTMLTemplate = `<!DOCTYPE html>
<html lang="en">
<head>
<meta charset="UTF-8">
<title>Resource Report: {{.SuiteName}}</title>
<style>
  body { font-family: monospace; background: #1a1a2e; color: #e0e0e0; padding: 20px; }
  h1 { color: #4da6ff; }
  .summary { background: #16213e; padding: 15px; border-radius: 8px; margin-bottom: 20px; }
  .summary span { margin-right: 30px; }
  table { border-collapse: collapse; width: 100%; }
  th { background: #0f3460; color: #4da6ff; text-align: left; padding: 8px 12px; }
  td { padding: 8px 12px; border-bottom: 1px solid #16213e; }
  tr:hover { background: #16213e; }
  .passed { color: #3bda48; }
  .failed { color: #ff4757; }
  .skipped { color: #ffa502; }
</style>
</head>
<body>
<h1>Resource Report: {{.SuiteName}}</h1>
<div class="summary">
  <span>Total: {{.TotalTests}}</span>
  <span class="passed">Passed: {{.Passed}}</span>
  <span class="failed">Failed: {{.Failed}}</span>
  <span class="skipped">Skipped: {{.Skipped}}</span>
  <span>Duration: {{printf "%.1f" .TotalDurationMs}}ms</span>
  <span>Peak Heap: {{.PeakHeapAlloc}} bytes</span>
  <span>Slowest: {{.SlowestTest}} ({{printf "%.1f" .SlowestTimeMs}}ms)</span>
</div>
<table>
<thead>
  <tr>
    <th>Test</th>
    <th>Status</th>
    <th>Duration (ms)</th>
    <th>Heap Delta (bytes)</th>
    <th>Total Alloc (bytes)</th>
    <th>Goroutines</th>
  </tr>
</thead>
<tbody>
{{range .Tests}}
  <tr>
    <td>{{.Name}}</td>
    <td class="{{.Status}}">{{.Status}}</td>
    <td>{{printf "%.1f" .DurationMs}}</td>
    <td>{{.HeapAllocDelta}}</td>
    <td>{{.TotalAlloc}}</td>
    <td>{{.Goroutines}}</td>
  </tr>
{{end}}
</tbody>
</table>
</body>
</html>`

func WriteConsolidatedHTMLReport(report ConsolidatedReport, path string) error {
	tmpl, err := template.New("consolidated").Parse(consolidatedHTMLTemplate)
	if err != nil {
		return fmt.Errorf("failed to parse HTML template: %w", err)
	}

	f, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("failed to create HTML file: %w", err)
	}
	defer f.Close()

	if err := tmpl.Execute(f, report); err != nil {
		return fmt.Errorf("failed to execute HTML template: %w", err)
	}

	return nil
}

func WriteHTMLReport(report SuiteReport, path string) error {
	tmpl, err := template.New("report").Parse(perSuiteHTMLTemplate)
	if err != nil {
		return fmt.Errorf("failed to parse HTML template: %w", err)
	}

	f, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("failed to create HTML file: %w", err)
	}
	defer f.Close()

	if err := tmpl.Execute(f, report); err != nil {
		return fmt.Errorf("failed to execute HTML template: %w", err)
	}

	return nil
}
