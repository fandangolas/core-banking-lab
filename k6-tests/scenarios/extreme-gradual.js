import { sleep } from 'k6';
import { setupAccounts, randomBankingOperation } from '../lib/banking.js';

/**
 * EXTREME Load Test (Gradual Ramp-Up)
 * Purpose: Test system at 5000 VUs without hitting somaxconn limit
 * Strategy: Slower ramp-up + connection reuse
 * VUs: 50-5000 users (gradual increase)
 * Duration: 8 minutes (longer to accommodate gradual ramp)
 */

export const options = {
  stages: [
    { duration: '1m', target: 100 },    // Warm up
    { duration: '1m', target: 500 },    // Gradual increase
    { duration: '1m', target: 1500 },   // Keep ramping
    { duration: '1m', target: 3000 },   // Getting there
    { duration: '1m', target: 5000 },   // EXTREME: 5000 users!
    { duration: '2m', target: 5000 },   // Sustain extreme load
    { duration: '30s', target: 100 },   // Recovery
    { duration: '30s', target: 0 },     // Ramp down
  ],
  thresholds: {
    http_req_duration: ['p(95)<10000'],
    http_req_failed: ['rate<0.30'],
  },
  // Enable connection reuse to avoid somaxconn limit
  noConnectionReuse: false,
  // Batch configuration to reduce connection churn
  batch: 10,
  batchPerHost: 5,
};

export function setup() {
  console.log('💥 Starting EXTREME test with gradual ramp-up...');
  console.log('🔧 Connection reuse ENABLED to work within somaxconn limits');
  console.log('🔥 This approach is more realistic for production workloads');

  // Create 100 accounts with $1,000,000 initial balance
  const accountIds = setupAccounts(100, 1000000);

  console.log(`✅ Created ${accountIds.length} accounts`);
  console.log('🚀 Launching gradual extreme load attack...');
  return { accountIds };
}

export default function (data) {
  const { accountIds } = data;

  // Perform rapid banking operations
  randomBankingOperation(accountIds);

  // Minimal think time for extreme load (0-100ms)
  sleep(Math.random() * 0.1);
}

export function teardown(data) {
  const { accountIds } = data;
  console.log(`✅ EXTREME gradual test completed!`);
  console.log(`System tested with ${accountIds.length} accounts under MASSIVE 5000 VU load`);
  console.log('📊 Gradual ramp-up should result in much fewer connection errors');
  console.log('🎯 This is a more realistic production scenario!');
}
