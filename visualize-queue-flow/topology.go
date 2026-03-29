package main

import (
	"fmt"
	"math"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// ─── Topology L1 ── static ASCII network diagram ──────────────────────────────

type topologyL1Model struct {
	sim      *Sim
	ticks    int
	maxTicks int
}

func (m topologyL1Model) Init() tea.Cmd { return doTick() }

func (m topologyL1Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
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

func (m topologyL1Model) View() string {
	s := m.sim

	pubActive := s.PubRate > 2
	conActive := s.ConRate > 2

	pubColor := colBorder
	if pubActive {
		pubColor = colPub
	}
	conColor := colBorder
	if conActive {
		conColor = colCon
	}
	qColor := colQueue
	if s.QueueFill() >= 0.85 {
		qColor = colDrop
	}

	// Node boxes
	pubNode := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color(pubColor)).
		Padding(1, 2).
		Width(18).
		Align(lipgloss.Center).
		Render(
			lipgloss.NewStyle().Foreground(lipgloss.Color(pubColor)).Bold(true).Render("node-a") + "\n" +
				styleDim().Render("publisher") + "\n\n" +
				stylePub().Render(fmt.Sprintf("↑ %.1f/s", s.PubRate)) + "\n" +
				styleDim().Render(fmt.Sprintf("sent %d", s.TotalPub)))

	queueNode := lipgloss.NewStyle().
		Border(lipgloss.ThickBorder()).
		BorderForeground(lipgloss.Color(qColor)).
		Padding(1, 2).
		Width(20).
		Align(lipgloss.Center).
		Render(
			lipgloss.NewStyle().Foreground(lipgloss.Color(qColor)).Bold(true).Render("task_queue") + "\n" +
				styleDim().Render("amq.default") + "\n\n" +
				lipgloss.NewStyle().Foreground(lipgloss.Color(qColor)).Render(
					fmt.Sprintf("%d / %d", s.QueueDepth, s.QueueCapacity)) + "\n" +
				fillBarAuto(14, s.QueueFill()))

	conNode := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color(conColor)).
		Padding(1, 2).
		Width(18).
		Align(lipgloss.Center).
		Render(
			lipgloss.NewStyle().Foreground(lipgloss.Color(conColor)).Bold(true).Render("node-b") + "\n" +
				styleDim().Render("consumer") + "\n\n" +
				styleCon().Render(fmt.Sprintf("↓ %.1f/s", s.ConRate)) + "\n" +
				styleDim().Render(fmt.Sprintf("recv %d", s.TotalCon)))

	// Edge labels
	pubEdge := styleDim().Render("  AMQP publish\n") +
		stylePub().Render("  ───────────▶")
	conEdge := styleDim().Render("  AMQP consume\n") +
		styleCon().Render("  ───────────▶")

	row := lipgloss.JoinHorizontal(lipgloss.Center,
		pubNode, pubEdge, queueNode, conEdge, conNode)

	// Protocol details footer
	footer := styleDim().Render("  protocol: AMQP 0-9-1  ·  broker: RabbitMQ  ·  exchange: amq.default  ·  binding: task_queue")

	title := gradientStr("  Network Topology  ·  RabbitMQ Pipeline", colPub, colCon)
	return "\n" + title + "\n\n" + row + "\n\n" + footer + "\n"
}

func runTopologyL1() {
	runProgram(topologyL1Model{sim: NewSim(), maxTicks: 160})
}

// ─── Topology L2 ── animated messages on the edge ─────────────────────────────

type topologyL2Model struct {
	sim      *Sim
	ticks    int
	maxTicks int
	// edge animation: particles on pub→queue edge and queue→con edge
	pubEdgeDots []float64
	conEdgeDots []float64
}

func (m topologyL2Model) Init() tea.Cmd { return doTick() }

