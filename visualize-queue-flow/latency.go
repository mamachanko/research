package main

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// Latency buckets: 0-20, 20-40, 40-60, 60-80, 80-100, 100-120, 120-150, 150+ms
var latBuckets = []struct {
	lo, hi float64
	label  string
}{
	{0, 20, "0-20ms"},
	{20, 40, "20-40ms"},
	{40, 60, "40-60ms"},
	{60, 80, "60-80ms"},
	{80, 100, "80-100ms"},
	{100, 120, "100-120ms"},
	{120, 150, "120-150ms"},
	{150, 1e9, "150ms+"},
}

func bucketLatencies(lats []float64) []int {
	counts := make([]int, len(latBuckets))
	for _, l := range lats {
		for i, b := range latBuckets {
			if l >= b.lo && l < b.hi {
				counts[i]++
				break
			}
		}
	}
	return counts
}

// ─── Latency L1 ── static histogram ───────────────────────────────────────────

type latencyL1Model struct {
	sim      *Sim
	ticks    int
	maxTicks int
}

func (m latencyL1Model) Init() tea.Cmd { return doTick() }

func (m latencyL1Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg.(type) {
	case tea.KeyMsg:
		return m, tea.Quit
	case tickMsg:
		m.sim.Step()
		m.ticks++
		if m.ticks >= m.maxTicks {
			return m, tea.Quit
		}
		return m, doTick()
	}
	return m, nil
}

func (m latencyL1Model) View() string {
	s := m.sim
	counts := bucketLatencies(s.LatHist)

	mx := 0
	for _, c := range counts {
		if c > mx {
			mx = c
		}
	}
	if mx == 0 {
		mx = 1
	}

	const barW = 28
	var lines []string
	for i, b := range latBuckets {
		fill := float64(counts[i]) / float64(mx)
		bColor := colMsg
		switch {
		case b.lo >= 100:
			bColor = colDrop
		case b.lo >= 60:
			bColor = colYellow
		}
		bar := lipgloss.NewStyle().Foreground(lipgloss.Color(bColor)).
			Render(strings.Repeat("█", int(fill*barW))) +
			styleDim().Render(strings.Repeat("░", barW-int(fill*barW)))
		label := fmt.Sprintf("%-10s", b.label)
		count := fmt.Sprintf("%4d", counts[i])
		lines = append(lines, styleDim().Render(label)+" "+bar+" "+styleText().Render(count))
	}

	avgLat := s.AvgLatency(50)
	p95 := 0.0
	if len(s.LatHist) > 0 {
		sorted := make([]float64, len(s.LatHist))
		copy(sorted, s.LatHist)
		// simple insertion sort (small slice)
		for i := 1; i < len(sorted); i++ {
			for j := i; j > 0 && sorted[j] < sorted[j-1]; j-- {
				sorted[j], sorted[j-1] = sorted[j-1], sorted[j]
			}
		}
		p95 = sorted[int(float64(len(sorted))*0.95)]
	}

	stats := styleDim().Render("avg: ") + styleMsg().Render(fmt.Sprintf("%.1fms", avgLat)) + "  " +
		styleDim().Render("p95: ") + styleMsg().Render(fmt.Sprintf("%.1fms", p95)) + "  " +
		styleDim().Render("samples: ") + styleText().Render(fmt.Sprintf("%d", len(s.LatHist)))

	title := gradientStr("  Latency Distribution  ·  task_queue", colMsg, colCon)

	box := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color(colBorder)).
		Padding(1, 2).
		Render(strings.Join(lines, "\n") + "\n\n" + stats)

	return "\n" + title + "\n\n" + box + "\n"
}

func runLatencyL1() {
	runProgram(latencyL1Model{sim: NewSim(), maxTicks: 160})
}

// ─── Latency L2 ── animated updating histogram ────────────────────────────────

type latencyL2Model struct {
	sim      *Sim
	ticks    int
	maxTicks int
	// Rolling window for smoother display
	windowCounts [8]int
	windowTotal  int
}

