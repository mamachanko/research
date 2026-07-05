# Notes — Industrial Micrographics Tool

## 2026-07-05

- Started investigation from the prompt: a Python tool that an agent can call to create industrial-style micrographic assets on demand, intended for embedding in rich HTML responses for software engineers.
- Chose a fresh research folder: `industrial-micrographics-tool`.
- Interpreted "micrographics" as small, embeddable SVG assets: status plates, service diagrams, pipelines, traces, deployment/callout cards, and data-flow sketches.
- Primary constraint: assets should be easy for an agent to generate without a browser, binary dependency, or full design stack.
- Decided to prototype a dependency-free Python CLI that writes deterministic SVG. SVG is the best target for HTML skills because it can be inlined, styled, cached, diffed, and copied as text.
- Implemented `industrial_micrographics.py`, a dependency-free Python CLI producing SVG for three initial kinds: `pipeline`, `status`, and `trace`.
- Generated three sample SVGs to validate the visual vocabulary and output paths.
- Design language used: dark grid background, hairline panels, monospace labels, high-contrast accent colors, small metadata stamps, and deterministic trace variation.
- Wrote final README report with goal, prototype details, integration pattern, findings, limitations, and next steps.
- Added root README entry using the UTC completion time.
