import http from 'k6/http';
import { check, sleep } from 'k6';

const BASE_URL = __ENV.BASE_URL || 'http://localhost:8080';
const CONCURRENT_REQUESTS = Number(__ENV.CONCURRENT_REQUESTS || 10);
const TOTAL_REQUESTS = Number(__ENV.TOTAL_REQUESTS || 9000);
const REQUEST_DELAY_SECONDS = Number(__ENV.REQUEST_DELAY_SECONDS || 0.1);

const endpoints = [
    { method: 'GET', path: '/health' },
    { method: 'GET', path: '/api/v1/users' },
    { method: 'POST', path: '/api/v1/users' },
    { method: 'GET', path: '/api/v1/users/123' },
    { method: 'PUT', path: '/api/v1/users/123' },
    { method: 'POST', path: '/load' },
    { method: 'GET', path: '/connections' },
];

export const options = {
    scenarios: {
        fiber_load: {
            executor: 'shared-iterations',
            vus: CONCURRENT_REQUESTS,
            iterations: TOTAL_REQUESTS,
            maxDuration: '15m',
        },
    },
    thresholds: {
        http_req_failed: ['rate<0.05'],
        http_req_duration: ['p(95)<1000'],
    },
};

export function setup() {
    const health = http.get(`${BASE_URL}/health`);
    const ok = check(health, {
        'health endpoint disponível': (res) => res.status === 200,
    });

    if (!ok) {
        throw new Error(
            `Fiber app indisponível em ${BASE_URL}. Rode: kubectl port-forward service/fiber-prometheus-app 8080:8080`
        );
    }
}

export default function () {
    const endpoint = endpoints[Math.floor(Math.random() * endpoints.length)];
    const url = `${BASE_URL}${endpoint.path}`;

    let response;
    if (endpoint.method === 'POST' || endpoint.method === 'PUT') {
        const payload = JSON.stringify({
            name: 'Test User',
            email: 'test@example.com',
        });

        response = http.request(endpoint.method, url, payload, {
            headers: { 'Content-Type': 'application/json' },
            tags: { endpoint: `${endpoint.method} ${endpoint.path}` },
        });
    } else {
        response = http.request(endpoint.method, url, null, {
            tags: { endpoint: `${endpoint.method} ${endpoint.path}` },
        });
    }

    check(response, {
        'status válido': (res) => res.status >= 200 && res.status < 500,
    });

    sleep(REQUEST_DELAY_SECONDS);
}