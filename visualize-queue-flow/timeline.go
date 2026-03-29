package main

import (
	"fmt"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// ─── Timeline L1 ── static colored event log ──────────────────────────────────

type timelineL1Model struct {
	sim      *Sim
	ticks    int
	maxTicks int
}

func (m timelineL1Model) Init() tea.Cmd { return doTick() }

func (m timelineL1Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
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

func eventStyle(et EventType) (icon, label string, s lipgloss.Style) {
	switch et {
	case EvPublished:
		return "▲", "PUB", stylePub()
	case EvQueued:
		return "→", "QUE", styleQueue()
	case EvConsumed:
		return "▼", "CON", styleCon()
	case EvDropped:
		return "✗", "DRP", styleDrop()
	}
	return "?", "???", styleDim()
}

func (m timelineL1Model) View() string {
	s := m.sim
	events := s.LastEvents(16)

	var lines []string
	for _, ev := range events {
		icon, label, st := eventStyle(ev.Type)
		ts := ev.Time.Format("15:04:05.000")
		msg := st.Render(fmt.Sprintf("%s %-3s", icon, label)) +
			styleDim().Render("  "+ts) +
			styleText().Render(fmt.Sprintf("  msg#%04d", ev.MsgID))
		if ev.Type == EvConsumed && ev.Latency > 0 {
			msg += styleDim().Render(fmt.Sprintf("  lat: %.0fms", float64(ev.Latency)/float64(time.Millisecond)))
		}
		lines = append(lines, "  "+msg)
	}

	evLog := strings.Join(lines, "\n")
	if evLog == "" {
		evLog = styleDim().Render("  waiting for events...")
	}

	// Summary counts
	nPub, nCon, nDrop := 0, 0, 0
	for _, ev := range s.Events {
		switch ev.Type {
		case EvPublished:
			nPub++
		case EvConsumed:
			nCon++
		case EvDropped:
			nDrop++
		}
	}

	summary := stylePub().Render(fmt.Sprintf("▲ %d published", nPub)) + "  " +
		styleCon().Render(fmt.Sprintf("▼ %d consumed", nCon)) + "  " +
		styleDrop().Render(fmt.Sprintf("✗ %d dropped", nDrop))

	title := gradientStr("  Event Timeline  ·  task_queue", colPub, colCon)

	box := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color(colBorder)).
		Padding(1, 1).
		Width(62).
		Render(evLog + "\n\n" + "  " + summary)

	return "\n" + title + "\n\n" + box + "\n"
}

func runTimelineL1() {
	runProgram(timelineL1Model{sim: NewSim(), maxTicks: 160})
}

// ─── Timeline L2 ── streaming feed with fade-by-age ───────────────────────────

type timelineL2Model struct {
	sim      *Sim
	ticks    int
	maxTicks int
	entries  []tlEntry
}

type tlEntry struct {
	event   Event
	ageTick int // tick when added
}

func (m timelineL2Model) Init() tea.Cmd { return doTick() }

func (m timelineL2Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg.(type) {
	case tea.KeyMsg:
		return m, tea.Quit
	case tickMsg:
		m.sim.Step()
		m.ticks++
		if m.ticks >= m.maxTicks {
			return m, tea.Quit
		}
		// Add new events to the feed
		for _, ev := range m.sim.LastEvents(5) {
			// Avoid duplicate entries
			alreadyIn := false
			for _, e := range m.entries {
				if e.event.MsgID == ev.MsgID && e.event.Type == ev.Type {
					alreadyIn = true
					break
				}
			}
			if !alreadyIn {
				m.entries = append(m.entries, tlEntry{ev, m.ticks})
			}
		}
		// Keep last 20 entries
		if len(m.entries) > 20 {
			m.entries = m.entries[len(m.entries)-20:]
		}
		return m, doTick()
	}
	return m, nil
}

func (m timelineL2Model) View() string {
	const maxAge = 30 // ticks before fully dim

	var lines []string
	for _, e := range m.entries {
		age := m.ticks - e.ageTick
		icon, label, st := eventStyle(e.event.Type)
		ts := e.event.Time.Format("15:04:05.000")

		line := st.Render(fmt.Sprintf("%s %-3s", icon, label)) +
			styleDim().Render("  "+ts) +
			styleText().Render(fmt.Sprintf("  msg#%04d", e.event.MsgID))
		if e.event.Type == EvConsumed && e.event.Latency > 0 {
			line += styleDim().Render(fmt.Sprintf("  %.0fms", float64(e.event.Latency)/float64(time.Millisecond)))
		}

		// Age fade indicator
		freshness := 1.0 - float64(age)/float64(maxAge)
		if freshness < 0 {
			freshness = 0
		}
		ageBar := ""
		switch {
		case freshness > 0.7:
			ageBar = styleMsg().Render("▐")
		case freshness > 0.4:
			ageBar = styleDim().Render("▌")
		default:
			ageBar = styleDim().Render("·")
		}

		lines = append(lines, "  "+ageBar+" "+line)
	}

	feed := strings.Join(lines, "\n")
	if feed == "" {
		feed = styleDim().Render("  waiting for events...")
	}

	// Ticker arrow
	arrow := ""
	for i := 0; i < (m.ticks/2)%4; i++ {
		arrow += ">"
	}

	rateStr := stylePub().Render(fmt.Sprintf("▲ %.1f/s", m.sim.PubRate)) + "  " +
		styleCon().Render(fmt.Sprintf("▼ %.1f/s", m.sim.ConRate)) + "  " +
		styleQueue().Render(fmt.Sprintf("depth %d", m.sim.QueueDepth)) + "  " +
		styleMsg().Render(arrow)

	title := gradientStr("  Event Feed  ·  Streaming  ·  task_queue", colPub, colCon)

	box := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color(colBorder)).
		Padding(1, 1).
		Width(68).
		Render(feed + "\n\n  " + rateStr)

	return "\n" + title + "\n\n" + box + "\n"
}

