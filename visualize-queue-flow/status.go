package main

import (
	"fmt"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// ─── Status L1 ── two simple bordered boxes ───────────────────────────────────

type statusL1Model struct {
	sim      *Sim
	ticks    int
	maxTicks int
}

func (m statusL1Model) Init() tea.Cmd { return doTick() }

func (m statusL1Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
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

func (m statusL1Model) View() string {
	s := m.sim

	pubState := "● IDLE"
	if s.PubRate > 2 {
		pubState = "▲ PUBLISHING"
	}

	conState := "● IDLE"
	if s.ConRate > 2 {
		conState = "▼ CONSUMING"
	}

	pubBox := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color(colPub)).
		Padding(1, 3).
		Width(28).
		Render(
			stylePub().Bold(true).Render("PUBLISHER") + "\n" +
				styleDim().Render("node-a  ·  amqp://localhost") + "\n\n" +
				lipgloss.NewStyle().Foreground(lipgloss.Color(colPub)).Render(pubState) + "\n\n" +
				styleDim().Render("exchange: ") + styleText().Render("amq.default") + "\n" +
				styleDim().Render("sent:     ") + styleText().Render(fmt.Sprintf("%d messages", s.TotalPub)) + "\n" +
				styleDim().Render("rate:     ") + stylePub().Render(fmt.Sprintf("%.1f msg/s", s.PubRate)))

	conBox := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color(colCon)).
		Padding(1, 3).
		Width(28).
		Render(
			styleCon().Bold(true).Render("CONSUMER") + "\n" +
				styleDim().Render("node-b  ·  amqp://localhost") + "\n\n" +
				lipgloss.NewStyle().Foreground(lipgloss.Color(colCon)).Render(conState) + "\n\n" +
				styleDim().Render("queue:    ") + styleText().Render("task_queue") + "\n" +
				styleDim().Render("received: ") + styleText().Render(fmt.Sprintf("%d messages", s.TotalCon)) + "\n" +
				styleDim().Render("rate:     ") + styleCon().Render(fmt.Sprintf("%.1f msg/s", s.ConRate)))

	queueBox := lipgloss.NewStyle().
		Border(lipgloss.NormalBorder()).
		BorderForeground(lipgloss.Color(colQueue)).
		Padding(0, 2).
		Align(lipgloss.Center).
		Width(16).
		Render(
			styleQueue().Bold(true).Render("QUEUE") + "\n" +
				styleQueue().Render(fmt.Sprintf("%d / %d", s.QueueDepth, s.QueueCapacity)) + "\n" +
				fillBarAuto(10, s.QueueFill()) + "\n" +
				styleDrop().Render(fmt.Sprintf("dropped: %d", s.TotalDrop)))

	row := lipgloss.JoinHorizontal(lipgloss.Center,
		pubBox, "  ", queueBox, "  ", conBox)

	title := gradientStr("  App Status  ·  RabbitMQ Pipeline", colPub, colCon)
	return "\n" + title + "\n\n" + row + "\n"
}

func runStatusL1() {
	runProgram(statusL1Model{sim: NewSim(), maxTicks: 160})
}

// ─── Status L2 ── animated spinners + live counters ───────────────────────────

var spinFrames = []string{"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏"}

type statusL2Model struct {
	sim      *Sim
	ticks    int
	maxTicks int
}

func (m statusL2Model) Init() tea.Cmd { return doTick() }