func (m latencyL2Model) Init() tea.Cmd { return doTick() }

func (m latencyL2Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg.(type) {
	case tea.KeyMsg:
		return m, tea.Quit
	case tickMsg:
		m.sim.Step()
		m.ticks++
		if m.ticks >= m.maxTicks {
			return m, tea.Quit
		}
		// Update rolling counts from last 40 samples
		recent := m.sim.LatHist
		if len(recent) > 40 {
			recent = recent[len(recent)-40:]
		}
		counts := bucketLatencies(recent)
		for i, c := range counts {
			m.windowCounts[i] = c
		}
		m.windowTotal = len(recent)
		return m, doTick()
	}
	return m, nil
}

func (m latencyL2Model) View() string {
	s := m.sim
	counts := m.windowCounts[:]
	total := m.windowTotal

	mx := 0
	for _, c := range counts {
		if c > mx {
			mx = c
		}
	}
	if mx == 0 {
		mx = 1
	}

	const barW = 32
	var lines []string
	for i, b := range latBuckets {
		fill := float64(counts[i]) / float64(mx)
		pct := 0.0
		if total > 0 {
			pct = float64(counts[i]) / float64(total) * 100
		}

		bColor := colMsg
		switch {
		case b.lo >= 100:
			bColor = colDrop
		case b.lo >= 60:
			bColor = colYellow
		case b.lo >= 40:
			bColor = colQueue
		}

		// Animated: bar length changes smoothly tick-by-tick
		bar := lipgloss.NewStyle().Foreground(lipgloss.Color(bColor)).
			Render(strings.Repeat("█", int(fill*barW))) +
			styleDim().Render(strings.Repeat("░", barW-int(fill*barW)))

		label := fmt.Sprintf("%-10s", b.label)
		pctStr := fmt.Sprintf("%5.1f%%", pct)
		cntStr := fmt.Sprintf("%3d", counts[i])
		lines = append(lines, styleDim().Render(label)+" "+bar+
			"  "+styleText().Render(cntStr)+
			"  "+lipgloss.NewStyle().Foreground(lipgloss.Color(bColor)).Render(pctStr))
	}

	avgLat := s.AvgLatency(40)

	// Latency trend sparkline
	var window20 []float64
	if len(s.LatHist) > 20 {
		window20 = s.LatHist[len(s.LatHist)-20:]
	} else {
		window20 = s.LatHist
	}
	latSpark := styleMsg().Render(sparklineAuto(window20, 40))
	trendLabel := styleDim().Render("  recent latency trend: ") + latSpark

	// Queue depth influence note
	depthNote := styleDim().Render(fmt.Sprintf("  queue depth: %d/%d  → higher depth = higher latency",
		s.QueueDepth, s.QueueCapacity))

	stats := "  " + styleDim().Render("window: last 40 samples  ") +
		styleDim().Render("avg: ") + styleMsg().Render(fmt.Sprintf("%.1fms", avgLat)) + "  " +
		styleDim().Render("depth contrib: ") + styleMsg().Render(fmt.Sprintf("+%.0fms", float64(s.QueueDepth)*2.5))

	title := gradientStr("  Latency Distribution  ·  Live  ·  40-sample window", colMsg, colCon)

	box := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color(colBorder)).
		Padding(1, 2).
		Render(strings.Join(lines, "\n") + "\n\n" + trendLabel + "\n" + depthNote + "\n" + stats)

	return "\n" + title + "\n\n" + box + "\n"
}

func runLatencyL2() {
	runProgram(latencyL2Model{sim: NewSim(), maxTicks: 200})
}

// ─── Latency L3 ── 2D heat grid: bucket × time ────────────────────────────────

type latencyL3Model struct {
	sim      *Sim
	ticks    int
	maxTicks int
	// heat grid: [timeSlot][bucket] = count
	grid     [][8]int
	maxCount int
	// Snapshot every 4 ticks
	snapTick int
}

func (m latencyL3Model) Init() tea.Cmd { return doTick() }

