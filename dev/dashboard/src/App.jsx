import React, { useEffect, useRef, useState } from 'react';
import {
  LineChart,
  Line,
  XAxis,
  YAxis,
  Tooltip,
  CartesianGrid,
  Legend,
  ResponsiveContainer
} from 'recharts';
import { runBatchedSimulation, createAccount, deposit, withdraw, transfer } from './simulator';

const COLORS = [
  '#3b82f6', // blue-500
  '#10b981', // emerald-500
  '#f59e0b', // amber-500
  '#ef4444', // red-500
  '#8b5cf6', // violet-500
  '#06b6d4', // cyan-500
  '#f97316'  // orange-500
];

const Card = ({ children, className = '' }) => (
  <div className={`bg-white rounded-xl shadow-lg border border-gray-200 ${className}`}>
    {children}
  </div>
);

const Button = ({ children, onClick, disabled, variant = 'primary', className = '' }) => {
  const baseClasses = 'px-6 py-2.5 rounded-lg font-medium transition-all duration-200 focus:outline-none focus:ring-2 focus:ring-offset-2';
  const variants = {
    primary: disabled 
      ? 'bg-gray-300 text-gray-500 cursor-not-allowed' 
      : 'bg-blue-600 hover:bg-blue-700 text-white focus:ring-blue-500 shadow-lg hover:shadow-xl transform hover:-translate-y-0.5',
    secondary: 'bg-gray-100 hover:bg-gray-200 text-gray-700 focus:ring-gray-500'
  };
  
  return (
    <button 
      onClick={onClick}
      disabled={disabled}
      className={`${baseClasses} ${variants[variant]} ${className}`}
    >
      {children}
    </button>
  );
};

const Input = ({ label, ...props }) => (
  <div className="flex flex-col space-y-1">
    <label className="text-sm font-medium text-gray-700">{label}</label>
    <input 
      {...props}
      className="px-3 py-2 border border-gray-300 rounded-lg focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-transparent disabled:bg-gray-100 disabled:text-gray-500"
    />
  </div>
);

const Select = ({ label, children, ...props }) => (
  <div className="flex flex-col space-y-1">
    <label className="text-sm font-medium text-gray-700">{label}</label>
    <select 
      {...props}
      className="px-3 py-2 border border-gray-300 rounded-lg focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-transparent bg-white"
    >
      {children}
    </select>
  </div>
);

const Checkbox = ({ label, ...props }) => (
  <label className="flex items-center space-x-3 p-3 rounded-lg hover:bg-gray-50 transition-colors cursor-pointer">
    <input 
      type="checkbox" 
      {...props}
      className="w-4 h-4 text-blue-600 rounded focus:ring-blue-500 focus:ring-2 disabled:opacity-50"
    />
    <span className="text-sm font-medium text-gray-700">{label}</span>
  </label>
);

const ProgressBar = ({ current, total }) => {
  const percentage = total > 0 ? (current / total) * 100 : 0;
  return (
    <div className="w-full">
      <div className="flex justify-between text-sm font-medium text-gray-700 mb-2">
        <span>Progresso</span>
        <span>{current} / {total} ({Math.round(percentage)}%)</span>
      </div>
      <div className="w-full bg-gray-200 rounded-full h-3 overflow-hidden">
        <div 
          className="bg-gradient-to-r from-blue-500 to-blue-600 h-3 rounded-full transition-all duration-300 ease-out"
          style={{ width: `${percentage}%` }}
        />
      </div>
    </div>
  );
};

