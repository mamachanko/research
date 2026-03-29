package main

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// ─── Throughput L1 ── static bar chart ────────────────────────────────────────

type throughputL1Model struct {
	sim      *Sim
	ticks    int
	maxTicks int
}

func (m throughputL1Model) Init() tea.Cmd { return doTick() }

func (m throughputL1Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
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

func (m throughputL1Model) View() string {
	s := m.sim
	const chartW = 50
	const chartH = 8

	chart := barChart(s.PubHist, chartW, chartH)

	// Y-axis labels
	mx := 0.0
	for _, v := range s.PubHist {
		if v > mx {
			mx = v
		}
	}
	if mx == 0 {
		mx = 6
	}

	yLabels := ""
	for r := chartH - 1; r >= 0; r-- {
		val := mx * float64(chartH-r) / float64(chartH)
		yLabels += fmt.Sprintf("%4.1f│\n", val)
	}
	yLabels += "     └" + strings.Repeat("─", chartW)

	title := gradientStr("  Publish Throughput  ·  msg/tick", colPub, colMsg)
	chartBlock := chart + "\n" + styleDim().Render("     └"+strings.Repeat("─", chartW))

	current := stylePub().Render(fmt.Sprintf("  current: %.2f msg/tick", s.PubRate)) + "  " +
		styleDim().Render(fmt.Sprintf("total: %d", s.TotalPub))

	box := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color(colBorder)).
		Padding(1, 2).
		Render(chartBlock + "\n" + current)

	return "\n" + title + "\n\n" + box + "\n"
}

func runThroughputL1() {
	runProgram(throughputL1Model{sim: NewSim(), maxTicks: 180})
}

// ─── Throughput L2 ── scrolling real-time publish chart ───────────────────────

type throughputL2Model struct {
	sim      *Sim
	ticks    int
	maxTicks int
}

func (m throughputL2Model) Init() tea.Cmd { return doTick() }

func (m throughputL2Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
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

func (m throughputL2Model) View() string {
	s := m.sim
	const chartW = 54
	const chartH = 10

	// Scrolling pub chart
	pubChart := barChart(s.PubHist, chartW, chartH)
	// Scrolling con chart (mirrored, growing down)
	conRows := make([]string, chartH)
	mx := 0.0
	for _, v := range s.ConHist {
		if v > mx {
			mx = v
		}
	}
	for _, v := range s.PubHist {
		if v > mx {
			mx = v
		}
	}
	if mx == 0 {
		mx = 6
	}
	_ = mx

	// Dual chart (publish up, consume down mirror)
	pubLines := strings.Split(pubChart, "\n")

	conMirror := make([]float64, len(s.ConHist))
	for i, v := range s.ConHist {
		conMirror[i] = v
	}
	conChart := barChart(conMirror, chartW, chartH)
	conLines := strings.Split(conChart, "\n")

	// Reverse con lines for mirror effect
	for i, j := 0, len(conLines)-1; i < j; i, j = i+1, j-1 {
		conLines[i], conLines[j] = conLines[j], conLines[i]
	}
	_ = conRows

	axis := styleDim().Render("     ├" + strings.Repeat("─", chartW))
	mirrorAxis := styleDim().Render("     ├" + strings.Repeat("─", chartW))

	pubHeader := stylePub().Render("▲ publish rate") + styleDim().Render(fmt.Sprintf("  %.2f/tick  total %d", s.PubRate, s.TotalPub))
	conHeader := styleCon().Render("▼ consume rate") + styleDim().Render(fmt.Sprintf("  %.2f/tick  total %d", s.ConRate, s.TotalCon))

	var sb strings.Builder
	sb.WriteString(pubHeader + "\n")
	for _, l := range pubLines {
		sb.WriteString(stylePub().Render("  │") + l + "\n")
	}
	sb.WriteString(axis + "\n")
	for _, l := range conLines {
		sb.WriteString(styleCon().Render("  │") + l + "\n")
	}
	sb.WriteString(mirrorAxis + "\n")
	sb.WriteString(conHeader)

	// Current balance indicator
	diff := s.PubRate - s.ConRate
	var balance string
	if diff > 0.5 {
		balance = styleDrop().Render(fmt.Sprintf("  ▲ pub ahead by %.1f/tick → queue filling", diff))
	} else if diff < -0.5 {
		balance = styleCon().Render(fmt.Sprintf("  ▼ con ahead by %.1f/tick → queue draining", -diff))
	} else {
		balance = styleMsg().Render("  ◆ rates balanced")
	}

	title := gradientStr("  Throughput  ·  Publish vs Consume  ·  Scrolling", colPub, colCon)

	box := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color(colBorder)).
		Padding(1, 2).
		Render(sb.String() + "\n" + balance)

	return "\n" + title + "\n\n" + box + "\n"
}

