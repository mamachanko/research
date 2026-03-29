package main

import (
	"flag"
	"fmt"
	"math"
	"math/rand"
	"os"
	"strings"
	"time"

	colorful "github.com/lucasb-eyer/go-colorful"
)

// ANSI helpers
func rgb(r, g, b uint8) string {
	return fmt.Sprintf("\x1b[38;2;%d;%d;%dm", r, g, b)
}
func reset() string { return "\x1b[0m" }
func clearLine() string { return "\r\x1b[2K" }
func hideCursor() string { return "\x1b[?25l" }
func showCursor() string { return "\x1b[?25h" }

// makeGradient returns n colors blended between stops in HCL space.
func makeGradient(stops []colorful.Color, n int) []colorful.Color {
	if n <= 0 {
		return nil
	}
	if len(stops) < 2 {
		out := make([]colorful.Color, n)
		for i := range out {
			out[i] = stops[0]
		}
		return out
	}
	out := make([]colorful.Color, n)
	segments := len(stops) - 1
	for i := 0; i < n; i++ {
		t := float64(i) / float64(n-1)
		seg := int(t * float64(segments))
		if seg >= segments {
			seg = segments - 1
		}
		localT := t*float64(segments) - float64(seg)
		out[i] = stops[seg].BlendHcl(stops[seg+1], localT).Clamped()
	}
	return out
}

// SpinnerConfig defines a spinner variation.
type SpinnerConfig struct {
	Name        string
	Label       string
	Chars       []rune
	ColorStops  []colorful.Color
	Width       int
	FPS         int
	Duration    time.Duration
	CycleColors bool
	BirthDelay  float64 // max random birth delay in seconds
	BGColor     *[3]uint8
}

var hex = func(h string) colorful.Color {
	c, _ := colorful.Hex(h)
	return c
}

var variations = []SpinnerConfig{
	{
		// 1. Crush Classic — faithful recreation
		Name:        "crush-classic",
		Label:       "Thinking",
		Chars:       []rune("0123456789abcdefABCDEF~!@#$£€%^&*()+=_"),
		ColorStops:  []colorful.Color{hex("#FF00CC"), hex("#00FFFF")},
		Width:       12,
		FPS:         20,
		Duration:    5 * time.Second,
		CycleColors: true,
		BirthDelay:  0.8,
	},
	{
		// 2. Matrix Rain — green binary cascade
		Name:        "matrix-rain",
		Label:       "Decrypting",
		Chars:       []rune("01001101010110100110100110101100111001010"),
		ColorStops:  []colorful.Color{hex("#003300"), hex("#00FF41"), hex("#AAFFCC")},
		Width:       14,
		FPS:         25,
		Duration:    5 * time.Second,
		CycleColors: false,
		BirthDelay:  0.3,
	},
	{
		// 3. Fire — warm flame characters
		Name:        "fire",
		Label:       "Burning",
		Chars:       []rune("▲△◆◇░▒▓▴▵∧"),
		ColorStops:  []colorful.Color{hex("#8B0000"), hex("#FF4500"), hex("#FF8C00"), hex("#FFD700")},
		Width:       10,
		FPS:         15,
		Duration:    5 * time.Second,
		CycleColors: true,
		BirthDelay:  0.5,
	},
	{
		// 4. Ocean Wave — flowing water characters
		Name:        "ocean-wave",
		Label:       "Loading",
		Chars:       []rune("≋~≈∿⌇〜∼"),
		ColorStops:  []colorful.Color{hex("#003366"), hex("#0066CC"), hex("#00CED1"), hex("#7FFFD4")},
		Width:       12,
		FPS:         18,
		Duration:    5 * time.Second,
		CycleColors: true,
		BirthDelay:  0.6,
	},
	{
		// 5. Retro Braille — dense amber braille patterns
		Name:        "retro-braille",
		Label:       "Processing",
		Chars:       []rune("⣾⣽⣻⢿⡿⣟⣯⣷⠿⣀⣤⣶⣿⡀⠁⠂⠄⠸⠰⠠"),
		ColorStops:  []colorful.Color{hex("#5C2E00"), hex("#FF8C00"), hex("#FFD700")},
		Width:       16,
		FPS:         20,
		Duration:    5 * time.Second,
		CycleColors: false,
		BirthDelay:  0.9,
	},
	{
		// 6. Neon Glitch — katakana with electric colors
		Name:        "neon-glitch",
		Label:       "Glitching",
		Chars:       []rune("アイウエオカキクケコサシスセソタチツテトナニヌネノ"),
		ColorStops:  []colorful.Color{hex("#FF00FF"), hex("#7B2FBE"), hex("#00F0FF")},
		Width:       10,
		FPS:         22,
		Duration:    5 * time.Second,
		CycleColors: true,
		BirthDelay:  0.4,
	},
	{
		// 7. Rainbow Minimal — simple ASCII, full spectrum
		Name:        "rainbow-minimal",
		Label:       "Running",
		Chars:       []rune(".+*!|/-\\:;=?><[]{}()"),
		ColorStops:  []colorful.Color{hex("#FF0000"), hex("#FF7F00"), hex("#FFFF00"), hex("#00FF00"), hex("#0000FF"), hex("#8B00FF")},
		Width:       14,
		FPS:         20,
		Duration:    5 * time.Second,
		CycleColors: true,
		BirthDelay:  0.5,
	},
	{
		// 8. Snow/Ice — snowflake Unicode, frosty palette
		Name:        "snow-ice",
		Label:       "Freezing",
		Chars:       []rune("❄✦✧✶❋*·°•"),
		ColorStops:  []colorful.Color{hex("#003366"), hex("#4488BB"), hex("#87CEEB"), hex("#FFFFFF")},
		Width:       12,
		FPS:         16,
		Duration:    5 * time.Second,
		CycleColors: true,
		BirthDelay:  1.0,
	},
}

