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

  useEffect(() => {
    const fetchMetrics = async () => {
      const res = await fetch('/metrics');
      const metrics = await res.json();
      const newMetrics = metrics.slice(lastCount.current);
      lastCount.current = metrics.length;

      if (newMetrics.length === 0) return;

      const now = Date.now();
      const cutoff = now - 5 * 60 * 1000; // last 5 minutes
      const newEndpoints = new Set();

      setData(prev => {
        const filtered = prev.filter(d => d.time >= cutoff);
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
    const id = setInterval(fetchMetrics, 1000);
    return () => clearInterval(id);
  }, []);

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
      <LineChart width={800} height={300} data={data}>
        <CartesianGrid stroke="#ccc" />
        <XAxis
          dataKey="time"
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
