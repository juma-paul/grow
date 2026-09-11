#!/usr/bin/env python3
"""Generate a latency-vs-concurrency SVG chart from benchmark CSV.

Usage:
    go run scripts/benchmark.go | tee results.csv
    python3 scripts/chart_latency.py results.csv -o docs/latency.svg

Reads CSV with columns: conc,p50_ms,p95_ms,p99_ms,avg_ms,rps,errors,cpu_delta_s
Produces a dual-axis SVG: latency (left) and RPS (right).
"""

import argparse
import csv
import sys

SVG_W, SVG_H = 720, 400
MARGIN = {"top": 40, "right": 70, "bottom": 50, "left": 60}
PLOT_W = SVG_W - MARGIN["left"] - MARGIN["right"]
PLOT_H = SVG_H - MARGIN["top"] - MARGIN["bottom"]

COLORS = {"p50": "#34d399", "p95": "#fbbf24", "p99": "#f87171", "rps": "#818cf8"}


def read_csv(path):
    rows = []
    with open(path) as f:
        reader = csv.DictReader(f)
        for row in reader:
            rows.append({
                "conc": int(row["conc"]),
                "p50": float(row["p50_ms"]),
                "p95": float(row["p95_ms"]),
                "p99": float(row["p99_ms"]),
                "rps": float(row["rps"]),
            })
    return rows


def scale(val, lo, hi, out_lo, out_hi):
    if hi == lo:
        return (out_lo + out_hi) / 2
    return out_lo + (val - lo) / (hi - lo) * (out_hi - out_lo)


def polyline(points, color, label):
    pts = " ".join(f"{x:.1f},{y:.1f}" for x, y in points)
    return (
        f'<polyline points="{pts}" fill="none" stroke="{color}" '
        f'stroke-width="2" stroke-linejoin="round"/>\n'
        + "".join(
            f'<circle cx="{x:.1f}" cy="{y:.1f}" r="3" fill="{color}"/>\n'
            for x, y in points
        )
    )


def generate_svg(rows, out):
    concs = [r["conc"] for r in rows]
    max_lat = max(r["p99"] for r in rows) * 1.15
    max_rps = max(r["rps"] for r in rows) * 1.15

    def x(conc):
        idx = concs.index(conc)
        return MARGIN["left"] + idx / max(len(concs) - 1, 1) * PLOT_W

    def y_lat(v):
        return MARGIN["top"] + PLOT_H - scale(v, 0, max_lat, 0, PLOT_H)

    def y_rps(v):
        return MARGIN["top"] + PLOT_H - scale(v, 0, max_rps, 0, PLOT_H)

    svg = [
        f'<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 {SVG_W} {SVG_H}" '
        f'font-family="system-ui, sans-serif" font-size="12">\n',
        f'<rect width="{SVG_W}" height="{SVG_H}" fill="#0f1117" rx="8"/>\n',
    ]

    # Grid lines
    for i in range(5):
        gy = MARGIN["top"] + i * PLOT_H / 4
        svg.append(
            f'<line x1="{MARGIN["left"]}" y1="{gy:.0f}" '
            f'x2="{MARGIN["left"] + PLOT_W}" y2="{gy:.0f}" '
            f'stroke="#1e2130" stroke-width="1"/>\n'
        )
        lat_val = max_lat * (4 - i) / 4
        svg.append(
            f'<text x="{MARGIN["left"] - 8}" y="{gy + 4:.0f}" '
            f'text-anchor="end" fill="#6b7280">{lat_val:.0f}</text>\n'
        )
        rps_val = max_rps * (4 - i) / 4
        svg.append(
            f'<text x="{MARGIN["left"] + PLOT_W + 8}" y="{gy + 4:.0f}" '
            f'text-anchor="start" fill="#818cf8">{rps_val:.0f}</text>\n'
        )

    # X-axis labels
    for conc in concs:
        cx = x(conc)
        svg.append(
            f'<text x="{cx:.0f}" y="{MARGIN["top"] + PLOT_H + 20}" '
            f'text-anchor="middle" fill="#6b7280">{conc}</text>\n'
        )

    # Axis titles
    svg.append(
        f'<text x="{SVG_W / 2}" y="{SVG_H - 5}" text-anchor="middle" '
        f'fill="#9ca3af" font-size="13">Concurrency</text>\n'
    )
    svg.append(
        f'<text x="15" y="{SVG_H / 2}" text-anchor="middle" '
        f'fill="#9ca3af" font-size="13" transform="rotate(-90,15,{SVG_H / 2})">'
        f'Latency (ms)</text>\n'
    )
    svg.append(
        f'<text x="{SVG_W - 15}" y="{SVG_H / 2}" text-anchor="middle" '
        f'fill="#818cf8" font-size="13" transform="rotate(90,{SVG_W - 15},{SVG_H / 2})">'
        f'Requests/sec</text>\n'
    )

    # Title
    svg.append(
        f'<text x="{SVG_W / 2}" y="24" text-anchor="middle" '
        f'fill="#e5e7eb" font-size="16" font-weight="600">'
        f'Latency vs Concurrency</text>\n'
    )

    # Data lines
    for key in ("p50", "p95", "p99"):
        pts = [(x(r["conc"]), y_lat(r[key])) for r in rows]
        svg.append(polyline(pts, COLORS[key], key))

    rps_pts = [(x(r["conc"]), y_rps(r["rps"])) for r in rows]
    svg.append(
        f'<polyline points="{" ".join(f"{px:.1f},{py:.1f}" for px, py in rps_pts)}" '
        f'fill="none" stroke="{COLORS["rps"]}" stroke-width="2" '
        f'stroke-dasharray="6,3" stroke-linejoin="round"/>\n'
    )
    svg.append(
        "".join(
            f'<circle cx="{px:.1f}" cy="{py:.1f}" r="3" fill="{COLORS["rps"]}"/>\n'
            for px, py in rps_pts
        )
    )

    # Legend
    lx = MARGIN["left"] + 10
    ly = MARGIN["top"] + 12
    for i, (key, color) in enumerate(COLORS.items()):
        svg.append(
            f'<rect x="{lx + i * 70}" y="{ly - 8}" width="12" height="12" '
            f'rx="2" fill="{color}"/>\n'
            f'<text x="{lx + i * 70 + 16}" y="{ly + 2}" fill="#d1d5db" '
            f'font-size="11">{key}</text>\n'
        )

    svg.append("</svg>\n")

    with open(out, "w") as f:
        f.writelines(svg)
    print(f"Chart written to {out}")


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("csv_file", help="benchmark CSV output")
    parser.add_argument("-o", "--output", default="docs/latency.svg")
    args = parser.parse_args()

    rows = read_csv(args.csv_file)
    if not rows:
        print("No data rows found", file=sys.stderr)
        sys.exit(1)

    generate_svg(rows, args.output)


if __name__ == "__main__":
    main()