func (m latencyL3Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg.(type) {
	case tea.KeyMsg:
		return m, tea.Quit
	case tickMsg:
		m.sim.Step()
		m.ticks++
		if m.ticks >= m.maxTicks {
			return m, tea.Quit
		}

		// Snapshot every 4 ticks
		if m.ticks-m.snapTick >= 4 {
			m.snapTick = m.ticks
			// Take last 8 latency samples for this time slot
			recent := m.sim.LatHist
			if len(recent) > 8 {
				recent = recent[len(recent)-8:]
			}
			counts := bucketLatencies(recent)
			var slot [8]int
			for i, c := range counts {
				slot[i] = c
				if c > m.maxCount {
					m.maxCount = c
				}
			}
			m.grid = append(m.grid, slot)
			// Keep last 40 time slots
			if len(m.grid) > 40 {
				m.grid = m.grid[1:]
			}
		}

		return m, doTick()
	}
	return m, nil
}

func (m latencyL3Model) View() string {
	s := m.sim
	nSlots := len(m.grid)
	if nSlots == 0 {
		return "\n  collecting data...\n"
	}

	mx := m.maxCount
	if mx == 0 {
		mx = 1
	}

	// Build the heat grid: rows=buckets, cols=time slots
	var gridLines []string
	for bi, b := range latBuckets {
		label := fmt.Sprintf("%-10s │", b.label)
		var row strings.Builder
		row.WriteString(styleDim().Render(label))

		for _, slot := range m.grid {
			intensity := float64(slot[bi]) / float64(mx)
			row.WriteString(heatCell(intensity))
		}

		// Pad to 40 slots
		for i := nSlots; i < 40; i++ {
			row.WriteString(styleDim().Render(" "))
		}

		// Summary bar on the right
		totalInBucket := 0
		for _, slot := range m.grid {
			totalInBucket += slot[bi]
		}
		row.WriteString(styleDim().Render("│ ") + styleText().Render(fmt.Sprintf("%4d", totalInBucket)))
		gridLines = append(gridLines, row.String())
	}

	// Time axis
	timeAxis := styleDim().Render(strings.Repeat("─", 10) + "┼" + strings.Repeat("─", 40) + "┼──────")
	header := styleDim().Render(strings.Repeat(" ", 11) + "│") +
		styleMsg().Render(center("← time (newest right) →", 40)) +
		styleDim().Render("│ total")

	// Color legend
	legend := "  " + styleDim().Render("intensity: ")
	intensities := []float64{0, 0.15, 0.3, 0.5, 0.7, 0.85, 1.0}
	for _, iv := range intensities {
		legend += heatCell(iv)
	}
	legend += "  " + styleDim().Render("low → high")

	// Stats
	avgLat := s.AvgLatency(40)
	depth := s.QueueDepth

	var maxLat float64
	if len(s.LatHist) > 0 {
		for _, l := range s.LatHist {
			if l > maxLat {
				maxLat = l
			}
		}
	}

	stats := styleDim().Render("  avg: ") + styleMsg().Render(fmt.Sprintf("%.1fms", avgLat)) + "  " +
		styleDim().Render("max: ") + styleDrop().Render(fmt.Sprintf("%.1fms", maxLat)) + "  " +
		styleDim().Render("queue depth: ") + styleQueue().Render(fmt.Sprintf("%d", depth)) + "  " +
		styleDim().Render("samples: ") + styleText().Render(fmt.Sprintf("%d", len(s.LatHist)))

	title := gradientStr("  Latency Heat Map  ·  Bucket × Time  ·  Live", colMsg, colCon)

	box := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color(colBorder)).
		Padding(1, 1).
		Render(header + "\n" + timeAxis + "\n" +
			strings.Join(gridLines, "\n") + "\n" +
			timeAxis + "\n\n" + legend + "\n\n" + stats)

	return "\n" + title + "\n\n" + box + "\n"
}

func runLatencyL3() {
	runProgram(latencyL3Model{sim: NewSim(), maxTicks: 220})
}
