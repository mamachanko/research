# Translating Yoji Shinkawa's Style to Pen Plotters

An investigation into how the distinctive ink brush aesthetic of Yoji Shinkawa (character designer for Metal Gear Solid, Death Stranding) could be reproduced or approximated using pen plotter hardware and generative algorithms.

## Background

Yoji Shinkawa is renowned for his sumi-e-influenced character illustrations, created primarily with Pentel brush pens on paper. His style is defined by bold gestural strokes, extreme contrast, expressive negative space, and a dynamic tension between organic looseness and mechanical precision. Hideo Kojima has described how Shinkawa works by "grasping the essence of the image with his brush" before adding detail — silhouette first, refinement second.

The challenge: pen plotters are fundamentally different tools. They draw continuous vector paths with uniform-width pens, have no native pressure sensitivity, and operate through precise mechanical motion rather than gestural human impulse. Translating Shinkawa's style requires decomposing it into reproducible operations while preserving its essential character.

## Shinkawa's Style: Key Visual Elements

| Element | Description | Plotter Challenge |
|---------|-------------|-------------------|
| **Bold contour strokes** | Thick, confident outlines that define silhouettes in 1-3 strokes | Requires thick brush pen in plotter arm; path planning for minimal stroke count |
| **Variable line weight** | Single strokes taper from broad to hair-thin | Z-axis height variation, or simulate with multiple parallel paths |
| **Dry brush texture** | Visible bristle separation in fast strokes | Parallel displaced lines with random gaps |
| **Ink wash / dilution** | Grey tonal areas for atmosphere and depth | Hatching density variation; diluted ink in plotter pen |
| **White correction fluid** | Used as positive mark-making on top of black | Second pass with white pen (gel pen, Posca marker) |
| **Ink splatter** | Energetic dots and flicks near stroke endpoints | Stipple patterns; tiny random marks near path terminations |
| **Negative space** | White areas are compositionally active; only 20-30% of the surface carries ink | Threshold-based: only draw in the darkest regions |
| **Weight and gravity** | Characters feel grounded, never floating | Compositional concern — ensure bottom-heavy density |
| **Mechanical precision** | Guns, armor, mechanical parts rendered with clean detail amid organic chaos | Separate layer with technical pen; precise geometric paths |

## Translation Strategies

### Strategy 1: Multi-Layer Decomposition

The most practical approach decomposes a Shinkawa-style composition into separate plotter layers, each plotted with a different pen:

1. **Layer 4 — Tonal hatching** (fine 0.1-0.3mm technical pen): Diagonal parallel lines with spacing inversely proportional to image darkness. Provides underlying tonal structure.
2. **Layer 2 — Flow field strokes** (medium 0.5-0.8mm felt-tip): Perlin noise-guided gestural strokes concentrated in dark areas. Provides the organic, sweeping energy.
3. **Layer 3 — Dry brush texture** (fine pen or worn brush pen): Parallel displaced strokes with random gaps simulating bristle separation. Adds Shinkawa's characteristic rough texture.
4. **Layer 1 — Bold contours** (thick brush pen, 1.0mm+): Edge-detected contour paths with variable width. The defining silhouette lines.
5. **Layer 5 — Splatter accents** (brush pen, light contact): Random dots and short flicks near dark areas. Adds kinetic energy.

Plotting in this order (background to foreground) lets each layer build on the previous, mimicking how Shinkawa layers ink washes, then strokes, then correction fluid.

### Strategy 2: Flow Field Gesture Simulation

