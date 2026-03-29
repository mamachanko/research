package main

import (
	"fmt"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// ─── Full Dashboard ────────────────────────────────────────────────────────────
//
//  Layout (120×36):
//
//  ┌─ title bar ──────────────────────────────────────────────────────────────┐
//  │  PUBLISHER node-a        QUEUE task_queue        CONSUMER node-b        │
//  │                                                                          │
//  │  ┌── queue depth gauge ──────────────────────────────────────────────┐  │
//  │  │  [████████████████████░░░░░░░░░░░░]  12/20  60%  ▲ FILLING        │  │
//  │  └───────────────────────────────────────────────────────────────────┘  │
//  │                                                                          │
//  │  ┌── flow ─────────────────────────────────────────────────────────┐   │
//  │  │  PUB ●──●──────[▓▓▓▓▓░░░░░░]──●──── CON                        │   │
//  │  └─────────────────────────────────────────────────────────────────┘   │
//  │                                                                          │
//  │  ┌── pub panel ─────┐  ┌── throughput ────────┐  ┌── events ─────────┐ │
//  │  │  PUBLISHER       │  │  ▲ pub ▼ con          │  │  ▲ PUB  #0034    │ │
//  │  │  rate: 3.2/s     │  │  ██  █  ██  ██       │  │  → QUE  #0034    │ │
//  │  │  sent: 120       │  │  ██  █  ██  ██       │  │  ▼ CON  #0031    │ │
//  │  │  CONSUMER        │  │  ...                 │  │  ...              │ │
//  │  │  rate: 2.8/s     │  └──────────────────────┘  └───────────────────┘ │
//  │  └──────────────────┘                                                   │
//  └──────────────────────────────────────────────────────────────────────────┘

type dashboardModel struct {
	sim         *Sim
	ticks       int
	maxTicks    int
	w, h        int
	pubEdgeDots []float64
	conEdgeDots []float64
	entries     []tlEntry
}

func (m dashboardModel) Init() tea.Cmd { return doTick() }

func (m dashboardModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
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

		// Edge dots
		if m.sim.PubRate > 1 {
			m.pubEdgeDots = append(m.pubEdgeDots, 0.0)
		}
		if m.sim.ConRate > 1 && m.sim.QueueDepth > 0 {
			m.conEdgeDots = append(m.conEdgeDots, 0.0)
		}
		alive := m.pubEdgeDots[:0]
		for _, d := range m.pubEdgeDots {
			d += 0.15
			if d < 1.05 {
				alive = append(alive, d)
			}
		}
		m.pubEdgeDots = alive
		alive2 := m.conEdgeDots[:0]
		for _, d := range m.conEdgeDots {
			d += 0.15
			if d < 1.05 {
				alive2 = append(alive2, d)
			}
		}
		m.conEdgeDots = alive2
		if len(m.pubEdgeDots) > 6 {
			m.pubEdgeDots = m.pubEdgeDots[len(m.pubEdgeDots)-6:]
		}
		if len(m.conEdgeDots) > 6 {
			m.conEdgeDots = m.conEdgeDots[len(m.conEdgeDots)-6:]
		}

		// Event feed
		for _, ev := range m.sim.LastEvents(6) {
			already := false
			for _, e := range m.entries {
				if e.event.MsgID == ev.MsgID && e.event.Type == ev.Type {
					already = true
					break
				}
			}
			if !already {
				m.entries = append(m.entries, tlEntry{ev, m.ticks})
			}
		}
		if len(m.entries) > 30 {
			m.entries = m.entries[len(m.entries)-30:]
		}

		return m, doTick()
	}
	return m, nil
}

