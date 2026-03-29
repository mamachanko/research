package main

import (
	"fmt"
	"math"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// ─── Flow L1 ── static topology diagram ───────────────────────────────────────

type flowL1Model struct {
	sim      *Sim
	ticks    int
	maxTicks int
	w, h     int
}

func (m flowL1Model) Init() tea.Cmd { return doTick() }

func (m flowL1Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
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

func (m flowL1Model) View() string {
	s := m.sim
	fill := s.QueueFill()

	pubState := "IDLE"
	if s.PubRate > 2 {
		pubState = "ACTIVE"
	}
	conState := "IDLE"
	if s.ConRate > 2 {
		conState = "ACTIVE"
	}

	pubColor := colPub
	if pubState == "ACTIVE" {
		pubColor = "#FF79C6"
	}
	conColor := colCon
	if conState == "ACTIVE" {
		conColor = "#69FF47"
	}
	queueColor := colQueue
	if fill >= 0.85 {
		queueColor = colDrop
	}

	pubBox := lipgloss.NewStyle().
		Border(lipgloss.DoubleBorder()).
		BorderForeground(lipgloss.Color(pubColor)).
		Padding(0, 2).
		Align(lipgloss.Center).
		Width(16).
		Render(
			lipgloss.NewStyle().Foreground(lipgloss.Color(pubColor)).Bold(true).Render("PUBLISHER") + "\n" +
				styleDim().Render("node-a") + "\n" +
				lipgloss.NewStyle().Foreground(lipgloss.Color(pubColor)).Render(pubState) + "\n" +
				styleDim().Render(fmt.Sprintf("sent: %d", s.TotalPub)))

	queueBox := lipgloss.NewStyle().
		Border(lipgloss.ThickBorder()).
		BorderForeground(lipgloss.Color(queueColor)).
		Padding(0, 2).
		Align(lipgloss.Center).
		Width(18).
		Render(
			lipgloss.NewStyle().Foreground(lipgloss.Color(queueColor)).Bold(true).Render("QUEUE") + "\n" +
				styleDim().Render("task_queue") + "\n" +
				lipgloss.NewStyle().Foreground(lipgloss.Color(queueColor)).
					Render(fmt.Sprintf("%d / %d", s.QueueDepth, s.QueueCapacity)) + "\n" +
				fillBarAuto(12, fill))

	conBox := lipgloss.NewStyle().
		Border(lipgloss.DoubleBorder()).
		BorderForeground(lipgloss.Color(conColor)).
		Padding(0, 2).
		Align(lipgloss.Center).
		Width(16).
		Render(
			lipgloss.NewStyle().Foreground(lipgloss.Color(conColor)).Bold(true).Render("CONSUMER") + "\n" +
				styleDim().Render("node-b") + "\n" +
				lipgloss.NewStyle().Foreground(lipgloss.Color(conColor)).Render(conState) + "\n" +
				styleDim().Render(fmt.Sprintf("recv: %d", s.TotalCon)))

	arrow := styleDim().Render("──────▶")

	row := lipgloss.JoinHorizontal(lipgloss.Center, pubBox, arrow, queueBox, arrow, conBox)

	pubRateStr := stylePub().Render(fmt.Sprintf("%.1f msg/s", s.PubRate))
	conRateStr := styleCon().Render(fmt.Sprintf("%.1f msg/s", s.ConRate))

	title := gradientStr("  Message Queue Topology", colPub, colCon)
	footer := "  " + pubRateStr + styleDim().Render("  ──  exchange: amq.default  ──  ") + conRateStr

	return "\n" + title + "\n\n" + row + "\n\n" + footer + "\n"
}

func runFlowL1() {
	runProgram(flowL1Model{sim: NewSim(), maxTicks: 160})
}

// ─── Flow L2 ── animated single-row particle stream ───────────────────────────

type flowL2Model struct {
	sim      *Sim
	ticks    int
	maxTicks int
	w, h     int
}

func (m flowL2Model) Init() tea.Cmd { return doTick() }

func (m flowL2Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
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

func (m flowL2Model) View() string {
	s := m.sim
	width := 76
	if m.w > 0 && m.w < width+6 {
		width = m.w - 6
	}

	// Zone boundaries
	pubEnd := 10
	qL := width * 35 / 100
	qR := width * 65 / 100
	conStart := width * 90 / 100

	// Build the particle row
	cells := make([]rune, width)
	for i := range cells {
		cells[i] = ' '
	}
	cellColor := make([]string, width)

	// Draw zone markers
	for i := pubEnd; i < qL; i++ {
		cells[i] = '─'
		cellColor[i] = colBorder
	}
	for i := qR; i < conStart; i++ {
		cells[i] = '─'
		cellColor[i] = colBorder
	}

	// Draw queue zone
	queueDepthVis := s.QueueDepth
	for i := qL; i < qR; i++ {
		relPos := i - qL
		qZoneW := qR - qL
		fillN := queueDepthVis * qZoneW / s.QueueCapacity
		if relPos < fillN {
			cells[i] = '▓'
			cellColor[i] = colQueue
		} else {
			cells[i] = '░'
			cellColor[i] = colBorder
		}
	}

	// Draw particles
	for _, p := range s.Particles {
		col := int(p.Pos * float64(width-1))
		if col < 0 || col >= width {
			continue
		}
		c := colMsg
		if p.Pos < float64(pubEnd)/float64(width) {
			c = colPub
		} else if p.InQueue {
			c = colQueue
		} else if p.Pos > float64(conStart)/float64(width) {
			c = colCon
		}
		cells[col] = p.Char
		cellColor[col] = c
	}

	// Render the particle row
	var rowSB strings.Builder
	for i, ch := range cells {
		if cellColor[i] != "" {
			rowSB.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color(cellColor[i])).Render(string(ch)))
		} else {
			rowSB.WriteRune(ch)
		}
	}
	particleRow := rowSB.String()

	// Zone labels
	pubLabel := center(stylePub().Render("PUB"), pubEnd)
	queueLabel := center(styleQueue().Render(fmt.Sprintf("QUEUE %d/%d", s.QueueDepth, s.QueueCapacity)), qR-qL)
	conLabel := center(styleCon().Render("CON"), width-conStart)
	spacerL := strings.Repeat(" ", qL-pubEnd)
	spacerR := strings.Repeat(" ", conStart-qR)
	labelRow := pubLabel + spacerL + queueLabel + spacerR + conLabel

	// Rate row
	pubRate := stylePub().Render(fmt.Sprintf("▶ %.1f msg/s", s.PubRate))
	conRate := styleCon().Render(fmt.Sprintf("%.1f msg/s ▶", s.ConRate))
	qDepth := styleQueue().Render(fmt.Sprintf("depth: %d", s.QueueDepth))
	mid := (width - lipgloss.Width(qDepth)) / 2
	rateRow := pubRate + strings.Repeat(" ", mid-lipgloss.Width(pubRate)) + qDepth +
		strings.Repeat(" ", width-mid-lipgloss.Width(qDepth)-lipgloss.Width(conRate)) + conRate

	title := gradientStr("  Message Flow  (amq.default → task_queue)", colPub, colCon)

	box := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color(colBorder)).
		Padding(1, 1).
		Render(labelRow + "\n" + particleRow + "\n" + rateRow)

	return "\n" + title + "\n\n" + box + "\n"
}