func (m statusL2Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
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

func (m statusL2Model) View() string {
	s := m.sim
	spin := spinFrames[m.ticks%len(spinFrames)]

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

	// Animated rate bar (width varies with rate)
	maxRate := 6.0
	pubBarW := int(s.PubRate / maxRate * 18)
	if pubBarW > 18 {
		pubBarW = 18
	}
	conBarW := int(s.ConRate / maxRate * 18)
	if conBarW > 18 {
		conBarW = 18
	}
	pubBar := stylePub().Render(strings.Repeat("▰", pubBarW)) + styleDim().Render(strings.Repeat("▱", 18-pubBarW))
	conBar := styleCon().Render(strings.Repeat("▰", conBarW)) + styleDim().Render(strings.Repeat("▱", 18-conBarW))

	pubStateStr := styleDim().Render("IDLE")
	if pubActive {
		pubStateStr = stylePub().Render("PUBLISHING")
	}

	conStateStr := styleDim().Render("IDLE")
	if conActive {
		conStateStr = styleCon().Render("CONSUMING")
	}

	pubBox := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color(func() string {
			if pubActive {
				return colPub
			}
			return colBorder
		}())).
		Padding(1, 2).
		Width(30).
		Render(
			pubSpin + "  " + stylePub().Bold(true).Render("PUBLISHER") + "\n" +
				styleDim().Render("node-a") + "\n\n" +
				pubStateStr + "\n\n" +
				styleDim().Render("rate  ") + pubBar + "\n" +
				styleDim().Render("      ") + stylePub().Render(fmt.Sprintf("%.2f msg/s", s.PubRate)) + "\n\n" +
				styleDim().Render("sent:  ") + stylePub().Bold(true).Render(fmt.Sprintf("%d", s.TotalPub)))

	conBox := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color(func() string {
			if conActive {
				return colCon
			}
			return colBorder
		}())).
		Padding(1, 2).
		Width(30).
		Render(
			conSpin + "  " + styleCon().Bold(true).Render("CONSUMER") + "\n" +
				styleDim().Render("node-b") + "\n\n" +
				conStateStr + "\n\n" +
				styleDim().Render("rate  ") + conBar + "\n" +
				styleDim().Render("      ") + styleCon().Render(fmt.Sprintf("%.2f msg/s", s.ConRate)) + "\n\n" +
				styleDim().Render("recv:  ") + styleCon().Bold(true).Render(fmt.Sprintf("%d", s.TotalCon)))

	// Queue center
	qFill := s.QueueFill()
	qColor := colQueue
	if qFill >= 0.85 {
		qColor = colDrop
	}
	qBar := fillBarAuto(12, qFill)
	queuePanel := lipgloss.NewStyle().
		Border(lipgloss.DoubleBorder()).
		BorderForeground(lipgloss.Color(qColor)).
		Padding(0, 2).
		Width(20).
		Align(lipgloss.Center).
		Render(
			lipgloss.NewStyle().Foreground(lipgloss.Color(qColor)).Bold(true).Render("QUEUE") + "\n" +
				styleDim().Render("task_queue") + "\n\n" +
				qBar + "\n" +
				lipgloss.NewStyle().Foreground(lipgloss.Color(qColor)).
					Render(fmt.Sprintf("%d / %d", s.QueueDepth, s.QueueCapacity)) + "\n\n" +
				styleDim().Render("dropped") + "\n" +
				styleDrop().Bold(true).Render(fmt.Sprintf("%d", s.TotalDrop)))

	row := lipgloss.JoinHorizontal(lipgloss.Center, pubBox, "  ", queuePanel, "  ", conBox)

	title := gradientStr("  Application Status  ·  Live", colPub, colCon)
	return "\n" + title + "\n\n" + row + "\n"
}

func runStatusL2() {
	runProgram(statusL2Model{sim: NewSim(), maxTicks: 200})
}

// ─── Status L3 ── rich panels with connection details and recent messages ──────

type statusL3Model struct {
	sim      *Sim
	ticks    int
	maxTicks int
}

func (m statusL3Model) Init() tea.Cmd { return doTick() }

