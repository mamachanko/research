# Research Notes: Pen Plotter Brush Techniques for Shinkawa-Style Art

## Research Process

1. Searched for variable line weight and expressive stroke techniques for pen plotters
2. Searched for hatching, stippling, and tonal techniques in generative plotter art
3. Searched for brush pen, ink wash, and painterly plotter projects
4. Searched for software tools: vsketch, vpype, axidraw ecosystem
5. Searched for algorithms: flow fields, Perlin noise, image-to-vector
6. Searched specifically for Yoji Shinkawa's process and style characteristics

## Key Findings

### Shinkawa's Actual Process
- Uses Pentel brush pens (XFL2V Sukiho) on A4 copy paper
- Scans and adjusts in Photoshop
- Uses correction fluid for white highlights
- Diluted ink for wash effects, not accurate shading but compositional impact
- Strong contrast, negative space, few straight lines, dynamic movement

### Most Promising Plotter Techniques for Shinkawa Style
- Z-axis pressure control (Calligraphy Z research paper) for variable stroke weight
- Pentel Color Brush Pen on plotter - speed controls opacity
- Rubber band hack for additional pen pressure on AxiDraw
- Multiple pen swapping: thick brush pen + fineliner for bold outlines vs detail
- Flow fields with Perlin noise for organic stroke paths
- vpype-vectrace for image-to-vector tracing
- StippleGen for tonal areas via weighted Voronoi stippling
- HalftonePAL for various dithering approaches

### Software Stack
- vpype: SVG optimization, pipeline processing, extensible via plugins
- vsketch: Python generative art environment, Processing-like API
- Inkscape + AxiDraw plugin: direct plotting, includes hatch fill
- canvas-sketch (JS), nannou (Rust), p5.js alternatives
- StippleGen, linedraw, HalftonePAL for image conversion
- Drawing Bot V3 for raster-to-plot conversion

### Physical Media Notes
- Pilot Parallel pen for calligraphic thick/thin variation
- Pentel Pocket Brush Pen responds to Z-axis height changes
- Princeton Select/Velvetouch flat shader brushes for watercolor plotting
- Paper variance matters hugely for ink bleed effects
- Fountain pens offer ink variety
