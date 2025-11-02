import { sleep } from 'k6';
import { setupAccounts, randomBankingOperation } from '../lib/banking.js';

/**
 * Stress Test
 * Purpose: Find the breaking point of the system
 * VUs: 10-200 users
 * Duration: 10 minutes
 * Use case: Determine system limits and recovery behavior
 */

export const options = {
  stages: [
    { duration: '2m', target: 50 },    // Ramp up to 50 users
    { duration: '3m', target: 100 },   // Increase to 100 users
    { duration: '2m', target: 200 },   // Push to 200 users (stress)
    { duration: '2m', target: 100 },   // Scale back to 100
    { duration: '1m', target: 0 },     // Ramp down
  ],
  thresholds: {
    http_req_duration: ['p(95)<3000'],  // 95% < 3s (relaxed threshold)
    http_req_failed: ['rate<0.10'],      // Error rate < 10%
  },
};

export function setup() {
  console.log('💪 Starting stress test...');
  console.log('This will push the system to its limits...');

  // Create 20 accounts with $100,000 initial balance
  const accountIds = setupAccounts(20, 100000);

  console.log(`✅ Created ${accountIds.length} accounts`);
  return { accountIds };
}

export default function (data) {
  const { accountIds } = data;

  // Perform rapid banking operations
  randomBankingOperation(accountIds);

  // Reduced think time for stress (0.1-0.5 seconds)
  sleep(Math.random() * 0.4 + 0.1);
}

export function teardown(data) {
  const { accountIds } = data;
  console.log(`✅ Stress test completed`);
  console.log(`System survived ${accountIds.length} accounts under heavy load`);
}