func (m dashboardModel) View() string {
	s := m.sim
	spin := spinFrames[m.ticks%len(spinFrames)]
	totalW := 116

	// ── Title bar ────────────────────────────────────────────────────────────
	pubActive := s.PubRate > 1.5
	conActive := s.ConRate > 1.5

	pubSpin := styleDim().Render("·")
	if pubActive {
		pubSpin = stylePub().Render(spin)
	}
	conSpin := styleDim().Render("·")
	if conActive {
		conSpin = styleCon().Render(spin)
	}

	qColor := colQueue
	if s.QueueFill() >= 0.85 {
		qColor = colDrop
	}

	titleContent := pubSpin + " " + stylePub().Bold(true).Render("node-a  PUBLISHER") +
		styleDim().Render("  ←  amq.default / task_queue  →  ") +
		styleCon().Bold(true).Render("CONSUMER  node-b") + " " + conSpin

	titleBar := lipgloss.NewStyle().
		Background(lipgloss.Color("#1e1f29")).
		Foreground(lipgloss.Color(colText)).
		Width(totalW).
		Padding(0, 2).
		Render(gradientStr("  mqvis dashboard  ·  ", colPub, colCon) +
			styleDim().Render("RabbitMQ pipeline monitor") +
			strings.Repeat(" ", 30) + titleContent)

	// ── Queue depth gauge (full width) ───────────────────────────────────────
	fill := s.QueueFill()
	barColor := colCon
	if fill >= 0.85 {
		barColor = colDrop
	} else if fill >= 0.55 {
		barColor = colYellow
	}
	gaugeW := totalW - 8
	bar := fillBar(gaugeW, fill, barColor, colBorder)
	pct := int(fill * 100)

	var fillStatus string
	diff := s.PubRate - s.ConRate
	if diff > 0.5 {
		fillStatus = styleDrop().Render(fmt.Sprintf("  ▲ filling (+%.1f/tick)", diff))
	} else if diff < -0.5 {
		fillStatus = styleCon().Render(fmt.Sprintf("  ▼ draining (−%.1f/tick)", -diff))
	} else {
		fillStatus = styleMsg().Render("  ◆ stable")
	}

	depthF := make([]float64, len(s.DepthHist))
	for i, v := range s.DepthHist {
		depthF[i] = float64(v)
	}
	depthSpark := lipgloss.NewStyle().Foreground(lipgloss.Color(barColor)).
		Render(sparklineAuto(depthF, gaugeW))

	gaugeContent := styleDim().Render("  queue depth  ") +
		lipgloss.NewStyle().Foreground(lipgloss.Color(barColor)).Bold(true).
			Render(fmt.Sprintf("%d/%d  %d%%", s.QueueDepth, s.QueueCapacity, pct)) +
		fillStatus + "\n" +
		"  " + bar + "\n" +
		"  " + depthSpark

	gaugeBox := lipgloss.NewStyle().
		Border(lipgloss.NormalBorder()).
		BorderForeground(lipgloss.Color(barColor)).
		Padding(0, 1).
		Width(totalW - 2).
		Render(gaugeContent)

	// ── Flow animation (full width) ───────────────────────────────────────────
	flowW := totalW - 6
	const qL, qR, conS = 0.32, 0.68, 0.88

	cells := make([]rune, flowW)
	cellColors := make([]string, flowW)
	for i := range cells {
		cells[i] = ' '
	}

	qColL := int(qL * float64(flowW))
	qColR := int(qR * float64(flowW))
	qFillN := s.QueueDepth * (qColR - qColL) / s.QueueCapacity

	for i := qColL; i < qColR; i++ {
		relPos := i - qColL
		if relPos < qFillN {
			cells[i] = '▓'
			cellColors[i] = qColor
		} else {
			cells[i] = '░'
			cellColors[i] = colBorder
		}
	}

	for _, p := range s.Particles {
		col := int(p.Pos * float64(flowW-1))
		if col < 0 || col >= flowW {
			continue
		}
		c := colMsg
		if p.Pos < qL {
			c = colPub
		} else if p.InQueue {
			c = qColor
		} else if p.Pos > conS {
			c = colCon
		}
		cells[col] = p.Char
		cellColors[col] = c
	}

	var flowSB strings.Builder
	for i, ch := range cells {
		if cellColors[i] != "" {
			flowSB.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color(cellColors[i])).Render(string(ch)))
		} else {
			flowSB.WriteRune(ch)
		}
	}

	pubLabel := stylePub().Render(fmt.Sprintf("PUB %.1f/s", s.PubRate))
	qLabel := lipgloss.NewStyle().Foreground(lipgloss.Color(qColor)).Render(fmt.Sprintf("QUEUE %d/%d", s.QueueDepth, s.QueueCapacity))
	conLabel := styleCon().Render(fmt.Sprintf("%.1f/s CON", s.ConRate))

	qLabelStart := qColL + (qColR-qColL)/2 - lipgloss.Width(qLabel)/2
	labelRow := pubLabel +
		strings.Repeat(" ", qLabelStart-lipgloss.Width(pubLabel)) +
		qLabel +
		strings.Repeat(" ", flowW-qLabelStart-lipgloss.Width(qLabel)-lipgloss.Width(conLabel)) +
		conLabel

	flowBox := lipgloss.NewStyle().
		Border(lipgloss.NormalBorder()).
		BorderForeground(lipgloss.Color(colBorder)).
		Padding(0, 1).
		Width(totalW - 2).
		Render(labelRow + "\n" + flowSB.String())

	// ── Bottom row: status | throughput | events ──────────────────────────────
	const statusW = 28
	const throughW = 48
	const eventsW = 34

	// Status panel
	pubBarW := int(s.PubRate / 6.0 * 16)
	if pubBarW > 16 {
		pubBarW = 16
	}
	conBarW := int(s.ConRate / 6.0 * 16)
	if conBarW > 16 {
		conBarW = 16
	}
	pubRateBar := stylePub().Render(strings.Repeat("▰", pubBarW)) + styleDim().Render(strings.Repeat("▱", 16-pubBarW))
	conRateBar := styleCon().Render(strings.Repeat("▰", conBarW)) + styleDim().Render(strings.Repeat("▱", 16-conBarW))

	pubStatSpin := styleDim().Render("·")
	if pubActive {
		pubStatSpin = stylePub().Render(spin)
	}
	conStatSpin := styleDim().Render("·")
	if conActive {
		conStatSpin = styleCon().Render(spin)
	}

	statusContent :=
		pubStatSpin + " " + stylePub().Bold(true).Render("PUBLISHER  node-a") + "\n" +
			styleDim().Render("rate  ") + pubRateBar + "\n" +
			styleDim().Render("      ") + stylePub().Render(fmt.Sprintf("%.2f/s", s.PubRate)) + "\n" +
			styleDim().Render("sent  ") + stylePub().Bold(true).Render(fmt.Sprintf("%d", s.TotalPub)) + "\n" +
			styleDim().Render("drop  ") + styleDrop().Render(fmt.Sprintf("%d", s.TotalDrop)) + "\n\n" +
			conStatSpin + " " + styleCon().Bold(true).Render("CONSUMER   node-b") + "\n" +
			styleDim().Render("rate  ") + conRateBar + "\n" +
			styleDim().Render("      ") + styleCon().Render(fmt.Sprintf("%.2f/s", s.ConRate)) + "\n" +
			styleDim().Render("recv  ") + styleCon().Bold(true).Render(fmt.Sprintf("%d", s.TotalCon)) + "\n" +
			styleDim().Render("lat   ") + styleMsg().Render(fmt.Sprintf("%.0fms avg", s.AvgLatency(20)))

	statusBox := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color(colBorder)).
		Padding(0, 1).
		Width(statusW).
		Render(statusContent)

	// Throughput dual chart
	const tChartW = throughW - 4
	const tChartH = 9
	dChart := dualBarChart(s.PubHist, s.ConHist, tChartW, tChartH, colPub, colCon)
	dLines := strings.Split(dChart, "\n")
	var dRendered strings.Builder
	for _, l := range dLines {
		dRendered.WriteString("  " + l + "\n")
	}
	dRendered.WriteString(styleDim().Render("  " + strings.Repeat("─", tChartW)) + "\n")
	dRendered.WriteString(
		stylePub().Render("  █ pub") + styleDim().Render("  ") +
			styleCon().Render("█ con") + styleDim().Render("  ") +
			stylePub().Render(fmt.Sprintf("%.1f/s", s.PubRate)) +
			styleDim().Render(" vs ") +
			styleCon().Render(fmt.Sprintf("%.1f/s", s.ConRate)))

	throughputBox := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color(colBorder)).
		Padding(0, 1).
		Width(throughW).
		Render(styleDim().Render("throughput") + "\n" + dRendered.String())

	// Event feed
	var evLines []string
	for _, e := range m.entries {
		icon, label, st := eventStyle(e.event.Type)
		ts := e.event.Time.Format("15:04:05")
		latStr := ""
		if e.event.Type == EvConsumed && e.event.Latency > 0 {
			latStr = " " + styleDim().Render(fmt.Sprintf("%.0fms", float64(e.event.Latency)/float64(time.Millisecond)))
		}
		age := m.ticks - e.ageTick
		if age > 20 {
			st = styleDim()
		}
		line := st.Render(fmt.Sprintf("%s %-3s", icon, label)) +
			styleDim().Render(" "+ts) +
			st.Render(fmt.Sprintf(" #%04d", e.event.MsgID)) + latStr
		evLines = append(evLines, line)
	}
	if len(evLines) > 11 {
		evLines = evLines[len(evLines)-11:]
	}

	evContent := styleDim().Render("event feed") + "\n" + strings.Join(evLines, "\n")
	eventsBox := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color(colBorder)).
		Padding(0, 1).
		Width(eventsW).
		Render(evContent)

	bottomRow := lipgloss.JoinHorizontal(lipgloss.Top,
		statusBox, "  ", throughputBox, "  ", eventsBox)

	return titleBar + "\n" +
		gaugeBox + "\n" +
		flowBox + "\n" +
		bottomRow + "\n"
}

func runDashboard() {
	runProgram(dashboardModel{sim: NewSim(), maxTicks: 240})
}
