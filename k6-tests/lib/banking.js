import http from 'k6/http';
import { check } from 'k6';

const BASE_URL = __ENV.BASE_URL || 'http://localhost:8080';

/**
 * Create a new bank account
 * @param {string} owner - Account owner name
 * @returns {object} Response with account ID
 */
export function createAccount(owner) {
  const payload = JSON.stringify({ owner });
  const params = {
    headers: { 'Content-Type': 'application/json' },
    tags: { operation: 'create_account' },
  };

  const res = http.post(`${BASE_URL}/accounts`, payload, params);

  check(res, {
    'account created': (r) => r.status === 201,
    'has account id': (r) => r.json('id') !== undefined,
  });

  return res;
}

/**
 * Get account balance
 * @param {number} accountId - Account ID
 * @returns {object} Response with balance
 */
export function getBalance(accountId) {
  const params = {
    tags: { operation: 'get_balance' },
  };

  const res = http.get(`${BASE_URL}/accounts/${accountId}/balance`, params);

  check(res, {
    'balance retrieved': (r) => r.status === 200,
    'has balance': (r) => r.json('balance') !== undefined,
  });

  return res;
}

/**
 * Deposit money into account
 * @param {number} accountId - Account ID
 * @param {number} amount - Amount to deposit
 * @returns {object} Response
 */
export function deposit(accountId, amount) {
  const payload = JSON.stringify({ amount });
  const params = {
    headers: { 'Content-Type': 'application/json' },
    tags: { operation: 'deposit' },
  };

  const res = http.post(`${BASE_URL}/accounts/${accountId}/deposit`, payload, params);

  check(res, {
    'deposit successful': (r) => r.status === 202,  // Async processing returns 202 Accepted
    'operation accepted': (r) => r.json('operation_id') !== undefined,
  });

  return res;
}

/**
 * Withdraw money from account
 * @param {number} accountId - Account ID
 * @param {number} amount - Amount to withdraw
 * @returns {object} Response
 */
export function withdraw(accountId, amount) {
  const payload = JSON.stringify({ amount });
  const params = {
    headers: { 'Content-Type': 'application/json' },
    tags: { operation: 'withdraw' },
  };

  const res = http.post(`${BASE_URL}/accounts/${accountId}/withdraw`, payload, params);

  check(res, {
    'withdraw successful': (r) => r.status === 200 || r.status === 400,
  });

  return res;
}

/**
 * Transfer money between accounts
 * @param {number} fromId - Source account ID
 * @param {number} toId - Destination account ID
 * @param {number} amount - Amount to transfer
 * @returns {object} Response
 */
export function transfer(fromId, toId, amount) {
  const payload = JSON.stringify({ from: fromId, to: toId, amount });
  const params = {
    headers: { 'Content-Type': 'application/json' },
    tags: { operation: 'transfer' },
  };

  const res = http.post(`${BASE_URL}/accounts/transfer`, payload, params);

  check(res, {
    'transfer attempted': (r) => r.status === 200 || r.status === 400,
  });

  return res;
}

/**
 * Setup test accounts with initial balances
 * @param {number} count - Number of accounts to create
 * @param {number} initialBalance - Initial balance for each account
 * @returns {array} Array of account IDs
 */
export function setupAccounts(count, initialBalance) {
  const accountIds = [];

  for (let i = 0; i < count; i++) {
    const res = createAccount(`TestUser${i}`);
    if (res.status === 201) {
      const accountId = res.json('id');
      accountIds.push(accountId);

      // Add initial balance
      if (initialBalance > 0) {
        deposit(accountId, initialBalance);
      }
    }
  }

  return accountIds;
}

/**
 * Random banking operation for mixed workload
 * @param {array} accountIds - Array of available account IDs
 */
export function randomBankingOperation(accountIds) {
  const operation = Math.random();
  const randomAccountId = accountIds[Math.floor(Math.random() * accountIds.length)];

  if (operation < 0.25) {
    // 25% - Get balance
    getBalance(randomAccountId);
  } else if (operation < 0.50) {
    // 25% - Deposit
    const amount = Math.floor(Math.random() * 10000) + 100;
    deposit(randomAccountId, amount);
  } else if (operation < 0.75) {
    // 25% - Withdraw
    const amount = Math.floor(Math.random() * 1000) + 50;
    withdraw(randomAccountId, amount);
  } else {
    // 25% - Transfer
    const fromId = randomAccountId;
    const toId = accountIds[Math.floor(Math.random() * accountIds.length)];
    if (fromId !== toId) {
      const amount = Math.floor(Math.random() * 500) + 10;
      transfer(fromId, toId, amount);
    }
  }
}
