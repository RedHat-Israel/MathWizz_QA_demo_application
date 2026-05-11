const fs = require('fs');
const path = require('path');

function mergeReports(reportDir) {
  const files = fs.readdirSync(reportDir).filter((f) => f.endsWith('.json'));
  const suites = [];
  let totalTests = 0, passed = 0, failed = 0, skipped = 0;
  let totalDurationMs = 0, peakHeapAlloc = 0;
  let slowestTest = '', slowestTimeMs = 0;

  for (const file of files) {
    const data = JSON.parse(fs.readFileSync(path.join(reportDir, file), 'utf-8'));
    suites.push(data);
    totalTests += data.total_tests || 0;
    passed += data.passed || 0;
    failed += data.failed || 0;
    skipped += data.skipped || 0;
    totalDurationMs += data.total_duration_ms || 0;
    if ((data.peak_heap_alloc || 0) > peakHeapAlloc) peakHeapAlloc = data.peak_heap_alloc;
    if ((data.slowest_time_ms || 0) > slowestTimeMs) {
      slowestTimeMs = data.slowest_time_ms;
      slowestTest = data.slowest_test;
    }
  }

  return {
    timestamp: new Date().toISOString(),
    total_tests: totalTests,
    passed, failed, skipped,
    total_duration_ms: totalDurationMs,
    peak_heap_alloc: peakHeapAlloc,
    slowest_test: slowestTest,
    slowest_time_ms: slowestTimeMs,
    suites,
  };
}

function generateConsolidatedHTMLReport(consolidated, outputPath) {
  const suiteSections = consolidated.suites.map((suite) => {
    const rows = (suite.tests || [])
      .map(
        (t) => `  <tr>
    <td>${t.name}</td>
    <td class="${t.status}">${t.status}</td>
    <td>${(t.duration_ms || 0).toFixed(1)}</td>
    <td>${t.heap_alloc_delta || 0}</td>
    <td>${t.total_alloc || 0}</td>
    <td>${t.goroutines || 0}</td>
  </tr>`
      )
      .join('\n');

    return `<h2>${suite.suite_name}</h2>
<div class="suite-summary">
  <span>Tests: ${suite.total_tests}</span>
  <span class="passed">Passed: ${suite.passed}</span>
  <span class="failed">Failed: ${suite.failed}</span>
  <span>Duration: ${(suite.total_duration_ms || 0).toFixed(1)}ms</span>
  <span>Peak Heap: ${suite.peak_heap_alloc || 0} bytes</span>
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
${rows}
</tbody>
</table>`;
  }).join('\n');

  const html = `<!DOCTYPE html>
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
  <span>Total: ${consolidated.total_tests}</span>
  <span class="passed">Passed: ${consolidated.passed}</span>
  <span class="failed">Failed: ${consolidated.failed}</span>
  <span class="skipped">Skipped: ${consolidated.skipped}</span>
  <span>Duration: ${consolidated.total_duration_ms.toFixed(1)}ms</span>
  <span>Peak Heap: ${consolidated.peak_heap_alloc} bytes</span>
  <span>Slowest: ${consolidated.slowest_test} (${consolidated.slowest_time_ms.toFixed(1)}ms)</span>
  <span>Suites: ${consolidated.suites.length}</span>
</div>
${suiteSections}
</body>
</html>`;

  fs.writeFileSync(outputPath, html, 'utf-8');
}

module.exports = { mergeReports, generateConsolidatedHTMLReport };
