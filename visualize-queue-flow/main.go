package main

import (
	"flag"
	"fmt"
	"os"
	"time"

	tea "github.com/charmbracelet/bubbletea"
)

// tickMsg is sent every frame
type tickMsg struct{}

// doTick schedules the next frame at ~20 fps
func doTick() tea.Cmd {
	return tea.Tick(50*time.Millisecond, func(time.Time) tea.Msg {
		return tickMsg{}
	})
}

// runProgram is a convenience wrapper that starts a Bubble Tea program
func runProgram(m tea.Model) {
	p := tea.NewProgram(m, tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func main() {
	mode := flag.String("mode", "", "visualization mode")
	flag.Parse()

	modes := map[string]func(){
		// Queue depth meter
		"queue-l1": runQueueL1,
		"queue-l2": runQueueL2,
		"queue-l3": runQueueL3,
		// Message flow animation
		"flow-l1": runFlowL1,
		"flow-l2": runFlowL2,
		"flow-l3": runFlowL3,
		// Publisher / Consumer status panels
		"status-l1": runStatusL1,
		"status-l2": runStatusL2,
		"status-l3": runStatusL3,
		// Throughput graphs
		"throughput-l1": runThroughputL1,
		"throughput-l2": runThroughputL2,
		"throughput-l3": runThroughputL3,
		// Event timeline
		"timeline-l1": runTimelineL1,
		"timeline-l2": runTimelineL2,
		"timeline-l3": runTimelineL3,
		// Network topology
		"topology-l1": runTopologyL1,
		"topology-l2": runTopologyL2,
		"topology-l3": runTopologyL3,
		// Latency heat map
		"latency-l1": runLatencyL1,
		"latency-l2": runLatencyL2,
		"latency-l3": runLatencyL3,
		// Full dashboard
		"dashboard": runDashboard,
	}

	if fn, ok := modes[*mode]; ok {
		fn()
		return
	}

	fmt.Fprintln(os.Stderr, "Usage: mqvis -mode <mode>")
	fmt.Fprintln(os.Stderr, "")
	fmt.Fprintln(os.Stderr, "  Queue depth:   queue-l1  queue-l2  queue-l3")
	fmt.Fprintln(os.Stderr, "  Message flow:  flow-l1   flow-l2   flow-l3")
	fmt.Fprintln(os.Stderr, "  Status panels: status-l1 status-l2 status-l3")
	fmt.Fprintln(os.Stderr, "  Throughput:    throughput-l1 throughput-l2 throughput-l3")
	fmt.Fprintln(os.Stderr, "  Timeline:      timeline-l1 timeline-l2 timeline-l3")
	fmt.Fprintln(os.Stderr, "  Topology:      topology-l1 topology-l2 topology-l3")
	fmt.Fprintln(os.Stderr, "  Latency:       latency-l1 latency-l2 latency-l3")
	fmt.Fprintln(os.Stderr, "  Dashboard:     dashboard")
	os.Exit(1)
}
