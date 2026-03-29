# Research

Research projects carried out by AI tools.

Each directory here is a separate research project carried out by an LLM tool - usually Claude Code. Every single line of text and code was written by an LLM.

Times shown are in UTC.

### [Visualize Queue Flow](https://github.com/mamachanko/research/tree/main/visualize-queue-flow#readme) (2026-03-29 09:20)

An exploration of 22 terminal visualizations for a two-app RabbitMQ pipeline (publisher → queue → consumer), built with Bubble Tea and Lip Gloss. Eight conceptual areas (queue depth, message flow, status panels, throughput, event timeline, network topology, latency distribution, full dashboard) were each implemented at three complexity levels, from a single colored progress bar to a particle burst stream to a 2D latency heat grid, all recorded as animated GIFs with VHS.

Key findings:
- Queue depth is the anchor metric — color-ramped fill bars communicate urgency faster than numbers, and every other panel becomes more legible when paired with it.
- Particle flow (messages as moving dots slowing into a queue zone) is the most immediately intuitive visualization of async buffering; even the static L1 version conveys the topology at a glance.
- 2D heat grids (latency bucket × time) reveal temporal patterns that snapshot histograms hide — rising queue pressure visibly shifts the hot-zone upward into higher-latency buckets.
- Sinusoidal pub/consume rates on different periods create natural fill/drain cycles with no manual scripting, keeping every recording visually dynamic.
- Playwright-bundled ffmpeg lacks the GIF muxer; Pillow's alpha-composite + palette-quantize pipeline is a clean drop-in replacement for VHS's native GIF output.

### [Spinner Design Variations](https://github.com/mamachanko/research/tree/main/spinner-design-variations#readme) (2026-03-29 05:52)

An exploration of animated terminal spinner designs inspired by the [crush CLI](https://github.com/charmbracelet/crush) gradient spinner. Eight variations were built in Go using HCL color blending via `go-colorful`, covering themes from Matrix Rain and Fire to Ocean Wave and Neon Glitch, each recorded as a GIF with VHS.

Key findings:
- HCL color blending produces perceptually smoother gradients than RGB, especially across hue boundaries.
- Character set density (braille, katakana vs. ASCII punctuation) strongly influences the perceived "weight" and mood of a spinner independent of color.
- Staggered birth delays (`BirthDelay`) have an outsized effect on personality — long delays feel organic, short ones feel snappy.
- Pre-rendering all frames upfront keeps the animation loop allocation-free and avoids per-frame color math.
