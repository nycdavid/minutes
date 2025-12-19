package datacollector

import (
	"fmt"
	"time"
)

type (
	Mod func(*dataCollector)

	dataCollector struct {
		hbchan  chan Heartbeat
		buckets map[string][]Heartbeat
	}

	Heartbeat struct {
		App       string
		Timestamp time.Time
	}
)

func WithHeartbeatChannel(hbchan chan Heartbeat) Mod {
	return func(dataCollector *dataCollector) {
		dataCollector.hbchan = hbchan
	}
}

func New(opts ...Mod) *dataCollector {
	dc := &dataCollector{
		buckets: make(map[string][]Heartbeat),
	}

	for _, opt := range opts {
		opt(dc)
	}

	if dc.hbchan == nil {
		dc.hbchan = make(chan Heartbeat)
	}

	return dc
}

func (d *dataCollector) Run() {
	for {
		select {
		case hb, ok := <-d.hbchan:
			if !ok {
				return
			}

			d.buckets[hb.App] = append(d.buckets[hb.App], hb)
		}
	}
}

func (d *dataCollector) flush() {
	for app, hbs := range d.buckets {
		for _, hb := range hbs {
			fmt.Println(hb)
			// convert to session
			// persist to disk
		}

		// empty bucket
		delete(d.buckets, app)
	}
}

func (d *dataCollector) Buckets() map[string][]Heartbeat {
	return d.buckets
}
