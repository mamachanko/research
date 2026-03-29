# Spinner Design Variations - Notes

## Objective
Explore the design space of terminal spinners inspired by crush CLI's animated gradient spinner.

## Crush Spinner Architecture (from provided context)
- Random chars from `"0123456789abcdefABCDEF~!@#$£€%^&*()+=_"`
- Gradient blended in HCL color space via `go-colorful`
- Pre-rendered frames (static: 10 frames, cycling: width*2 frames)
- Staggered character "birth" with random delays up to 1s
- 20 FPS via Bubble Tea tick
- Configurable: size, label, colors, gradient direction, cycle mode

## Spinner Variations Planned

1. **Crush Classic** - Faithful recreation with hex+symbol chars, magenta→cyan gradient
2. **Matrix Rain** - Binary/hex chars, pure green, fast flicker
3. **Fire** - Flame chars (▲△◆◇░▒▓), red→yellow gradient, slow pulse
4. **Ocean Wave** - Wave chars (≋~≈∿), deep blue→teal, smooth cycle
5. **Retro Braille** - Braille dot patterns, amber/orange mono, dense feel
6. **Neon Glitch** - Katakana/symbols, pink→purple→blue, fast chaotic
7. **Rainbow Minimal** - Simple ASCII punctuation, full ROYGBIV rainbow
8. **Snow/Ice** - Snowflake Unicode (❄✦✧✶❋), white→light-blue gradient

## Tools Used
- Go 1.24 with Bubble Tea + Lipgloss
- VHS v0.9.0 for terminal recording → GIF
- `go-colorful` for HCL gradient blending

## Log

### 2026-03-29
- Checked available tooling: VHS v0.9.0 installed, Go 1.24
- Created Go module `spinners`
- Implemented all 8 variations as flags on a single binary (`./spinners -n N`)
- Each variation runs for 5 seconds then exits (for VHS recording)

## Hurdles / Lessons Learned

### VHS 0.9.0 path parsing quirk
VHS 0.9.0 lexer treats `/` as a regex delimiter, so absolute paths in `Output` commands
fail: `Output /home/user/...` is tokenized as REGEX `/home/` + identifier `user` + etc.
Fix: Use simple relative filenames like `Output spinner1.gif` and run VHS from the output dir.
Also: filenames starting with a digit followed by `-` (like `1-crush-classic.gif`) also fail because
the lexer tokenizes `1` as a number and `-` as MINUS. Used underscores: `Output spinner1.gif`.

### Chromium for VHS
VHS uses go-rod (headless browser) to render the terminal. No system Chromium was available.
The Playwright Chromium at `/root/.cache/ms-playwright/chromium-1194/chrome-linux/chrome` worked
after symlinking it to `/usr/bin/chromium` (which is on go-rod's LookPath list).
Required `VHS_NO_SANDBOX=true` env var to run as root without sandbox.

### PNG output is a frames directory
VHS's `Output foo.png` creates a *directory* `foo.png/` containing `frame-text-NNNNN.png` and
`frame-cursor-NNNNN.png` for each recorded frame. Picked frame 40 as a representative mid-animation snapshot.

### HCL vs RGB blending
go-colorful's `BlendHcl` avoids the muddy gray midpoint that RGB interpolation produces across
complementary hues. Most visible in the rainbow and ocean gradients.

### Birth delay personality
BirthDelay has outsized impact on "feel": 0.3s = snappy/immediate, 1.0s = organic/deliberate.
Snow and braille use 1.0s; matrix uses 0.3s. Matches each theme's character.

### Character width (CJK)
Katakana characters (neon-glitch variation) are full-width (2 terminal columns each).
With width=10, the spinner actually spans 20 columns. VHS terminal width of 800px accommodated this fine.

## Final Outputs
- 8 GIFs in output/ (146KB - 407KB each)
- 8 preview PNGs in output/ (3-6KB each, extracted from VHS frame directories)
- 8 .tape files in tapes/
