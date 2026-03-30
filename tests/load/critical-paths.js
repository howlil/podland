// tests/load/critical-paths.js
// k6 load testing script for Podland platform
// Run with: k6 run tests/load/critical-paths.js

import http from 'k6/http';
import { check, sleep } from 'k6';
import { Rate } from 'k6/metrics';

// Custom metric for error rate
const errorRate = new Rate('errors');

export const options = {
  vus: 100,           // 100 concurrent users
  duration: '5m',     // 5 minutes
  thresholds: {
    http_req_duration: ['p(95)<500'], // p95 < 500ms
    http_req_failed: ['rate==0'],     // 0% errors
    errors: ['rate<0.01'],            // <1% custom errors
  },
};

const BASE_URL = __ENV.BASE_URL || 'http://localhost:8080';

export default function () {
  // 1. Auth flow (login)
  const loginRes = http.post(`${BASE_URL}/api/auth/login`);

  const loginSuccess = check(loginRes, {
    'login status is 200': (r) => r.status === 200,
    'login has token': (r) => r.json('access_token') !== '',
  });

  errorRate.add(!loginSuccess);

  if (!loginSuccess) {
    console.error('Login failed');
    return;
  }

  const token = loginRes.json('access_token');
  const headers = { 'Authorization': `Bearer ${token}` };

  sleep(1);

  // 2. VM create
  const createRes = http.post(
    `${BASE_URL}/api/vms`,
    JSON.stringify({
      name: `load-test-vm-${__VU}`,
      os: 'ubuntu-2204',
      tier: 'micro'
    }),
    { headers }
  );

  check(createRes, {
    'create VM status is 201': (r) => r.status === 201,
    'create VM has id': (r) => r.json('id') !== '',
  });

  const vmID = createRes.json('id');
  sleep(1);

  // 3. VM start
  const startRes = http.post(
    `${BASE_URL}/api/vms/${vmID}/start`,
    null,
    { headers }
  );

  check(startRes, {
    'start VM status is 200': (r) => r.status === 200,
  });

  sleep(1);

  // 4. Metrics fetch
  const metricsRes = http.get(
    `${BASE_URL}/api/vms/${vmID}/metrics?range=24h`,
    { headers }
  );

  check(metricsRes, {
    'metrics status is 200': (r) => r.status === 200,
  });

  sleep(1);

  // 5. VM stop
  const stopRes = http.post(
    `${BASE_URL}/api/vms/${vmID}/stop`,
    null,
    { headers }
  );

  check(stopRes, {
    'stop VM status is 200': (r) => r.status === 200,
  });

  sleep(1);

  // 6. VM delete
  const deleteRes = http.del(
    `${BASE_URL}/api/vms/${vmID}`,
    null,
    { headers }
  );

  check(deleteRes, {
    'delete VM status is 200': (r) => r.status === 200,
  });

  sleep(1);
}