func runTimelineL2() {
	runProgram(timelineL2Model{sim: NewSim(), maxTicks: 200})
}

// ─── Timeline L3 ── rich categorized timeline with type counts ─────────────────

type timelineL3Model struct {
	sim        *Sim
	ticks      int
	maxTicks   int
	entries    []tlEntry
	counts     map[EventType]int
}

func (m timelineL3Model) Init() tea.Cmd {
	if m.counts == nil {
		m.counts = map[EventType]int{}
	}
	return doTick()
}

func (m timelineL3Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if m.counts == nil {
		m.counts = map[EventType]int{}
	}
	switch msg.(type) {
	case tea.KeyMsg:
		return m, tea.Quit
	case tickMsg:
		m.sim.Step()
		m.ticks++
		if m.ticks >= m.maxTicks {
			return m, tea.Quit
		}
		for _, ev := range m.sim.LastEvents(8) {
			alreadyIn := false
			for _, e := range m.entries {
				if e.event.MsgID == ev.MsgID && e.event.Type == ev.Type {
					alreadyIn = true
					break
				}
			}
			if !alreadyIn {
				m.entries = append(m.entries, tlEntry{ev, m.ticks})
				m.counts[ev.Type]++
			}
		}
		if len(m.entries) > 25 {
			m.entries = m.entries[len(m.entries)-25:]
		}
		return m, doTick()
	}
	return m, nil
}