func (m statusL3Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
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

func (m statusL3Model) View() string {
	s := m.sim
	spin := spinFrames[m.ticks%len(spinFrames)]

	// Build recent activity from events
	recentPubMsgs := []string{}
	recentConMsgs := []string{}
	for _, ev := range s.LastEvents(30) {
		ts := ev.Time.Format("15:04:05.000")
		switch ev.Type {
		case EvPublished:
			recentPubMsgs = append(recentPubMsgs, styleDim().Render(ts)+" "+stylePub().Render(fmt.Sprintf("#%d", ev.MsgID)))
		case EvConsumed:
			lat := fmt.Sprintf("%.0fms", float64(ev.Latency)/float64(time.Millisecond))
			recentConMsgs = append(recentConMsgs, styleDim().Render(ts)+" "+styleCon().Render(fmt.Sprintf("#%d", ev.MsgID))+" "+styleDim().Render(lat))
		}
	}
	// keep last 6 of each
	if len(recentPubMsgs) > 6 {
		recentPubMsgs = recentPubMsgs[len(recentPubMsgs)-6:]
	}
	if len(recentConMsgs) > 6 {
		recentConMsgs = recentConMsgs[len(recentConMsgs)-6:]
	}

	pubActive := s.PubRate > 1.5
	pubBorderCol := colBorder
	if pubActive {
		pubBorderCol = colPub
	}
	pubSpin := styleDim().Render("·")
	if pubActive {
		pubSpin = stylePub().Render(spin)
	}

	pubRecent := strings.Join(recentPubMsgs, "\n")
	if pubRecent == "" {
		pubRecent = styleDim().Render("  (no recent messages)")
	}

	pubBox := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color(pubBorderCol)).
		Padding(1, 2).
		Width(38).
		Render(
			pubSpin + "  " + stylePub().Bold(true).Render("PUBLISHER") + "  " + styleDim().Render("node-a") + "\n" +
				styleDim().Render("amqp://localhost:5672") + "\n" +
				styleDim().Render("exchange: amq.default  routing_key: task") + "\n\n" +

				styleDim().Render("status   ") + func() string {
				if pubActive {
					return stylePub().Render("● CONNECTED / PUBLISHING")
				}
				return styleDim().Render("● CONNECTED / IDLE")
			}() + "\n" +
				styleDim().Render("rate     ") + stylePub().Render(fmt.Sprintf("%.2f msg/s", s.PubRate)) + "\n" +
				styleDim().Render("sent     ") + styleText().Render(fmt.Sprintf("%d", s.TotalPub)) + "\n" +
				styleDim().Render("dropped  ") + styleDrop().Render(fmt.Sprintf("%d", s.TotalDrop)) + "\n\n" +
				styleDim().Render("recent publishes") + "\n" +
				pubRecent)

	conActive := s.ConRate > 1.5
	conBorderCol := colBorder
	if conActive {
		conBorderCol = colCon
	}
	conSpin := styleDim().Render("·")
	if conActive {
		conSpin = styleCon().Render(spin)
	}

	conRecent := strings.Join(recentConMsgs, "\n")
	if conRecent == "" {
		conRecent = styleDim().Render("  (no recent messages)")
	}

	avgLat := s.AvgLatency(20)

	conBox := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color(conBorderCol)).
		Padding(1, 2).
		Width(38).
		Render(
			conSpin + "  " + styleCon().Bold(true).Render("CONSUMER") + "  " + styleDim().Render("node-b") + "\n" +
				styleDim().Render("amqp://localhost:5672") + "\n" +
				styleDim().Render("queue: task_queue  prefetch: 5") + "\n\n" +

				styleDim().Render("status   ") + func() string {
				if conActive {
					return styleCon().Render("● CONNECTED / CONSUMING")
				}
				return styleDim().Render("● CONNECTED / IDLE")
			}() + "\n" +
				styleDim().Render("rate     ") + styleCon().Render(fmt.Sprintf("%.2f msg/s", s.ConRate)) + "\n" +
				styleDim().Render("received ") + styleText().Render(fmt.Sprintf("%d", s.TotalCon)) + "\n" +
				styleDim().Render("avg lat  ") + styleMsg().Render(fmt.Sprintf("%.1f ms", avgLat)) + "\n\n" +
				styleDim().Render("recent consumes") + "\n" +
				conRecent)

	// Center queue status
	qFill := s.QueueFill()
	qColor := colQueue
	if qFill >= 0.85 {
		qColor = colDrop
	}
	pubSpark := sparklineAuto(s.PubHist, 16)
	conSpark := sparklineAuto(s.ConHist, 16)

	queuePanel := lipgloss.NewStyle().
		Border(lipgloss.ThickBorder()).
		BorderForeground(lipgloss.Color(qColor)).
		Padding(0, 1).
		Width(20).
		Render(
			center(lipgloss.NewStyle().Foreground(lipgloss.Color(qColor)).Bold(true).Render("QUEUE"), 16) + "\n" +
				center(styleDim().Render("task_queue"), 16) + "\n\n" +
				fillBarAuto(14, qFill) + "\n" +
				lipgloss.NewStyle().Foreground(lipgloss.Color(qColor)).
					Render(fmt.Sprintf("%d/%d  %d%%", s.QueueDepth, s.QueueCapacity, int(qFill*100))) + "\n\n" +
				stylePub().Render("pub  ") + stylePub().Render(pubSpark) + "\n" +
				styleCon().Render("con  ") + styleCon().Render(conSpark) + "\n\n" +
				styleDim().Render("avg lat") + "\n" +
				styleMsg().Render(fmt.Sprintf("%.0f ms", avgLat)))

	row := lipgloss.JoinHorizontal(lipgloss.Top, pubBox, "  ", queuePanel, "  ", conBox)

	title := gradientStr("  Application Status  ·  Full Details  ·  Live", colPub, colCon)
	return "\n" + title + "\n\n" + row + "\n"
}

func runStatusL3() {
	runProgram(statusL3Model{sim: NewSim(), maxTicks: 200})
}
