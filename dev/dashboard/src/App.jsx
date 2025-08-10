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

  return (
    <div style={{ width: '100%', height: 400 }}>
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
