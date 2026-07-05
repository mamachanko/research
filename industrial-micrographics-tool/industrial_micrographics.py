#!/usr/bin/env python3
"""Generate small industrial-style SVG micrographics for agent HTML responses.

The tool intentionally has no third-party dependencies. It emits compact SVGs that
can be saved as files or inlined directly into rich HTML responses.
"""
from __future__ import annotations

import argparse
import hashlib
import html
import math
import random
from dataclasses import dataclass
from pathlib import Path
from typing import Sequence

PALETTES = {
    "amber": {
        "bg": "#111315",
        "panel": "#191d20",
        "panel2": "#20262a",
        "line": "#f2a900",
        "muted": "#6e7781",
        "text": "#e8edf2",
        "subtext": "#9aa4ad",
        "accent": "#ffcf5a",
        "danger": "#ff5d4d",
        "ok": "#55d187",
    },
    "cyan": {
        "bg": "#0b1117",
        "panel": "#111b24",
        "panel2": "#172634",
        "line": "#27d7ff",
        "muted": "#526579",
        "text": "#e8f6ff",
        "subtext": "#92a8b8",
        "accent": "#8ff0ff",
        "danger": "#ff6678",
        "ok": "#63e6a6",
    },
    "lime": {
        "bg": "#0f130d",
        "panel": "#171d13",
        "panel2": "#202817",
        "line": "#a6e22e",
        "muted": "#64705d",
        "text": "#f0f7e8",
        "subtext": "#a4b197",
        "accent": "#d8ff70",
        "danger": "#ff6b4a",
        "ok": "#7dff9d",
    },
}


@dataclass(frozen=True)
class Svg:
    width: int
    height: int
    parts: list[str]

    def add(self, value: str) -> None:
        self.parts.append(value)

    def render(self) -> str:
        return "\n".join(
            [
                f'<svg xmlns="http://www.w3.org/2000/svg" width="{self.width}" height="{self.height}" viewBox="0 0 {self.width} {self.height}" role="img">',
                *self.parts,
                "</svg>",
            ]
        )


def esc(value: str) -> str:
    return html.escape(value, quote=True)


def seed_from(*values: str) -> int:
    digest = hashlib.sha256("|".join(values).encode()).hexdigest()
    return int(digest[:12], 16)


def style_block(palette: dict[str, str]) -> str:
    return f"""<style>
    .bg {{ fill: {palette['bg']}; }}
    .panel {{ fill: {palette['panel']}; stroke: {palette['muted']}; stroke-width: 1; }}
    .panel2 {{ fill: {palette['panel2']}; stroke: {palette['line']}; stroke-width: 1; }}
    .grid {{ stroke: {palette['muted']}; stroke-width: .7; opacity: .28; }}
    .hair {{ stroke: {palette['muted']}; stroke-width: 1; fill: none; opacity: .7; }}
    .wire {{ stroke: {palette['line']}; stroke-width: 2; fill: none; stroke-linecap: round; stroke-linejoin: round; }}
    .wire-muted {{ stroke: {palette['muted']}; stroke-width: 1.5; fill: none; stroke-linecap: round; }}
    .accent {{ stroke: {palette['accent']}; fill: none; }}
    .txt {{ font-family: ui-monospace, SFMono-Regular, Menlo, Consolas, monospace; fill: {palette['text']}; font-size: 12px; letter-spacing: .06em; }}
    .sub {{ font-family: ui-monospace, SFMono-Regular, Menlo, Consolas, monospace; fill: {palette['subtext']}; font-size: 9px; letter-spacing: .08em; }}
    .tiny {{ font-family: ui-monospace, SFMono-Regular, Menlo, Consolas, monospace; fill: {palette['subtext']}; font-size: 7px; letter-spacing: .1em; }}
    .ok {{ fill: {palette['ok']}; }}
    .danger {{ fill: {palette['danger']}; }}
    .linefill {{ fill: {palette['line']}; }}
    </style>"""


def base_svg(width: int, height: int, palette: dict[str, str], title: str) -> Svg:
    svg = Svg(width, height, [])
    svg.add(f"<title>{esc(title)}</title>")
    svg.add(style_block(palette))
    svg.add(f'<rect class="bg" x="0" y="0" width="{width}" height="{height}" rx="12"/>')
    for x in range(12, width, 16):
        svg.add(f'<line class="grid" x1="{x}" y1="8" x2="{x}" y2="{height-8}"/>')
    for y in range(12, height, 16):
        svg.add(f'<line class="grid" x1="8" y1="{y}" x2="{width-8}" y2="{y}"/>')
    svg.add(f'<rect class="hair" x="6" y="6" width="{width-12}" height="{height-12}" rx="10"/>')
    return svg


def label(svg: Svg, x: int, y: int, text: str, cls: str = "txt") -> None:
    svg.add(f'<text class="{cls}" x="{x}" y="{y}">{esc(text)}</text>')


def panel(svg: Svg, x: int, y: int, w: int, h: int, label_text: str) -> None:
    svg.add(f'<rect class="panel" x="{x}" y="{y}" width="{w}" height="{h}" rx="8"/>')
    label(svg, x + 10, y + 18, label_text[:18].upper(), "sub")
    svg.add(f'<line class="hair" x1="{x+8}" y1="{y+24}" x2="{x+w-8}" y2="{y+24}"/>')


