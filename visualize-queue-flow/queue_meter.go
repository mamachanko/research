package main

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// ─── Queue Meter L1 ── simple progress bar ────────────────────────────────────

type queueL1Model struct {
	sim      *Sim
	ticks    int
	maxTicks int
	w, h     int
}

func (m queueL1Model) Init() tea.Cmd { return doTick() }

func (m queueL1Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.w, m.h = msg.Width, msg.Height
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

func (m queueL1Model) View() string {
	s := m.sim
	fill := s.QueueFill()
	barW := 48

	var statusStr string
	if s.PubRate > s.ConRate+0.3 {
		statusStr = styleDrop().Render("▲ FILLING")
	} else if s.ConRate > s.PubRate+0.3 {
		statusStr = styleCon().Render("▼ DRAINING")
	} else {
		statusStr = styleMsg().Render("◆ STABLE")
	}

	bar := fillBarAuto(barW, fill)
	pct := int(fill * 100)

	content := styleBold().Render("Queue Depth") + "\n\n" +
		styleDim().Render("exchange: amq.default  ·  queue: task_queue") + "\n\n" +
		"  " + bar + styleText().Render(fmt.Sprintf(" %3d%%  %d/%d", pct, s.QueueDepth, s.QueueCapacity)) + "\n\n" +
		styleDim().Render(fmt.Sprintf("  published: %-6d  consumed: %-6d  dropped: %d",
			s.TotalPub, s.TotalCon, s.TotalDrop)) + "   " + statusStr

	box := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color(colBorder)).
		Padding(1, 3).
		Render(content)

	return "\n\n" + box
}

func runQueueL1() {
	runProgram(queueL1Model{sim: NewSim(), maxTicks: 160})
}

// ─── Queue Meter L2 ── animated gauge + rates on each side ───────────────────

type queueL2Model struct {
	sim      *Sim
	ticks    int
	maxTicks int
	w, h     int
}

func (m queueL2Model) Init() tea.Cmd { return doTick() }

func (m queueL2Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.w, m.h = msg.Width, msg.Height
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

func (m queueL2Model) View() string {
	s := m.sim
	fill := s.QueueFill()

	// Choose bar color dynamically
	barColor := colCon
	if fill >= 0.85 {
		barColor = colDrop
	} else if fill >= 0.55 {
		barColor = colYellow
	}

	barW := 36
	bar := fillBar(barW, fill, barColor, colBorder)

	// Animated fill indicator: pulse on active publish/consume
	pubArrow := stylePub().Render(fmt.Sprintf("%.1f/s ──▶", s.PubRate))
	conArrow := styleCon().Render(fmt.Sprintf("◀── %.1f/s", s.ConRate))

	// Tick-based blinking cursor on the active end
	blink := ""
	if m.ticks%8 < 4 {
		blink = lipgloss.NewStyle().Foreground(lipgloss.Color(barColor)).Render("▌")
	} else {
		blink = " "
	}

	queueLabel := styleQueue().Bold(true).Render("  QUEUE  ")
	fillLabel := styleText().Render(fmt.Sprintf("%d / %d", s.QueueDepth, s.QueueCapacity))
	pctLabel := lipgloss.NewStyle().Foreground(lipgloss.Color(barColor)).Bold(true).
		Render(fmt.Sprintf(" %3d%%", int(fill*100)))

	line1 := center(queueLabel+"  "+styleDim().Render("amq.default / task_queue"), 70)
	line2 := "  " + pubArrow + " [" + bar + blink + "] " + conArrow
	line3 := center(fillLabel+pctLabel, 70)

	// Drop warning
	dropLine := ""
	if s.TotalDrop > 0 {
		dropLine = "\n  " + styleDrop().Render(fmt.Sprintf("⚠  %d messages dropped (queue full)", s.TotalDrop))
	}

	content := line1 + "\n\n" + line2 + "\n" + line3 + dropLine

	box := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color(barColor)).
		Padding(1, 2).
		Width(72).
		Render(content)

	pubBox := lipgloss.NewStyle().
		Border(lipgloss.NormalBorder()).
		BorderForeground(lipgloss.Color(colPub)).
		Padding(0, 2).
		Render(stylePub().Bold(true).Render("PUBLISHER") + "\n" +
			styleDim().Render("sent: ") + styleText().Render(fmt.Sprintf("%d", s.TotalPub)))

	conBox := lipgloss.NewStyle().
		Border(lipgloss.NormalBorder()).
		BorderForeground(lipgloss.Color(colCon)).
		Padding(0, 2).
		Render(styleCon().Bold(true).Render("CONSUMER") + "\n" +
			styleDim().Render("recv: ") + styleText().Render(fmt.Sprintf("%d", s.TotalCon)))

	bottomRow := lipgloss.JoinHorizontal(lipgloss.Top, pubBox, "  ", conBox)
	return "\n" + box + "\n\n" + bottomRow + "\n"
}

