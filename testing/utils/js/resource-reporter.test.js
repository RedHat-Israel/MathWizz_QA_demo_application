const fs = require('fs');
const path = require('path');
const os = require('os');
const { mergeReports, generateConsolidatedHTMLReport } = require('./html-report');

describe('Resource Reporter', () => {
  let reportDir;

  beforeEach(() => {
    reportDir = fs.mkdtempSync(path.join(os.tmpdir(), 'testutils-reports-'));
  });

  afterEach(() => {
    fs.rmSync(reportDir, { recursive: true, force: true });
  });

  describe('mergeReports and generateConsolidatedHTMLReport', () => {
    it('merges multiple JSON reports into one consolidated HTML', () => {
      const suite1 = {
        suite_name: 'Web-Server Suite',
        total_tests: 2,
        passed: 2,
        failed: 0,
        skipped: 0,
        total_duration_ms: 100.0,
        peak_heap_alloc: 4096,
        slowest_test: 'solver test',
        slowest_time_ms: 80.0,
        tests: [
          { name: 'solver test', status: 'passed', duration_ms: 80.0, heap_alloc_delta: 1024 },
          { name: 'handler test', status: 'passed', duration_ms: 20.0, heap_alloc_delta: 512 },
        ],
      };

      const suite2 = {
        suite_name: 'Frontend Tests',
        total_tests: 1,
        passed: 1,
        failed: 0,
        skipped: 0,
        total_duration_ms: 50.0,
        peak_heap_alloc: 8192,
        slowest_test: 'render test',
        slowest_time_ms: 50.0,
        tests: [
          { name: 'render test', status: 'passed', duration_ms: 50.0, heap_alloc_delta: 2048 },
        ],
      };

      fs.writeFileSync(path.join(reportDir, 'web-server.json'), JSON.stringify(suite1, null, 2));
      fs.writeFileSync(path.join(reportDir, 'frontend.json'), JSON.stringify(suite2, null, 2));

      const consolidated = mergeReports(reportDir);
      expect(consolidated.total_tests).toBe(3);
      expect(consolidated.passed).toBe(3);
      expect(consolidated.suites).toHaveLength(2);
      expect(consolidated.peak_heap_alloc).toBe(8192);
      expect(consolidated.slowest_test).toBe('solver test');

      const htmlPath = path.join(reportDir, 'resource-report.html');
      generateConsolidatedHTMLReport(consolidated, htmlPath);

      const html = fs.readFileSync(htmlPath, 'utf-8');
      expect(html).toContain('Web-Server Suite');
      expect(html).toContain('Frontend Tests');
      expect(html).toContain('solver test');
      expect(html).toContain('render test');
      expect(html).toContain('<!DOCTYPE html>');
      expect(html).toContain('MathWizz Resource Report');
    });
  });
});
