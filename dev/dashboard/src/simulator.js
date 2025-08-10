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

export { BASE_URL };
