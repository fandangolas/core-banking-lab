import React, { useEffect, useState } from 'react';
import { LineChart, Line, XAxis, YAxis, Tooltip, CartesianGrid } from 'recharts';

export default function App() {
  const [data, setData] = useState([]);

  useEffect(() => {
    const fetchMetrics = async () => {
      const res = await fetch('/metrics');
      const json = await res.json();
      const formatted = json.map(m => {
        const name = m.endpoint || m.Endpoint;
        const durStr = String(m.duration || m.Duration || '');
        return {
          name,
          duration: parseFloat(durStr.replace(/[^0-9.]/g, ''))
        };
      });
      setData(formatted);
    };
    fetchMetrics();
    const id = setInterval(fetchMetrics, 1000);
    return () => clearInterval(id);
  }, []);

  return (
    <div style={{ width: '100%', height: 400 }}>
      <LineChart width={600} height={300} data={data}>
        <CartesianGrid stroke="#ccc" />
        <XAxis dataKey="name" />
        <YAxis />
        <Tooltip />
        <Line type="monotone" dataKey="duration" stroke="#8884d8" />
      </LineChart>
    </div>
  );
}
