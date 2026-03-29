package main

import (
	"fmt"
	"math"
	"strings"

	"github.com/charmbracelet/lipgloss"
	colorful "github.com/lucasb-eyer/go-colorful"
)

// Dracula-inspired palette
const (
	colPub    = "#BD93F9" // purple  — publisher
	colCon    = "#50FA7B" // green   — consumer
	colQueue  = "#FFB86C" // orange  — queue
	colMsg    = "#8BE9FD" // cyan    — in-flight messages
	colDrop   = "#FF5555" // red     — dropped / error
	colDim    = "#6272A4" // dim     — secondary text
	colText   = "#F8F8F2" // white   — primary text
	colBorder = "#44475A" // grey    — borders
	colYellow = "#F1FA8C" // yellow  — warning
)

func stylePub() lipgloss.Style    { return lipgloss.NewStyle().Foreground(lipgloss.Color(colPub)) }
func styleCon() lipgloss.Style    { return lipgloss.NewStyle().Foreground(lipgloss.Color(colCon)) }
func styleQueue() lipgloss.Style  { return lipgloss.NewStyle().Foreground(lipgloss.Color(colQueue)) }
func styleMsg() lipgloss.Style    { return lipgloss.NewStyle().Foreground(lipgloss.Color(colMsg)) }
func styleDrop() lipgloss.Style   { return lipgloss.NewStyle().Foreground(lipgloss.Color(colDrop)) }
func styleDim() lipgloss.Style    { return lipgloss.NewStyle().Foreground(lipgloss.Color(colDim)) }
func styleText() lipgloss.Style   { return lipgloss.NewStyle().Foreground(lipgloss.Color(colText)) }
func styleBold() lipgloss.Style   { return lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color(colText)) }

func boxStyle(w int) lipgloss.Style {
	return lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color(colBorder)).
		Padding(0, 1).
		Width(w)
}

// fillBar returns a colored progress bar of given width and fill ratio 0..1
func fillBar(width int, fill float64, filledColor, emptyColor string) string {
	n := int(math.Round(fill * float64(width)))
	if n > width {
		n = width
	}
	filled := lipgloss.NewStyle().Foreground(lipgloss.Color(filledColor)).Render(strings.Repeat("█", n))
	empty := lipgloss.NewStyle().Foreground(lipgloss.Color(emptyColor)).Render(strings.Repeat("░", width-n))
	return filled + empty
}

// fillBarColor picks green/yellow/red automatically
func fillBarAuto(width int, fill float64) string {
	var c string
	switch {
	case fill >= 0.85:
		c = colDrop
	case fill >= 0.55:
		c = colYellow
	default:
		c = colCon
	}
	return fillBar(width, fill, c, colBorder)
}

// sparkline renders a series of values as a single-row braille chart
// values should be normalised to [0, max]
func sparkline(vals []float64, width int, max float64) string {
	if len(vals) == 0 || max <= 0 {
		return strings.Repeat("▁", width)
	}
	chars := []rune{'▁', '▂', '▃', '▄', '▅', '▆', '▇', '█'}
	// sample vals to fit width
	step := float64(len(vals)) / float64(width)
	var sb strings.Builder
	for i := 0; i < width; i++ {
		idx := int(float64(i) * step)
		if idx >= len(vals) {
			idx = len(vals) - 1
		}
		v := vals[idx]
		lvl := int(v / max * float64(len(chars)-1))
		if lvl < 0 {
			lvl = 0
		}
		if lvl >= len(chars) {
			lvl = len(chars) - 1
		}
		sb.WriteRune(chars[lvl])
	}
	return sb.String()
}

// sparklineF is sparkline but for float slices, computing max automatically
func sparklineAuto(vals []float64, width int) string {
	mx := 0.0
	for _, v := range vals {
		if v > mx {
			mx = v
		}
	}
	if mx == 0 {
		mx = 1
	}
	return sparkline(vals, width, mx)
}

