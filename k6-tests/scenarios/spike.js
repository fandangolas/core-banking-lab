import { sleep } from 'k6';
import { setupAccounts, randomBankingOperation } from '../lib/banking.js';

/**
 * Spike Test
 * Purpose: Test system behavior under sudden extreme load
 * VUs: 10-1000 users (sudden spike)
 * Duration: 5 minutes
 * Use case: Validate system handles traffic spikes and recovers gracefully
 */

export const options = {
  stages: [
    { duration: '30s', target: 50 },    // Normal load
    { duration: '30s', target: 1000 },  // Spike to 1000 users!
    { duration: '3m', target: 1000 },   // Sustain spike
    { duration: '30s', target: 50 },    // Recovery
    { duration: '30s', target: 0 },     // Ramp down
  ],
  thresholds: {
    http_req_duration: ['p(95)<5000'],   // 95% < 5s (very relaxed)
    http_req_failed: ['rate<0.20'],       // Error rate < 20% (expect some failures)
  },
};

export function setup() {
  console.log('🚀 Starting SPIKE test with 1000 VUs...');
  console.log('⚠️  This will push the system to its absolute limits!');

  // Create 50 accounts with $500,000 initial balance
  const accountIds = setupAccounts(50, 500000);

  console.log(`✅ Created ${accountIds.length} accounts`);
  console.log('🔥 Prepare for extreme load...');
  return { accountIds };
}

export default function (data) {
  const { accountIds } = data;

  // Perform rapid banking operations
  randomBankingOperation(accountIds);

  // Very short think time for spike (0-200ms)
  sleep(Math.random() * 0.2);
}

export function teardown(data) {
  const { accountIds } = data;
  console.log(`✅ Spike test completed!`);
  console.log(`System tested with ${accountIds.length} accounts under extreme 1000 VU load`);
  console.log('Check metrics to see how the system handled the spike 📊');
}
