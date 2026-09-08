import {
  ComposedChart,
  Bar,
  XAxis,
  YAxis,
  CartesianGrid,
  Tooltip,
  ResponsiveContainer,
  Cell,
} from "recharts";
import { useEventStore } from "../store";

export default function CostGraph() {
  const costs = useEventStore((s) => s.costs);

  if (costs.length === 0) return null;

  return (
    <div className="shrink-0 border-t border-zinc-700 px-6 py-3">
      <div className="mb-1 text-sm font-medium text-zinc-400">
        Cost per append
      </div>
      <ResponsiveContainer width="100%" height={160}>
        <ComposedChart
          data={costs}
          margin={{ top: 4, right: 4, bottom: 0, left: -20 }}
        >
          <CartesianGrid
            strokeDasharray="3 3"
            stroke="rgba(63, 63, 70, 0.5)"
            vertical={false}
          />
          <XAxis
            dataKey="op"
            tick={{ fontSize: 12, fill: "#71717a" }}
            tickLine={false}
            axisLine={{ stroke: "#3f3f46" }}
          />
          <YAxis
            tick={{ fontSize: 12, fill: "#71717a" }}
            tickLine={false}
            axisLine={false}
            allowDecimals={false}
          />
          <Tooltip
            contentStyle={{
              backgroundColor: "#18181b",
              border: "1px solid #3f3f46",
              borderRadius: 8,
              fontSize: 13,
              color: "#e4e4e7",
            }}
            labelStyle={{ color: "#a1a1aa" }}
            itemStyle={{ color: "#e4e4e7" }}
            labelFormatter={(op) => `Append #${op}`}
            formatter={(value, _name, entry) => [
              value,
              (entry.payload as { isResize: boolean }).isResize ? "Cost (resize)" : "Cost",
            ]}
            cursor={{ fill: "rgba(63, 63, 70, 0.3)" }}
          />
          <Bar dataKey="cost" radius={[3, 3, 0, 0]} maxBarSize={40}>
            {costs.map((entry, i) => (
              <Cell
                key={i}
                fill={entry.isResize ? "#f59e0b" : "#10b981"}
                fillOpacity={0.85}
              />
            ))}
          </Bar>
        </ComposedChart>
      </ResponsiveContainer>
    </div>
  );
}
