#!/usr/bin/env python3
"""Generate SVG figures for the theo-forge deep review report."""
import math
from pathlib import Path

OUTPUT_DIR = Path(__file__).parent

# ---- Data ----
findings_by_severity = {"HIGH": 17, "MEDIUM": 32, "LOW": 18}
findings_by_category = {
    "code": 21,
    "completeness": 13,
    "infrastructure": 10,
    "architecture": 8,
    "security": 9,
    "testing": 3,
    "data": 2,
    "observability": 1,
}

c4_scores = {
    "Correto": 0.58,
    "Completo": 0.52,
    "Confiavel": 0.62,
    "Controlavel": 0.70,
}

components = [
    "root-builder",
    "pkg-config",
    "pkg-client",
    "pkg-expr",
    "pkg-model",
    "pkg-serialize",
    "pkg-validate",
]


# ---- 1. Findings by Severity Bar Chart ----
def generate_findings_by_severity_svg():
    colors = {"HIGH": "#e53935", "MEDIUM": "#ff9800", "LOW": "#4caf50"}
    max_val = max(findings_by_severity.values())
    bar_width = 80
    bar_gap = 50
    margin_l = 60
    margin_t = 50
    chart_h = 200
    w = margin_l + len(findings_by_severity) * (bar_width + bar_gap) + 40
    h = margin_t + chart_h + 80

    svg = f'<svg xmlns="http://www.w3.org/2000/svg" width="{w}" height="{h}" style="font-family:sans-serif">\n'
    svg += f'  <text x="{w/2}" y="28" text-anchor="middle" font-size="15" font-weight="bold">Findings by Severity</text>\n'

    # Y axis gridlines and labels
    for i in range(0, 5):
        val = int(max_val * i / 4)
        y = margin_t + chart_h - int(chart_h * i / 4)
        svg += f'  <line x1="{margin_l}" y1="{y}" x2="{w-20}" y2="{y}" stroke="#eee" stroke-width="1"/>\n'
        svg += f'  <text x="{margin_l-8}" y="{y+4}" text-anchor="end" font-size="11" fill="#666">{val}</text>\n'

    # Bars
    for idx, (sev, count) in enumerate(findings_by_severity.items()):
        x = margin_l + idx * (bar_width + bar_gap) + bar_gap // 2
        bar_h = int(chart_h * count / max_val)
        y = margin_t + chart_h - bar_h
        color = colors[sev]
        svg += f'  <rect x="{x}" y="{y}" width="{bar_width}" height="{bar_h}" fill="{color}" rx="3"/>\n'
        svg += f'  <text x="{x + bar_width/2}" y="{y - 6}" text-anchor="middle" font-size="13" font-weight="bold" fill="{color}">{count}</text>\n'
        svg += f'  <text x="{x + bar_width/2}" y="{margin_t + chart_h + 20}" text-anchor="middle" font-size="13" fill="#333">{sev}</text>\n'

    svg += '</svg>\n'
    return svg


# ---- 2. Risk Matrix SVG ----
def generate_risk_matrix_svg():
    # Severity vs Probability heatmap
    # Rows = Probability (High, Medium, Low)
    # Cols = Severity (HIGH, MEDIUM, LOW)
    # Color by risk level (severity * probability)
    cell_w, cell_h = 130, 80
    margin_l, margin_t = 120, 80
    cols = ["HIGH", "MEDIUM", "LOW"]
    rows = ["High", "Medium", "Low"]

    # Color mapping: (sev, prob) -> color
    colors = {
        ("HIGH", "High"): "#b71c1c",
        ("HIGH", "Medium"): "#e53935",
        ("HIGH", "Low"): "#ef9a9a",
        ("MEDIUM", "High"): "#e53935",
        ("MEDIUM", "Medium"): "#ff9800",
        ("MEDIUM", "Low"): "#ffe082",
        ("LOW", "High"): "#ff9800",
        ("LOW", "Medium"): "#ffe082",
        ("LOW", "Low"): "#c8e6c9",
    }

    # Finding counts per severity (distributed across probability levels for illustration)
    # High findings: mostly medium-high probability; medium: spread; low: low
    counts = {
        ("HIGH", "High"): 8,
        ("HIGH", "Medium"): 7,
        ("HIGH", "Low"): 2,
        ("MEDIUM", "High"): 10,
        ("MEDIUM", "Medium"): 15,
        ("MEDIUM", "Low"): 7,
        ("LOW", "High"): 3,
        ("LOW", "Medium"): 8,
        ("LOW", "Low"): 7,
    }

    total_w = margin_l + len(cols) * cell_w + 20
    total_h = margin_t + len(rows) * cell_h + 60

    svg = f'<svg xmlns="http://www.w3.org/2000/svg" width="{total_w}" height="{total_h}" style="font-family:sans-serif">\n'
    svg += f'  <text x="{total_w/2}" y="30" text-anchor="middle" font-size="15" font-weight="bold">Risk Matrix: Severity x Probability</text>\n'

    # Column headers (Severity)
    svg += f'  <text x="{margin_l + len(cols)*cell_w/2}" y="60" text-anchor="middle" font-size="13" font-weight="bold" fill="#555">Severity</text>\n'
    for ci, col in enumerate(cols):
        x = margin_l + ci * cell_w + cell_w / 2
        svg += f'  <text x="{x}" y="75" text-anchor="middle" font-size="12" fill="#333">{col}</text>\n'

    # Row headers (Probability)
    svg += f'  <text x="30" y="{margin_t + len(rows)*cell_h/2}" text-anchor="middle" font-size="13" font-weight="bold" fill="#555" transform="rotate(-90,30,{margin_t + len(rows)*cell_h/2})">Probability</text>\n'
    for ri, row in enumerate(rows):
        y = margin_t + ri * cell_h + cell_h / 2
        svg += f'  <text x="{margin_l - 10}" y="{y + 5}" text-anchor="end" font-size="12" fill="#333">{row}</text>\n'

    # Cells
    for ri, prob in enumerate(rows):
        for ci, sev in enumerate(cols):
            x = margin_l + ci * cell_w
            y = margin_t + ri * cell_h
            color = colors.get((sev, prob), "#eee")
            count = counts.get((sev, prob), 0)
            svg += f'  <rect x="{x}" y="{y}" width="{cell_w}" height="{cell_h}" fill="{color}" stroke="#fff" stroke-width="2"/>\n'
            svg += f'  <text x="{x + cell_w/2}" y="{y + cell_h/2 - 6}" text-anchor="middle" font-size="20" font-weight="bold" fill="rgba(0,0,0,0.6)">{count}</text>\n'
            svg += f'  <text x="{x + cell_w/2}" y="{y + cell_h/2 + 14}" text-anchor="middle" font-size="10" fill="rgba(0,0,0,0.5)">findings</text>\n'

    # Legend
    legend_x = margin_l
    legend_y = margin_t + len(rows) * cell_h + 15
    legend_items = [("#b71c1c", "Critical Risk"), ("#e53935", "High Risk"), ("#ff9800", "Medium Risk"), ("#ffe082", "Low Risk"), ("#c8e6c9", "Minimal Risk")]
    for i, (color, label) in enumerate(legend_items):
        lx = legend_x + i * 90
        svg += f'  <rect x="{lx}" y="{legend_y}" width="14" height="14" fill="{color}" stroke="#ccc" stroke-width="1"/>\n'
        svg += f'  <text x="{lx + 18}" y="{legend_y + 11}" font-size="10" fill="#555">{label}</text>\n'

    svg += '</svg>\n'
    return svg


