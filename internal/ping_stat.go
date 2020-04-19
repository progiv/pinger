package internal

import (
	"fmt"
	"math"
	"sync"
	"time"
)

type pingStats struct {
	min      float64
	max      float64
	sum      float64
	sumSqr   float64
	sent     int
	received int
	errors   int
	sync.Mutex
}

func (p *pingStats) Receive(time time.Duration) {
	p.Lock()
	defer p.Unlock()
	t := float64(time.Milliseconds())
	if t < p.min || p.received == 0 {
		p.min = t
	}
	if t > p.max || p.received == 0 {
		p.max = t
	}
	p.received++
	p.sum += t
	p.sumSqr += t * t
}

func (p *pingStats) Send() {
	p.Lock()
	defer p.Unlock()
	p.sent++
}

func (p *pingStats) Error() {
	p.Lock()
	defer p.Unlock()
	p.errors++
}

func (p *pingStats) Report() string {
	p.Lock()
	defer p.Unlock()
	var loss float64 = 0
	if p.received > 0 {
		loss = (float64(p.sent) - float64(p.received)) / float64(p.sent)
	}
	report := fmt.Sprintf("%d packets transmitted, %d received +%d errors, %.1f%% loss", p.sent, p.received, p.errors, loss)
	if p.received > 0 {
		mean := p.sum / float64(p.received)
		variance := math.Sqrt(p.sumSqr - mean*mean)
		return fmt.Sprintf("%s\nrtt min/avg/max/mdev %v/%v/%.2f/%.2f ms", report, p.min, p.max, mean, variance)
	}
	return report
}