def render_pipeline(args: argparse.Namespace) -> str:
    palette = PALETTES[args.palette]
    steps = [s.strip() for s in args.steps.split(",") if s.strip()] or ["INPUT", "QUEUE", "WORKER", "OUTPUT"]
    svg = base_svg(args.width, args.height, palette, args.title)
    label(svg, 16, 24, args.title.upper())
    label(svg, 18, args.height - 13, "PIPELINE / FLOW / HANDOFF", "tiny")
    usable_w = args.width - 44
    gap = 14
    node_w = max(54, int((usable_w - gap * (len(steps) - 1)) / len(steps)))
    y = args.height // 2 - 24
    for i, step in enumerate(steps):
        x = 22 + i * (node_w + gap)
        panel(svg, x, y, node_w, 54, step)
        svg.add(f'<circle class="linefill" cx="{x+12}" cy="{y+39}" r="3"/>')
        label(svg, x + 23, y + 43, f"S{i+1:02d}", "tiny")
        if i < len(steps) - 1:
            x2 = x + node_w + gap
            mid_y = y + 27
            svg.add(f'<path class="wire" d="M{x+node_w} {mid_y} L{x2-8} {mid_y}"/>')
            svg.add(f'<path class="linefill" d="M{x2-8} {mid_y-4} L{x2} {mid_y} L{x2-8} {mid_y+4} Z"/>')
    return svg.render()


def render_status(args: argparse.Namespace) -> str:
    palette = PALETTES[args.palette]
    metrics = [m.strip() for m in args.metrics.split(",") if m.strip()] or ["latency:42ms", "errors:0", "queue:17", "cpu:61%"]
    svg = base_svg(args.width, args.height, palette, args.title)
    label(svg, 16, 24, args.title.upper())
    label(svg, 18, args.height - 13, "STATUS PLATE / OPERATIONS", "tiny")
    cols = 2
    cell_w = (args.width - 44) // cols
    cell_h = 42
    for i, metric in enumerate(metrics[:6]):
        name, _, value = metric.partition(":")
        x = 22 + (i % cols) * cell_w
        y = 42 + (i // cols) * (cell_h + 10)
        svg.add(f'<rect class="panel2" x="{x}" y="{y}" width="{cell_w-10}" height="{cell_h}" rx="6"/>')
        svg.add(f'<circle class="{"ok" if i % 3 != 1 else "linefill"}" cx="{x+12}" cy="{y+15}" r="4"/>')
        label(svg, x + 24, y + 18, name.upper()[:16], "sub")
        label(svg, x + 24, y + 34, (value or "--").upper()[:14], "txt")
    return svg.render()


def render_trace(args: argparse.Namespace) -> str:
    palette = PALETTES[args.palette]
    rng = random.Random(seed_from(args.title, args.palette, str(args.width), str(args.height)))
    svg = base_svg(args.width, args.height, palette, args.title)
    label(svg, 16, 24, args.title.upper())
    left, top, w, h = 22, 46, args.width - 44, args.height - 74
    svg.add(f'<rect class="panel" x="{left}" y="{top}" width="{w}" height="{h}" rx="8"/>')
    points = []
    for i in range(32):
        x = left + 10 + i * ((w - 20) / 31)
        y = top + h / 2 + math.sin(i * .55) * h * .22 + rng.uniform(-h * .08, h * .08)
        points.append(f"{x:.1f},{y:.1f}")
    svg.add(f'<polyline class="wire" points="{" ".join(points)}"/>')
    for i in range(0, 32, 5):
        x, y = [float(v) for v in points[i].split(",")]
        svg.add(f'<circle class="linefill" cx="{x:.1f}" cy="{y:.1f}" r="2.5"/>')
    label(svg, 18, args.height - 13, "SIGNAL TRACE / TELEMETRY", "tiny")
    return svg.render()


def build_parser() -> argparse.ArgumentParser:
    parser = argparse.ArgumentParser(description="Generate industrial-style SVG micrographics.")
    parser.add_argument("kind", choices=["pipeline", "status", "trace"])
    parser.add_argument("--title", default="service map")
    parser.add_argument("--palette", choices=sorted(PALETTES), default="amber")
    parser.add_argument("--width", type=int, default=420)
    parser.add_argument("--height", type=int, default=180)
    parser.add_argument("--steps", default="api,queue,worker,db")
    parser.add_argument("--metrics", default="latency:42ms,errors:0,queue:17,cpu:61%")
    parser.add_argument("--out", type=Path, default=None, help="Write SVG to this path; stdout if omitted.")
    return parser


def main(argv: Sequence[str] | None = None) -> int:
    parser = build_parser()
    args = parser.parse_args(argv)
    renderers = {
        "pipeline": render_pipeline,
        "status": render_status,
        "trace": render_trace,
    }
    svg = renderers[args.kind](args)
    if args.out:
        args.out.write_text(svg + "\n", encoding="utf-8")
    else:
        print(svg)
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
