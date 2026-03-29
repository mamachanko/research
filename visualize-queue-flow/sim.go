package main

import (
	"math"
	"math/rand"
	"time"
)

// EventType classifies simulation events
type EventType int

const (
	EvPublished EventType = iota
	EvQueued
	EvConsumed
	EvDropped
)

func (e EventType) String() string {
	switch e {
	case EvPublished:
		return "PUBLISHED"
	case EvQueued:
		return "QUEUED"
	case EvConsumed:
		return "CONSUMED"
	case EvDropped:
		return "DROPPED"
	}
	return "UNKNOWN"
}

func (e EventType) Short() string {
	switch e {
	case EvPublished:
		return "PUB"
	case EvQueued:
		return "QUE"
	case EvConsumed:
		return "CON"
	case EvDropped:
		return "DRP"
	}
	return "???"
}

// Event is one simulation event
type Event struct {
	Time    time.Time
	Type    EventType
	MsgID   int
	Latency time.Duration
}

// FlowParticle is one in-flight message for flow animations
type FlowParticle struct {
	ID      int
	Pos     float64 // 0.0=publisher, 1.0=consumer
	Speed   float64
	InQueue bool
	Char    rune
}

// Sim holds the full simulation state
type Sim struct {
	Tick int

	// Queue
	QueueDepth    int
	QueueCapacity int

	// Totals
	TotalPub  int
	TotalCon  int
	TotalDrop int

	// Current rates (msgs/tick, smoothed)
	PubRate float64
	ConRate float64

	// Histories (capped at 60)
	DepthHist []int
	PubHist   []float64
	ConHist   []float64
	LatHist   []float64 // ms

	// Events (last 100)
	Events []Event

	// Flow particles
	Particles []FlowParticle
	nextID    int
}

// NewSim constructs a ready-to-use simulation
func NewSim() *Sim {
	return &Sim{QueueCapacity: 20}
}

// Step advances the simulation by one tick (~50 ms)
func (s *Sim) Step() {
	s.Tick++
	t := float64(s.Tick)

	// Pub rate: 0.5–5.5 msgs/tick, sinusoidal + noise
	s.PubRate = 3.0 + 2.5*math.Sin(t*0.07) + rand.Float64()*0.8 - 0.4
	if s.PubRate < 0 {
		s.PubRate = 0
	}

	// Con rate: slightly different period/phase
	s.ConRate = 2.8 + 1.8*math.Sin(t*0.05+1.2) + rand.Float64()*0.6 - 0.3
	if s.ConRate < 0 {
		s.ConRate = 0
	}

	// Publish messages
	pubChars := []rune{'●', '○', '◉', '◎', '◆'}
	nPub := int(s.PubRate + rand.Float64())
	for i := 0; i < nPub; i++ {
		s.nextID++
		id := s.nextID
		s.TotalPub++
		s.addEvent(Event{Time: time.Now(), Type: EvPublished, MsgID: id})

		if s.QueueDepth < s.QueueCapacity {
			s.QueueDepth++
			s.addEvent(Event{Time: time.Now(), Type: EvQueued, MsgID: id})
			s.Particles = append(s.Particles, FlowParticle{
				ID:    id,
				Pos:   rand.Float64() * 0.05,
				Speed: 0.018 + rand.Float64()*0.008,
				Char:  pubChars[id%len(pubChars)],
			})
		} else {
			s.TotalDrop++
			s.addEvent(Event{Time: time.Now(), Type: EvDropped, MsgID: id})
		}
	}

	// Consume messages
	nCon := int(s.ConRate + rand.Float64())
	if nCon > s.QueueDepth {
		nCon = s.QueueDepth
	}
	for i := 0; i < nCon; i++ {
		s.QueueDepth--
		s.TotalCon++
		latMs := 15.0 + rand.Float64()*60.0 + float64(s.QueueDepth)*2.5
		s.LatHist = appendF(s.LatHist, latMs, 200)
		s.addEvent(Event{
			Time:    time.Now(),
			Type:    EvConsumed,
			MsgID:   s.TotalCon,
			Latency: time.Duration(latMs * float64(time.Millisecond)),
		})
	}

	// Advance particles
	s.stepParticles(nCon)

	// Histories
	s.DepthHist = appendI(s.DepthHist, s.QueueDepth, 60)
	s.PubHist = appendF(s.PubHist, s.PubRate, 60)
	s.ConHist = appendF(s.ConHist, s.ConRate, 60)
}

func (s *Sim) stepParticles(consumed int) {
	const qL, qR = 0.35, 0.65

	// Release consumed particles from queue (oldest first)
	released := 0
	for i := range s.Particles {
		if s.Particles[i].InQueue && released < consumed {
			s.Particles[i].InQueue = false
			s.Particles[i].Pos = qR + rand.Float64()*0.02
			released++
		}
	}

	alive := s.Particles[:0]
	for _, p := range s.Particles {
		switch {
		case p.InQueue:
			// Slight brownian drift while queued
			p.Pos += (rand.Float64() - 0.5) * 0.003

		case p.Pos < qL:
			// Travelling toward queue
			p.Pos += p.Speed
			if p.Pos >= qL {
				p.InQueue = true
				p.Pos = qL + rand.Float64()*(qR-qL)
			}

		default:
			// Travelling toward consumer
			p.Pos += p.Speed * 1.8
		}

		if p.Pos < 1.08 {
			alive = append(alive, p)
		}
	}

	if len(alive) > 64 {
		alive = alive[len(alive)-64:]
	}
	s.Particles = alive
}

func (s *Sim) addEvent(e Event) {
	s.Events = append(s.Events, e)
	if len(s.Events) > 100 {
		s.Events = s.Events[1:]
	}
}

// QueueFill returns 0.0–1.0
func (s *Sim) QueueFill() float64 {
	if s.QueueCapacity == 0 {
		return 0
	}
	return float64(s.QueueDepth) / float64(s.QueueCapacity)
}

// LastEvents returns the last n events
func (s *Sim) LastEvents(n int) []Event {
	if len(s.Events) <= n {
		return s.Events
	}
	return s.Events[len(s.Events)-n:]
}

// AvgLatency returns average latency over the last n samples
func (s *Sim) AvgLatency(n int) float64 {
	h := s.LatHist
	if len(h) > n {
		h = h[len(h)-n:]
	}
	if len(h) == 0 {
		return 0
	}
	var sum float64
	for _, v := range h {
		sum += v
	}
	return sum / float64(len(h))
}

// MaxDepth returns the highest depth recorded
func (s *Sim) MaxDepth() int {
	mx := 0
	for _, d := range s.DepthHist {
		if d > mx {
			mx = d
		}
	}
	return mx
}

func appendI(s []int, v, cap int) []int {
	s = append(s, v)
	if len(s) > cap {
		s = s[1:]
	}
	return s
}

func appendF(s []float64, v float64, cap int) []float64 {
	s = append(s, v)
	if len(s) > cap {
		s = s[1:]
	}
	return s
}
