# Visualize Queue Flow — Notes

## Objective

Explore terminal visualizations of two apps communicating over a RabbitMQ-style queue:
- **Publisher** (`node-a`) publishes messages to the queue
- **Consumer** (`node-b`) reads messages from the queue

Goal: visually represent the activity — queue depth, message flow, rates, latency — across
a spectrum from very simple to complex/fancy, recorded as GIFs with VHS.

## Visualization Areas Identified

After thinking through what's interesting about two apps sharing a queue, 8 distinct
conceptual areas emerged:

1. **Queue Depth Meter** — the queue as a resource: how full, filling or draining?
2. **Message Flow Animation** — messages as physical objects traveling through a pipe
3. **Publisher/Consumer Status Panels** — what each app is doing right now
4. **Throughput Graph** — publish and consume rates over time (are they in balance?)
5. **Event Timeline** — scrolling log of individual message lifecycle events
6. **Network Topology** — graph view of nodes connected by their shared queue
7. **Latency Distribution** — how long messages wait before being consumed
8. **Full Dashboard** — all of the above in one coordinated screen

## Complexity Levels

Each area was implemented at three levels:

- **L1 (Simple)**: Single concept rendered with lipgloss borders + color. No movement beyond data
  updates. Emphasises clarity — one key metric visible at a glance.
- **L2 (Medium)**: The focal element animates. Moving characters, blinking spinners, scrolling
  bars, color-gradient transitions. Adds temporal feel without clutter.
- **L3 (Complex)**: Multi-section layout. Sparklines, heat grids, particle streams, rich
  per-component metrics. Requires wider/taller terminal.

Plus one combined **Dashboard** mode.

## Implementation Details

### Simulation Engine (`sim.go`)
- Oscillating pub/consume rates using `sin()` with different periods/phases
- This naturally creates queue filling and draining cycles
- Rates: pub = `3.0 + 2.5·sin(0.07t) + noise`, con = `2.8 + 1.8·sin(0.05t+1.2) + noise`
- FlowParticles track individual messages with `Pos ∈ [0, 1]`
- Queued particles pause in the `[0.35, 0.65]` zone until consumed

### Color palette (Dracula-inspired)
- Publisher: `#BD93F9` (purple)
- Consumer: `#50FA7B` (green)
- Queue: `#FFB86C` (orange)
- In-flight messages: `#8BE9FD` (cyan)
- Dropped / error: `#FF5555` (red)
- Background: `#0D1117` (near-black)

### Notable rendering techniques
- `sparklineAuto()`: braille-height block chars `▁▂▃▄▅▆▇█` for scrolling history
- `dualBarChart()`: interleaved pub/con vertical bars on same Y-scale
- `fillBarAuto()`: auto-colors green/yellow/red based on fill %
- `heatCell()`: background-colored single space through 8-stop color ramp
- `gradientStr()`: per-character HCL gradient titles via go-colorful
- `renderEdge()`: dot particles traveling along a `───●───` edge line
- FlowParticle system: 64-particle cap, `InQueue` flag pauses them mid-pipe

## Tools

- Go 1.24 + Bubble Tea v1.2.4 + Lipgloss v1.0.0 + go-colorful v1.4.0
- VHS v0.9.0 for terminal recording → PNG frame directory
- Python 3 + Pillow for PNG frame → animated GIF assembly (ffmpeg workaround)
- Playwright Chromium (headless) for VHS rendering

## Log

### 2026-03-29

- Confirmed VHS v0.9.0 + Go 1.24 available; symlinked Playwright chromium, installed ttyd
- Designed and implemented simulation engine with sinusoidal oscillating rates
- Implemented all 22 visualizations in 9 Go source files (~2500 lines total)
- Fixed one compile error: `const w = 90` + constant fold in `int()` is invalid in Go
- All 22 VHS tapes recorded successfully to PNG frame directories
- GIF assembly via Pillow: `Image.alpha_composite(text, cursor)` → palette-quantized GIF
- 4 GIFs exceeded 2MB at every-2nd-frame; reduced to every-3rd or every-4th frame

## Hurdles / Lessons Learned

### Playwright ffmpeg is stripped to VP8/WebM only
VHS uses ffmpeg for GIF encoding via a `palettegen`/`paletteuse` two-pass filter_complex.
The Playwright-bundled ffmpeg (`ffmpeg-linux`) disables everything except VP8, WebM, PNG,
MJPEG — no GIF muxer, no palette filters. System `apt-get install ffmpeg` failed due to
broken package dependencies in this environment.

**Workaround**: Remove `Output *.gif` from tapes, keep `Output *.png` (which produces a
frame directory), then assemble GIFs with Python/Pillow: alpha-composite text+cursor frames,
convert to 256-color palette mode, and save as animated GIF.

### VHS 0.9.0 constant quirks (inherited from spinner project)
- Absolute paths in `Output` fail (lexer treats `/` as regex delimiter) → use relative names
- Filenames starting with digit-hyphen fail → use underscores (e.g., `queue_l1.tape`)

### Go: `int()` of non-integer constant expression is a compile error
`const w = 90; const qL = 0.32; int(float64(w)*qL)` is a constant expression = `28.8`,
and `int(28.8)` on a constant is illegal. Fix: declare `w` as `w := 90` (variable).

### Bubble Tea alt-screen + VHS
VHS injects keypresses via a pty, but `tea.WithAltScreen()` means the alternate screen
buffer is used. This works fine with VHS — the frame captures the full alt-screen content.
`tea.Quit` is triggered by maxTicks, so each recording self-terminates cleanly.

### Dashboard layout: Bubble Tea vs. static lipgloss
The dashboard uses `lipgloss.JoinHorizontal` / `JoinVertical` for layout, which is simpler
than a full Bubble Tea layout manager. For fixed-size UIs (known terminal width from VHS)
this works well. For truly responsive layouts, something like `bubbletea-layout` or manual
`tea.WindowSizeMsg` handling per-panel would be needed.

### Latency heat map: 2D data is compelling
The `latency-l3` heat grid (bucket × time) is the most information-dense visualization —
it shows not just the current distribution but how it shifts over time as queue pressure
changes. High queue depth visibly shifts the color hot-zone upward into the 80-120ms buckets.

### Flow particle system: cap is important
Without capping at 64 particles, the queue zone becomes solid orange when pub >> con,
which obscures the individual message granularity. The cap keeps the visualization readable
while still conveying "queue is full".

### Dual-bar throughput chart: phase reveals lag
The interleaved pub/con bar chart makes it easy to spot when consume rate lags publish rate
(orange zone in queue gauge rises). The rate histories have different phases on their
sinusoids — this produces natural "queue filling" and "queue draining" episodes, making
all visualizations show interesting dynamics rather than steady state.
