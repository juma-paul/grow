import {
  ComposedChart,
  Bar,
  Line,
  XAxis,
  YAxis,
  CartesianGrid,
  Tooltip,
  ResponsiveContainer,
  Cell,
} from "recharts";
import { useEventStore } from "../store";

function cssVar(name: string): string {
  return getComputedStyle(document.documentElement).getPropertyValue(name).trim();
}

export default function CostGraph() {
  const costs = useEventStore((s) => s.costs);

  if (costs.length === 0) return null;

  const surface1 = cssVar("--surface-1");
  const border = cssVar("--border");
  const text0 = cssVar("--text-0");
  const text1 = cssVar("--text-1");
  const text2 = cssVar("--text-2");

  return (
    <div className="shrink-0 border-t border-[var(--border)] px-6 py-3">
      <div className="mb-1 text-sm font-medium text-[var(--text-1)]">
        Cost per append
      </div>
      <ResponsiveContainer width="100%" height={160}>
        <ComposedChart
          data={costs}
          margin={{ top: 4, right: 4, bottom: 0, left: -20 }}
        >
          <CartesianGrid
            strokeDasharray="3 3"
            stroke={`${border}80`}
            vertical={false}
          />
          <XAxis
            dataKey="op"
            tick={{ fontSize: 12, fill: text2 }}
            tickLine={false}
            axisLine={{ stroke: border }}
          />
          <YAxis
            tick={{ fontSize: 12, fill: text2 }}
            tickLine={false}
            axisLine={false}
            allowDecimals
          />
          <Tooltip
            contentStyle={{
              backgroundColor: surface1,
              border: `1px solid ${border}`,
              borderRadius: 8,
              fontSize: 13,
              color: text0,
            }}
            labelStyle={{ color: text1 }}
            itemStyle={{ color: text0 }}
            labelFormatter={(op) => `Append #${op}`}
            formatter={(value, name, entry) => {
              if (name === "amortized") return [value, "Amortized avg"];
              const payload = entry.payload as { isResize: boolean };
              return [value, payload.isResize ? "Cost (resize)" : "Cost"];
            }}
            cursor={{ fill: `${border}4D` }}
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
          <Line
            dataKey="amortized"
            type="monotone"
            stroke="#8b5cf6"
            strokeWidth={2}
            dot={false}
            isAnimationActive={false}
          />
        </ComposedChart>
      </ResponsiveContainer>
    </div>
  );
}