// frame holds one pre-rendered string per position.
type frame struct {
	cells []string // colored char strings
}

// preRender computes all animation frames for a config.
func preRender(cfg SpinnerConfig, totalFrames int) []frame {
	rng := rand.New(rand.NewSource(42))
	chars := cfg.Chars
	width := cfg.Width
	gradient := makeGradient(cfg.ColorStops, width)

	// birth offsets: each position has a random delay before it appears
	birthFrames := make([]int, width)
	for i := range birthFrames {
		birthFrames[i] = int(cfg.BirthDelay * float64(cfg.FPS) * rng.Float64())
	}

	// random chars per position per frame (pre-chosen)
	charGrid := make([][]rune, totalFrames)
	for f := 0; f < totalFrames; f++ {
		charGrid[f] = make([]rune, width)
		for i := 0; i < width; i++ {
			charGrid[f][i] = chars[rng.Intn(len(chars))]
		}
	}

	frames := make([]frame, totalFrames)
	for f := 0; f < totalFrames; f++ {
		cells := make([]string, width)
		for i := 0; i < width; i++ {
			// color cycling: offset gradient position per frame
			colorIdx := i
			if cfg.CycleColors {
				offset := (f * width / totalFrames)
				colorIdx = (i + offset) % width
			}
			c := gradient[colorIdx]
			r8, g8, b8 := toRGB8(c)
			col := rgb(r8, g8, b8)

			if f < birthFrames[i] {
				cells[i] = col + "." + reset()
			} else {
				cells[i] = col + string(charGrid[f][i]) + reset()
			}
		}
		frames[f] = frame{cells: cells}
	}
	return frames
}

// toRGB8 converts colorful.Color to 8-bit RGB components.
func toRGB8(c colorful.Color) (uint8, uint8, uint8) {
	r := uint8(math.Round(c.R * 255))
	g := uint8(math.Round(c.G * 255))
	b := uint8(math.Round(c.B * 255))
	return r, g, b
}

// ellipsisCycle cycles through "", ".", "..", "..."
func ellipsisAt(step, fps int) string {
	period := fps / 3
	if period < 1 {
		period = 1
	}
	n := (step / period) % 4
	return strings.Repeat(".", n)
}

func runSpinner(cfg SpinnerConfig) {
	fps := cfg.FPS
	totalFrames := fps * 2 // cycle length (2 seconds per cycle)
	if cfg.CycleColors {
		totalFrames = cfg.Width * 2
	}
	frames := preRender(cfg, totalFrames)

	// label styling
	labelStyle := "\x1b[1m" // bold
	dimStyle := "\x1b[2m"   // dim for ellipsis

	fmt.Print(hideCursor())
	defer fmt.Print(showCursor())

	ticker := time.NewTicker(time.Second / time.Duration(fps))
	defer ticker.Stop()

	start := time.Now()
	step := 0

	for range ticker.C {
		if time.Since(start) >= cfg.Duration {
			break
		}
		f := frames[step%len(frames)]

		var sb strings.Builder
		sb.WriteString(clearLine())
		for _, cell := range f.cells {
			sb.WriteString(cell)
		}
		sb.WriteString("  ")
		sb.WriteString(labelStyle)
		sb.WriteString(cfg.Label)
		sb.WriteString(reset())
		sb.WriteString(dimStyle)
		sb.WriteString(ellipsisAt(step, fps))
		sb.WriteString(reset())

		fmt.Print(sb.String())
		step++
	}

	// Final clear
	fmt.Print(clearLine())
	fmt.Println()
}

func main() {
	num := flag.Int("n", 1, "Spinner variation (1-8)")
	list := flag.Bool("list", false, "List all variations")
	flag.Parse()

	if *list {
		for i, v := range variations {
			fmt.Printf("%d. %s\n", i+1, v.Name)
		}
		return
	}

	idx := *num - 1
	if idx < 0 || idx >= len(variations) {
		fmt.Fprintf(os.Stderr, "Invalid spinner number: %d (valid: 1-%d)\n", *num, len(variations))
		os.Exit(1)
	}

	runSpinner(variations[idx])
}
