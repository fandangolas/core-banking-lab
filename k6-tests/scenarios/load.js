import { sleep } from 'k6';
import { setupAccounts, randomBankingOperation } from '../lib/banking.js';

/**
 * Load Test
 * Purpose: Test system performance under expected load
 * VUs: 10-50 users
 * Duration: 5 minutes
 * Use case: Validate system can handle normal production load
 */

export const options = {
  stages: [
    { duration: '1m', target: 20 },   // Ramp up to 20 users
    { duration: '3m', target: 50 },   // Increase to 50 users
    { duration: '1m', target: 0 },    // Ramp down to 0
  ],
  thresholds: {
    http_req_duration: ['p(95)<1000', 'p(99)<2000'], // 95% < 1s, 99% < 2s
    http_req_failed: ['rate<0.05'],                   // Error rate < 5%
    'http_req_duration{operation:deposit}': ['p(95)<800'],
    'http_req_duration{operation:withdraw}': ['p(95)<800'],
    'http_req_duration{operation:transfer}': ['p(95)<1000'],
    'http_req_duration{operation:get_balance}': ['p(95)<500'],
  },
};

export function setup() {
  console.log('📊 Starting load test...');
  console.log('Setting up 10 test accounts with initial balance...');

  // Create 10 accounts with $50,000 initial balance
  const accountIds = setupAccounts(10, 50000);

  console.log(`✅ Created ${accountIds.length} accounts`);
  return { accountIds };
}

export default function (data) {
  const { accountIds } = data;

  // Perform random banking operations
  randomBankingOperation(accountIds);

  // Think time between operations (0.5-2 seconds)
  sleep(Math.random() * 1.5 + 0.5);
}

export function teardown(data) {
  const { accountIds } = data;
  console.log(`✅ Load test completed with ${accountIds.length} test accounts`);
}