// barChart renders a vertical bar chart (single character height = 8 levels)
// Returns `rows` lines, each `width` wide
func barChart(vals []float64, width, rows int) string {
	if len(vals) == 0 {
		return ""
	}
	mx := 0.0
	for _, v := range vals {
		if v > mx {
			mx = v
		}
	}
	if mx == 0 {
		mx = 1
	}

	type cell struct{ r rune; c string }
	grid := make([][]cell, rows)
	for r := range grid {
		grid[r] = make([]cell, width)
		for c := range grid[r] {
			grid[r][c] = cell{' ', ""}
		}
	}

	// Sample vals to fit width
	step := float64(len(vals)) / float64(width)
	for col := 0; col < width; col++ {
		idx := int(float64(col) * step)
		if idx >= len(vals) {
			idx = len(vals) - 1
		}
		v := vals[idx]
		fillH := v / mx * float64(rows)
		fillFull := int(fillH)
		partial := fillH - float64(fillFull)

		// Fill full rows from bottom
		for r := 0; r < fillFull && r < rows; r++ {
			rowIdx := rows - 1 - r
			grid[rowIdx][col] = cell{'█', colMsg}
		}
		// Partial top block
		if fillFull < rows {
			blocks := []rune{' ', '▁', '▂', '▃', '▄', '▅', '▆', '▇', '█'}
			lvl := int(partial * float64(len(blocks)-1))
			if lvl > 0 {
				rowIdx := rows - 1 - fillFull
				grid[rowIdx][col] = cell{blocks[lvl], colMsg}
			}
		}
	}

	var lines []string
	for _, row := range grid {
		var sb strings.Builder
		for _, c := range row {
			if c.c != "" {
				sb.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color(c.c)).Render(string(c.r)))
			} else {
				sb.WriteRune(c.r)
			}
		}
		lines = append(lines, sb.String())
	}
	return strings.Join(lines, "\n")
}

// dualBarChart renders two series side by side on the same scale
func dualBarChart(a, b []float64, width, rows int, ca, cb string) string {
	if len(a) == 0 && len(b) == 0 {
		return ""
	}
	mx := 0.0
	for _, v := range a {
		if v > mx {
			mx = v
		}
	}
	for _, v := range b {
		if v > mx {
			mx = v
		}
	}
	if mx == 0 {
		mx = 1
	}

	type cell struct{ r rune; c string }
	grid := make([][]cell, rows)
	for r := range grid {
		grid[r] = make([]cell, width)
		for c := range grid[r] {
			grid[r][c] = cell{' ', ""}
		}
	}

	renderSeries := func(vals []float64, col int, color string) {
		if len(vals) == 0 {
			return
		}
		step := float64(len(vals)) / float64(width/2)
		idx := int(float64(col) * step)
		if idx >= len(vals) {
			idx = len(vals) - 1
		}
		v := vals[idx]
		fillH := v / mx * float64(rows)
		fillFull := int(fillH)
		for r := 0; r < fillFull && r < rows; r++ {
			grid[rows-1-r][col] = cell{'█', color}
		}
	}

	// Interleave a and b columns
	for col := 0; col < width; col++ {
		if col%2 == 0 {
			renderSeries(a, col, ca)
		} else {
			renderSeries(b, col, cb)
		}
	}

	var lines []string
	for _, row := range grid {
		var sb strings.Builder
		for _, c := range row {
			if c.c != "" {
				sb.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color(c.c)).Render(string(c.r)))
			} else {
				sb.WriteRune(c.r)
			}
		}
		lines = append(lines, sb.String())
	}
	return strings.Join(lines, "\n")
}

// gradientStr renders text with a horizontal color gradient between two hex colors
func gradientStr(text, from, to string) string {
	c0, _ := colorful.Hex(from)
	c1, _ := colorful.Hex(to)
	runes := []rune(text)
	n := len(runes)
	if n == 0 {
		return ""
	}
	var sb strings.Builder
	for i, r := range runes {
		t := float64(i) / float64(n-1)
		if n == 1 {
			t = 0
		}
		blended := c0.BlendHcl(c1, t).Clamped()
		hex := fmt.Sprintf("#%02X%02X%02X",
			uint8(blended.R*255), uint8(blended.G*255), uint8(blended.B*255))
		sb.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color(hex)).Render(string(r)))
	}
	return sb.String()
}

// center pads s to width w
func center(s string, w int) string {
	vis := lipgloss.Width(s)
	if vis >= w {
		return s
	}
	pad := (w - vis) / 2
	return strings.Repeat(" ", pad) + s + strings.Repeat(" ", w-vis-pad)
}

// heatCell returns a colored block based on intensity 0..1
func heatCell(intensity float64) string {
	// black → dark blue → blue → cyan → green → yellow → red
	colors := []string{
		"#1E1E2E", "#1A1A4E", "#2B5CE6", "#00B4FF",
		"#50FA7B", "#F1FA8C", "#FFB86C", "#FF5555",
	}
	i := int(intensity * float64(len(colors)-1))
	if i < 0 {
		i = 0
	}
	if i >= len(colors) {
		i = len(colors) - 1
	}
	return lipgloss.NewStyle().Background(lipgloss.Color(colors[i])).Render(" ")
}