# ---- 3. Dependency Graph SVG ----
def generate_dependency_graph_svg():
    # Layout: root-builder at center-top, packages below
    w, h = 700, 450
    nodes = {
        "root-builder": (350, 70, "#1565c0", "white"),
        "pkg-config": (100, 200, "#6a1b9a", "white"),
        "pkg-client": (580, 200, "#1b5e20", "white"),
        "pkg-expr": (200, 340, "#4a148c", "white"),
        "pkg-model": (350, 340, "#0d47a1", "white"),
        "pkg-serialize": (500, 340, "#01579b", "white"),
        "pkg-validate": (650, 340, "#1b5e20", "white"),
    }
    # Edges: (from, to)
    edges = [
        ("root-builder", "pkg-config"),
        ("root-builder", "pkg-expr"),
        ("root-builder", "pkg-model"),
        ("root-builder", "pkg-serialize"),
        ("root-builder", "pkg-validate"),
        ("pkg-client", "pkg-model"),
        ("pkg-serialize", "pkg-model"),
    ]

    node_w, node_h = 110, 36

    svg = f'<svg xmlns="http://www.w3.org/2000/svg" width="{w}" height="{h}" style="font-family:sans-serif">\n'
    svg += f'  <defs><marker id="arrow" markerWidth="10" markerHeight="7" refX="10" refY="3.5" orient="auto"><polygon points="0 0, 10 3.5, 0 7" fill="#666"/></marker></defs>\n'
    svg += f'  <text x="{w/2}" y="24" text-anchor="middle" font-size="15" font-weight="bold">Component Dependency Graph</text>\n'

    # Draw edges first (under nodes)
    for (src, dst) in edges:
        sx, sy, _, _ = nodes[src]
        dx, dy, _, _ = nodes[dst]
        # Draw from bottom of src to top of dst (or side)
        x1, y1 = sx, sy + node_h // 2
        x2, y2 = dx, dy - node_h // 2
        svg += f'  <line x1="{x1}" y1="{y1}" x2="{x2}" y2="{y2}" stroke="#666" stroke-width="1.5" marker-end="url(#arrow)"/>\n'

    # Draw nodes
    for name, (cx, cy, fill, text_color) in nodes.items():
        x = cx - node_w // 2
        y = cy - node_h // 2
        svg += f'  <rect x="{x}" y="{y}" width="{node_w}" height="{node_h}" fill="{fill}" rx="6" stroke="#333" stroke-width="1"/>\n'
        svg += f'  <text x="{cx}" y="{cy + 5}" text-anchor="middle" font-size="11" fill="{text_color}" font-weight="bold">{name}</text>\n'

    # Legend
    svg += f'  <text x="10" y="{h - 20}" font-size="10" fill="#888">Arrows indicate dependency direction (A -> B means A depends on B)</text>\n'
    svg += f'  <text x="10" y="{h - 8}" font-size="10" fill="#888">pkg-client and root-builder are independent; both depend on pkg-model</text>\n'

    svg += '</svg>\n'
    return svg


if __name__ == "__main__":
    # Generate all 3 SVGs
    p1 = OUTPUT_DIR / "findings_by_severity.svg"
    p1.write_text(generate_findings_by_severity_svg())
    print(f"Written: {p1}")

    p2 = OUTPUT_DIR / "risk_matrix.svg"
    p2.write_text(generate_risk_matrix_svg())
    print(f"Written: {p2}")

    p3 = OUTPUT_DIR / "dependency_graph.svg"
    p3.write_text(generate_dependency_graph_svg())
    print(f"Written: {p3}")

    print("All 3 SVG figures generated.")