const StatCard = ({ title, success, error }) => (
  <div className="bg-gradient-to-br from-gray-50 to-gray-100 rounded-lg p-4 border border-gray-200">
    <h3 className="font-semibold text-gray-700 text-sm uppercase tracking-wide mb-3">{title}</h3>
    <div className="flex justify-between items-center">
      <div className="text-center">
        <div className="text-2xl font-bold text-green-600">{success || 0}</div>
        <div className="text-xs text-gray-500">Sucessos</div>
      </div>
      <div className="text-center">
        <div className="text-2xl font-bold text-red-600">{error || 0}</div>
        <div className="text-xs text-gray-500">Erros</div>
      </div>
    </div>
  </div>
);

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
      try {
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
      } catch (error) {
        console.error('Erro ao buscar métricas:', error);
      }
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
      strokeWidth={2}
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

      <div className="max-w-7xl mx-auto px-6 py-8 space-y-8">
        {/* Configuração de Simulação */}
        <Card className="p-6">
          <h2 className="text-lg font-semibold text-gray-900 mb-6">⚙️ Configuração da Simulação</h2>
          
          <div className="grid grid-cols-1 md:grid-cols-3 gap-6 mb-6">
            <Input
              label="Número de Requests"
              type="number"
              min="1"
              max="100000"
              value={reqCount}
              onChange={e => setReqCount(Number(e.target.value))}
              disabled={isRunning}
            />
            <Input
              label="Tamanho do Bloco"
              type="number"
              min="1"
              max="1000"
              value={blockSize}
              onChange={e => setBlockSize(Number(e.target.value))}
              disabled={isRunning}
            />
            <Input
              label="Duração do Bloco (ms)"
              type="number"
              min="100"
              max="10000"
              value={blockDuration}
              onChange={e => setBlockDuration(Number(e.target.value))}
              disabled={isRunning}
            />
          </div>

          <div className="mb-6">
            <h3 className="text-sm font-medium text-gray-700 mb-3">Tipos de Operação</h3>
            <div className="grid grid-cols-2 md:grid-cols-3 gap-2">
              <Checkbox
                label="🏦 Criar Conta"
                checked={ops.create}
                onChange={e => setOps({ ...ops, create: e.target.checked })}
                disabled={isRunning}
              />
              <Checkbox
                label="💰 Depósito"
                checked={ops.deposit}
                onChange={e => setOps({ ...ops, deposit: e.target.checked })}
                disabled={isRunning}
              />
              <Checkbox
                label="💸 Saque"
                checked={ops.withdraw}
                onChange={e => setOps({ ...ops, withdraw: e.target.checked })}
                disabled={isRunning}
              />
              <Checkbox
                label="🔄 Transferência"
                checked={ops.transfer}
                onChange={e => setOps({ ...ops, transfer: e.target.checked })}
                disabled={isRunning}
              />
              <Checkbox
                label="🎲 Aleatório"
                checked={randomOp}
                onChange={e => setRandomOp(e.target.checked)}
                disabled={isRunning}
              />
            </div>
          </div>

          <div className="flex justify-center">
            <Button
              onClick={handleRun}
              disabled={isRunning}
              className="px-12"
            >
              {isRunning ? (
                <div className="flex items-center space-x-2">
                  <div className="w-4 h-4 border-2 border-white border-t-transparent rounded-full animate-spin"></div>
                  <span>Executando...</span>
                </div>
              ) : (
                '🚀 Disparar Simulação'
              )}
            </Button>
          </div>

          {isRunning && (
            <div className="mt-6">
              <ProgressBar current={progress.current} total={progress.total} />
            </div>
          )}
        </Card>

        {/* Estatísticas */}
        {Object.keys(endpointStats).length > 0 && (
          <Card className="p-6">
            <h2 className="text-lg font-semibold text-gray-900 mb-6">📊 Estatísticas por Endpoint</h2>
            <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-4">
              {Object.entries(endpointStats).map(([ep, stats]) => (
                <StatCard
                  key={ep}
                  title={ep}
                  success={stats.success}
                  error={stats.error}
                />
              ))}
            </div>
          </Card>
        )}

        {/* Controles do Gráfico */}
        <Card className="p-6">
          <h2 className="text-lg font-semibold text-gray-900 mb-6">📈 Monitoramento em Tempo Real</h2>
          
          <div className="grid grid-cols-1 md:grid-cols-5 gap-4 mb-6">
            <Input
              label="Início"
              type="time"
              value={startTime}
              onChange={e => setStartTime(e.target.value)}
            />
            <Input
              label="Fim"
              type="time"
              value={endTime}
              onChange={e => setEndTime(e.target.value)}
            />
            <Select
              label="Janela de Tempo"
              value={windowMinutes}
              onChange={e => setWindowMinutes(Number(e.target.value))}
            >
              <option value={1}>1 minuto</option>
              <option value={5}>5 minutos</option>
              <option value={10}>10 minutos</option>
              <option value={30}>30 minutos</option>
              <option value={60}>1 hora</option>
            </Select>
            <Select
              label="Taxa de Atualização"
              value={refreshMs}
              onChange={e => setRefreshMs(Number(e.target.value))}
            >
              <option value={1000}>1 segundo</option>
              <option value={10000}>10 segundos</option>
              <option value={30000}>30 segundos</option>
              <option value={60000}>1 minuto</option>
            </Select>
          </div>

          {/* Gráfico */}
          <div className="h-96 w-full">
            <ResponsiveContainer width="100%" height="100%">
              <LineChart data={displayData}>
                <CartesianGrid strokeDasharray="3 3" stroke="#e5e7eb" />
                <XAxis
                  dataKey="time"
                  type="number"
                  domain={[displayStart, displayEnd]}
                  tickFormatter={t => new Date(t).toLocaleTimeString()}
                  stroke="#6b7280"
                />
                <YAxis allowDecimals={false} stroke="#6b7280" />
                <Tooltip 
                  labelFormatter={t => new Date(t).toLocaleTimeString()}
                  contentStyle={{ 
                    backgroundColor: 'white', 
                    border: '1px solid #e5e7eb',
                    borderRadius: '8px',
                    boxShadow: '0 10px 15px -3px rgba(0, 0, 0, 0.1)'
                  }}
                />
                <Legend />
                {lines}
              </LineChart>
            </ResponsiveContainer>
          </div>
        </Card>
      </div>
    </div>
  );
}