func (m topologyL2Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg.(type) {
	case tea.KeyMsg:
		return m, tea.Quit
	case tickMsg:
		m.sim.Step()
		m.ticks++
		if m.ticks >= m.maxTicks {
			return m, tea.Quit
		}

		// Add dots to pub→queue edge based on pub rate
		if m.sim.PubRate > 1 {
			m.pubEdgeDots = append(m.pubEdgeDots, 0.0)
		}
		// Add dots to queue→con edge based on con rate
		if m.sim.ConRate > 1 && m.sim.QueueDepth > 0 {
			m.conEdgeDots = append(m.conEdgeDots, 0.0)
		}

		// Advance
		alive := m.pubEdgeDots[:0]
		for _, d := range m.pubEdgeDots {
			d += 0.12
			if d < 1.05 {
				alive = append(alive, d)
			}
		}
		m.pubEdgeDots = alive

		alive2 := m.conEdgeDots[:0]
		for _, d := range m.conEdgeDots {
			d += 0.12
			if d < 1.05 {
				alive2 = append(alive2, d)
			}
		}
		m.conEdgeDots = alive2

		if len(m.pubEdgeDots) > 8 {
			m.pubEdgeDots = m.pubEdgeDots[len(m.pubEdgeDots)-8:]
		}
		if len(m.conEdgeDots) > 8 {
			m.conEdgeDots = m.conEdgeDots[len(m.conEdgeDots)-8:]
		}

		return m, doTick()
	}
	return m, nil
}

func renderEdge(dots []float64, width int, dotColor string) string {
	cells := make([]rune, width)
	for i := range cells {
		cells[i] = '─'
	}
	for _, d := range dots {
		col := int(d * float64(width-1))
		if col >= 0 && col < width {
			cells[col] = '●'
		}
	}
	var sb strings.Builder
	for i, ch := range cells {
		if ch == '●' {
			sb.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color(dotColor)).Render(string(ch)))
		} else {
			_ = i
			sb.WriteString(styleDim().Render(string(ch)))
		}
	}
	return sb.String()
}

