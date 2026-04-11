# Research

Research projects carried out by AI tools.

Each directory here is a separate research project carried out by an LLM tool - usually Claude Code. Every single line of text and code was written by an LLM.

Times shown are in UTC.

### [Pi VPS-Only Local Access](https://github.com/mamachanko/research/tree/main/pi-vps-only-local-access#readme) (2026-04-05 15:20)

Investigated the strict deployment constraint that Pi (the coding agent) cannot be installed locally and must run only on a VPS. Produced a VPS-only architecture and setup guide using SSH local port-forwarding for browser access from local machines.

Key findings:
- The local machine only needs a browser and SSH client; Pi runtime stays exclusively on VPS
- Binding Pi and web services to `127.0.0.1` on VPS plus SSH `-L` forwarding is the safest default
- Public HTTPS access should be optional and gated by an auth layer on the VPS reverse proxy

### [Pi Coding Agent Web Client on VPS](https://github.com/mamachanko/research/tree/main/pi-coding-agent-web-client-vps#readme) (2026-04-05 15:09)

Investigated a corrected scope: running Pi (the coding agent) on a VPS while using it from a local browser through SSH local port-forwarding. Produced a loopback-first deployment pattern with compose/systemd examples and optional TLS reverse-proxying.

Key findings:
- For single-user operation, SSH `-L` forwarding to VPS loopback is the cleanest and safest default
- Keeping both web UI and agent API bound to `127.0.0.1` materially reduces exposure risk
- A reverse proxy with TLS/auth should be treated as an optional collaboration layer, not the base requirement

### [Pi Web Client via VPS Port-Forwarding](https://github.com/mamachanko/research/tree/main/pi-web-client-vps-port-forward#readme) (2026-04-05 15:05)

Investigated how to run a browser-accessed web client on a VPS while securely reaching a Raspberry Pi on a private/home network via reverse tunneling. Produced a concrete deployment pattern using autossh reverse port-forwarding, VPS reverse proxying, and hardening guidance.

Key findings:
- A reverse SSH tunnel initiated by the Pi avoids opening inbound home router ports
- Keeping the forwarded endpoint bound to `127.0.0.1` on the VPS reduces exposure and pairs cleanly with HTTPS reverse proxy routes
- `autossh` with systemd provides a practical, resilient baseline for persistent connectivity

### [CF Self-Redeploy](https://github.com/mamachanko/research/tree/main/cf-self-redeploy#readme) (2026-04-05 15:30)

An investigation into whether a Spring Boot application deployed to Cloud Foundry can programmatically redeploy itself (or a clone) with different properties, profiles, and service bindings using the CF V3 API. Includes proof-of-concept Java code using the cf-java-client library.

Key findings:
- Yes, a CF app can deploy clones of itself via the CF V3 API using the cf-java-client library
- The most efficient approach is copying the app's compiled droplet (`POST /v3/droplets?source_guid=...`), which avoids re-uploading the JAR and re-staging entirely
- The running app can discover its own identity (app GUID, space ID, CF API endpoint) from the `VCAP_APPLICATION` environment variable injected by CF
- Different Spring profiles, env vars, service bindings, and routes can be set on each clone independently
- Fork-bomb prevention (recursive self-cloning) must be explicitly guarded against

### [Shinkawa Pen Plotter Style](https://github.com/mamachanko/research/tree/main/shinkawa-pen-plotter-style#readme) (2026-04-02 12:45)

An investigation into translating Yoji Shinkawa's ink brush aesthetic (Metal Gear Solid, Death Stranding) to pen plotter output. Analyzed his visual style — bold sumi-e-influenced contours, variable line weight, dry brush texture, ink splatter — and developed algorithmic strategies to approximate each element using plotter hardware. Includes a proof-of-concept Python script that generates multi-layer plotter-ready SVGs.

Key findings:
- Multi-layer decomposition (contours, flow strokes, dry brush, hatching, splatter) with per-layer pen swaps is the most practical approach
- Physical pen choice (actual Pentel brush pens in the plotter arm) matters more than algorithm sophistication
- Perlin noise flow fields seeded in dark image regions naturally produce gestural, Shinkawa-like stroke quality
- Plotter imperfections (ink flow variation, brush flex, paper texture) are features that add organic quality
- High-contrast "less is more" compositions (drawing only the darkest 20-30%) best capture Shinkawa's economy of mark-making

### [3D Pen Plotter — Continuous Stroke](https://github.com/mamachanko/research/tree/main/3d-pen-plotter-continuous#readme) (2026-03-31 11:20)

An exploration of drawing three-dimensional forms with a single continuous swirling stroke — not wireframes, but flowing parametric curves that wrap around 3D surfaces. When projected with perspective, the line's natural density variation (bunching at edges, spreading in the middle) makes the brain perceive spheres, tori, vases, and knots. Twelve shapes rendered as SVGs with depth-varying line weight and opacity.

Key findings:
- A single spiral on a sphere's surface, projected with perspective, reads unmistakably as a 3D sphere from line density alone
- Line weight variation (14:1 thick-to-thin ratio) combined with opacity is the strongest depth cue — no hidden-line removal needed
- Three different sphere spirals (pole-to-pole, loxodrome, Fibonacci) each produce distinct artistic character while all clearly reading as "sphere"
- The technique extends to any parametric surface: tori, cones, vases, Mobius strips, Klein bottles, and knots

### [Isoline 3D Visualization](https://github.com/mamachanko/research/tree/main/isoline-3d-visualization#readme) (2026-03-30 04:10)

An exploration of eight distinct approaches to visualizing 3D objects using only isolines: horizontal slicing, Joy Division ridgelines, rotating parametric curves, depth-layered contours, cross-hatched multi-axis contours, radial coordinate isolines, parametric surface grids, and animated morphing contours. Each technique is demonstrated on multiple mathematical surfaces with Python/matplotlib.

Key findings:
- Occlusion (Joy Division style) creates the strongest immediate 3D perception from pure line work
- Cross-hatching multiple contour sets from different axes mimics classical engraving and reveals curvature direction
- Depth-modulated line weight (thicker = closer) is the simplest enhancement to add depth to any contour plot
- Parametric iso-curves are the only viable approach for non-orientable surfaces like Klein bottles and Mobius strips

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