func runQueueL2() {
	runProgram(queueL2Model{sim: NewSim(), maxTicks: 180})
}

// ─── Queue Meter L3 ── sparkline history + gauge + stats ─────────────────────

type queueL3Model struct {
	sim      *Sim
	ticks    int
	maxTicks int
	w, h     int
}

func (m queueL3Model) Init() tea.Cmd { return doTick() }

func (m queueL3Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.w, m.h = msg.Width, msg.Height
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

func (m queueL3Model) View() string {
	s := m.sim
	fill := s.QueueFill()
	const w = 60

	// Title
	title := gradientStr("  Queue Depth Monitor", colPub, colMsg)
	sub := styleDim().Render("  amq.default · task_queue · cap=20")

	// Sparkline for depth history
	depthF := make([]float64, len(s.DepthHist))
	for i, v := range s.DepthHist {
		depthF[i] = float64(v)
	}
	sparkColor := colMsg
	if fill >= 0.85 {
		sparkColor = colDrop
	} else if fill >= 0.55 {
		sparkColor = colYellow
	}
	spark := lipgloss.NewStyle().Foreground(lipgloss.Color(sparkColor)).
		Render(sparklineAuto(depthF, w-4))
	sparkLabel := styleDim().Render("  depth / time  (last 60 ticks)") + "\n  " + spark

	// Main gauge
	barColor := colCon
	if fill >= 0.85 {
		barColor = colDrop
	} else if fill >= 0.55 {
		barColor = colYellow
	}
	bar := fillBar(w-4, fill, barColor, colBorder)
	gauge := "  " + bar

	// Depth numbers
	pct := int(fill * 100)
	depthStr := lipgloss.NewStyle().Foreground(lipgloss.Color(barColor)).Bold(true).
		Render(fmt.Sprintf("%d/%d  %d%%", s.QueueDepth, s.QueueCapacity, pct))

	mx := s.MaxDepth()
	avgDepth := 0.0
	if len(s.DepthHist) > 0 {
		for _, v := range s.DepthHist {
			avgDepth += float64(v)
		}
		avgDepth /= float64(len(s.DepthHist))
	}

	stats1 := styleDim().Render("  cur: ") + depthStr +
		styleDim().Render("  max: ") + styleText().Render(fmt.Sprintf("%d", mx)) +
		styleDim().Render("  avg: ") + styleText().Render(fmt.Sprintf("%.1f", avgDepth))

	// Rate history sparklines
	pubSpark := stylePub().Render(sparklineAuto(s.PubHist, (w-4)/2-1))
	conSpark := styleCon().Render(sparklineAuto(s.ConHist, (w-4)/2-1))
	rateRow := "  " + pubSpark + " " + conSpark

	stats2 := stylePub().Render(fmt.Sprintf("  pub %.1f/s", s.PubRate)) +
		styleDim().Render("  ·  ") +
		styleCon().Render(fmt.Sprintf("con %.1f/s", s.ConRate)) +
		styleDim().Render("  ·  ") +
		styleDrop().Render(fmt.Sprintf("dropped %d", s.TotalDrop))

	divider := styleDim().Render("  " + strings.Repeat("─", w-4))

	content := title + "\n" + sub + "\n\n" +
		sparkLabel + "\n\n" +
		divider + "\n" +
		gauge + "\n" +
		stats1 + "\n\n" +
		styleDim().Render("  pub rate ──────── con rate") + "\n" +
		rateRow + "\n" +
		stats2

	box := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color(barColor)).
		Padding(1, 1).
		Width(w + 4).
		Render(content)

	return "\n" + box + "\n"
}

func runQueueL3() {
	runProgram(queueL3Model{sim: NewSim(), maxTicks: 200})
}