func (m topologyL2Model) View() string {
	s := m.sim
	const edgeW = 14

	pubColor := colBorder
	if s.PubRate > 2 {
		pubColor = colPub
	}
	conColor := colBorder
	if s.ConRate > 2 {
		conColor = colCon
	}
	qColor := colQueue
	if s.QueueFill() >= 0.85 {
		qColor = colDrop
	}

	spin := spinFrames[m.ticks%len(spinFrames)]

	pubSpin := styleDim().Render("·")
	if s.PubRate > 2 {
		pubSpin = stylePub().Render(spin)
	}
	conSpin := styleDim().Render("·")
	if s.ConRate > 2 {
		conSpin = styleCon().Render(spin)
	}

	pubNode := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color(pubColor)).
		Padding(0, 2).
		Width(16).
		Align(lipgloss.Center).
		Render(
			pubSpin + " " + lipgloss.NewStyle().Foreground(lipgloss.Color(pubColor)).Bold(true).Render("node-a") + "\n" +
				styleDim().Render("publisher") + "\n" +
				stylePub().Render(fmt.Sprintf("↑ %.1f/s", s.PubRate)))

	queueNode := lipgloss.NewStyle().
		Border(lipgloss.ThickBorder()).
		BorderForeground(lipgloss.Color(qColor)).
		Padding(0, 2).
		Width(18).
		Align(lipgloss.Center).
		Render(
			lipgloss.NewStyle().Foreground(lipgloss.Color(qColor)).Bold(true).Render("task_queue") + "\n" +
				fillBarAuto(12, s.QueueFill()) + "\n" +
				lipgloss.NewStyle().Foreground(lipgloss.Color(qColor)).Render(
					fmt.Sprintf("%d/%d", s.QueueDepth, s.QueueCapacity)))

	conNode := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color(conColor)).
		Padding(0, 2).
		Width(16).
		Align(lipgloss.Center).
		Render(
			conSpin + " " + lipgloss.NewStyle().Foreground(lipgloss.Color(conColor)).Bold(true).Render("node-b") + "\n" +
				styleDim().Render("consumer") + "\n" +
				styleCon().Render(fmt.Sprintf("↓ %.1f/s", s.ConRate)))

	// Animated edges (horizontal strings)
	pubEdgeTop := styleDim().Render("publish")
	pubEdgeStr := renderEdge(m.pubEdgeDots, edgeW, colPub) + stylePub().Render("▶")
	conEdgeTop := styleDim().Render("consume")
	conEdgeStr := renderEdge(m.conEdgeDots, edgeW, colCon) + styleCon().Render("▶")

	// Compose row manually with vertical alignment
	pubH := lipgloss.Height(pubNode)
	queueH := lipgloss.Height(queueNode)
	conH := lipgloss.Height(conNode)
	maxH := pubH
	if queueH > maxH {
		maxH = queueH
	}
	if conH > maxH {
		maxH = conH
	}

	midRow := maxH / 2

	pubLines := strings.Split(pubNode, "\n")
	queueLines := strings.Split(queueNode, "\n")
	conLines := strings.Split(conNode, "\n")

	padLines := func(lines []string, h int) []string {
		w := 0
		for _, l := range lines {
			if lipgloss.Width(l) > w {
				w = lipgloss.Width(l)
			}
		}
		for len(lines) < h {
			lines = append(lines, strings.Repeat(" ", w))
		}
		return lines
	}

	pubLines = padLines(pubLines, maxH)
	queueLines = padLines(queueLines, maxH)
	conLines = padLines(conLines, maxH)

	var rows []string
	for r := 0; r < maxH; r++ {
		pubPart := ""
		if r < len(pubLines) {
			pubPart = pubLines[r]
		}
		qPart := ""
		if r < len(queueLines) {
			qPart = queueLines[r]
		}
		conPart := ""
		if r < len(conLines) {
			conPart = conLines[r]
		}

		edge1 := strings.Repeat(" ", edgeW+1)
		edge2 := strings.Repeat(" ", edgeW+1)
		if r == midRow {
			edge1 = pubEdgeStr
			edge2 = conEdgeStr
		} else if r == midRow-1 {
			pubW := lipgloss.Width(pubEdgeTop)
			edge1 = strings.Repeat(" ", (edgeW+1-pubW)/2) + pubEdgeTop
			conW := lipgloss.Width(conEdgeTop)
			edge2 = strings.Repeat(" ", (edgeW+1-conW)/2) + conEdgeTop
		}

		rows = append(rows, pubPart+edge1+qPart+edge2+conPart)
	}

	title := gradientStr("  Network Topology  ·  Live  ·  Animated", colPub, colCon)
	return "\n" + title + "\n\n" + strings.Join(rows, "\n") + "\n"
}

func runTopologyL2() {
	runProgram(topologyL2Model{sim: NewSim(), maxTicks: 200})
}

// ─── Topology L3 ── full graph with per-node metrics and animated edges ────────

type topologyL3Model struct {
	sim         *Sim
	ticks       int
	maxTicks    int
	pubEdgeDots []float64
	conEdgeDots []float64
}

func (m topologyL3Model) Init() tea.Cmd { return doTick() }

func (m topologyL3Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg.(type) {
	case tea.KeyMsg:
		return m, tea.Quit
	case tickMsg:
		m.sim.Step()
		m.ticks++
		if m.ticks >= m.maxTicks {
			return m, tea.Quit
		}

		nPubDots := int(m.sim.PubRate)
		for i := 0; i < nPubDots; i++ {
			m.pubEdgeDots = append(m.pubEdgeDots, float64(i)/float64(nPubDots+1)*0.3)
		}
		nConDots := int(m.sim.ConRate)
		if m.sim.QueueDepth > 0 {
			for i := 0; i < nConDots; i++ {
				m.conEdgeDots = append(m.conEdgeDots, float64(i)/float64(nConDots+1)*0.3)
			}
		}

		alive := m.pubEdgeDots[:0]
		for _, d := range m.pubEdgeDots {
			d += 0.08
			if d < 1.05 {
				alive = append(alive, d)
			}
		}
		m.pubEdgeDots = alive

		alive2 := m.conEdgeDots[:0]
		for _, d := range m.conEdgeDots {
			d += 0.08
			if d < 1.05 {
				alive2 = append(alive2, d)
			}
		}
		m.conEdgeDots = alive2

		if len(m.pubEdgeDots) > 12 {
			m.pubEdgeDots = m.pubEdgeDots[len(m.pubEdgeDots)-12:]
		}
		if len(m.conEdgeDots) > 12 {
			m.conEdgeDots = m.conEdgeDots[len(m.conEdgeDots)-12:]
		}

		return m, doTick()
	}
	return m, nil
}

