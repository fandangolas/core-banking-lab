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

// Função original mantida para compatibilidade
export function runSimulation(totalRequests, {
  requestFn,
  batchSize = 1,
  intervalMs = 1000,
  onMetric,
} = {}) {
  if (typeof requestFn !== 'function') {
    throw new Error('requestFn must be provided');
  }

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

  const sequential = () => {
    if (completed + failed >= totalRequests) return;
    const start = performance.now();
    requestFn()
      .then(() => {
        completed++;
        emit({ ok: true, duration: performance.now() - start, timestamp: Date.now() });
        if (completed + failed < totalRequests) {
          setTimeout(sequential, intervalMs);
        }
      })
      .catch(err => {
        failed++;
        emit({ ok: false, error: err.message, duration: performance.now() - start, timestamp: Date.now() });
        if (completed + failed < totalRequests) {
          setTimeout(sequential, intervalMs);
        }
      });
  };

  const batched = () => {
    const timer = setInterval(() => {
      const remaining = totalRequests - completed - failed;
      const count = Math.min(batchSize, remaining);
      for (let i = 0; i < count; i++) {
        const start = performance.now();
        requestFn()
          .then(() => {
            completed++;
            emit({ ok: true, duration: performance.now() - start, timestamp: Date.now() });
            if (completed + failed >= totalRequests) {
              clearInterval(timer);
            }
          })
          .catch(err => {
            failed++;
            emit({ ok: false, error: err.message, duration: performance.now() - start, timestamp: Date.now() });
            if (completed + failed >= totalRequests) {
              clearInterval(timer);
            }
          });
      }
      if (completed + failed >= totalRequests) {
        clearInterval(timer);
      }
    }, intervalMs);
  };

  if (batchSize === 1) {
    sequential();
  } else {
    batched();
  }
}

// Nova função que replica exatamente a lógica do simulador Go
export async function runBatchedSimulation(totalRequests, {
  requestFn,
  blockSize = 500,
  blockDuration = 2000, // 2 segundos
  abortSignal,
  onProgress,
  onMetric,
} = {}) {
  if (typeof requestFn !== 'function') {
    throw new Error('requestFn must be provided');
  }

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

  // Função para calcular quantas operações executar neste bloco
  const opsThisBlock = (sent, total, block) => {
    if (total - sent < block) {
      return total - sent;
    }
    return block;
  };

  console.log(`Iniciando simulação: ${totalRequests} requests, blocos de ${blockSize}, duração ${blockDuration}ms`);

  // Loop principal que replica a lógica do Go
  for (let sent = 0; sent < totalRequests;) {
    // Verifica se foi cancelado
    if (abortSignal && abortSignal.aborted) {
      throw new Error('Abortado pelo usuário');
    }

    const n = opsThisBlock(sent, totalRequests, blockSize);
    console.log(`Executando bloco: ${n} requests (${sent + 1}-${sent + n} de ${totalRequests})`);
    
    // Calcula o intervalo entre requests neste bloco
    const interval = blockDuration / n;
    
    // Executa todas as requests deste bloco
    const blockPromises = [];
    
    for (let i = 0; i < n; i++) {
      // Agenda cada request com o intervalo apropriado
      const requestPromise = new Promise((resolve) => {
        setTimeout(async () => {
          // Verifica novamente se foi cancelado
          if (abortSignal && abortSignal.aborted) {
            resolve();
            return;
          }

          const start = performance.now();
          try {
            await requestFn();
            completed++;
            emit({ 
              ok: true, 
              duration: performance.now() - start, 
              timestamp: Date.now() 
            });
            onProgress && onProgress(completed + failed, totalRequests);
            resolve();
          } catch (err) {
            if (err.message === 'Abortado pelo usuário') {
              resolve();
              return;
            }
            failed++;
            emit({ 
              ok: false, 
              error: err.message, 
              duration: performance.now() - start, 
              timestamp: Date.now() 
            });
            onProgress && onProgress(completed + failed, totalRequests);
            resolve(); // Resolve mesmo com erro para não travar o bloco
          }
        }, i * interval);
      });
      
      blockPromises.push(requestPromise);
    }
    
    // Aguarda todas as requests do bloco terminarem
    await Promise.all(blockPromises);
    
    sent += n;
    console.log(`Bloco concluído. Total processado: ${completed + failed}/${totalRequests}`);

    // Verifica se foi cancelado após o bloco
    if (abortSignal && abortSignal.aborted) {
      throw new Error('Abortado pelo usuário');
    }
  }

  console.log(`Simulação concluída. Sucessos: ${completed}, Falhas: ${failed}`);
  return { completed, failed };
}

export { BASE_URL };