[Perlin noise flow fields](https://www.tylerxhobbs.com/words/flow-fields) can approximate Shinkawa's gestural quality. The algorithm:

1. Load a reference image (or generate a silhouette programmatically)
2. Build a 2D flow field using Perlin noise for angle at each grid point
3. Optionally bias the flow field to follow image gradients (contour-driven flow)
4. Seed particles in dark areas; trace each along the flow field to generate stroke paths
5. Stroke length, width, and density all scale with local image darkness

This produces sweeping, curved paths that naturally concentrate in shadow areas and thin out in highlights — capturing Shinkawa's economy of mark-making.

### Strategy 3: Dry Brush Stroke Modeling

A single "brush stroke" path is expanded into N parallel sub-paths (bristles), each with:
- Slight perpendicular offset from the center line
- Random gaps where the bristle lifts (simulating dry brush / ink depletion)
- Taper at start and end (spread narrows to a point)
- Perlin noise displacement for organic irregularity

With 3-7 bristle paths per stroke and a gap probability of ~10-15%, the result reads as a single rough brush stroke when plotted with a fine pen.

### Strategy 4: Contour-Driven Composition

Rather than hatching the entire image, extract only the strongest edges (high Canny/Sobel threshold) and draw only those. This naturally produces Shinkawa's high-contrast, negative-space-heavy look — most of the paper stays white, with bold dark strokes defining only essential forms.

### Strategy 5: Physical Pen and Paper Choices

The plotter's physical media contributes enormously:

- **Pentel Brush Pen (XFL2V)** — Shinkawa's own tool. Works in AxiDraw with a 3D-printed adapter. Natural variation from brush flex.
- **Pilot Parallel Pen** — Broad/thin strokes depending on nib angle. Rotation during plotting creates Shinkawa-like weight variation.
- **Faber-Castell Pitt Artist Pen (brush tip)** — Soft tip deforms under plotter pressure for natural variation.
- **Pre-wetted paper** — Ink bleeds and feathers, approximating ink wash diffusion.
- **Heavyweight paper (200+ gsm)** — Better ink absorption, holds up to multiple passes.
- **Manual ink loading** — As artist [LIA](https://www.liaworks.com/theprojects/mechanical-interventions-an-investigation-into-generative-plotter-painting/) demonstrates, manually loading varying amounts of ink onto a plotter-held brush during plotting creates natural saturation variation.

## Proof of Concept

The included Python script `shinkawa_plotter.py` demonstrates all five techniques. It:

- Accepts any input image (or generates a demo silhouette)
- Produces 5 separate layer SVGs + 1 combined SVG
- All output is plotter-ready (paths only, no fills, proper mm dimensions)

### Usage

```bash
pip install numpy Pillow scipy

# With your own image:
python shinkawa_plotter.py input_photo.png

# Demo mode (synthetic figure):
python shinkawa_plotter.py
```

### Output Files

| File | Layer | Recommended Pen |
|------|-------|----------------|
| `*_layer1_contours.svg` | Bold silhouette contours | Pentel Brush Pen or thick felt-tip (1.0mm+) |
| `*_layer2_flow.svg` | Flow field gestural strokes | Medium felt-tip (0.5-0.8mm) |
| `*_layer3_drybrush.svg` | Dry brush texture simulation | Worn brush pen or fine felt-tip (0.3mm) |
| `*_layer4_hatching.svg` | Tonal hatching for shadows | Fine technical pen (0.1-0.3mm) |
| `*_layer5_splatter.svg` | Ink splatter accents | Brush pen with light Z-axis contact |
| `*_combined.svg` | All layers merged | Single pen (loses multi-pen benefit) |

## Software Ecosystem

Relevant tools for further development:

| Tool | Purpose | Link |
|------|---------|------|
| **vpype** | SVG optimization, layer management, path reordering for plotters | [github.com/abey79/vpype](https://github.com/abey79/vpype) |
| **vsketch** | Generative plotter art framework (Processing-like API, built on vpype) | [github.com/abey79/vsketch](https://github.com/abey79/vsketch) |
| **vpype-flow-imager** | Flow-field image vectorization plugin for vpype | vpype plugin ecosystem |
| **hatched** | Half-toning with hatching lines for vpype | vpype plugin ecosystem |
| **occult** | Hidden line removal for 3D plotter art | vpype plugin ecosystem |
| **AxiDraw** | Popular pen plotter hardware with Inkscape plugin | [axidraw.com](https://axidraw.com) |
| **Processing / p5.js** | Generative art environments with SVG export | [processing.org](https://processing.org) |
| **plotterfun** | Interactive browser tool for image-to-plotter conversion | [mitxela.com/projects/plotting](https://mitxela.com/projects/plotting) |

## Academic References

Relevant research on digital sumi-e simulation:

- **Strassmann (1986)** — Pioneered bristle-based brush simulation with ink transfer between neighboring bristles
- **Lattice Boltzmann ink diffusion** — Models ink spread on paper as fluid dynamics, applicable to understanding how ink will behave with plotter brush pens
- **Contour-driven sumi-e rendering** ([ScienceDirect](https://www.sciencedirect.com/science/article/abs/pii/S009784931000186X)) — Maps scanned brush footprint textures along computed stroke trajectories
- **Artist Agent (arXiv 1206.4634)** — Reinforcement learning approach to automatic stroke generation in oriental ink painting
- **Learning Hatching for Pen-and-Ink Illustration** ([ACM](https://dl.acm.org/doi/10.1145/2077341.2077342)) — Algorithm for learning hatching styles from artist examples

## Key Findings

1. **Multi-layer decomposition is the most practical approach.** Separating contours, tonal fill, texture, and accents into distinct plotter layers — each with its own pen — comes closest to Shinkawa's layered ink technique (wash → stroke → correction fluid).

2. **Physical pen choice matters more than algorithm sophistication.** Putting an actual Pentel brush pen in the plotter arm, using heavyweight paper, and varying Z-axis height produces more authentic results than any amount of algorithmic stroke simulation with a technical pen.

3. **Flow fields naturally capture gestural quality.** Perlin noise flow fields produce sweeping, curved strokes that read as intentional and gestural rather than mechanical — especially when seeded preferentially in dark image regions.

4. **Embrace plotter imperfection.** Ink flow variation, paper texture interaction, brush pen flex, and slight mechanical wobble are *features*, not bugs. They add the organic quality that makes Shinkawa's work feel alive. Over-optimizing paths removes this.

5. **Less is more — match Shinkawa's economy.** The strongest results come from high-contrast compositions where most of the paper stays white. Drawing only in the darkest 20-30% of the image, with bold confident strokes, captures his aesthetic better than comprehensive detail.

## Sources

- [Cook and Becker — Artist Spotlight: Yoji Shinkawa](https://www.cookandbecker.com/en/article/322/artist-spotlight-yoji-shinkawa.html)
- [Sabukaru — Yoji Shinkawa: The Art Director of Metal Gear Solid](https://sabukaru.online/articles/yoji-shinkawa-the-art-director-of-metal-gear-solid)
- [Shmuplations — Yoji Shinkawa 2001 Developer Interview](https://shmuplations.com/yojishinkawa/)
- [Brush Pens and Yoji Shinkawa — Applied Media Aesthetics](https://michaelgibbsama.wordpress.com/2013/10/18/brush-pens-and-yoji-shinkawa/)
- [Tyler Hobbs — Flow Fields](https://www.tylerxhobbs.com/words/flow-fields)
- [Matt DesLauriers — Pen Plotter Art & Algorithms](https://mattdesl.svbtle.com/pen-plotter-1)
- [LIA — Mechanical Interventions](https://www.liaworks.com/theprojects/mechanical-interventions-an-investigation-into-generative-plotter-painting/)
- [Pen Plotter Artwork Blog — Multiple Line Widths](https://penplotterartwork.com/blog/2021/10/28/multiple-line-widths-example-pen-plot-art/)
- [Cameron Sun — Generative Art: Pen Plotting an Old Family Photo](https://www.csun.io/2021/12/29/plotting-old-pictures.html)
- [Sighack — Getting Creative with Perlin Noise Fields](https://sighack.com/post/getting-creative-with-perlin-noise-fields)
- [awesome-plotters (GitHub)](https://github.com/beardicus/awesome-plotters)
