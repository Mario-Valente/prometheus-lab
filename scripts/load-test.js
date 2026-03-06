import http from 'k6/http';
import { check, sleep } from 'k6';

export let options = {
  stages: [
    { duration: '30s', target: 10 },   // Ramp-up: 0 → 10 VUs em 30s
    { duration: '1m30s', target: 50 }, // Stress: 10 → 50 VUs em 90s
    { duration: '20s', target: 0 },    // Ramp-down: 50 → 0 VUs em 20s
  ],
  thresholds: {
    http_req_duration: ['p(95)<500', 'p(99)<1000'],
    http_req_failed: ['rate<0.1'],
  },
};

const BASE_URL = 'http://prometheus-app.default.svc.cluster.local:8080';

export default function () {
  let requestId = Math.floor(Math.random() * 10000);

  let response = http.post(
    `${BASE_URL}/request/${requestId}`,
    null,
    { timeout: '30s' }
  );

  check(response, {
    'status é 200': (r) => r.status === 200,
    'resposta < 1s': (r) => r.timings.duration < 1000,
  });

  sleep(1);
}
