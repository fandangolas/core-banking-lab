const BASE_URL = import.meta.env.VITE_API_BASE_URL || 'http://localhost:8080';

async function request(url, options = {}) {
  const res = await fetch(`${BASE_URL}${url}`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    ...options,
  });

  if (!res.ok) {
    const error = await res.text();
    throw new Error(error || `Request failed with status ${res.status}`);
  }

  try {
    return await res.json();
  } catch {
    return null;
  }
}

export async function createAccount(owner) {
  return request('/accounts', {
    body: JSON.stringify({ owner }),
  });
}

export async function deposit(id, amount) {
  return request(`/accounts/${id}/deposit`, {
    body: JSON.stringify({ amount }),
  });
}

export async function withdraw(id, amount) {
  return request(`/accounts/${id}/withdraw`, {
    body: JSON.stringify({ amount }),
  });
}

export async function transfer(from, to, amount) {
  return request('/accounts/transfer', {
    body: JSON.stringify({ from, to, amount }),
  });
}

export function runSimulation(totalRequests, {
  requestFn,
  batchSize,
  intervalMs = 0,
  onMetric,
} = {}) {
  if (typeof requestFn !== 'function') {
    throw new Error('requestFn must be provided');
  }

  batchSize = batchSize || totalRequests;

  let completed = 0;
  let failed = 0;

  const emit = metric => {
    onMetric && onMetric(metric);
    if (metric.error) {
      console.error('simulation error', metric.error);
    } else {
      console.log('simulation metric', metric);
    }
  };

  const runBatch = async () => {
    const remaining = totalRequests - completed - failed;
    const count = Math.min(batchSize, remaining);

    await Promise.all(
      Array.from({ length: count }, () => {
        const start = performance.now();
        return requestFn()
          .then(() => {
            completed++;
            emit({ ok: true, duration: performance.now() - start, timestamp: Date.now() });
          })
          .catch(err => {
            failed++;
            emit({ ok: false, error: err.message, duration: performance.now() - start, timestamp: Date.now() });
          });
      })
    );
  };

  return (async function run() {
    while (completed + failed < totalRequests) {
      await runBatch();
      if (intervalMs > 0 && completed + failed < totalRequests) {
        await new Promise(resolve => setTimeout(resolve, intervalMs));
      }
    }
  })();
}

export { BASE_URL };
