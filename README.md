# Research

Research projects carried out by AI tools.

Each directory here is a separate research project carried out by an LLM tool - usually Claude Code. Every single line of text and code was written by an LLM.

Times shown are in UTC.

### [Spinner Design Variations](https://github.com/mamachanko/research/tree/main/spinner-design-variations#readme) (2026-03-29 05:52)

An exploration of animated terminal spinner designs inspired by the [crush CLI](https://github.com/charmbracelet/crush) gradient spinner. Eight variations were built in Go using HCL color blending via `go-colorful`, covering themes from Matrix Rain and Fire to Ocean Wave and Neon Glitch, each recorded as a GIF with VHS.

Key findings:
- HCL color blending produces perceptually smoother gradients than RGB, especially across hue boundaries.
- Character set density (braille, katakana vs. ASCII punctuation) strongly influences the perceived "weight" and mood of a spinner independent of color.
- Staggered birth delays (`BirthDelay`) have an outsized effect on personality — long delays feel organic, short ones feel snappy.
- Pre-rendering all frames upfront keeps the animation loop allocation-free and avoids per-frame color math.
