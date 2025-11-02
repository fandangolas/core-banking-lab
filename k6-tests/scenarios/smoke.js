import { sleep } from 'k6';
import { createAccount, deposit, withdraw, transfer, getBalance } from '../lib/banking.js';

/**
 * Smoke Test
 * Purpose: Verify the system works under minimal load
 * VUs: 1-5 users
 * Duration: 1 minute
 * Use case: Quick sanity check before running larger tests
 */

export const options = {
  stages: [
    { duration: '10s', target: 2 },  // Ramp up to 2 users
    { duration: '40s', target: 2 },  // Stay at 2 users
    { duration: '10s', target: 0 },  // Ramp down to 0 users
  ],
  thresholds: {
    http_req_duration: ['p(95)<500'],  // 95% of requests should be below 500ms
    http_req_failed: ['rate<0.01'],     // Error rate should be below 1%
  },
};

export function setup() {
  console.log('🔥 Starting smoke test...');

  // Create test accounts
  const account1 = createAccount('SmokeTestUser1');
  const account2 = createAccount('SmokeTestUser2');

  return {
    account1Id: account1.json('id'),
    account2Id: account2.json('id'),
  };
}

export default function (data) {
  const { account1Id, account2Id } = data;

  // Test deposit
  deposit(account1Id, 10000);
  sleep(1);

  // Test balance check
  getBalance(account1Id);
  sleep(1);

  // Test withdraw
  withdraw(account1Id, 2000);
  sleep(1);

  // Test transfer
  transfer(account1Id, account2Id, 1000);
  sleep(1);

  // Check balances
  getBalance(account1Id);
  getBalance(account2Id);
  sleep(1);
}

export function teardown(data) {
  console.log('✅ Smoke test completed');
}
