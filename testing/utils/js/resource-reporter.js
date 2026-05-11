const fs = require('fs');
const path = require('path');
const { mergeReports, generateConsolidatedHTMLReport } = require('./html-report');

class ResourceReporter {
  constructor(globalConfig, reporterOptions) {
    this._globalConfig = globalConfig;
    this._options = reporterOptions || {};
    this._reportDir = this._options.reportDir || path.resolve(__dirname, '..', '..', 'reports');
    this._tests = [];
    this._suiteStart = Date.now();
    this._memBefore = process.memoryUsage();
    this._peakHeap = 0;
  }

  onTestResult(_test, testResult) {
    const memAfter = process.memoryUsage();

    for (const result of testResult.testResults) {
      const heapDelta = memAfter.heapUsed - this._memBefore.heapUsed;
      this._tests.push({
        name: result.fullName,
        status: result.status,
        duration_ms: result.duration || 0,
        heap_alloc_before: this._memBefore.heapUsed,
        heap_alloc_after: memAfter.heapUsed,
        heap_alloc_delta: heapDelta,
        total_alloc: memAfter.heapTotal,
        goroutines: 0,
      });

      if (memAfter.heapUsed > this._peakHeap) {
        this._peakHeap = memAfter.heapUsed;
      }
    }

    this._memBefore = process.memoryUsage();
  }

  onRunComplete(_contexts, results) {
    const totalDuration = Date.now() - this._suiteStart;

    let slowestTest = '';
    let slowestTime = 0;
    let passed = 0;
    let failed = 0;
    let skipped = 0;

    for (const t of this._tests) {
      if (t.status === 'passed') passed++;
      else if (t.status === 'failed') failed++;
      else skipped++;

      if (t.duration_ms > slowestTime) {
        slowestTime = t.duration_ms;
        slowestTest = t.name;
      }
    }

    const report = {
      suite_name: 'Frontend Tests',
      timestamp: new Date().toISOString(),
      total_tests: this._tests.length,
      passed,
      failed,
      skipped,
      total_duration_ms: totalDuration,
      peak_heap_alloc: this._peakHeap,
      slowest_test: slowestTest,
      slowest_time_ms: slowestTime,
      tests: this._tests,
    };

    fs.mkdirSync(this._reportDir, { recursive: true });

    const timestamp = new Date().toISOString().replace(/[:.]/g, '-').slice(0, 19);
    const baseName = `Frontend Tests-${timestamp}`;

    const jsonPath = path.join(this._reportDir, `${baseName}.json`);
    fs.writeFileSync(jsonPath, JSON.stringify(report, null, 2), 'utf-8');
    console.log(`\nResource report (JSON): ${jsonPath}`);

    const consolidated = mergeReports(this._reportDir);
    const htmlPath = path.join(this._reportDir, 'resource-report.html');
    generateConsolidatedHTMLReport(consolidated, htmlPath);
    console.log(`Resource report (HTML): ${htmlPath}`);
  }
}

module.exports = ResourceReporter;
