import { BarChart, Bar, XAxis, YAxis, CartesianGrid, Tooltip, ResponsiveContainer } from 'recharts';
import type { UsageSummary } from '../types';

interface CostChartProps {
  data: UsageSummary[];
}

export function CostChart({ data }: CostChartProps) {
  return (
    <ResponsiveContainer width="100%" height={300}>
      <BarChart data={data}>
        <CartesianGrid strokeDasharray="3 3" />
        <XAxis dataKey="period" />
        <YAxis />
        <Tooltip formatter={(value: number) => `$${value.toFixed(4)}`} />
        <Bar dataKey="total_cost_usd" fill="#3b82f6" name="Cost (USD)" />
      </BarChart>
    </ResponsiveContainer>
  );
}