func (m topologyL3Model) View() string {
	s := m.sim
	const edgeW = 16
	spin := spinFrames[m.ticks%len(spinFrames)]

	pubActive := s.PubRate > 1.5
	conActive := s.ConRate > 1.5

	pubBorder := colBorder
	if pubActive {
		pubBorder = colPub
	}
	conBorder := colBorder
	if conActive {
		conBorder = colCon
	}
	qColor := colQueue
	if s.QueueFill() >= 0.85 {
		qColor = colDrop
	}

	pubSpin := styleDim().Render("·")
	if pubActive {
		pubSpin = stylePub().Render(spin)
	}
	conSpin := styleDim().Render("·")
	if conActive {
		conSpin = styleCon().Render(spin)
	}

	pubPubSpark := stylePub().Render(sparklineAuto(s.PubHist, 14))

	pubPanel := lipgloss.NewStyle().
		Border(lipgloss.DoubleBorder()).
		BorderForeground(lipgloss.Color(pubBorder)).
		Padding(0, 2).
		Width(22).
		Render(
			pubSpin + " " + stylePub().Bold(true).Render("node-a  PUBLISHER") + "\n" +
				styleDim().Render("amqp://localhost:5672") + "\n" +
				styleDim().Render("exchange: amq.default") + "\n\n" +
				styleDim().Render("rate    ") + stylePub().Render(fmt.Sprintf("%.2f/s", s.PubRate)) + "\n" +
				styleDim().Render("total   ") + styleText().Render(fmt.Sprintf("%d", s.TotalPub)) + "\n" +
				styleDim().Render("dropped ") + styleDrop().Render(fmt.Sprintf("%d", s.TotalDrop)) + "\n\n" +
				styleDim().Render("rate history") + "\n" +
				pubPubSpark)

	qFill := s.QueueFill()
	qBar := fillBarAuto(16, qFill)
	depthF := make([]float64, len(s.DepthHist))
	for i, v := range s.DepthHist {
		depthF[i] = float64(v)
	}
	depthSpark := styleQueue().Render(sparklineAuto(depthF, 16))

	queuePanel := lipgloss.NewStyle().
		Border(lipgloss.ThickBorder()).
		BorderForeground(lipgloss.Color(qColor)).
		Padding(0, 2).
		Width(24).
		Render(
			lipgloss.NewStyle().Foreground(lipgloss.Color(qColor)).Bold(true).Render("QUEUE  task_queue") + "\n" +
				styleDim().Render("amq.default / task_queue") + "\n\n" +
				qBar + "\n" +
				lipgloss.NewStyle().Foreground(lipgloss.Color(qColor)).
					Render(fmt.Sprintf("%d/%d  %d%%", s.QueueDepth, s.QueueCapacity, int(qFill*100))) + "\n\n" +
				styleDim().Render("depth history") + "\n" +
				depthSpark + "\n\n" +
				styleDim().Render("pressure  ") +
				func() string {
					d := s.PubRate - s.ConRate
					if d > 0.5 {
						return styleDrop().Render(fmt.Sprintf("▲ +%.1f", d))
					} else if d < -0.5 {
						return styleCon().Render(fmt.Sprintf("▼ −%.1f", -d))
					}
					return styleMsg().Render("◆ stable")
				}())

	conSpark := styleCon().Render(sparklineAuto(s.ConHist, 14))
	avgLat := s.AvgLatency(20)

	conPanel := lipgloss.NewStyle().
		Border(lipgloss.DoubleBorder()).
		BorderForeground(lipgloss.Color(conBorder)).
		Padding(0, 2).
		Width(22).
		Render(
			conSpin + " " + styleCon().Bold(true).Render("node-b  CONSUMER") + "\n" +
				styleDim().Render("amqp://localhost:5672") + "\n" +
				styleDim().Render("queue: task_queue") + "\n\n" +
				styleDim().Render("rate    ") + styleCon().Render(fmt.Sprintf("%.2f/s", s.ConRate)) + "\n" +
				styleDim().Render("total   ") + styleText().Render(fmt.Sprintf("%d", s.TotalCon)) + "\n" +
				styleDim().Render("avg lat ") + styleMsg().Render(fmt.Sprintf("%.0fms", avgLat)) + "\n\n" +
				styleDim().Render("rate history") + "\n" +
				conSpark)

	// Animated edges between panels
	pubEdgeStr := renderEdge(m.pubEdgeDots, edgeW, colPub) + stylePub().Render("▶")
	conEdgeStr := renderEdge(m.conEdgeDots, edgeW, colCon) + styleCon().Render("▶")

	// Compose with edges at mid-height
	pubLines := strings.Split(pubPanel, "\n")
	queueLines := strings.Split(queuePanel, "\n")
	conLines := strings.Split(conPanel, "\n")

	maxH := len(pubLines)
	if len(queueLines) > maxH {
		maxH = len(queueLines)
	}
	if len(conLines) > maxH {
		maxH = len(conLines)
	}

	pad := func(lines []string, h int) []string {
		w := 0
		for _, l := range lines {
			if lw := lipgloss.Width(l); lw > w {
				w = lw
			}
		}
		for len(lines) < h {
			lines = append(lines, strings.Repeat(" ", w))
		}
		return lines
	}

	pubLines = pad(pubLines, maxH)
	queueLines = pad(queueLines, maxH)
	conLines = pad(conLines, maxH)
	mid := maxH / 2

	// Edge annotations
	pubEdgeLabel := styleDim().Render("publish →")
	conEdgeLabel := styleDim().Render("consume →")
	pubRateLabel := stylePub().Render(fmt.Sprintf("%.1f/s", s.PubRate))
	conRateLabel := styleCon().Render(fmt.Sprintf("%.1f/s", s.ConRate))

	var rows []string
	for r := 0; r < maxH; r++ {
		pL := pubLines[r]
		qL := queueLines[r]
		cL := conLines[r]

		e1 := strings.Repeat(" ", edgeW+1)
		e2 := strings.Repeat(" ", edgeW+1)

		switch r {
		case mid - 1:
			e1 = center(pubEdgeLabel, edgeW) + " "
			e2 = center(conEdgeLabel, edgeW) + " "
		case mid:
			e1 = pubEdgeStr
			e2 = conEdgeStr
		case mid + 1:
			e1 = center(pubRateLabel, edgeW) + " "
			e2 = center(conRateLabel, edgeW) + " "
		}

		rows = append(rows, pL+e1+qL+e2+cL)
	}

	title := gradientStr("  Network Topology  ·  Full Graph  ·  Live", colPub, colCon)

	// Angle-based edge path label
	angleStr := styleDim().Render(fmt.Sprintf(
		"  ∠ pub→queue latency %.0fms  ·  ∠ queue pressure %d%%",
		s.AvgLatency(10),
		int(math.Round(s.QueueFill()*100))))

	return "\n" + title + "\n\n" + strings.Join(rows, "\n") + "\n" + angleStr + "\n"
}

func runTopologyL3() {
	runProgram(topologyL3Model{sim: NewSim(), maxTicks: 220})
}