func runFlowL2() {
	runProgram(flowL2Model{sim: NewSim(), maxTicks: 200})
}

// ─── Flow L3 ── multi-row particle burst system ────────────────────────────────

type flowL3Model struct {
	sim      *Sim
	ticks    int
	maxTicks int
	w, h     int
	pubBurst int // ticks remaining for publisher burst effect
	conBurst int // ticks remaining for consumer burst effect
	prevPub  int
	prevCon  int
}

func (m flowL3Model) Init() tea.Cmd { return doTick() }

func (m flowL3Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
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

		// Detect new publishes/consumes for burst effect
		if m.sim.TotalPub > m.prevPub {
			m.pubBurst = 6
		}
		if m.sim.TotalCon > m.prevCon {
			m.conBurst = 6
		}
		m.prevPub = m.sim.TotalPub
		m.prevCon = m.sim.TotalCon
		if m.pubBurst > 0 {
			m.pubBurst--
		}
		if m.conBurst > 0 {
			m.conBurst--
		}

		return m, doTick()
	}
	return m, nil
}

func (m flowL3Model) renderParticleRows(rows, width int) []string {
	s := m.sim
	const qL, qR, conS = 0.32, 0.68, 0.88

	lines := make([][]rune, rows)
	colors := make([][]string, rows)
	for r := range lines {
		lines[r] = make([]rune, width)
		colors[r] = make([]string, width)
		for c := range lines[r] {
			lines[r][c] = ' '
		}
	}

	// Draw queue zone (across all rows)
	qColL := int(qL * float64(width))
	qColR := int(qR * float64(width))
	for r := 0; r < rows; r++ {
		for c := qColL; c <= qColR; c++ {
			lines[r][c] = '·'
			colors[r][c] = colBorder
		}
	}

	// Draw particles spread across rows
	for _, p := range s.Particles {
		col := int(p.Pos * float64(width-1))
		if col < 0 || col >= width {
			continue
		}
		row := int(math.Abs(float64(p.ID%rows)))
		if row >= rows {
			row = rows - 1
		}

		ch := p.Char
		var c string
		switch {
		case p.Pos < qL:
			c = colPub
			ch = '●'
		case p.InQueue:
			c = colQueue
			ch = '▪'
		case p.Pos > conS:
			c = colCon
			ch = '◉'
		default:
			c = colMsg
			ch = '○'
		}
		lines[row][col] = ch
		colors[row][col] = c
	}

	result := make([]string, rows)
	for r := range lines {
		var sb strings.Builder
		for c, ch := range lines[r] {
			if colors[r][c] != "" {
				sb.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color(colors[r][c])).Render(string(ch)))
			} else {
				sb.WriteRune(ch)
			}
		}
		result[r] = sb.String()
	}
	return result
}

