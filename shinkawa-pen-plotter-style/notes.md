# Research Notes: Translating Yoji Shinkawa's Style to Pen Plotters

## Research Process

### Phase 1: Analyzing Shinkawa's Style

Searched for detailed breakdowns of Yoji Shinkawa's artistic techniques, materials, and process.

**Key findings about his style:**

- Primary tool: Pentel Brush Pen (XFL2V Sukiho type), felt-tip pens
- Works on Muji sketchbooks, sometimes regular copy paper
- Strong sumi-e (Japanese ink painting) influence
- Process: rapid brush pen sketches -> scanned -> Photoshop finishing
- Cited influences: Yoshitaka Amano, Hayao Miyazaki, Yoshikazu Yasuhiko, Frank Miller, Moebius, Aubrey Beardsley
- Hideo Kojima notes Shinkawa is "extremely fast at drawing silhouettes" — grasps the essence first, then adds detail

**Visual characteristics identified:**
1. **Bold, gestural brush strokes** — few straight lines, everything flows
2. **Extreme contrast** — deep blacks against white, very little mid-tone
3. **Expressive negative space** — white space is compositionally active
4. **Variable line weight** — from hair-thin to broad swaths in a single stroke
5. **Dry brush texture** — bristle separation visible in fast strokes
6. **Ink dilution** — occasional grey washes for atmospheric depth
7. **White correction fluid** — used as a positive mark-making tool on top of ink
8. **Gravity and weight** — characters feel grounded, never floating
9. **Speed marks and splatter** — kinetic energy captured in ink artifacts
10. **Mechanical precision amid organic chaos** — detailed mechanical parts (guns, armor) rendered precisely while organic forms (cloth, hair, muscle) are loose and gestural

### Phase 2: Pen Plotter Capabilities and Constraints

Researched what pen plotters can and cannot do, with focus on expressive output.

**Fundamental constraints:**
- Plotters draw continuous vector paths — no filled regions natively
- Line weight is determined by the physical pen, not software
- No true pressure sensitivity (though Z-axis height can approximate it)
- Speed is limited — complex drawings can take hours
- Pen must be raised/lowered (binary contact vs. no contact)

**Techniques for expressiveness:**
1. **Brush pens on plotters** — Pentel brush pens, Pilot Parallel pens work in plotter arms
2. **Z-axis manipulation** — varying pen height changes brush contact area
3. **Speed variation** — some plotters allow variable speed; slow = more ink deposit
4. **Multiple passes** — drawing the same area multiple times builds up ink
5. **Hatching density** — closer lines = darker tone
6. **Paper manipulation** — pre-wetting, warping, or texturing the paper
7. **Manual ink loading** — artist LIA manually loads brush with varying ink amounts during plotting
8. **Pen swapping** — different pens for different line weights in layers

**Software/tools discovered:**
- vpype: Swiss-Army-knife CLI for plotter vector graphics (Python, MIT license)
- vsketch: Generative plotter art environment built on vpype (Processing-like API)
- AxiDraw: Popular plotter hardware with Inkscape plugin
- Processing / p5.js: Common for generative art, outputs SVG
- OpenCV: Edge detection, contour extraction
- Potrace: Bitmap-to-vector tracing (but no hatching support)
- plotterfun (mitxela.com): Interactive tool with multiple rendering algorithms

**Relevant algorithms:**
- Perlin noise flow fields (Tyler Hobbs' seminal essay)
- Agent-based hatching (John Proudlock's "goat" algorithm)
- Lattice Boltzmann ink diffusion simulation
- Canny edge detection -> contour tracing
- TSP (Traveling Salesman) art for continuous-line portraits
- Spiral/squiggle density modulation

### Phase 3: Bridging the Gap — Style Translation Strategies

After analyzing both domains, I identified these translation strategies:

**Strategy 1: Bold Contour Extraction**
- Use Canny edge detection or neural edge detection on reference images
- Filter for only the strongest edges (high threshold) to match Shinkawa's economy of line
- Convert edges to vector paths with variable-width strokes
- Plot with a fat brush pen for bold outlines

**Strategy 2: Flow Field Gesture Simulation**
- Use Perlin noise flow fields to generate gestural stroke paths
- Constrain flow fields to follow image contours (contour-driven flow)
- Vary line density based on image darkness
- Produces organic, sweeping marks reminiscent of brush strokes

**Strategy 3: Multi-Layer Approach**
- Layer 1: Bold silhouette outlines (thick brush pen)
- Layer 2: Internal detail lines (fine pen)
- Layer 3: Tonal hatching (medium pen, dense in shadows)
- Layer 4: Gestural speed marks (brush pen, fast/light)
- Swap pens between layers

**Strategy 4: Dry Brush Simulation**
- Draw multiple thin parallel lines with slight random offset
- Leave gaps to simulate bristle separation
- Taper the line bundle toward stroke ends
- Perlin noise displacement along the stroke path

**Strategy 5: Ink Splatter and Speed Marks**
- Generate random dot clusters near stroke endpoints
- Use stippling in radial patterns from stroke termination points
- Small flicked lines radiating from fast-moving stroke areas

**Strategy 6: Negative Space Composition**
- Start from a reference image, identify the darkest 20-30% of areas
- Only draw in those areas — leave the rest as white space
- This naturally creates Shinkawa's high-contrast look

### Phase 4: Academic Research on Sumi-e Simulation

Found significant academic work on digital sumi-e rendering:

- **Strassmann's bristle model**: 1D array of bristles, each with grey value, spread under pressure, ink transfer between neighbors
- **Lattice Boltzmann ink diffusion**: Models ink spread on paper using fluid dynamics
- **Particle-based methods (Shi et al.)**: Pseudo-Brownian motion for ink particle flow
- **Reinforcement learning brush agent (arXiv 1206.4634)**: RL agent learns optimal brush trajectories
- **Contour-driven sumi-e rendering**: Maps real brush footprint textures along computed trajectories
- **B-spline brush modeling**: Variable offset approximation of cubic B-splines for stroke shape

Most of these produce raster output, but the stroke trajectory computation is directly applicable to generating plotter paths.

## Key Insight

The most promising approach combines:
1. **Image analysis** (edge detection, tonal mapping) to determine *where* to draw
2. **Flow fields** to determine *how* strokes move (direction, curvature)
3. **Stroke modeling** to determine *what* each stroke looks like (width variation, taper, bristle separation)
4. **Physical pen choice** to add real ink texture the algorithm can't fully control

The happy accident of pen plotter work — ink flow variation, paper texture interaction, slight mechanical imprecision — actually helps achieve the organic quality of Shinkawa's work. The plotter's imperfections become features.