func (m timelineL3Model) View() string {
	if m.counts == nil {
		m.counts = map[EventType]int{}
	}

	// Stats bar at top
	statsBar := stylePub().Bold(true).Render(fmt.Sprintf("▲ PUB  %5d", m.counts[EvPublished])) + "  " +
		styleQueue().Bold(true).Render(fmt.Sprintf("→ QUE  %5d", m.counts[EvQueued])) + "  " +
		styleCon().Bold(true).Render(fmt.Sprintf("▼ CON  %5d", m.counts[EvConsumed])) + "  " +
		styleDrop().Bold(true).Render(fmt.Sprintf("✗ DRP  %5d", m.counts[EvDropped]))

	// Rate bars
	maxR := 6.0
	pubW := int(m.sim.PubRate / maxR * 20)
	if pubW > 20 {
		pubW = 20
	}
	conW := int(m.sim.ConRate / maxR * 20)
	if conW > 20 {
		conW = 20
	}
	pubRateBar := stylePub().Render(strings.Repeat("▰", pubW)) + styleDim().Render(strings.Repeat("▱", 20-pubW))
	conRateBar := styleCon().Render(strings.Repeat("▰", conW)) + styleDim().Render(strings.Repeat("▱", 20-conW))

	rateSection := stylePub().Render("pub ") + pubRateBar + stylePub().Render(fmt.Sprintf(" %.2f/s", m.sim.PubRate)) + "\n" +
		styleCon().Render("con ") + conRateBar + styleCon().Render(fmt.Sprintf(" %.2f/s", m.sim.ConRate))

	divider := styleDim().Render(strings.Repeat("─", 66))

	// Timeline entries grouped by type with colored rows
	var pubLines, queLines, conLines, drpLines []string
	for _, e := range m.entries {
		ts := e.event.Time.Format("15:04:05.000")
		age := m.ticks - e.ageTick
		faded := age > 15

		latStr := ""
		if e.event.Type == EvConsumed && e.event.Latency > 0 {
			latStr = fmt.Sprintf(" %3.0fms", float64(e.event.Latency)/float64(time.Millisecond))
		}

		entry := fmt.Sprintf("#%04d %s%s", e.event.MsgID, ts, latStr)
		icon, _, st := eventStyle(e.event.Type)
		if faded {
			st = styleDim()
		}
		rendered := st.Render(icon + " " + entry)

		switch e.event.Type {
		case EvPublished:
			pubLines = append(pubLines, rendered)
		case EvQueued:
			queLines = append(queLines, rendered)
		case EvConsumed:
			conLines = append(conLines, rendered)
		case EvDropped:
			drpLines = append(drpLines, rendered)
		}
	}

	// Keep last 5 of each
	trimLast := func(sl []string, n int) []string {
		if len(sl) > n {
			return sl[len(sl)-n:]
		}
		return sl
	}
	pubLines = trimLast(pubLines, 5)
	queLines = trimLast(queLines, 5)
	conLines = trimLast(conLines, 5)
	drpLines = trimLast(drpLines, 3)

	renderCol := func(header string, lines []string, w int) string {
		content := header + "\n" + strings.Repeat("─", w) + "\n"
		for _, l := range lines {
			content += l + "\n"
		}
		return content
	}

	colPubStr := renderCol(stylePub().Bold(true).Render("PUBLISHED"), pubLines, 26)
	colQueStr := renderCol(styleQueue().Bold(true).Render("QUEUED"), queLines, 26)
	colConStr := renderCol(styleCon().Bold(true).Render("CONSUMED"), conLines, 26)
	colDrpStr := renderCol(styleDrop().Bold(true).Render("DROPPED"), drpLines, 26)

	cols := lipgloss.JoinHorizontal(lipgloss.Top,
		colPubStr+"  ", colQueStr+"  ", colConStr+"  ", colDrpStr)

	title := gradientStr("  Event Timeline  ·  Categorized  ·  task_queue", colPub, colCon)

	box := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color(colBorder)).
		Padding(1, 2).
		Render(statsBar + "\n\n" + rateSection + "\n\n" + divider + "\n\n" + cols)

	return "\n" + title + "\n\n" + box + "\n"
}

func runTimelineL3() {
	runProgram(timelineL3Model{sim: NewSim(), maxTicks: 220, counts: map[EventType]int{}})
}
