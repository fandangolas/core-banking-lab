import React, { useEffect, useRef, useState } from 'react';
import {
  LineChart,
  Line,
  XAxis,
  YAxis,
  Tooltip,
  CartesianGrid,
  Legend
} from 'recharts';
import { runBatchedSimulation, createAccount, deposit, withdraw, transfer } from './simulator';

const COLORS = [
  '#8884d8',
  '#82ca9d',
  '#ff7300',
  '#ff0000',
  '#00C49F',
  '#FFBB28',
  '#FF8042'
];

export default function App() {
  const [data, setData] = useState([]);
  const [colorMap, setColorMap] = useState({});
  const lastCount = useRef(0);
  const [windowMinutes, setWindowMinutes] = useState(5);
  const [refreshMs, setRefreshMs] = useState(1000);
  const [startTime, setStartTime] = useState('');
  const [endTime, setEndTime] = useState('');
  const [reqCount, setReqCount] = useState(1000);
  const [blockSize, setBlockSize] = useState(100);
  const [blockDuration, setBlockDuration] = useState(2000);
  const [ops, setOps] = useState({
    create: false,
    deposit: false,
    withdraw: false,
    transfer: false
  });
  const [randomOp, setRandomOp] = useState(false);
  const [accountIds, setAccountIds] = useState([]);
  const [endpointStats, setEndpointStats] = useState({});
  const [isRunning, setIsRunning] = useState(false);
  const [progress, setProgress] = useState({ current: 0, total: 0 });

  useEffect(() => {
    const fetchMetrics = async () => {
      const res = await fetch('/metrics');
      const metrics = await res.json();
      const newMetrics = metrics.slice(lastCount.current);
      lastCount.current = metrics.length;

      if (newMetrics.length === 0) return;

      const now = Date.now();
      const retentionCutoff = now - 24 * 60 * 60 * 1000; // keep last 24h
      const newEndpoints = new Set();

      setData(prev => {
        const filtered = prev.filter(d => d.time >= retentionCutoff);
        let entry = filtered[filtered.length - 1];
        if (!entry || now - entry.time >= 1000) {
          entry = { time: now };
          filtered.push(entry);
        }

        newMetrics.forEach(m => {
          const ep = m.endpoint || m.Endpoint;
          entry[ep] = (entry[ep] || 0) + 1;
          newEndpoints.add(ep);
        });

        // assign colors for new endpoints
        if (newEndpoints.size > 0) {
          setColorMap(prevColors => {
            const updated = { ...prevColors };
            newEndpoints.forEach(ep => {
              if (!updated[ep]) {
                const color = COLORS[Object.keys(updated).length % COLORS.length];
                updated[ep] = color;
              }
            });
            return updated;
          });
        }

        return filtered;
      });
    };

    fetchMetrics();
    const id = setInterval(fetchMetrics, refreshMs);
    return () => clearInterval(id);
  }, [refreshMs]);

  const now = Date.now();
  let displayStart;
  let displayEnd;
  let displayData = data;
  if (startTime && endTime) {
    const today = new Date();
    const [sh, sm] = startTime.split(':').map(Number);
    const [eh, em] = endTime.split(':').map(Number);
    displayStart = new Date(
      today.getFullYear(),
      today.getMonth(),
      today.getDate(),
      sh,
      sm
    ).getTime();
    displayEnd = new Date(
      today.getFullYear(),
      today.getMonth(),
      today.getDate(),
      eh,
      em
    ).getTime();
    displayData = data.filter(d => d.time >= displayStart && d.time <= displayEnd);
  } else {
    displayEnd = now;
    displayStart = now - windowMinutes * 60 * 1000;
    displayData = data.filter(d => d.time >= displayStart);
  }

  const lines = Object.entries(colorMap).map(([ep, color]) => (
    <Line
      key={ep}
      type="monotone"
      dataKey={ep}
      stroke={color}
      dot={false}
      isAnimationActive={false}
    />
  ));

  const handleRun = async () => {
    const selected = Object.entries(ops)
      .filter(([, v]) => v)
      .map(([k]) => k);

    if (!randomOp && selected.length === 0) {
      alert('Selecione pelo menos uma operação ou marque "Aleatório"');
      return;
    }

    if (isRunning) {
      alert('Simulação já em execução!');
      return;
    }

    setIsRunning(true);
    setProgress({ current: 0, total: reqCount });

    let ids = accountIds.slice();

    // Garante que temos pelo menos algumas contas criadas para operações que não sejam create
    const needsAccounts = randomOp || selected.some(op => op !== 'create');
    if (needsAccounts && ids.length < 10) {
      console.log('Criando contas iniciais...');
      for (let i = ids.length; i < 10; i++) {
        try {
          const { id } = await createAccount(`User${Date.now()}_${i}`);
          ids.push(id);
          await deposit(id, Math.floor(Math.random() * 10000) + 5000); // Saldo inicial entre R$50 e R$150
        } catch (err) {
          console.error('Erro ao criar conta inicial:', err);
        }
      }
      setAccountIds(ids);
    }

    const requestFn = () => {
      let selectedOps = selected;
      if (randomOp) {
        selectedOps = ['create', 'deposit', 'withdraw', 'transfer'];
      }
      
      const op = selectedOps[Math.floor(Math.random() * selectedOps.length)];
      let promise;
      
      switch (op) {
        case 'create': {
          const username = `User${Date.now()}_${Math.floor(Math.random() * 10000)}`;
          promise = createAccount(username).then(({ id }) => {
            ids.push(id);
            setAccountIds([...ids]);
            // Adiciona saldo inicial
            const initialAmount = Math.floor(Math.random() * 5000) + 1000;
            return deposit(id, initialAmount);
          });
          break;
        }
        case 'deposit': {
          if (ids.length === 0) return Promise.resolve();
          const id = ids[Math.floor(Math.random() * ids.length)];
          const amount = Math.floor(Math.random() * 1000) + 100;
          promise = deposit(id, amount);
          break;
        }
        case 'withdraw': {
          if (ids.length === 0) return Promise.resolve();
          const id = ids[Math.floor(Math.random() * ids.length)];
          const amount = Math.floor(Math.random() * 500) + 50;
          promise = withdraw(id, amount);
          break;
        }
        case 'transfer': {
          if (ids.length < 2) return Promise.resolve();
          let from = ids[Math.floor(Math.random() * ids.length)];
          let to = ids[Math.floor(Math.random() * ids.length)];
          while (to === from && ids.length > 1) {
            to = ids[Math.floor(Math.random() * ids.length)];
          }
          const amount = Math.floor(Math.random() * 300) + 10;
          promise = transfer(from, to, amount);
          break;
        }
        default:
          promise = Promise.resolve();
      }
      
      return promise
        .then(res => {
          setEndpointStats(prev => ({
            ...prev,
            [op]: {
              success: (prev[op]?.success || 0) + 1,
              error: prev[op]?.error || 0,
            },
          }));
          setProgress(prev => ({ ...prev, current: prev.current + 1 }));
          return res;
        })
        .catch(err => {
          setEndpointStats(prev => ({
            ...prev,
            [op]: {
              success: prev[op]?.success || 0,
              error: (prev[op]?.error || 0) + 1,
            },
          }));
          setProgress(prev => ({ ...prev, current: prev.current + 1 }));
          console.error(`Erro na operação ${op}:`, err);
          throw err;
        });
    };

    try {
      await runBatchedSimulation(reqCount, {
        requestFn,
        blockSize,
        blockDuration,
        onProgress: (current, total) => {
          setProgress({ current, total });
        }
      });
    } catch (error) {
      console.error('Erro na simulação:', error);
    } finally {
      setIsRunning(false);
    }
  };

  return (
    <div style={{ width: '100%', height: 400 }}>
      <div style={{ marginBottom: '1rem' }}>
        <div style={{ marginBottom: '1rem' }}>
          <label style={{ marginRight: '1rem' }}>
            Nº de requests:
            <input
              type="number"
              min="1"
              max="100000"
              value={reqCount}
              onChange={e => setReqCount(Number(e.target.value))}
              disabled={isRunning}
            />
          </label>
          <label style={{ marginRight: '1rem' }}>
            Tamanho do bloco:
            <input
              type="number"
              min="1"
              max="1000"
              value={blockSize}
              onChange={e => setBlockSize(Number(e.target.value))}
              disabled={isRunning}
            />
          </label>
          <label style={{ marginRight: '1rem' }}>
            Duração do bloco (ms):
            <input
              type="number"
              min="100"
              max="10000"
              value={blockDuration}
              onChange={e => setBlockDuration(Number(e.target.value))}
              disabled={isRunning}
            />
          </label>
        </div>
        <div style={{ marginBottom: '1rem' }}>
          <label style={{ marginRight: '0.5rem' }}>
            <input
              type="checkbox"
              checked={ops.create}
              onChange={e => setOps({ ...ops, create: e.target.checked })}
              disabled={isRunning}
            />
            Criar conta
          </label>
          <label style={{ marginRight: '0.5rem' }}>
            <input
              type="checkbox"
              checked={ops.deposit}
              onChange={e => setOps({ ...ops, deposit: e.target.checked })}
              disabled={isRunning}
            />
            Depósito
          </label>
          <label style={{ marginRight: '0.5rem' }}>
            <input
              type="checkbox"
              checked={ops.withdraw}
              onChange={e => setOps({ ...ops, withdraw: e.target.checked })}
              disabled={isRunning}
            />
            Saque
          </label>
          <label style={{ marginRight: '0.5rem' }}>
            <input
              type="checkbox"
              checked={ops.transfer}
              onChange={e => setOps({ ...ops, transfer: e.target.checked })}
              disabled={isRunning}
            />
            Transferência
          </label>
          <label style={{ marginRight: '0.5rem' }}>
            <input
              type="checkbox"
              checked={randomOp}
              onChange={e => setRandomOp(e.target.checked)}
              disabled={isRunning}
            />
            Aleatório (ignora seleções acima)
          </label>
          <button 
            style={{ 
              marginLeft: '1rem',
              backgroundColor: isRunning ? '#ccc' : '#007bff',
              color: 'white',
              border: 'none',
              padding: '8px 16px',
              borderRadius: '4px',
              cursor: isRunning ? 'not-allowed' : 'pointer'
            }} 
            onClick={handleRun}
            disabled={isRunning}
          >
            {isRunning ? 'Executando...' : 'Disparar'}
          </button>
        </div>
        {isRunning && (
          <div style={{ marginBottom: '1rem' }}>
            <div>Progresso: {progress.current} / {progress.total}</div>
            <div style={{ 
              width: '100%', 
              height: '20px', 
              backgroundColor: '#f0f0f0', 
              borderRadius: '10px',
              overflow: 'hidden'
            }}>
              <div style={{
                width: `${progress.total > 0 ? (progress.current / progress.total) * 100 : 0}%`,
                height: '100%',
                backgroundColor: '#007bff',
                transition: 'width 0.3s ease'
              }} />
            </div>
          </div>
        )}
      </div>
      <div style={{ marginBottom: '1rem' }}>
        <label>
          Início:
          <input
            type="time"
            value={startTime}
            onChange={e => setStartTime(e.target.value)}
          />
        </label>
        <label style={{ marginLeft: '0.5rem' }}>
          Fim:
          <input
            type="time"
            value={endTime}
            onChange={e => setEndTime(e.target.value)}
          />
        </label>
        <label style={{ marginLeft: '0.5rem' }}>
          Faixa:
          <select
            value={windowMinutes}
            onChange={e => setWindowMinutes(Number(e.target.value))}
          >
            <option value={1}>1min</option>
            <option value={5}>5min</option>
            <option value={10}>10min</option>
            <option value={30}>30min</option>
            <option value={60}>1h</option>
          </select>
        </label>
        <label style={{ marginLeft: '0.5rem' }}>
          Atualização:
          <select
            value={refreshMs}
            onChange={e => setRefreshMs(Number(e.target.value))}
          >
            <option value={1000}>1s</option>
            <option value={10000}>10s</option>
            <option value={30000}>30s</option>
            <option value={60000}>1min</option>
          </select>
        </label>
      </div>
      <div style={{ marginBottom: '1rem' }}>
        <table>
          <thead>
            <tr>
              <th>Endpoint</th>
              <th>Sucesso</th>
              <th>Erro</th>
            </tr>
          </thead>
          <tbody>
            {Object.entries(endpointStats).map(([ep, c]) => (
              <tr key={ep}>
                <td>{ep}</td>
                <td>{c.success || 0}</td>
                <td>{c.error || 0}</td>
              </tr>
            ))}
          </tbody>
        </table>
      </div>
      <LineChart width={800} height={300} data={displayData}>
        <CartesianGrid stroke="#ccc" />
        <XAxis
          dataKey="time"
          type="number"
          domain={[displayStart, displayEnd]}
          tickFormatter={t => new Date(t).toLocaleTimeString()}
        />
        <YAxis allowDecimals={false} />
        <Tooltip labelFormatter={t => new Date(t).toLocaleTimeString()} />
        <Legend />
        {lines}
      </LineChart>
    </div>
  );
}