func runThroughputL2() {
	runProgram(throughputL2Model{sim: NewSim(), maxTicks: 200})
}

// ─── Throughput L3 ── dual-color interleaved bar chart + sparklines ────────────

type throughputL3Model struct {
	sim      *Sim
	ticks    int
	maxTicks int
}

func (m throughputL3Model) Init() tea.Cmd { return doTick() }

func (m throughputL3Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
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

func (m throughputL3Model) View() string {
	s := m.sim
	const chartW = 60
	const chartH = 12

	// Dual interleaved bar chart
	dualChart := dualBarChart(s.PubHist, s.ConHist, chartW, chartH, colPub, colCon)

	// Max rate for scale
	mx := 0.0
	for _, v := range s.PubHist {
		if v > mx {
			mx = v
		}
	}
	for _, v := range s.ConHist {
		if v > mx {
			mx = v
		}
	}
	if mx == 0 {
		mx = 6
	}

	// Y-axis
	yLabels := []string{}
	for r := chartH; r >= 0; r-- {
		if r%3 == 0 {
			val := mx * float64(r) / float64(chartH)
			yLabels = append(yLabels, fmt.Sprintf("%4.1f", val))
		} else {
			yLabels = append(yLabels, "    ")
		}
	}

	chartLines := strings.Split(dualChart, "\n")
	var renderedChart strings.Builder
	for i, l := range chartLines {
		prefix := "    "
		if i < len(yLabels) {
			prefix = styleDim().Render(yLabels[i]) + styleDim().Render("│")
		}
		renderedChart.WriteString(prefix + l + "\n")
	}
	renderedChart.WriteString(styleDim().Render("    └" + strings.Repeat("─", chartW)))

	// Depth sparkline
	depthF := make([]float64, len(s.DepthHist))
	for i, v := range s.DepthHist {
		depthF[i] = float64(v)
	}
	depthSpark := styleQueue().Render(sparklineAuto(depthF, chartW))

	// Stats row
	avgLat := s.AvgLatency(20)
	diff := s.PubRate - s.ConRate

	var pressureStr string
	if diff > 0.5 {
		pressureStr = styleDrop().Render(fmt.Sprintf("▲ +%.1f filling", diff))
	} else if diff < -0.5 {
		pressureStr = styleCon().Render(fmt.Sprintf("▼ −%.1f draining", -diff))
	} else {
		pressureStr = styleMsg().Render("◆ balanced")
	}

	legend := stylePub().Render("█ publish") + "  " +
		styleCon().Render("█ consume") + "  " +
		styleQueue().Render("─ depth") + "  " +
		pressureStr

	statsLine := stylePub().Render(fmt.Sprintf("pub %.1f/t  %d tot", s.PubRate, s.TotalPub)) + "  │  " +
		styleCon().Render(fmt.Sprintf("con %.1f/t  %d tot", s.ConRate, s.TotalCon)) + "  │  " +
		styleQueue().Render(fmt.Sprintf("depth %d/%d", s.QueueDepth, s.QueueCapacity)) + "  │  " +
		styleMsg().Render(fmt.Sprintf("lat %.0fms", avgLat))

	title := gradientStr("  Throughput Graph  ·  Dual Rate  ·  Live", colPub, colCon)

	box := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color(colBorder)).
		Padding(1, 2).
		Render(
			legend + "\n\n" +
				renderedChart.String() + "\n" +
				styleDim().Render("    ·") + depthSpark + "\n\n" +
				statsLine)

	return "\n" + title + "\n\n" + box + "\n"
}

func runThroughputL3() {
	runProgram(throughputL3Model{sim: NewSim(), maxTicks: 200})
}