func (m flowL3Model) View() string {
	s := m.sim
	w := 90
	rows := 5
	qL, qR := 0.32, 0.68

	// Publisher burst indicator
	pubIndicator := "  "
	if m.pubBurst > 0 {
		burst := []string{"✸", "✷", "✦", "✧", "·"}
		pubIndicator = stylePub().Render(burst[m.pubBurst%len(burst)]) + " "
	}

	// Consumer burst indicator
	conIndicator := "  "
	if m.conBurst > 0 {
		burst := []string{"◈", "◇", "◆", "◇", "·"}
		conIndicator = " " + styleCon().Render(burst[m.conBurst%len(burst)])
	}

	// Publisher node
	pubActive := m.pubBurst > 0
	pubBorderColor := colBorder
	if pubActive {
		pubBorderColor = colPub
	}
	pubSpinner := []string{"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏"}
	pubSpin := ""
	if pubActive {
		pubSpin = stylePub().Render(pubSpinner[m.ticks%len(pubSpinner)])
	} else {
		pubSpin = styleDim().Render("·")
	}

	pubPanel := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color(pubBorderColor)).
		Padding(0, 1).
		Width(16).
		Render(
			stylePub().Bold(true).Render("PUBLISHER") + " " + pubSpin + "\n" +
				styleDim().Render("node-a") + "\n" +
				stylePub().Render(fmt.Sprintf("▲ %.1f/s", s.PubRate)) + "\n" +
				styleDim().Render(fmt.Sprintf("tot: %d", s.TotalPub)))

	// Queue panel
	queueFill := s.QueueFill()
	qFillBar := fillBarAuto(10, queueFill)
	qColor := colQueue
	if queueFill >= 0.85 {
		qColor = colDrop
	}
	queuePanel := lipgloss.NewStyle().
		Border(lipgloss.ThickBorder()).
		BorderForeground(lipgloss.Color(qColor)).
		Padding(0, 1).
		Width(20).
		Render(
			lipgloss.NewStyle().Foreground(lipgloss.Color(qColor)).Bold(true).Render("QUEUE") + "\n" +
				styleDim().Render("task_queue") + "\n" +
				qFillBar + "\n" +
				lipgloss.NewStyle().Foreground(lipgloss.Color(qColor)).
					Render(fmt.Sprintf("%d/%d  %d%%", s.QueueDepth, s.QueueCapacity, int(queueFill*100))) + "\n" +
				styleDrop().Render(fmt.Sprintf("dropped: %d", s.TotalDrop)))

	// Consumer node
	conActive := m.conBurst > 0
	conBorderColor := colBorder
	if conActive {
		conBorderColor = colCon
	}
	conSpinner := []string{"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏"}
	conSpin := ""
	if conActive {
		conSpin = styleCon().Render(conSpinner[(m.ticks+5)%len(conSpinner)])
	} else {
		conSpin = styleDim().Render("·")
	}

	conPanel := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color(conBorderColor)).
		Padding(0, 1).
		Width(16).
		Render(
			styleCon().Bold(true).Render("CONSUMER") + " " + conSpin + "\n" +
				styleDim().Render("node-b") + "\n" +
				styleCon().Render(fmt.Sprintf("▼ %.1f/s", s.ConRate)) + "\n" +
				styleDim().Render(fmt.Sprintf("tot: %d", s.TotalCon)))

	topRow := lipgloss.JoinHorizontal(lipgloss.Top, pubPanel, "  ", queuePanel, "  ", conPanel)

	// Particle stream
	streamRows := m.renderParticleRows(rows, w)

	// Legend
	legend := "  " +
		stylePub().Render("● pub") + "  " +
		styleMsg().Render("○ transit") + "  " +
		styleQueue().Render("▪ queued") + "  " +
		styleCon().Render("◉ consumed") + "  " +
		pubIndicator + stylePub().Render("burst") +
		conIndicator + styleCon().Render("absorbed")

	streamBox := lipgloss.NewStyle().
		Border(lipgloss.NormalBorder()).
		BorderForeground(lipgloss.Color(colBorder)).
		Padding(0, 1).
		Width(w + 2).
		Render(strings.Join(streamRows, "\n"))

	qZoneLabel := center(styleQueue().Render("queue zone"), int(float64(w)*(qR-qL)))
	leftPad := strings.Repeat(" ", int(float64(w)*qL)+3)
	zoneMarker := leftPad + qZoneLabel

	title := gradientStr("  Message Flow  ·  Particle Stream  ·  amq.default", colPub, colCon)

	return "\n" + title + "\n\n" + topRow + "\n\n" + zoneMarker + "\n" + streamBox + "\n" + legend + "\n"
}

func runFlowL3() {
	runProgram(flowL3Model{sim: NewSim(), maxTicks: 240})
}
