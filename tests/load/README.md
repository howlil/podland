# Load Testing Guide - Podland

This directory contains k6 load testing scripts for the Podland platform.

## Prerequisites

### Install k6

**macOS:**
```bash
brew install k6
```

**Windows (with Scoop):**
```bash
scoop install k6
```

**Linux:**
```bash
# Debian/Ubuntu
sudo gpg -k
sudo gpg --no-default-keyring --keyring /usr/share/keyrings/k6-archive-keyring.gpg --keyserver hkp://keyserver.ubuntu.com:80 --recv-keys C5AD17C747E3415A3642D57D77C6C491D6AC1D69
echo "deb [signed-by=/usr/share/keyrings/k6-archive-keyring.gpg] https://dl.k6.io/deb stable main" | sudo tee /etc/apt/sources.list.d/k6.list
sudo apt-get update
sudo apt-get install k6

# Fedora/RHEL
sudo dnf install https://dl.k6.io/rpm/repo.rpm
sudo dnf install k6
```

## Running Load Tests

### Basic Test Run

```bash
k6 run tests/load/critical-paths.js
```

### Run with Custom Base URL

```bash
BASE_URL=http://podland-app.com k6 run tests/load/critical-paths.js
```

### Run with JSON Output

```bash
k6 run --out json=results.json tests/load/critical-paths.js
```

### Run with Summary Export

```bash
k6 run --summary-export=summary.json tests/load/critical-paths.js
```

## Test Configuration

The default test configuration (`critical-paths.js`):

- **Virtual Users (VUs):** 100 concurrent users
- **Duration:** 5 minutes
- **Thresholds:**
  - p95 response time < 500ms
  - 0% HTTP errors
  - <1% custom errors

## Critical Paths Tested

1. **Authentication Flow**
   - Login and token acquisition

2. **VM Lifecycle**
   - VM creation
   - VM start
   - VM metrics fetch
   - VM stop
   - VM deletion

## Success Criteria

- [ ] p95 response time < 500ms (all endpoints)
- [ ] 0% error rate (no 5xx errors)
- [ ] 100 concurrent users sustained for 5 minutes
- [ ] All resources healthy (CPU < 80%, RAM < 80%)

## Analyzing Results

### Console Output

k6 provides real-time metrics in the console:

```
     ✓ login status is 200
     ✓ login has token
     ✓ create VM status is 201
     
     checks.........................: 100.00% ✓ 6000      ✗ 0
     data_received..................: 1.2 MB  2045 B/s
     data_sent......................: 800 KB  1365 B/s
     http_req_duration..............: avg=120ms min=50ms med=115ms p(90)=180ms p(95)=220ms
     http_reqs......................: 6000    10.17 req/s
     iteration_duration.............: avg=4.9s min=4s med=4.8s p(90)=5.5s p(95)=5.8s
     iterations.....................: 1000    1.70/s
     vus............................: 100     min=100     max=100
```

### Visualizing with k6 Cloud

For detailed analysis, you can upload results to k6 Cloud:

```bash
k6 run --out cloud tests/load/critical-paths.js
```

## Troubleshooting

### High Response Times

If p95 > 500ms:
1. Check database connection pool settings
2. Review Prometheus metrics for resource bottlenecks
3. Consider scaling the backend

### High Error Rates

If error rate > 0%:
1. Check application logs for errors
2. Verify database connections are not exhausted
3. Review rate limiting configuration

### Connection Refused Errors

If seeing connection errors:
1. Ensure the backend is running
2. Check firewall rules
3. Verify BASE_URL is correct

## Best Practices

1. **Run tests in staging environment** before production
2. **Monitor resources** during test execution
3. **Start with lower VU counts** and gradually increase
4. **Document baseline metrics** for comparison
5. **Run tests regularly** (CI/CD integration recommended)

## CI/CD Integration

Example GitHub Actions workflow:

```yaml
name: Load Tests
on: [push]
jobs:
  k6:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3
      - uses: grafana/k6-action@v0.2.0
        with:
          filename: tests/load/critical-paths.js
          token: ${{ secrets.GITHUB_TOKEN }}
```
