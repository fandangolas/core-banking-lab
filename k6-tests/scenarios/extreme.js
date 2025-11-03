import { sleep } from 'k6';
import { setupAccounts, randomBankingOperation } from '../lib/banking.js';

/**
 * EXTREME Load Test
 * Purpose: Find the absolute breaking point of the system
 * VUs: 50-5000 users (extreme spike)
 * Duration: 6 minutes
 * Use case: Discover true system limits under massive concurrent load
 */

export const options = {
  stages: [
    { duration: '1m', target: 100 },    // Warm up to 100 users
    { duration: '1m', target: 2500 },   // Rapid ramp to 2500 users
    { duration: '1m', target: 5000 },   // Push to EXTREME 5000 users!
    { duration: '2m', target: 5000 },   // Sustain extreme load
    { duration: '30s', target: 100 },   // Recovery
    { duration: '30s', target: 0 },     // Ramp down
  ],
  thresholds: {
    http_req_duration: ['p(95)<10000'],  // 95% < 10s (very relaxed)
    http_req_failed: ['rate<0.30'],       // Error rate < 30% (expect failures)
  },
};

export function setup() {
  console.log('💥 Starting EXTREME test with 5000 VUs...');
  console.log('⚠️  WARNING: This will push the system beyond normal limits!');
  console.log('🔥 Expect high resource usage and potential failures...');

  // Create 100 accounts with $1,000,000 initial balance
  const accountIds = setupAccounts(100, 1000000);

  console.log(`✅ Created ${accountIds.length} accounts`);
  console.log('🚀 Launching extreme load attack...');
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
  console.log(`✅ EXTREME test completed!`);
  console.log(`System tested with ${accountIds.length} accounts under MASSIVE 5000 VU load`);
  console.log('📊 Check metrics to see how the system handled the extreme stress');
  console.log('🔥 This represents approximately 50,000+ requests per minute at peak!');
}
