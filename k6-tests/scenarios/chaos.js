import { sleep } from 'k6';
import { setupAccounts, randomBankingOperation } from '../lib/banking.js';

/**
 * CHAOS Test - The Ultimate Stress Test
 * Purpose: Find the absolute destruction point of the system
 * VUs: 100-10000 users (chaos level)
 * Duration: 7 minutes
 * Use case: Discover catastrophic failure modes and recovery behavior
 */

export const options = {
  stages: [
    { duration: '1m', target: 200 },     // Warm up to 200 users
    { duration: '1m', target: 2500 },    // Ramp to 2500 users
    { duration: '1m', target: 5000 },    // Push to 5000 users
    { duration: '1m', target: 10000 },   // CHAOS: 10,000 users!
    { duration: '2m', target: 10000 },   // Sustain chaos
    { duration: '30s', target: 200 },    // Recovery
    { duration: '30s', target: 0 },      // Ramp down
  ],
  thresholds: {
    http_req_duration: ['p(95)<15000'],  // 95% < 15s (extremely relaxed)
    http_req_failed: ['rate<0.50'],       // Error rate < 50% (expect significant failures)
  },
};

export function setup() {
  console.log('💀 Starting CHAOS test with 10,000 VUs...');
  console.log('⚠️  WARNING: This WILL break things!');
  console.log('🔥 Monitoring system stability under catastrophic load...');

  // Create 150 accounts with $2,000,000 initial balance
  const accountIds = setupAccounts(150, 2000000);

  console.log(`✅ Created ${accountIds.length} accounts`);
  console.log('💣 Initiating chaos attack...');
  return { accountIds };
}

export default function (data) {
  const { accountIds } = data;

  // Perform banking operations
  randomBankingOperation(accountIds);

  // Minimal think time for chaos (0-50ms)
  sleep(Math.random() * 0.05);
}

export function teardown(data) {
  const { accountIds } = data;
  console.log(`💀 CHAOS test completed!`);
  console.log(`System tested with ${accountIds.length} accounts under CATASTROPHIC 10,000 VU load`);
  console.log('📊 Analyzing damage and recovery patterns...');
  console.log('🔥 This represents approximately 100,000+ requests per minute at peak!');
}
